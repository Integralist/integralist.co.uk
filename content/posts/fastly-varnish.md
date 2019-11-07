---
title: "Fastly Varnish"
date: 2017-11-02T13:00:00+01:00
categories:
  - "code"
  - "development"
  - "guide"
  - "performance"
tags:
  - "cdn"
  - "fastly"
  - "varnish"
  - "vcl"
draft: false
---

In this post I'm going to be explaining how the [Fastly CDN](https://www.fastly.com/) works, with regards to their 'programmatic edge' feature (i.e. the ability to execute code on cache servers nearest to your users).

Fastly utilizes free software and extends it to fit their purposes, but this extending of existing software can make things confusing when it comes to understanding what underlying features work and how they work.

- [Introduction](#1)
  - [Varnish Basics](#varnish-basics)
- [Varnish Default VCL](#2)
- [Fastly Default VCL](#3)
- [Custom VCL](#3.0)
  - [Be Careful](#be-careful)
- [Fastly TTLs](#3.1)
- [Caching Priority List](#caching-priority-list)
- [Fastly Default Cached Status Codes](#3.2)
- [Fastly Request Flow Diagram](#4)
  - [304 Not Modified](#304-not-modified)
  - [Update 2019.08.10](#update-2019-08-10)
- [Error Handling](#4.0)
- [State Variables](#4.1)
  - [Anonymous Objects](#anonymous-objects)
- [Persisting State](#4.2) (inc. clustering architecture)
  - [Terminology](#terminology)
  - [Clustering](#clustering)
  - [Shielding](#shielding)
  - [`is_cluster` vs `is_origin` vs `is_shield`](#is-cluster-vs-is-origin-vs-is-shield)
  - [Undocumented APIs](#undocumented-apis)
      - [`req.http.Fastly-FF`](#req-http-fastly-ff)
      - [`fastly_info.is_cluster_edge` and `fastly_info.is_cluster_shield`](#fastly-info-is-cluster-edge-and-fastly-info-is-cluster-shield)
      - [`fastly.ff.visits_this_service`](#fastly-ff-visits-this-service)
  - [Breadcrumb Trail](#breadcrumb-trail)
  - [Header Overflow Errors](#header-overflow-errors)
- [Hit for Pass](#5)
- [Serving Stale](#6)
  - [Stale for Client Devices](#stale-for-client-devices)
  - [Caveats of Fastly’s Shielding](#caveats-of-fastly-s-shielding)
  - [Different actions for different states](#different-actions-for-different-states)
  - [Why the difference?](#why-the-difference)
  - [The happy path (stale found in vcl_fetch)](#the-happy-path-stale-found-in-vcl-fetch)
  - [The longer path (stale found in vcl_deliver)](#the-longer-path-stale-found-in-vcl-deliver)
  - [The unhappy path (stale not found anywhere)](#the-unhappy-path-stale-not-found-anywhere)
- [Disable Caching](#7)
- [Logging](#8)
  - [Logging Memory Exhaustion](#logging-memory-exhaustion)
- [Restricting requests to another Fastly service](#8.1)
- [Custom Error Pages](#custom-error-pages)
- [Conclusion](#9)

<div id="1"></div>
## Introduction

[Varnish](https://varnish-cache.org/) is an open-source HTTP accelerator.  
More concretely it is a web application that acts like a [HTTP reverse-proxy](https://en.wikipedia.org/wiki/Reverse_proxy). 

You place Varnish in front of your application servers (those that are serving HTTP content) and it will cache that content for you. If you want more information on what Varnish cache can do for you, then I recommend reading through their [introduction article](https://varnish-cache.org/intro/index.html) (and watching the video linked there as well).

[Fastly](https://www.fastly.com/) is many things, but for most people they are a CDN provider who utilise a highly customised version of Varnish. This post is about Varnish and explaining a couple of specific features (such as hit-for-pass and serving stale) and how they work in relation to Fastly's implementation of Varnish.

One stumbling block for Varnish is the fact that it only accelerates HTTP, not HTTPS. In order to handle HTTPS you would need a TLS/SSL termination process sitting in front of Varnish to convert HTTPS to HTTP. Alternatively you could use a termination process (such as nginx) _behind_ Varnish to fetch the content from your origins over HTTPS and to return it as HTTP for Varnish to then process and cache.

> Note: Fastly helps both with the HTTPS problem, and also with scaling Varnish in general.

The reason for this post is because when dealing with Varnish and VCL it gets very confusing having to jump between official documentation for VCL and Fastly's specific implementation of it. Even more so because the version of Varnish Fastly are using is now quite old and yet they've also implemented some features from more recent Varnish versions. Meaning you end up getting in a muddle about what should and should not be the expected behaviour (especially around the general request flow cycle).

Ultimately this is not a "VCL 101". If you need help understanding anything mentioned in this post, then I recommend reading:

- [Varnish Book](http://book.varnish-software.com/4.0/)
- [Varnish Blog](https://info.varnish-software.com/blog)
- [Fastly Blog](https://www.fastly.com/blog)

> Fastly has a couple of _excellent_ articles on utilising the `Vary` HTTP header (highly recommended reading).

### Varnish Basics

Varnish is a 'state machine' and it switches between these states via calls to a `return` function (where you tell the `return` function which state to move to). The various states are:

- `recv`: request is received and can be inspected/modified.
- `hash`: generate a hash key from host/path and lookup key in cache.
- `hit`: hash key was found in the cache.
- `miss`: hash key was not found in the cache.
- `pass`: content should be fetched from origin, regardless of if it exists in cache or not, and response will not be cached.
- `pipe`: content should be fetched from origin, and no other VCL will be executed.
- `fetch`: content has been fetched, we can now inspect/modify it before delivering it to the user.
- `deliver`: content has been cached (or not, if we had used `return(pass)`) and ready to be delivered to the user.

For each state there is a corresponding subroutine that is executed. It has the form `vcl_<state>`, and so there is a `vcl_recv`, `vcl_hash`, `vcl_hit` etc.

So in `vcl_recv` to change state to "pass" you would execute `return(pass)`. If you were in `vcl_fetch` and wanted to move to `vcl_deliver`, then you would execute `return(deliver)`.

For a state such as `vcl_miss` you'll discover shortly (when we look at a 'request flow' diagram of Varnish) that when `vcl_miss` finishes executing it will then trigger a request to the origin/backend service to acquire the requested content. Once the content is requested, _then_ we end up at `vcl_fetch` where we can then inspect the response from the origin. 

This is why at the end of `vcl_miss` we change state by calling `return(fetch)`. It looks like we're telling Varnish to 'fetch' data but really we're saying move to the next logical state which is actually `vcl_fetch`.

> Note: `vcl_hash` is the only exception to this rule because it's not a _state_ per se, so you don't execute `return(hash)` but `return(lookup)`. This helps distinguish that we're performing an action and not a state change (i.e. we're going to _lookup_ in the cache). We could argue `vcl_miss`'s `return(fetch)` is the same, but from my understanding that's not the case.

<div id="2"></div>
## Varnish Default VCL

When using the open-source version of Varnish, you'll typically implement your own custom VCL logic (e.g. add code to `vcl_recv` or any of the other common VCL subroutines). But it's important to be aware that if you don't `return` an action (e.g. `return(pass)`, or trigger any of the other available Varnish 'states'), then Varnish will continue to execute its own built-in VCL logic (i.e. its built-in logic is _appended_ to your custom VCL).

You can view the 'default' (or 'builtin') logic for each version of Varnish via GitHub:

- [Varnish v2.1](https://github.com/varnishcache/varnish-cache/blob/2.1/bin/varnishd/default.vcl) (the version used by Fastly)
- [Varnish v3.0](https://github.com/varnishcache/varnish-cache/blob/3.0/bin/varnishd/default.vcl)
- [Varnish v4.0](https://github.com/varnishcache/varnish-cache/blob/4.0/bin/varnishd/builtin.vcl)
- [Varnish v5.0](https://github.com/varnishcache/varnish-cache/blob/5.0/bin/varnishd/builtin.vcl)

> Note: after v3 Varnish renamed the file from `default.vcl` to `builtin.vcl`.

But things are slightly different with Fastly's Varnish implementation (which is based off Varnish open-source version 2.1.5).

Specifically:

- no `return(pipe)` in `vcl_recv`, they do `pass` there
- some modifications to the `synthetic` in `vcl_error`

<div id="3"></div>
## Fastly Default VCL

On top of the built-in VCL the open-source version of Varnish uses, Fastly also includes its own 'default' VCL logic alongside your own additions.

When creating a new Fastly 'service', this default VCL is added automatically to your new service. You are then free to remove it completely and replace it with your own custom VCL if you like.

See the link below for what this default VCL looks like, but in there you'll notice code comments such as:

```
#--FASTLY RECV BEGIN

...code here...

#--FASTLY RECV END
```

> Note: those specific portions of the default code define _critical_ behaviour that needs to be defined whenever you want to write [your own custom VCL](#3.0).

Fastly has some guidelines around the use (or removal) of their default VCL which you can learn more about [here](https://docs.fastly.com/guides/vcl/mixing-and-matching-fastly-vcl-with-custom-vcl).

Below are some useful links to see Fastly's default VCL:

- [Fastly's Default VCL (full service context)](https://gist.github.com/Integralist/2e4a78fe92ec70d2e2709ff7be660669)
- [Fastly's Default VCL (each state split into separate files)](https://gist.github.com/Integralist/56cf991ae97551583d5a2f0d69f37788)

> Note: Fastly also has what they call a 'master' VCL which runs outside of what we (as customers) can see, and this VCL is used to help Fastly scale varnish (e.g. handle things like their custom clustering solution).

<div id="3.0"></div>
## Custom VCL

When adding your own custom VCL code you'll need to ensure that you add Fastly's critical default behaviours, otherwise things might not work as expected.

The way you add their defaults to your own custom VCL code is to add a specific type of code comment, for example:

```
sub vcl_recv {
  #FASTLY recv
}
```

See [their documentation](https://docs.fastly.com/vcl/custom-vcl/creating-custom-vcl/) for more details, but ultimately these code comments are 'macros' that get expanded into the actual default VCL code at compile time.

It can be useful to know what the default VCL code does (see [links in previous section](#3)) because it might affect where you place these macros within your own custom code (e.g. do you place it in the middle of your custom sub routines or at the start or the end).

This is important because, for example, the default behaviours Fastly defines for `vcl_recv` is to set a backend for your service. Your custom VCL can of course override that backend, but where you define your custom code that does that overriding might not function correctly if placed in the wrong place. 

In our case we had a conditional comment like `if (req.restarts == 0) { ...set backend... }` and then later on in our VCL we would trigger a request restart (e.g. `return(restart)`), but now `req.restarts` would be equal to `1` and not zero and so when the request restarted we wouldn't set the backend correctly and our request would end up being proxied to an unexpected backend that was selected by Fastly's default VCL!

Fastly selects the default backend based on the age of the backend. To quote Fastly directly...

> We choose the first backend that was created that doesn't have a conditional on it. If all of your backends have conditionals on them, I believe we then just use the first backend that was created. If a backend has a conditional on it, we assume it isn't a default. That backend is only set under the conditions defined, so then we look for the oldest backend defined that doesn’t have a conditional to make it the default.

### Be Careful!

We experienced a problem that broke our production site and it was related to implementing our own `vcl_hash` subroutine.

The problem wasn't as obvious as you might think. We didn't implement `vcl_hash` to change the hashing algorithm, but instead we wanted to add some debug log calls into it.

We looked at the VCL that Fastly generated before we added our own `vcl_hash` subroutine and that VCL looked like the following...

```
sub vcl_hash {
#--FASTLY HASH BEGIN
  #if unspecified fall back to normal
  {
    set req.hash += req.url;
    set req.hash += req.http.host;
    set req.hash += "#####GENERATION#####";
    return (hash);
  }
#--FASTLY HASH END
}
```

We thought "OK, that's what Fastly is generating, so we'll just let them continue generating that code, nothing special we need to do" ...wrong!

So we added the following code to our own VCL...

```
sub vcl_hash {
  #FASTLY hash

  call debug_info;

  return(hash)
}
```

The expectation was that the `#FASTLY hash` macro would still include all the code from inbetween `#--FASTLY HASH BEGIN` and `#--FASTLY HASH END` (see the earlier code snippet).

What actually ended up happening was that the Fastly macro dynamically changed itself to not include critical behaviours

Notice the `set req.hash += req.url;` and `set req.hash += req.http.host;` that they originally were generating? Yup. They were no longer included. This caused the system caching to blow up.

The code that was being generated now looked like the following...

```
#--FASTLY HASH BEGIN
# support purge all
set req.hash += req.vcl.generation;
#--FASTLY HASH END
```

So to fix the problem we had to put those missing settings manually back into our own `vcl_hash` subroutine...

```
sub vcl_hash {
  #FASTLY hash

  set req.hash += req.url;
  set req.hash += req.http.host;

  call debug_info;

  return(hash);
}
```

Interestingly Fastly's default VCL _doesn't_ require us to also set `set req.hash += "#####GENERATION#####";`, so they happily keep that part within their generated code 🤦

<div id="3.1"></div>
## Fastly TTLs

Fastly [has some rules](https://docs.fastly.com/guides/performance-tuning/controlling-caching) about how it determines a TTL for your content. In summary...

Their 'master' VCL sets a TTL of 120s ([this comes from Varnish](https://github.com/varnishcache/varnish-cache/blob/5.0/bin/varnishd/builtin.vcl#L158-L172) rather than Fastly) when no other VCL TTL has been defined and if no cache headers were sent by the origin.

Fastly does a similar thing with its own default VCL which it uses when you create a new service. It looks like the following and increases the default to 3600s (1hr):

```
if (beresp.http.Expires || beresp.http.Surrogate-Control ~ "max-age" || beresp.http.Cache-Control ~"(s-maxage|max-age)") {
  # keep the ttl here
} else {
  # apply the default ttl
  set beresp.ttl = 3600s;
}
```

> Note: 3600 isn't long enough to persist your cached content to disk, it will exist in-memory only. See their documentation on ["Why serving stale content may not work as expected"](https://docs.fastly.com/guides/performance-tuning/serving-stale-content#why-serving-stale-content-may-not-work-as-expected) for more information.

You can override this VCL with your own custom VCL, but it's also worth being aware of the priority ordering Fastly gives when presented with multiple ways to determine your content's cache TTL...

1. `beresp.ttl = 10s`: caches for 10s 
2. `Surrogate-Control: max-age=300` caches for 5 minutes 
3. `Cache-Control: max-age=10` caches for 10s 
4. `Expires: Fri, 28 June 2008 15:00:00 GMT` caches until this date has expired

As we can see from the above list, setting a TTL via VCL takes ultimate priority even if caching headers are provided by the origin server.

## Caching Priority List

Fastly [has some rules](https://docs.fastly.com/guides/tutorials/cache-control-tutorial) about the various caching response headers it respects and in what order this behaviour is applied. The following is a summary of these rules:

- `Surrogate-Control` determines proxy caching behaviour (takes priority over `Cache-Control`) †.
- `Cache-Control` determines client caching behaviour.
- `Cache-Control` determines both client/proxy caching behaviour if no `Surrogate-Control` †.
- `Cache-Control` determines both client/proxy caching behaviour if it includes both `max-age` and `s-maxage`.
- `Expires` determines both client/proxy caching behaviour if no `Cache-Control` or `Surrogate-Control` headers.
- `Expires` ignored if `Cache-Control` is also set (recommended to avoid `Expires`).
- `Pragma` is a legacy cache header only recommended if you need to support older HTTP/1.0 protocol.

> † _except_ when `Cache-Control` contains `private`.

Next in line is `Surrogate-Control` (see [my post on HTTP caching](/posts/http-caching/) for more information on this cache header), which takes priority over `Cache-Control`. The `Cache-Control` header itself takes priority over `Expires`.

> Note: if you ever want to debug Fastly and your custom VCL then it's recommended you create a 'reduced test case' using their [Fastly Fiddle](https://fiddle.fastlydemo.net/) tool. Be aware this tool shares code publicly so don't put secret codes or logic into it!

<div id="3.2"></div>
## Fastly Default Cached Status Codes

The CDN (Fastly) [doesn't cache all responses](https://docs.fastly.com/en/guides/http-status-codes-cached-by-default). It will not cache any responses with a status code in the `5xx` range, and it will only cache a tiny subset of responses with a status code in the `4xx` and `3xx` range.

The status codes it will cache by default are:

- `200 OK`
- `203 Non-Authoritative Information`
- `300 Multiple Choices`
- `301 Moved Permanently`
- `302 Moved Temporarily`
- `404 Not Found`
- `410 Gone`

> Note: in VCL you can allow _any_ response status code to be cached by executing `set beresp.cacheable = true;` within `vcl_fetch` (you can also change the status code if you like to _look_ like it was a different code with `set beresp.status = <new_status_code>;`).

<div id="4"></div>
## Fastly Request Flow Diagram

There are various request flow diagrams for Varnish ([example](http://book.varnish-software.com/4.0/_images/simplified_fsm.svg)) and generally they separate the request flow into two sections: request and backend. 

So handling the request, looking up the hash key in the cache, getting a hit or miss, or opening a pipe to the origin are all considered part of the "request" section. Whereas fetching of the content is considered part of the "backend" section.

The purpose of the distinction is because Varnish likes to handle backend fetches _asynchronously_. This means Varnish can serve stale data while a new version of the cached object is being fetched. This means less request queuing when the backend is slow.

But the issue with these diagrams is that they're not all the same. Changes between Varnish versions and also the difference in Fastly's implementation make identifying the right request flow tricky.

Below is a diagram of Fastly's VCL request flow (including its WAF and Clustering logic). This is a great reference for confirming how your VCL logic is expected to behave.

<a href="../../images/fastly-request-flow.png">
    <img src="../../images/fastly-request-flow.png">
</a>

### 304 Not Modified

Although not specifically mentioned in the above diagram it's worth noting that Fastly doesn't execute `vcl_fetch` when it receives a `304 Not Modified` from origin, but it will use any `Cache-Control` or `Surrogate-Control` values defined on that response to determine how long the stale object should now be kept in cache.

If no caching headers are sent with the `304 Not Modified` response, then the stale object's TTL is _refreshed_. This means its age is set back to zero and the original `max-age` TTL is enforced.

Ultimately this means if you were hoping to execute logic defined within `vcl_fetch` whenever a `304 Not Modified` was returned (e.g. dynamically modify the stale/cached object's TTL), then that isn't possible.

### Update 2019.08.10

Fastly reached out to me to let me know that this diagram is now incorrect. 

Specifically, the request flow for a hit-for-pass ([see below](#5) for details) _was_: 

```
RECV, HASH, HIT, PASS, DELIVER
``` 

Where `vcl_hit` would `return(pass)` once it had identified the cached object as being a hit-for-pass object.

It is now:

```
RECV, HASH, PASS, DELIVER
```

Where after the object is returned from `vcl_hash`'s lookup, it's immediately identified as being a HFP (hit-for-pass) and thus triggers `vcl_pass` as the next state, and finally `vcl_deliver`.

What this ultimately means is there is some redundant code later on in this blog post where I make reference to serving stale content. Specifically, I mention that `vcl_hit` required some custom VCL for checking the cached object's `cacheable` attribute for the purpose of identifying whether it's a hit-for-pass object or not. 

> Note: this `vcl_hit` code logic is still part of the free Varnish software, but it has been made redundant by Fastly's version.

<div id="4.0"></div>
## Error Handling

In Varnish you can trigger an error using the `error` directive, like so:

```
error 900 "Not found";
```

> Note: it's common to use the made-up `9xx` range for these error triggers (900, 901, 902 etc).

Once executed, Varnish will switch to the `vcl_error` state, where you can construct a _synthetic_ error to be returned (or do some other action like set a header and restart the request).

```
if (obj.status == 900) {
  set obj.status = 500;
  set obj.http.Content-Type = "text/html";
  synthetic {"<h1>Hmmm. Something went wrong.</h1>"};
  return(deliver);
}
```

In the above example we construct a synthetic error response where the status code is a `500 Internal Server Error`, we set the content-type to HTML and then we use the `synthetic` directive to manually construct some HTML to be the 'body' of our response. Finally we execute `return(deliver)` to jump over to the `vcl_deliver` state.

> UPDATE 2019.11.07: Fastly's Fiddle tool now shows a compiler error that suggests 8xx-9xx are codes used internally by Fastly and that we should use the 6xx range instead. 

Now, I wanted to talk briefly about error handling because there are situations where an error can occur, and it can cause Varnish to change to an _unexpected_ state. I'll give a real-life example of this...

We noticed that we were getting a raw `503 Backend Unavailable` error from Varnish displayed to our customers. This is odd? We have VCL code in `vcl_fetch` (the state that you move to once the response from the origin has been received by Fastly/Varnish) which checks the response status code for a 5xx and handles the error there. Why didn't that code run?

Well, it turns out that `vcl_fetch` is only executed if the backend/origin was considered 'available' (i.e. Fastly could make a request to it). In this scenario what happened was that our backend _was_ available but there was a network issue with one of Fastly's POPs which meant it was unable to route certain traffic, resulting in the backend appearing as 'unavailable'.

So what happens in those scenarios? In this case Varnish won't execute `vcl_fetch` because of course no request was ever made (how could Varnish make a request if it thinks the backend is unavailable), so instead Varnish jumps from `vcl_miss` (where the request to the backend would be initiated from) to `vcl_error`.

This means in order to handle that very specific error scenario, we'd need to have similar code for checking the status code (and trying to serve stale, see later in this article for more information on that) within `vcl_error`.

<div id="4.1"></div>
## State Variables

Each Varnish 'state' has a set of built-in variables you can use.

Below is a list of available variables and which states they're available to:

> Based on Varnish 3.0 (which is the only explicit documentation I could find on this), although you can see in various request flow diagrams for different Varnish versions the variables listed next to each state. But [this](http://book.varnish-software.com/3.0/VCL_functions.html#variable-availability-in-vcl) was the first explicit list I found. Fastly themselves recommend [this Varnish reference](https://varnish-cache.org/docs/2.1/reference/vcl.html#variables) but that doesn't indicate which variables are read vs write.

Here's a quick key for the various states:

- *R*: recv
- *F*: fetch
- *P*: pass
- *M*: miss
- *H*: hit
- *E*: error
- *D*: deliver
- *I*: pipe
- *#*: hash

||R|F|P|M|H|E|D|I|#|
|:---|---|---|---|---|---|---|---|---|---|
|`req.*`|R/W|R/W|R/W|R/W|R/W|R/W|R/W|R/W|R/W|
|`bereq.*`||R/W|R/W|R/W||||R/W||
|`obj.hits`|||||R||R|||
|`obj.ttl`|||||R/W|R/W||||
|`obj.grace`|||||R/W|||||
|`obj.*`|||||R|R/W||||
|`beresp.*`||R/W||||||||
|`resp.*`||||||R/W|R/W|||

> For the values assigned to each variable:  
> `R/W` is "Read and Write",  
> and `R` is "Read"  
> and `W` is "Write"

It's important to realise that the above matrix is based on Varnish and not Fastly's version of Varnish. But there's only one difference between them, which is the response object `resp` isn't available within `vcl_error` when using Fastly.

When you're dealing with `vcl_recv` you pretty much only ever interact with the `req` object. You generally will want to manipulate the incoming request _before_ doing anything else.

> Note: the only other reason for setting data on the `req` object is when you want to keep track of things (because, as we can see from the above table matrix, the `req` object is available to R/W from _all_ available states).

Once a lookup in the cache is complete (i.e. `vcl_hash`) we'll end up in either `vcl_miss` or `vcl_hit`. If you end up in `vcl_hit`, then generally you'll look at and work with the `obj` object (this `obj` is what is pulled from the cache - so you'll check properties such as `obj.cacheable` for dealing with things like 'hit-for-pass'). 

If you were to end up at `vcl_miss` instead, then you'll probably want to manipulate the `bereq` object and not the `req` object because manipulating the `req` object doesn't affect the request that will shortly be made to the origin. If you decide at this last moment you want to send an additional header to the origin, then you would set that header on the `bereq` and that would mean the request to origin would include that header.

> Note: this is where understanding the various state variables can be useful, as you might want to modify the `req` object for the sake of 'persisting' a change to another state, where as `bereq` modification will only live for the lifetime of the `vcl_miss` subroutine.

Once a request is made, the content is copied into the `beresp` variable and made available within the `vcl_fetch` state. You would likely want to modify this object in order to change its ttl or cache headers because this is the last chance you have to do that before the content is stored in the cache.

Finally, the `beresp` object is copied into `resp` and that is what's made available within the `vcl_deliver` state. This is the last chance you have for manipulating the response that the client will receive. Changes you make to this object doesn't affect what was stored in the cache (because that time, `vcl_fetch`, has already passed).

### Anonymous Objects

There are two scenarios in which a `pass` occurs. One is that you `return(pass)` from `vcl_recv` explicitly, and the other is that `vcl_recv` executes a `return(lookup)` (which is the default behaviour) and the lookup results in a hit-for-pass object. 

In both cases, an anonymous object is created, and the next customer-accessible VCL hook that will run is `vcl_pass`. Because the object is anonymous, any changes you make to it in `vcl_fetch` are not persisted beyond that one client request. 

The only difference between the behaviour of a `return(pass)` from `vcl_recv` and a `return(pass)` resulting from a hit-for-pass in the cache, is that `req.digest` will be set. Fastly's internal varnish engineering team state that `req.digest` is not an identifier for the object but rather a property of the request, which has been set simply because the request went through the hash process. 

An early `return(pass)` from `vcl_recv` doesn't go through `vcl_hash` and so no hash (`req.digest`) is added to the anonymous object. If there is a hash (`req.digest`) available on the object inside of `vcl_pass`, it doesn't mean you retain a reference to the cache object.

<div id="4.2"></div>
## Persisting State

Now that we know there are state variables available, and we understand generally when and why we would use them, let's now consider the problem of clustering (in the realm of Fastly's Varnish implementation) and how that plays an important part in understanding these behaviours. Because, if you don't understand Fastly's design you'll end up in a situation where data you're setting on these variables are being lost and you won't know why.

So let me give you a _real_ example: creating a HTTP header `X-VCL-Route` breadcrumb trail of the various states a request moves through (this is good for debugging, when you want to be sure your VCL logic is taking you down the correct path and through the expected state changes). 

To make this easier to understand I've created a diagram...

<a href="../../images/varnish-request-flow.png">
    <img src="../../images/varnish-request-flow.png">
</a>

In this diagram we can see the various states available to Varnish, but also we can see that the states are separated by an "edge" and "cluster" section.

> Note: this graph is a little misleading in that `vcl_error` should appear in both the "edge" sections, as well as the "cluster" section. We'll come back to this later on and explain why that is.

Now the directional lines drawn on the diagram represent the request flow you might see in a typical Varnish implementation (definitely in my case at any rate). 

Let's consider one of the example routes given: we can see a request comes into `vcl_recv` and from there it could trigger a `return(pass)` and so it would result in the request skipping the cache and going straight to `vcl_pass`, where it will then fetch the content from origin and subsequently end up in `vcl_fetch`. From there the content fetched from origin is stored in the cache and the cached content delivered to the client via `vcl_deliver`. 

That's one example route that could be taken. As you can see there are many more shown on the diagram, and many more I've not included.

But what's important to understand is that Fastly's infrastructure means that `vcl_recv`, `vcl_hash` and `vcl_deliver` are all executed on an edge node (the node nearest the client). Whereas the other states are executed on a "cluster" node (or cache node).

We can see in [Fastly's documentation](https://docs.fastly.com/guides/performance-tuning/request-collapsing), certain VCL subroutines run on the 'edge' node and some on the 'cluster' node (or _fetching_ node):

- Edge: 
  - `vcl_recv`, `vcl_hash` †, `vcl_deliver`, `vcl_log`, `vcl_error`  
- Cluster (fetching): 
  - `vcl_miss`, `vcl_hit`, `vcl_pass`, `vcl_fetch`, `vcl_error`

> † not documented, but Fastly support say that `vcl_hash` executes at the edge.

**UPDATE** (2019.11.07): Fastly's documentation is incorrect. `vcl_pass` does not run on the cluster-shield node and it will actually cause clustering behaviour to break (official the response was "if you pass in recv, you're saying no to cache lookup, and no to request collapsing, so there's no reason to cluster the request" followed by "big sigh, ok, I'm filing an issue" when pointed to their documentation). It would seem that you definitely _can_ end up in `vcl_pass` on a cluster-shield node, but it's harder to do, you would have to `return(pass)` from either `vcl_hit` or `vcl_miss`.


### Terminology

Fastly _internally_ uses a different naming for these 'caching nodes'. 

What I call an 'edge node' is a 'cluster edge' node, and what I call a 'cluster node' (or 'fetching' node, the node that effectively does the fetching of content) they call a 'cluster shield' node. 

We've not covered 'shielding' yet, but the concept of shielding is not the same thing as what they refer to as a 'shield' node! So you can see how there is an overlap in terminology that can cause great confusion!

That said it's _more_ confusing NOT to use their terminology (especially when talking with their engineering support teams) so I would recommend learning these nuances and becoming familiar with them rather than trying to fight against it.

### Clustering

Clustering is the co-ordination of two nodes within a POP to fulfill a request. The following image (and explanation) should help clarify this.

<a href="../../images/fastly-pop.png">
    <img src="../../images/fastly-pop.png">
</a>

With clustering enabled, the resource cache key/hash is used to identify a specific node (this is referred to as the "primary" node, or _fetching_ node). 

This clustering approach means multiple requests to different edge nodes would all go to the same "primary" cache node to _fetch_ the content from the origin (and _request collapsing_ would help protect the origin from traffic overload).

> Note: for resiliency/redundancy Fastly also internally has a _secondary_ cache node that caches content in case of failures, but that's more an internal implementation detail.

The primary cache node will store the content on-disk, while the edge cache node will store the content in-memory. 

Finally, any node within the cluster can be selected as the 'edge' node for an incoming request (its selected at random), hence the in-memory copy of the cached content could exist there and you only hit one cache server before a response is served back to the user.

The benefit of clustering (summarized in [this fastly community support post](https://support.fastly.com/hc/en-us/community/posts/360046680252-What-is-Clustering-)) is that your request only ever goes through (at _most_) two cache server nodes (the 'cluster edge' node and a 'cluster shield' node). 

If clustering was disabled, then the server node the client request was sent to within the edge POP would have to handle the _complete_ state request life cycle (e.g. recv/hash/fetch/deliver).

This would mean that _every_ node within the edge POP (and at the time of writing there are 64 nodes within a POP) would have to go through to origin in order to request your content. 

That is _much_ less efficient. Hence Fastly designed their 'clustering' solution to help improve the request/caching/origin performance.

When we `return(restart)` a request, we _break_ 'clustering'. This means the request will go back to the 'cluster edge' node and that node will handle the full request cycle (this is done for reasons of performance, such as finding stale content in the `vcl_deliver` state on the 'cluster edge' node and wanting to serve that stale content). 

But sometimes breaking clustering might be something you want to avoid (note: I've never found a reason for this!). In order to avoid breaking the clustering process you'll need to utilize the `Fastly-Force-Shield: 1` request header. This header will re-enable clustering so that we again use multiple server nodes within a POP when executing the different VCL subroutine states.

> Notice the naming in `Fastly-Force-Shield` is still the legacy 'shield' terminology.

I use the terminology "cluster node" to describe cache server nodes that do the _fetching_ of content (i.e. with clustering the cluster node does _NOT_ handle recv/hash/deliver, as they are handled by the 'edge' node). 

Fastly on the other hand has historically used the term "shield" node to describe that node, and although their documentation still refers to the behaviour as "shielding", it infact has changed verbally to being described as "clustering". 

Subsequently I've purposely avoided the historical "shield" terminology because it overlaps with a more _recent_ fastly feature also called [Shielding](https://docs.fastly.com/guides/performance-tuning/shielding) (hence Fastly changed the historical process to "clustering", so as to not confuse the now two distinct concepts).

OK, so two more _really_ important things to be aware of at this point:

1. Data added to the `req` object _cannot_ persist across boundaries (except for when initially moving from the edge to the cluster).
2. Data added to the `req` object _can_ persist a restart, but _not_ when they are added from the cluster environment.

For number 1. that means: `req` data you set in `vcl_recv` and `vcl_hash` will be available in states like `vcl_pass` and `vcl_miss`.

For number 2. that means: if you were in `vcl_deliver` and you set a value on `req` and then triggered a restart, the value would be available in `vcl_recv`. Yet, if you were in `vcl_miss` for example and you set `req.http.X-Foo` and let's say in `vcl_fetch` you look at the response from the origin and see the origin sent you back a 5xx status, you might decide you want to restart the request and try again. But if you were expecting `X-Foo` to be set on the `req` object when the code in `vcl_recv` was re-executed, you'd be wrong. That's because the header was set on the `req` object while it was in a state that is executed on a 'cluster shield' node; and so the `req` data set there doesn't persist a restart.

> Summary: modifications on the 'cluster shield' node don't persist to 'cluster edge' node.

If you're starting to think: "this makes things tricky", you'd be right :-)

The earlier diagram visualizes the approach for how a request inside of a POP will reach a specific cache node (i.e. "clustering") but it doesn't cover how "[shielding](https://docs.fastly.com/guides/performance-tuning/)" works, which effectively is a _nested_ clustering process. 

Let's dig into Shielding next...

### Shielding

Shielding is a designated POP that a request will flow through _before_ reaching your origin.

As mentioned above, clustering is the co-ordination of two nodes within a POP, and this 'clustering' happens within every POP (so the shield POP is no different from any other POP in its fundamental behaviour). 

The purpose of shielding is to give extra protection to your origin, because multiple users will arrive at different POPs (due to their locality) and so a POP in the UK might not have a cached object, while a POP in the USA might have a cached version of the resource. To help prevent the UK user from making a request back to the origin, we can select a POP nearest the origin to act as a single point of access. 

Now when the UK request comes through and there is no cached content within that POP, the request wont go to origin, it'll go to the shield POP which will hopefully have the content cached (if not then the shield POP sends the request onto the origin). 

But ultimately the content either already cached at the shield (or is about to be cached at the shield if it had no cache) will be bubbled back to the UK POP where the content will be cached there as well.

> Note: there isn't any real latency concern with using shielding because Fastly optimizes the network BGP routing between POPs. The only thing to ensure is that your shield POP is located next to your origin because Fastly can't optimize the connection from the shield POP to the origin (it can only optimize traffic within its own network).

This also gives us extra nomenclature to distinguish POPs. Before we learnt about shielding, we just knew there were 'POPs' but now we know that with shielding enabled we have 'edge' POPs and a singular 'shield' POP for a particular Fastly service.

I'll link [here](https://fiddle.fastlydemo.net/fiddle/72e0d619) to a Fastly-Fiddle (created by a Fastly engineer) that demonstrates the clustering/shielding flow. If that fiddle no longer exists by the time you read this then I've made a copy of it in a [gist](https://gist.github.com/Integralist/c08b1ab3e9dd508b1ccc5fe768d1a9b0). It's interesting to see how the various APIs for identifying a server node come together.

If you want to track extra information when using shielding, then using (in combination with either `req.backend.is_origin` or `!req.backend.is_shield`) the values from `server.datacenter` and `server.hostname` which can help you identify the POP as your shielding POP (remember there is only one POP that is designated as your shield, so this can come in handy).

> Note: remember that, although a statistically small chance, the edge POP that is reached by a client request could be the shield POP so your mechanism for checking if something is a shield needs to account for that scenario.

Additionally, there is [this](https://fiddle.fastlydemo.net/fiddle/d053c409) Fastly-Fiddle which clarifies the `req.backend.is_cluster` API which actually is different to similarly named APIs such as `req.backend.is_origin` and `req.backend.is_shield`, so let's dig into that quickly...

### `is_cluster` vs `is_origin` vs `is_shield`

There are a few properties hanging off the `req.backend` object in VCL...

- `is_cluster`: indicates when the request has come from a clustering node (e.g. 'cluster shield' node).
- `is_origin`: indicates if the request will be proxied to an origin server (e.g. your own backend application).
- `is_shield`: indicates if the request will be proxied to a shield POP (which happens when shielding is enabled).

If you try to access `is_origin` from within the `vcl_recv` state subroutine, for example, it will be cause a compiler error. This is because that API is only available to 'cluster shield' nodes (and specifically only states that would result in a request being proxied, meaning although `vcl_hit` runs on a 'cluster shield' node, that state would not have access to `is_origin`).

So depending on what you're trying to verify, it might be preferable to use the negated `is_shield` approach for checking if the request is going to be proxied to origin or a shield pop node.

### Undocumented APIs

- `req.http.Fastly-FF`
- `fastly_info.is_cluster_edge`
- `fastly_info.is_cluster_shield`
- `fastly.ff.visits_this_service`

#### `req.http.Fastly-FF`

The first API we'll look at is: `req.http.Fastly-FF` which indicates if a request has come from a Fastly server.

It's worth mentioning that it's not really safe to use `req.http.Fastly-FF` because it can be set by a client making the request, and so there is no guarantee of its accuracy. 

Also, the use of `req.http.Fastly-FF` can become complicated if you have multiple Fastly services _chained_ one after the other because it means `Fastly-FF` could be set by a Fastly service not owned by you (i.e. the reported value is misleading).

#### `fastly_info.is_cluster_edge` and `fastly_info.is_cluster_shield`

With regards to `is_cluster` there are also some additional _undocumented_ APIs we can use:

- `fastly_info.is_cluster_edge`: `true` if the current `vcl_` state subroutine is running on a 'cluster edge' node.
- `fastly_info.is_cluster_shield`: `true` if the current `vcl_` state subroutine is running on a 'cluster shield' node.

It's important to realize that `is_cluster_edge` will only ever report true from `vcl_deliver`, as (just like `req.backend.is_cluster`) we have to come _from_ a clustering/fetching node first. The `vcl_recv` state can't know if it's going to go into clustering at that stage of the request hence it reports as `false` there. 

With `vcl_fetch` it knows it has come from the 'cluster edge' node and thus we've gone into 'clustering' mode, hence it can report `is_cluster_shield` as `true`.

So as a more fleshed out example, if we tried to log all three cluster APIs in all VCL subroutines (and imagine we have clustering enabled with Fastly, which is the default behaviour), then we would find the following results...

- `vcl_recv`:
  - `req.backend.is_cluster`: no
  - `fastly_info.is_cluster_edge`: no
  - `fastly_info.is_cluster_shield`: no
- `vcl_hash`:
  - `req.backend.is_cluster`: no
  - `fastly_info.is_cluster_edge`: no
  - `fastly_info.is_cluster_shield`: no
- `vcl_miss`:
  - `req.backend.is_cluster`: no
  - `fastly_info.is_cluster_edge`: no
  - `fastly_info.is_cluster_shield`: yes
- `vcl_fetch`:
  - `req.backend.is_cluster`: no
  - `fastly_info.is_cluster_edge`: no
  - `fastly_info.is_cluster_shield`: yes
- `vcl_deliver`:
  - `req.backend.is_cluster`: yes
  - `fastly_info.is_cluster_edge`: yes
  - `fastly_info.is_cluster_shield`: no

#### `fastly.ff.visits_this_service`

The last undocumented API we'll look at will be `fastly.ff.visits_this_service` which indicates for each server node how many times it has seen the request currently being handled. This helps us to execute a piece of code only once (maybe authentication needs to happen at the edge only once).

Let's see what this looks like in a clustering scenario like shown a moment ago...

- `vcl_recv`:
  - `fastly.ff.visits_this_service`: `0` (we're on a 'cluster edge' node and we've never seen this request before)
- `vcl_hash`:
  - `fastly.ff.visits_this_service`: `0` (we're still on the same 'cluster edge' node so the reported value is the same)
- `vcl_miss`:
  - `fastly.ff.visits_this_service`: `1` (we've jumped to the 'cluster shield' node so we know it's been seen once before somewhere)
- `vcl_fetch`:
  - `fastly.ff.visits_this_service`: `1` (we're on the same 'cluster shield' node so the value is reported the same)
- `vcl_deliver`:
  - `fastly.ff.visits_this_service`: `0` (we're back onto the original 'cluster edge' node so the value is reported as 0 again).

> Note: when you introduce shielding you'll find that the value increases to `2` when we reach `vcl_recv` on the 'cluster edge' node inside the shield POP.


### Breadcrumb Trail

Let's now revisit our requirement, which was to create a breadcrumb trail using a HTTP header (this is where all this context becomes important).

First I'm going to show you the basic outline of the various vcl state subroutines and a set of function calls. Next I'll show you the code for those function calls and talk through what they're doing.

> Note: the code examples will be simplified for sake of brevity and to highlight the calls related to the breadcrumb trail functionality.

In essence though, we're using the understanding we have for the 'cluster edge' and 'cluster shield' _boundaries_ and ensuring that while on the 'cluster shield' we track information in Varnish objects (e.g. things like `req`, `obj`, `beresp` etc) that are able to persist crossing those boundaries.

This means when going from `vcl_fetch` to `vcl_deliver` we can't track information on the `req` object because it'll be lost when we move over to `vcl_deliver` and so we track it in the `beresp` object that `vcl_fetch` has access to. We do this because we know that `beresp` is _copied_ over to `vcl_deliver` and exposed inside of `vcl_deliver` as the `resp` object. 

Similarly for `vcl_error`, we track information in its `obj` reference because when we move over to `vcl_deliver` the `obj` reference will be copied over as the `resp` object. 

From `vcl_deliver` we're then free to copy the tracked information from the `X-VCL-Route` header (stored in the `resp` object) into the `req` object (in case we need to restart the request and keep tracking information after a restart).

Remember that `vcl_pass` and `vcl_miss` both run on the 'cluster shield' and so again tracking information on the `req` object won't persist when moving over to `vcl_deliver`. That's why although we track information on the `req` object within those state subroutines, we in fact (once we've reached `vcl_fetch`) copy those tracked values into the `beresp` object available to `vcl_fetch` for the reasons we described earlier.

OK, enough talking, let's see the code...

```
include "debug_info"

sub vcl_recv {
  call debug_info_recv;

  #FASTLY recv

  return(lookup);
}

sub vcl_hash {
  #FASTLY hash

  set req.hash += req.url;
  set req.hash += req.http.host;

  call debug_info_hash;

  return(hash);
}


sub vcl_miss {
  #FASTLY miss

  call debug_info_miss;

  return(fetch);
}

sub vcl_pass {
  #FASTLY pass

  call debug_info_pass;
}

sub vcl_fetch {
  #FASTLY fetch

  call debug_info_fetch;

  return(deliver);
}

sub vcl_error {
  #FASTLY error

  call debug_info_error;

  return(deliver);
}

sub vcl_deliver {
  call debug_info_deliver;

  ...other code...

  call debug_info_send;

  #FASTLY deliver

  return(deliver);
}
```

The thing to look out for (in the above code) are the function calls that start with `debug_info_` (e.g. `debug_info_recv`, `debug_info_hash` ...etc).

These functions are all defined within a VCL file called `debug_info.vcl`, which we import at the top of the example code (i.e. `include "debug_info"`). 

Once that file is imported we will have access to the various `debug_info_` functions that the VCL state subroutines are calling.

The other thing you'll notice is that we have _two_ separate function calls within `vcl_deliver`. We have the standard `debug_info_deliver` (as described a moment ago), but we also have `debug_info_send`.

The `debug_info_send` isn't strictly necessary (e.g. the code within it could be moved inside of `debug_info_deliver`) but I wanted to distinguish between functions that _collected_ data and this function that was responsible for _sending_ the collected data back within the response.

Let's now take a look at that `debug_info` VCL and then we'll explain what the code is doing...

```
# This file sets a X-VCL-Route response header.
#
# X-VCL-Route has a baseline format of:
#
# pop: <value>, node: <value>, state: <value>, host: <value>, path: <value>
#
# pop  = the POP where the request is currently passing through.
# node  = whether the cache node is a 'cluster edge' or 'cluster shield' node (see: https://www.integralist.co.uk/posts/fastly-varnish/#4.2).
# state = internal fastly variable that reports state flow as well as whether a request waited for request collapsing or whether it was clustered.
# host  = the full host name, without the path or query parameters.
# path   = the full path, including query parameters.
#
# Additional to this baseline we include information relevant to the subroutine state.

sub debug_info_recv {
  declare local var.context STRING;
  set var.context = "";

  if (req.restarts > 0) {
    set var.context = req.http.X-VCL-Route + ", ";
  }

  set req.http.X-VCL-Route = var.context + "VCL_RECV(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
    "node: cluster_edge, " +
    "state: " + fastly_info.state + ", " +
    "host: " + req.http.host + ", " +
    "path: " + req.url +
    ")";
}

sub debug_info_hash {
  set req.http.X-VCL-Route = req.http.X-VCL-Route + ", VCL_HASH(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
    "node: cluster_edge, " +
    "state: " + fastly_info.state + ", " +
    "host: " + req.http.host + ", " +
    "path: " + req.url +
    ")";
}

sub debug_info_miss {
  set req.http.X-PreFetch-Miss = ", VCL_MISS(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) " [" server.datacenter ", " server.hostname "], " +
    "node: cluster_" + if(fastly_info.is_cluster_shield, "shield", "edge") + ", " +
    "state: " + fastly_info.state + ", " +
    "host: " + bereq.http.host + ", " +
    "path: " + bereq.url +
    ")";
}

sub debug_info_pass {
  set req.http.X-PreFetch-Pass = ", " if(fastly_info.state ~ "^HITPASS", "VCL_HIT", "VCL_PASS") "(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
    "node: cluster_" + if(fastly_info.is_cluster_shield, "shield", "edge") + ", " +
    "state: " + fastly_info.state + ", " +
    "host: " + req.http.host + ", " +
    "path: " + req.url + ", " +
    ")";
}

sub debug_info_fetch {
  set beresp.http.X-Track-VCL-Route = req.http.X-VCL-Route;
  set beresp.http.X-PreFetch-Pass = req.http.X-PreFetch-Pass;
  set beresp.http.X-PreFetch-Miss = req.http.X-PreFetch-Miss;
  set beresp.http.X-PostFetch = ", VCL_FETCH(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
    "node: cluster_" + if(fastly_info.is_cluster_shield, "shield", "edge") + ", " +
    "state: " + fastly_info.state + ", " +
    "host: " + req.http.host + ", " +
    "path: " + req.url + ", " +
    "status: " + beresp.status + ", " +
    "stale: " + if(stale.exists, "exists", "none") + ", " +
    if(beresp.http.Cache-Control ~ "private", "cache_control: private, return: pass", "return: deliver") +
    ")";
}

sub debug_info_error {
  declare local var.error_page BOOL;
  set var.error_page = false;

  set obj.http.X-VCL-Route = req.http.X-VCL-Route + ", VCL_ERROR(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
    "node: cluster_" + if(fastly_info.is_cluster_shield, "shield", "edge") + ", " +
    "state: " + fastly_info.state + ", " +
    "host: " + req.http.host + ", " +
    "path: " + req.url + ", " +
    "status: " + obj.status + ", " +
    "stale: " + if(stale.exists, "exists", "none") + ", " +
    ")";
}

sub debug_info_deliver {
  # only track the previous route flow if we've come from vcl_fetch
  # otherwise we'll end up displaying the uncached request flow as
  # part of this cache hit request flow (which would be confusing).
  if (resp.http.X-Track-VCL-Route && fastly_info.state ~ "^(MISS|PASS)") {
    set req.http.X-VCL-Route = resp.http.X-Track-VCL-Route;

    if (resp.http.X-PreFetch-Pass) {
      set req.http.X-VCL-Route = req.http.X-VCL-Route + resp.http.X-PreFetch-Pass;
    }

    if (resp.http.X-PreFetch-Miss) {
      set req.http.X-VCL-Route = req.http.X-VCL-Route + resp.http.X-PreFetch-Miss;
    }

    if (resp.http.X-PostFetch) {
      set req.http.X-VCL-Route = req.http.X-VCL-Route + resp.http.X-PostFetch;
    }
  } elseif (fastly_info.state ~ "^ERROR") {
    # otherwise track in the request object any request flow information that has occurred from an error request flow 
    # which should include either the original vcl_fetch flow or the vcl_hit flow.
    set req.http.X-VCL-Route = resp.http.X-VCL-Route;
  } elseif (fastly_info.state ~ "^HIT($|-)") {
    # otherwise track the initial vcl_hit request flow.
    set req.http.X-VCL-Route = req.http.X-VCL-Route + ", VCL_HIT(" +
      "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
      "node: cluster_shield, " +
      "state: " + fastly_info.state + ", " +
      "host: " + req.http.host + ", " +
      "path: " + req.url + ", " +
      "status: " + resp.status + ", " +
      "cacheable: true, " +
      "return: deliver" +
      ")";
  }

  # used to extend the baseline X-VCL-Route (set below)
  declare local var.context STRING;
  set var.context = "";

  # there is one state subroutine that we have no way of tracking information for: vcl_hit
  # this is because the only object we have available with R/W access is the req object
  # and modifications to the req object don't persiste from cluster node to edge node (e.g. vcl_deliver)
  # this means we need to utilise fastly's internal state to see if we came from vcl_hit.
  #
  # we also use this internal state variable to help identify other states progressions such as
  # STALE (stale content found in case of an error) and ERROR (we've arrived to vcl_deliver from vcl_error).
  #
  # Documentation (fastly_info.state):
  # https://support.fastly.com/hc/en-us/community/posts/360040168391/comments/360004718351
  if (fastly_info.state ~ "^HITPASS") {
    set var.context = ", cacheable: uncacheable, return: pass";
  } elseif (fastly_info.state ~ "^ERROR") {
    set var.context = ", custom_error_page: " + resp.http.CustomErrorPage;
  }

  set req.http.X-VCL-Route = req.http.X-VCL-Route + ", VCL_DELIVER(" +
    "pop: " + if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield")) + " [" + server.datacenter + ", " + server.hostname + "], " +
    "node: cluster_edge, " +
    "state: " + fastly_info.state + ", " +
    "host: " + req.http.host + ", " +
    "path: " + req.url + ", " +
    "status: " + resp.status + ", " +
    "stale: " + if(stale.exists, "exists", "none") +
    var.context +
    ")";
}

# this subroutine must be placed BEFORE the vcl_deliver macro `#FASTLY deliver`
# otherwise the setting of Fastly-Debug within this subroutine will have no effect.
sub debug_info_send {
  unset resp.http.X-Track-VCL-Route;
  unset resp.http.X-VCL-Route;
  unset resp.http.X-PreFetch-Miss;
  unset resp.http.X-PreFetch-Pass;
  unset resp.http.X-PostFetch;

  if (req.http.X-Debug) {
    # ensure that when Fastly's own VCL executes it will be able to identify
    # the Fastly-Debug request header as enabled.
    set req.http.Fastly-Debug = "true";

    # other useful debug information
    set resp.http.Fastly-State = fastly_info.state;
    set resp.http.X-VCL-Route = req.http.X-VCL-Route;
  }
}
```

So I've already summarized the basic approach (i.e. ensure we set tracking information in Varnish objects that we know persist the boundary changes), but there are some interesting bits to take away from this code still...

Within `debug_info_recv`, because a request can be restarted, the breadcrumb trail may need to repeat the VCL states. So for example if a request was restarted we would want to see `VCL_RECV(...)` twice in the output to indicate that this was the flow of the request. To account for that I use a `var.context` variable to prepend the information from _before_ the request to the new `X-VCL-Route` header being created now the request has restarted.

Next you'll notice that we have a strange expression which we're assigning to a `pop` field. This field represents whether the request is being executed at an 'edge' POP or a 'shield' POP (in the case that the Fastly service has shielding enabled)...

```
if(req.backend.is_shield, "edge", if(fastly.ff.visits_this_service < 2, "edge", "shield"))
```

We use the `req.backend.is_shield` API to tell us if the server is going to send the request to a shield POP or not, but unfortunately it's not as straight forward as just saying "ok `is_shield` is false so we must be on a node within the shield POP already".

The reason being, if the Fastly service _doesn't_ have shielding enabled then `req.backend.is_shield` will _always_ return false! So that's why we have that nested conditional check...

```
if(fastly.ff.visits_this_service < 2, "edge", "shield")
```

Instead of just printing `shield` when `req.backend.is_shield` is false we will now first check `fastly.ff.visits_this_service` (which we [discussed earlier](#fastly-ff-visits-this-service)). This says "ok, if the returned value is less than two we know that we must be running within an edge POP, because if shielding was enabled and we were on a shield POP currently then the number of 'views' would be three or more at this point".

Although I don't demonstrate it here in this example code, `debug_info_hash` might need an extra conditional header check added to it to prevent tracking the `vcl_hash` execution. The reason you might want to do that is because `vcl_hash` is _always_ executed (even though it can in some scenarios be a no-op!).

For example, in my own VCL code at work we have a process whereby we'll restart a request if we have an error come back from the origin. The reason we restart the request is so we can serve a custom error page synthetically from the edge server. In order to do that we set a header in `vcl_deliver` before we restart (e.g. `X-Error`) and then we check it from within `vcl_recv` only if the `req.restarts == 1`. From that point in `vcl_recv` if `X-Error` is set we'll trigger `vcl_error` by calling the `error` directive, and within `vcl_error` we'll construct our synthetic custom error page.

Now the problem we have is that in the final `X-VCL-Route` sent back in the response we'll see `vcl_hash` in the flow just after the second `vcl_recv` (which happens because of the restart in order to handle the custom error page logic), and that just ...looks weird because it suggests to a viewer that `vcl_hash` had some sort of _purpose_ for running, but it doesn't!

So to avoid recording the `vcl_hash` execution I wrap the tracking code in a check for the `X-Error` header (e.g. `if (!req.http.X-BF-Error) { track the execution }`).

In `debug_info_fetch` you'll see that I don't assign to a `X-VCL-Route` header on the `beresp` object but a new `X-Track-VCL-Route` header. The reason for this is because the `beresp` object is going to be placed into the Varnish cache system once `vcl_fetch` has finished executing and we need to be careful about setting `X-VCL-Route` there. 

The reason for the concern is due to _new_ client requests. Imagine we set the tracking information onto the actual `X-VCL-Route` header on the `beresp` (which will be cached). 

When a new request is received for that same resource we'll go to `vcl_hit` because the resource is found in the cache and from there we'll end up at `vcl_deliver` which would normally just check for a `X-VCL-Route` on the `resp` object it has been passed. 

But in this scenario the `resp` object is the object pulled from the Varnish cache and so we might end up including the _uncached_ request flow in the final `X-VCL-Route` header sent back for this secondary request! 

That would be wrong because the secondary request should only report a `recv > hash > hit > deliver` and not `recv > hash > miss > fetch > recv > hash > hit > deliver` (which is what would happen otherwise!). See how easy it is for things to get dangerous and confusing!

With that in mind if we look at `debug_info_deliver` we'll see that we only track the 'uncached' request flow _if_ we can tell that we actually came from a cache MISS state...

```
if (resp.http.X-Track-VCL-Route && fastly_info.state ~ "^MISS") { track the data that was stored in X-Track VCL-Route }
```

> If you want more information on `fastly_info.state` see [this community comment](https://community.fastly.com/t/useful-variables-to-log/303/3).

You'll also see in that same conditional block we then reconstruct a `X-VCL-Route` from the various temporary headers, such as `X-PreFetch-Pass`, `X-PreFetch-Miss` and `X-PostFetch`.

Remember that in `vcl_hit` we have no object that we can track information on that will be persisted to the 'cluster edge' node where `vcl_deliver` is executed. This is because `vcl_hit` can go straight to `vcl_deliver` and there's no inbetween state like `vcl_fetch` where we can borrow a Varnish object for tracking purposes like we do with `vcl_miss` and `vcl_pass`.

So to solve that problem you'll see that we check to see if we came from a cache HIT state...

```
fastly_info.state ~ "^HIT($|-)"
```

> Note: this will catch multiple types of hit flows, such as `HIT` or `HIT-CLUSTER` or `HIT-STALE-CLUSTER`.

Finally, in `debug_info_send` you'll see that we make sure to `unset` any temporary response headers (e.g. `X-Track-VCL-Route`) so it doesn't confuse people scanning the response headers we send back. We also don't send this debug information back to the client unless they ask for it, which they can do by providing a `req.http.X-Debug` request header.

One interesting thing to note there is that in order to get Fastly to send back its own debug information the user needs to provide a `Fastly-Debug` request header, but to avoid having to rely on that knowledge we provide our own `X-Debug` (which we tell our engineers about) so they only have to remember one simpler request header to enable not only our own debugging information but Fastly's as well.

In order then for the Fastly debug information to be included we have to do two things:

1. manually `set req.http.Fastly-Debug = "true";` in our VCL code.
2. ensure our `debug_info_send` is executed _before_ the `#FASTLY deliver` macro (which is where they execute their `Fastly-Debug` logic).

OK, let's go back and revisit `vcl_error`...

The `vcl_error` subroutine is a tricky one because it exists on _both_ the 'cluster edge' and the 'cluster shield' nodes. 

Meaning if you execute `error 401` from `vcl_recv`, then `vcl_error` will execute in the context of the 'cluster edge' node; whereas if you executed an `error` from a 'cluster shield' node state like `vcl_fetch`, then `vcl_error` would execute in the context of the 'cluster shield' node.

Meaning, how you transition information between `vcl_error` and `vcl_deliver` could depend on whether you're on a 'cluster edge' node or a 'cluster shield' node.

This is why when on what are typically considered the 'cluster shield' nodes, I don't hardcode them a such. I use the following conditional check...

```
"node: cluster_" + if(fastly_info.is_cluster_shield, "shield", "edge") + ", " +
```

...this check means if the request is restarted, which we know this causes Fastly's clustering behaviour to stop and for the request to flow through the same 'cluster edge' server for all states, then we can change the reported node to be the 'cluster edge' instead of a 'cluster shield' for states such as MISS/PASS/FETCH.

> Note: this conditional check could be improved by also checking for the existence of the `Fastly-Force-Shield` header (which we talked about earlier). This is because if someone uses that header then it would force Fastly's clustering behaviour to not be stopped.

To help explain this I'm going to give another _real_ example, where I wanted to lookup some content in our cache and if it didn't exist I wanted to restart the request and use a different origin server to serve the content.

To do this I expected the route to go from `vcl_recv`, to `vcl_hash` and the lookup to fail so we would end up in `vcl_miss`. Now from `vcl_miss` I could have triggered a restart, but anything I set on the `req` object at that point (such as any breadcrumb data appended to `X-VCL-Route`) would have been lost as we transitioned from the 'cluster shield' node back to the 'cluster edge' node (where `vcl_recv` is).

I needed a way to persist the "miss" breadcrumb, so instead of returning a restart from `vcl_miss` I would trigger a custom error such as `error 901` and inside of `vcl_error` I would have the following logic:

```
set obj.http.X-VCL-Route = req.http.X-VCL-Route;

if (fastly_info.state ~ "^MISS") {
  set obj.http.X-VCL-Route = obj.http.X-VCL-Route ",VCL_MISS(" req.http.host req.url ")";
}

set obj.http.X-VCL-Route = obj.http.X-VCL-Route ",VCL_ERROR";

if (obj.status == 901) {
  set obj.http.X-VCL-Route = obj.http.X-VCL-Route ",VCL_ERROR(status: 908, return: deliver)";

  return(deliver);
}
```

When we trigger an error an object is created for us. On that object I set our `X-VCL-Route` and assign it whatever was inside `req.http.X-VCL-Route` at that time †

> † which would include `vcl_recv`, `vcl_hash` and `vcl_miss`. Remember the `req` object _does_ persist across the 'cluster edge'/'cluster shield' boundaries, but _only_ when going from `vcl_recv`. After that, anything set on `req` is lost when crossing boundaries.

Now we could have arrived at `vcl_error` from `vcl_recv` (e.g. if in our `vcl_recv` we had logic for checking Basic Authentication and none was found on the incoming request we could decide from `vcl_recv` to execute `error 401`) or we could have arrived at `vcl_error` from `vcl_miss` (as per our earlier example). So we need to check the internal Fastly state to identify this, hence checking `fastly_info.state ~ "^MISS"`.

After that we append to the `obj` object's `X-VCL-Route` header our current state (i.e. so we know we came into `vcl_error`). Finally we look at the status on the `obj` object and see it's a 901 custom status code and so we append that information in order to know what happened.

But you'll notice we don't restart the request from `vcl_error`, because if we did come from `vcl_miss` the data in `obj` would be lost because ultimately it was set on a 'cluster shield' node (as `vcl_error` would be running on the 'cluster shield' node when coming from `vcl_miss`).

Instead we `return(deliver)`, because all that data assigned to `obj` is guaranteed to be copied into `resp` for us to reference when transitioning to `vcl_deliver` on the 'cluster edge' node.

Once we're at `vcl_deliver` we continue to set breadcrumb tracking onto `req.http.X-VCL-Route` as we know that will persist a restart (as we're still going to be executing code on the 'cluster edge' node).

Phew! Well, that was NOT an easy task, and if you struggled to follow along. That's OK because it's hard to learn about all this stuff at once. Just keep coming back to this post as a reference point if you need to. 

Thankfully by going through this process there will be very little about Varnish's request flow that you won't now understand or have the ability to work around in future if the right problem scenario presents itself.

### Header Overflow Errors

One final thing to note about building a 'breadcrumb trail' like I've detailed above is that you'll need to be careful about how many times you restart your request flow. 

I've seen us restart our request up to three times (for multiple-backend failover resiliency) and it has resulted in a raw Varnish error of `Header overflow` to be returned in the response (the client won't even record any network data at all, it's as if the client compleletely flatlines).

Googling around I discovered some [Fastly documentation](https://docs.fastly.com/en/guides/resource-limits#request-and-header-limits) which makes reference to a `Header count`...

> Exceeding the limit results in a Header overflow error. A small portion of this limit is reserved for internal Fastly use, making the practical limit closer to 85.

To me this was quite an ambiguous statement and have the value set to `96` didn't help elucidate the situation at all either. I had no idea what this really meant. 

Was Fastly saying the overall request header size of `X-VCL-Route` was too large (hence a `400 Bad Request` would be the typical response expected for that type of error) or did it mean that I could only set that specific header 96 times and I was reaching that limit due to the number of restarts (surely not?) 

> Even if I restarted the request flow three times, there's only a total of _eight_ VCL states and they don't all get executed in a single request flow, so that's only a maximum of 24 times I would have called set on the `X-VCL-Route` header (way under the 96 limit).

Turns out it's a bit more a broad range of possible problem causes than just that.

The official Fastly response to my question about what this documentation means was as follows...

> `header count = 96` specifically means 96 headers can be present for an object (this includes the request and response headers). The header overflow can be caused by exceeding _any_ of the limits defined in the documentation throughout the whole VCL process.

So it's more likely that we _were_ hitting the http header size limit, but it was being transformed into this raw Varnish error instead.

At any rate this is something to keep in mind and be mindful of.

<div id="5"></div>
## Hit for Pass

You may notice in Varnish's built-in `vcl_fetch` the following logic:

```vcl
sub vcl_fetch {
    if (!beresp.cacheable) {
        return (pass);
    }
    if (beresp.http.Set-Cookie) {
        return (pass);
    }
    return (deliver);
}
```

Now typically when you `return(pass)` you do that in `vcl_recv` to indicate to Varnish you do not want to lookup the content in the cache and to skip ahead to fetching the content from the origin. But when you `return(pass)` from `vcl_fetch` it causes a slightly different behaviour. Effectively we're telling Varnish we don't want to cache the content we've received from the origin.

> In the case of the VCL logic above, we're not caching this content because we can see there is a cookie set (indicating possibly unique user content).

You'll probably also notice in some organisation's own custom `vcl_fetch` the following additional logic:

```vcl
if (beresp.http.Cache-Control ~ "private") {
  return(pass);
}
```

This content isn't cached simply because the backend has indicated (via the `Cache-Control` header) that the content is `private` and so should not be cached.

But you'll find that even though you've executed a `return(pass)` operation, Varnish will _still_ create an object and cache it.

The object it creates is called a "hit-for-pass" (if you look back at the Fastly request flow diagram above you'll see it referenced) and it is given a ttl of 120s (i.e. it'll be cached for 120 seconds).

> Note: the ttl can be changed using vcl but it should be kept small. Varnish implements a type known as a 'duration' and takes many forms: ms (milliseconds), s (seconds), m (minutes), h (hours), d (days), w (weeks), y (years). For example, `beresp.ttl = 1h`.

The reason Varnish creates an object and caches it is because if it _didn't_, when `return(pass)` is executed and the content subsequently is not cached, then if another request is made for that same resource, we would find "request collapsing" causes a performance issue for users.

Request collapsing is where Varnish blocks requests for what looks to be the same uncached resource. It does this in order to prevent overloading your origin. So for example, if there are ten requests for an uncached resource, it'll allow one request through to origin and block the other nine until the origin has responded and the content has been cached. The nine requests would then get the content from the cache. 

But in the case of uncachable content (e.g. content that uses cookies typically will contain content that is unique to the user requesting the resource) users are blocked waiting for an existing request for that resource to complete, only to find that as the resource is uncacheable the request needs to be made again (this cycle would repeat for each user requesting this unique/uncacheable resource).

As you can imagine, this is very bad because the requests for this uncachable content has resulted in sequential processing. 

So when we "pass" inside of `vcl_fetch` Varnish prevents this bad sequential processing. It does this by creating a "hit-for-pass" object which has a short ttl of 120s, and so for the next 120s any requests for this same resource will _not_ result in request collapsing (i.e. user requests to origin will not be blocked waiting for an already "in-flight" origin request to complete). Meaning, _all_ requests will be sent straight through to the origin.

In order for this processing to work, we need `vcl_hit` to check for a "hit-for-pass" object. If we check Varnish's built-in logic we can see:

```vcl
sub vcl_hit {
    if (!obj.cacheable) {
        return (pass);
    }
    return (deliver);
}
```

What this does is it checks whether the object we found in the cache has the attribute `cacheable` set to 'false', and if it does we'll not deliver that cached object to the user but instead skip ahead to fetching the resource again from origin.

By default this `cacheable` attribute is set to 'true', but when Varnish executes `return(pass)` from inside of `vcl_fetch` it caches the "hit-for-pass" object with the `cacheable` attribute set to 'false'.

The reason the ttl for a "hit-for-pass" object is supposed to be short is because, for that period of time, your origin is susceptible to multiple requests. So you don't want your origin to become overloaded by lots of traffic for uncacheable content.

It's important to note that an object can't be _re-cached_ (let's say the origin no longer sends a `private` directive) until either the hit-for-pass TTL expires _or_ the hit-for-pass object is purged.

> See [this Varnish blog post](https://info.varnish-software.com/blog/hit-for-pass-varnish-cache) for the full details.

<div id="6"></div>
## Serving Stale

If we get a 5xx error from our origins we don't cache them.

But instead of serving that 5xx to the user (or even a custom 500 error page), we'll attempt to locate a 'stale' version of the content and serve that to the user instead (i.e. 'stale' in this case means a resource that was requested and cached previously, but the object was marked as being something that could be served stale if its 'stale ttl' has yet to expire).

The reason we do this is because serving old (i.e. stale) content is better than serving an error.

In order to serve stale we need to add some conditional checks into our VCL logic.

You'll typically notice in both `vcl_fetch` and `vcl_deliver` there are checks for a 5xx status code in the response we got back from origin, and subsequently a further check for `stale.exists` if we found a match for a 5xx status code. It looks something like the following:

> Note: you don't have to run this code only from `vcl_deliver`. It can be beneficial to do this via `vcl_error` as well because if your origin is unreachable, then it means you'll want to check for stale in `vcl_error` as well. Fastly gives an example of this in [their documentation](https://docs.fastly.com/en/guides/serving-stale-content#serving-stale-content-on-errors).

```vcl
if (<object>.status >= 500 && <object>.status < 600) {
  if (stale.exists) {
    # ...
  }
}
```

> Where `<object>` is either `beresp` (`vcl_fetch`) or `resp` (`vcl_deliver`). 

When looking at the value for `stale.exists`, if it returns 'true', it is telling us that there is an object in the cache whose ttl has expired but as far as Varnish is concerned is still valid for serving as stale content. The way Varnish knows whether to keep a stale object around so it can be used for serving as stale content depends on the following settings:

```vcl
set beresp.stale_while_revalidate = <N>s;
set beresp.stale_if_error = <N>s;
```

> Where `<N>` is the amount of time in seconds you want to keep the object for after its ttl has expired.

- `stale_while_revalidate`: request received, obj found in cache, but ttl has expired (results in cache MISS), so we serve stale while we go to origin and request the content (†).
- `stale_if_error`: request received, obj found in cache, but ttl has expired, so we go to origin and the origin returns an error, so we serve stale.

> † if successful, new content is cached and the TTL is updated to whatever the cache response headers dictate.

Ultimately this means that `stale_if_error` will only ever be initialized if `stale_while_revalidate` fails. Which begs the question... 

Why use `stale_if_error` at all when we _could_ just set `stale_while_revalidate` to a large value, and it would effectively result in the same outcome: stale content served to the user for a set period of time?

This is a question I don't have an answer for, and I've yet to receive a satisfactory answer to from either Fastly or the dev community.

My own opinion on this is that if my origin is unable to serve a `200 OK` response after something like an 1hr of trying via `stale_while_revalidate` then something must be seriously wrong with my origin and so maybe that's the appropriate indicator for what value I should give to it in comparison to the value I would set for `stale_if_error` (which would be the longest value possible).

It feels like setting _both_ `stale_while_revalidate` and `stale_if_error` is kind of like 'cargo culting' (i.e. doing something because it's always been done and so people presume it's the correct way), but at the same time I would hate to find myself in a situation where there _was_ a subtle difference between the two which had a very niche scenario that broke my user's experience because I neglected to set both stale configurations.

Additionally, if these 'stale' settings aren't configured in VCL, then you'll need to provide them as part of Fastly's `Surrogate-Control` header (see [here](https://docs.fastly.com/guides/tutorials/cache-control-tutorial#surrogate-control) for the details, but be aware that VCL configured values will take precedence over response headers from origin):

```vcl
"Surrogate-Control": "max-age=123, stale-while-revalidate=172800, stale-if-error=172800"
```

The use of `beresp.stale_if_error` is effectively the same as Varnish's `beresp.grace`. But be careful if your VCL already utilises Varnish's original `grace` feature, because it will override any Fastly behaviour that is using the `stale_if_error`.

> You can find more details on Fastly's implementation [here](https://docs.fastly.com/guides/performance-tuning/serving-stale-content) as well as a blog post announcing this feature [here](https://www.fastly.com/blog/stale-while-revalidate-stale-if-error-available-today/). If you want details on Varnish 4.0's implementation of serving stale, see [this post](https://info.varnish-software.com/blog/grace-varnish-4-stale-while-revalidate-semantics-varnish).

You might find that you're not serving stale even though you would expect to be. This can be caused by a lack of [shielding](https://docs.fastly.com/guides/performance-tuning/shielding) (an additional Fastly feature that's designed to improve your hit ratio) as you can only serve a stale cached object if the request ended up being routed through the POP that has the content cached (which is more difficult without shielding enabled). 

> Another reason would be to use "[soft purging](https://docs.fastly.com/guides/purging/soft-purges)" rather than hard purges.

Lastly, there's one quirk of Fastly's caching implementation you might need to know about: if you specify a `max-age` of less than 61 minutes, then your content will only be persisted into memory (and there are many situations where a cached object in memory can be removed). To make sure the object is persisted (i.e. cached) on disk and thus available for a longer period of time for serving stale, you must set a `max-age` above 61 minutes.

It's worth reading [fastly's notes on why you might not be serving stale](https://docs.fastly.com/guides/performance-tuning/serving-stale-content#why-serving-stale-content-may-not-work-as-expected) which additionally includes notes on things like the fact that their implementation is an LRU (Least Recently Used) cache, and so even if a cached object's TTL has not expired it might be so infrequently requested/accessed that it'll be evicted from the cache any way! The LRU affects both fresh (i.e. non-expired) objects _and_ 'stale' objects (i.e. TTL has expired but they have stale config values that haven't expired).

### Stale for Client Devices

It's worth reiterating a segment of the earlier [caching priority list](#caching-priority-list) which is that `Cache-Control` _can_ include serving stale directives such as `stale-while-revalidate` and `stale-if-error`, but they are typically utilized with `Surrogate-Control` more than they are with `Cache-Control`. If Fastly receives no `Surrogate-Control` but it does get `Cache-Control` with those directives it _will_ presume those are defined for its benefit.

Client devices (e.g. web browsers) can respect those stale directives, but it's not very well supported currently (see [MDN compatibility table](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Browser_compatibility)).

### Caveats of Fastly's Shielding

Be careful with changes you make to a request as they could result in the lookup hash to change between the edge POP nodes and shield POP nodes. 

Also, be aware that the "backend" will change when shielding is enabled. Traditionally (i.e. without shielding) you defined your backend with a specific value (e.g. an S3 bucket or a domain such as `https://app.domain.com`) and it would stay set to that value unless you yourself implemented custom vcl logic to change its value. 

But with shielding enabled, the 'cluster edge' node will dynamically change the backend to be a shield node value (as it's effectively _always_ going to pass through that node if there is no cached content found). Once on the 'cluster edge' node within the shield POP, _its_ "backend" value is set to whatever your actual origin is (e.g. an S3 bucket).

It's probably best to only modify your backends dynamically whilst your VCL is executing on the shield (e.g. `if (!req.backend.is_shield)`, maybe abstract in a variable `declare local var.shield_node BOOL;`) and to also only `restart` a request in vcl_deliver when executing on a node within the shield POP. 

You might also need to modify vcl_hash so that the generated hash is consistent with the edge POP's 'cluster edge' node if your shield POP nodes happen to modify the request! Remember that modifying either the host or the path will cause a different cache key to be generated and so modifying that in either the edge POP _or_ the shield POP means modifying the relevant vcl_hash subroutine so the hashes are _consistent_ between them.

```
sub vcl_hash {
  # we do this because we want the cache key to be identical at the edge and
  # at the shield. Because the shield rewrites req.url (but not the edge), we
  # need align vcl_hash by using the original Host and URL.
  set req.hash += req.http.X-Original-URL;
  set req.hash += req.http.X-Original-Host;

  #FASTLY hash

  return(hash);
}
```

> Note: alternatively you could move rewriting of the URL to a state after the hash lookup, such as vcl_miss (e.g. modifying the `bereq` object).

Lastly, when enabling shielding, make sure to deploy your VCL code changes first _before_ enabling shielding. This way you avoid a race condition whereby a shield has old VCL (i.e. no conditional checks for either `Fastly-FF` or `req.backend.is_shield`) and thus tries to do something that should only happen on the edge cache node.

When using `Fastly-Debug:1` to inspect debug response headers, we might want to look at `fastly-state`, `fastly-debug-path` and `fastly-debug-ttl`. These would have values such as...

```
< fastly-state: HIT-STALE
< fastly-debug-path: (D cache-lhr6346-LHR 1563794040) (F cache-lhr6324-LHR 1563794019)
< fastly-debug-ttl: (H cache-lhr6346-LHR -10.999 31536000.000 20)
```

The `fastly-debug-path` suggests we delivered from the 'cluster edge' node `lhr6346`, while we fetched from the 'cluster shield' node `lhr6324`. The `fastly-debug-ttl` header suggests we got a HIT (`H`) from the 'cluster edge' node `lhr6346` but this is just a side-effect of the stale/cached content (coming back from the 'cluster shield' node) being stored in-memory on the 'cluster edge' node and so it's indicated as a HIT from the 'cluster edge' node when really it came from the 'cluster shield' node (the `fastly-debug-ttl` header is set on the 'cluster edge' node, which re-enforces this understanding).

What makes it confusing is that you don't necessarily know if the request went to the 'cluster shield' node (i.e. the fetching node) or whether the stale content actually came from the 'cluster edge' node's in-memory cache. The only way to be sure is to check the `fastly-state` response header and see if you got back `HIT-STALE` or `HIT-STALE-CLUSTER`.

Another confusing aspect to `fastly-debug-ttl` is that with regards to `stale-while-revalidate` you could end up seeing a `-` in the section where you might otherwise expect to see the grace period of the object (i.e. how long can it be served stale for while revalidating). This can occur when the origin server hasn't sent back either an `ETag` header or `Last-Modified` header. Fastly still serves stale content if the `stale-while-revalidate` TTL is still valid but the output of the `fastly-debug-ttl` can be confusing and/or misleading.

> Something else to note, while I write this August 2019 update is that the `fastly-debug-ttl` only every displays the 'grace' value when it comes to `stale-if-error`, meaning if you're trying to check if you're serving `stale-while-revalidate` by looking at the grace period you might get confused when you see the `stale-if-error` grace period (or worse a `-` value), this is because the `fastly-debug-ttl` header isn't as granular as it should be. Fastly have indicated that they intend on making updates to this header in the future in order for the values to be much clearer.

### Different actions for different states

So if we find a stale object, we need to deliver it to the user. But the action you take (as far as Fastly's implementation of Varnish is concerned) depends on which state Varnish currently is in (`vcl_fetch` or `vcl_deliver`). 

When you're in `vcl_fetch` you'll `return(deliver_stale)` and in `vcl_deliver` you'll `return(restart)`.

#### Why the difference?

The reason for this difference is to do with Fastly's Varnish implementation and how they handle 'clustering'.

According to Fastly you'll typically find, due to the way their 'clustering' works, that `vcl_fetch` and `vcl_deliver` generally run on _different_ servers (although that's not always the case, as we'll soon see) and so different servers will have different caches.

> "What we refer to as 'clustering' is when two servers within the same POP communicate with each other for cache optimization purposes" -- Fastly

#### The happy path (stale found in `vcl_fetch`)

If a stale object is found in `vcl_fetch`, then `deliver_stale` will send the found stale content to `vcl_deliver` with a status of 200. This means when `vcl_deliver` checks the status code it'll not see a 5xx and so it'll just deliver that stale content to the user.

#### The longer path (stale found in `vcl_deliver`)

Imagine `vcl_fetch` is running on server `A` and `vcl_deliver` is running on server `B` (due to Fastly's 'clustering' infrastructure) and you looked for a stale object in `vcl_fetch`. The stale object might not exist there and so you would end up passing whatever the origin's 5xx response was onto `vcl_deliver` running on server `B`. 

Now in `vcl_deliver` we check `stale.exists` and it might tell us that an object _was_ found, remember: this is a different server with a different cache. 

In this scenario we have a 5xx object that we're about to deliver to the client, but on this particular server (`B`) we've since discovered there is actually a stale object we can serve to the user instead. So how do we now give the user this stale object and not the 5xx content that came from origin?

In order to do that, we need to restart the request cycle. When we `return(restart)` in `vcl_deliver` it forces Fastly's version of Varnish to break its clustering behaviour and to now route the request completely through whatever server `vcl_deliver` is currently running on (in this example it'll ensure the entire request flow will go through server `B` - which we know has a stale object available).

This means we end up processing the request again, but this time when we make a request to origin and get a 5xx back, we will (in `vcl_fetch`) end up finding the stale object (remember: our earlier restart has forced clustering behaviour to be broken so we're routing through the same server `B`). Now we find `stale.exists` has matched in `vcl_fetch` we will `return(deliver_stale)` (which we failed to do previously due to the stale object not existing on server `A`) and this means the stale content with a status of 200 is passed to `vcl_deliver` and that will subsequently deliver the stale object to the client.

#### The unhappy path (stale not found anywhere)

In this scenario we don't find a stale object in either `vcl_fetch` or `vcl_deliver` and so we end up serving the 5xx content that we got from origin to the client. Although you may want to attempt to restart the request and use a custom header (e.g. `set req.http.X-Serve-500-Page = "true"`) in order to indicate to `vcl_recv` that you want to short-circuit the request cycle and serve a custom error page instead.

<div id="7"></div>
## Disable Caching

It's possible to disable caching for either the client or fastly, or both! But it gets confusing with all the various documentation pages fastly provides to know which one is the source of truth (useful information is spread across all of them).

In my experience I've found the following to be sufficient...

> Note: any time you want Fastly to cache your content use `Surrogate-Control`, and this will take precedence over `Cache-Control` _EXCEPT_ when `Cache-Control` has the value `private` included somewhere inside it. So generally if I'm doing that I'll just make sure I'm not sending a `Surrogate-Control` as it'll just be confusing for anyone reading that code.

- Disable Client Caching: `Cache-Control: no-store, must-revalidate` ([docs](https://docs.fastly.com/guides/tutorials/cache-control-tutorial#applying-different-cache-rules-for-fastly-and-browsers))
- Disable CDN Caching: `Cache-Control: private` ([docs](https://docs.fastly.com/guides/tutorials/cache-control-tutorial#do-not-cache))
- Disable ALL Caching: `Cache-Control: no-cache, no-store, private, must-revalidate, max-age=0, max-stale=0, post-check=0, pre-check=0` + `Pragma: no-cache` + `Expires: 0` ([docs](https://docs.fastly.com/guides/debugging/temporarily-disabling-caching))

<div id="8"></div>
## Logging

With Fastly, to set-up logging you'll need to use their UI, as this means they can configure the relevant integration with your log aggregation provider. But what people don't realise is that by default Fastly will generate a subroutine called `vcl_log`.

Now, if you don't specify a "log format", then the generated subroutine will look like this:

```
sub vcl_log {
#--FASTLY LOG START
  # default response conditions
  log {"syslog <service_id> <service_name> :: "} ;
#--FASTLY LOG END
}
```

If you _do_ provide a log format value, then it could look something like the following:

```
sub vcl_log {
#--FASTLY LOG START
  # default response conditions
  log {"syslog <service_id> <service_name> :: "} req.http.Fastly-Client-IP " [" strftime({"%Y-%m-%d %H:%M:%S"}, time.start) "." time.start.msec_frac {"] ""} cstr_escape(req.request) " " cstr_escape(req.url) " " cstr_escape(req.proto) {"" "} resp.status " " resp.body_bytes_written " " tls.client.protocol " " fastly_info.state " " req.http.user-agent;
#--FASTLY LOG END
}
```

The `vcl_log` subroutine executes _after_ `vcl_deliver`. So it's the _very last_ routine to be executed before Varnish completes the request (so this is even after it has delivered the response to the client). The reason this subroutine exists is because logging inside of `vcl_deliver` (which is how Fastly _used_ to work - i.e. they would auto-generate the log call inside of `vcl_deliver`) they wouldn't have certain response variables available, such as determining how long it took to send the first byte of the response to the client.

What confused me about the logging set-up was the fact that I didn't realise I was looking at things too much from an 'engineering' perspective. By that I mean, not all of the UI features Fastly provides exist just for my benefit :-) 

So what I was confusedly thinking was: "why is Fastly generating a single log call in `vcl_log`, is there a technical reason for that or a limitation with Varnish?" and it ended up simply just being that some Fastly users don't want to write their own VCL (or don't know anything about writing code) and so they are quite happy with knowing that the UI will trigger a single log call at the end of each request cycle.

Where (as a programmer) I was expecting to add log calls all over the place! But I never realised a `vcl_log` was being auto-generated for me by Fastly. Meaning... I never realised that an extra log call (after all the log calls I was manually adding to my custom VCL) was being executed.

So when I eventually stumbled across a Fastly documentation page talking about "duplicate logs", I started digging deeper into why that might be ...and that's where I discovered `vcl_log` was a thing.

Now, the documentation then goes on to mention using the UI to create a custom 'condition' to prevent Fastly's auto-generated log from being executed. This intrigued me. I was thinking "what is this magical 'condition feature' they have?". Well, turns out that I was again thinking too much like an engineer and not enough like a normal Joe Bloggs user who knows nothing about programming, as this 'feature' just generates a standard conditional 'if statement' code block around the auto-generated log call, like so:

```
sub vcl_log {
#--FASTLY LOG START
  # default response conditions
  # Response Condition: generate-a-log Prio: 10
  if( !req.url ) {
    log {"syslog <service_id> <service_name> :: "} ;
  }
#--FASTLY LOG END
}
```

As you can probably tell, this recommended condition will never match and so the log call isn't executed.

Mystery solved.

### Logging Memory Exhaustion

We discovered an issue with our logging code which meant we would see `(null)` log lines being generated. The cause of this turned out to be the Fastly 'workspace' (as they refer to it) would run out of memory while generating our log output that they would stream to either GCS or AWS S3.

> You can inspect the workspace using the undocumented `workspace.bytes_free` property, and this value will change as you make modifications to your request/response and other related objects.

The 'workspace' is the amount of memory allocated to _each_ in-flight request, and its value is set to 128k. So although the JSON object we're building is small, _prior_ to that, our request flow might be making lots of HTTP header modifications and all those things are counting towards the overall memory being consumed.

So ultimately be wary of making too many HTTP modifcations as you might discover you'll end up losing log data.

> Note: we have two seperate services, one production and one staging and we run the same VCL code in both. This requires us to have logic that checks whether our code is running in the stage environment or not.

```
declare local var.stage_service_id STRING;
declare local var.prod_service_id STRING;
declare local var.running_service_id STRING;

set var.stage_service_id = req.http.X-Fastly-ServiceID-Stage;
set var.prod_service_id = req.http.X-Fastly-ServiceID-Prod;
set var.running_service_id = var.stage_service_id;

if (req.service_id == var.prod_service_id) {
  set var.running_service_id = var.prod_service_id;
}

declare local var.json STRING;

set var.json = "{" +
  "%22http%22: {" +
    "%22body_size%22: %22" + resp.body_bytes_written + "%22," +
    "%22content_type%22: %22" + resp.http.Content-Type + "%22," +
    "%22host%22: %22" + req.http.host + "%22," +
    "%22method%22: %22" + req.method + "%22," +
    "%22path%22: %22" + json.escape(req.url) + "%22," +
    "%22protocol%22: %22" + req.proto + "%22," +
    "%22request_time%22: " + time.to_first_byte + "," +
    "%22served_stale%22: " + req.http.X-BF-Served-Stale + "," +
    "%22status_code%22: " + resp.status + "," +
    "%22tls_version%22: %22" + tls.client.protocol + "%22," +
    "%22uri%22: %22" + json.escape(if(req.http.Fastly-SSL, "https", "http") + "://" + req.http.host + req.url) + "%22," +
    "%22user_agent%22: %22" + json.escape(req.http.User-Agent) + "%22" +
  "}," +
  "%22network%22: {" +
    "%22client%22: {" +
      "%22ip%22: %22" + req.http.Fastly-Client-IP + "%22" +
    "}," +
    "%22server%22: {" +
      "%22state%22: %22" + fastly_info.state + "%22" +
    "}" +
  "}," +
  "%22timestamp%22: " + time.start.msec + "," +
  "%22timestamp_sec%22: " + time.start.sec + "," + # exists to support BigQuery
  "%22upstreams%22: {" +
    "%22service%22: %22" + req.http.X-BF-Backend + "%22" +
  "}"
"}";
```

Generating JSON structured output is not easy in VCL. The above VCL snippet demonstrates the 'manual' approach (encoding `"` as `%22` and concatenating across multiple lines). This was the best solution for us as the alternatives were unsuitable:

1. [github.com/fastly/vcl-json-generate](https://github.com/fastly/vcl-json-generate) is wildly verbose in comparison, although it's supposed to make things easier?
2. using "long string" variation (everything must be all on _one_ line!?), e.g. `{"{  "foo":""} req.http.X-Foo {"",  "bar":"} req.http.X-Bar {" }"};`

So manually constructing the JSON was what we opted for, and once you get used to the endless `%22` it's actually quite readable IMHO (when compared to the alternatives we just described).

Now here's how we worked-around the memory issue causing a `(null)` log line: before we trigger our `log` call we first check that Fastly's workspace hasn't been exhausted (the variable contains the error code raised by the last function). If we find an 'out of memory' error, then we won't attempt to call the `log` function (as that would result in the `(null)` for a log line).

> Documentation: [docs.fastly.com/vcl/variables/fastly-error/](https://docs.fastly.com/vcl/variables/fastly-error/)

It's important to reiterate that this isn't a _solution_, but a _work-around_. We don't know how many logs we're losing due to memory exhaustion (it could be lots!) so this is a problem we need to investigate further (as of October 2019, ...wow did I really start writing this post two years ago!!? time flies heh).

```
if (fastly.error != "ESESOOM") {
  log "syslog " var.running_service_id " logs_to_s3 :: " var.json;
  log "syslog " var.running_service_id " logs_to_gcs :: " var.json;
}
```

> Note: `logs_to_s3` and `logs_to_gcs` are references to the two different log streams we setup within the Fastly UI.

<div id="8.1"></div>
## Restricting requests to another Fastly service

We had a requirement where by we had a request flow that looked something like the following...

```
Client > Fastly (service: foo) > Fastly (service: bar) > Origin
```

In this request flow both the Fastly services were managed by us (i.e. managed by BuzzFeed) and we wanted to ensure that the service `bar` could only accept requests when they came via service `foo`.

We didn't want to rely on HTTP request headers as these can be spoofed very easily. Fastly suggested the following VCL based solution...

```
if (!req.http.myservice && (client.ip ~ fastly_ip_ranges))
 error 808;
}
```

> Note: `808` is a custom error status (see [Error Handling](#4.0) for more context about using non-standard status codes). I typically prefer to use the 9xx range.

The idea being we can check if the `client.ip` matches a known Fastly POP IP range. If it doesn't match that then we'll reject the request. Additionally, service `bar` would check to see if a special request header was set by service `foo` (e.g. `set req.http.myservice = "yes";`).

An alternative approach also suggested by Fastly was to replace the IP check for a request header check for a Fastly specific header `Fastly-FF` (which is only set when the request came from a Fastly POP):

```
if (req.http.fastly-ff && !req.http.myservice) {
  error 808;
}
```

## Custom Error Pages

Near the beginning of this post we looked at [error handling](#4.0) in VCL. I want to now revisit this topic with regards to serving your own custom error pages from the edge synthetically, so as to give you a real example of how this might be put together.

Below is example code that you can also run via [Fastly's Fiddle tool](https://fiddle.fastlydemo.net/fiddle/cfd09f79):

```
sub vcl_recv {
  if (req.restarts == 1) {
    if (req.http.X-Error) {
      error std.strtol(req.http.X-Error, 10);
    }
  }
}

sub vcl_deliver {
  if (req.restarts == 0 && resp.status >= 400) {
    if(fastly_info.state ~ "HIT($|-)") {
      set req.http.X-FastlyInfoStatus = fastly_info.state;
      set req.http.X-CachedObjectHitCounter = obj.hits;
    }
    set req.http.X-Error = resp.status;
    return(restart);
  }

  if (req.restarts == 1) {
    if (req.http.X-FastlyInfoStatus) {
      set resp.http.X-Cache = req.http.X-FastlyInfoStatus;
    }
    
    if (req.http.X-CachedObjectHitCounter) {
      set resp.http.X-Cache-Hits = std.strtol(req.http.X-CachedObjectHitCounter, 10);
    }
  }
}

vcl_error {
  synthetic {"
    <!doctype html>
    <html>
    <head>
      <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
      <title>Page Not Found</title>
      <meta name="generator" content="fastly">
    </head>
      <body>
        error happened
      </body>
    </html>
  "};

  return(deliver);
}
```

The flow is actually quite simple once you have these pieces put together. The principle of what we're trying to achieve is that we want to serve a custom error page from the edge server (rather than using the error page that may have be sent by an origin).

The way we'll do this is that when we get the error response from the origin we'll restart the request and upon restarting the request we'll trigger `vcl_error` (via the `error` directive) and construct a synthetic response.

Once we have that synthetic response we'll also need to make some modifications so that Fastly's default VCL doesn't make it look like we don't get cache HITs from future requests for this same resource. 

This otherwise happens because Fastly's default VCL has a basic check in `vcl_deliver` for an internal 'hit' state which does occur, but then because we restart the request and trigger `vcl_error`, that internal hit state is replaced by an 'error' state and so it ends up confusing Fastly's VCL logic.

Let's dig into the code a bit...

So first of all we need to check in `vcl_deliver` for an error status (and because we plan on restarting the request, we don't want to have an endless loop so we'll check the restart count as well):

```
if (req.restarts == 0 && resp.status >= 400) {
```

If we find we have indeed come from `vcl_hit`, then we store off contextual information about that cached object into the `req` so it will persist the restart and thus be usable later on when required:

```
set req.http.X-FastlyInfoStatus = fastly_info.state;
set req.http.X-CachedObjectHitCounter = obj.hits;
```

Regardless of whether we came from `vcl_hit` or if we came from `vcl_fetch`, we store off the error response status code into the `req` object so again we may access it after the request is restarted:

```
set req.http.X-Error = resp.status;
```

Finally, we'll restart the request. At this point let's now look at `vcl_recv` where we first check if we are dealing with a restart, and if so we extract the response error status code and pass it over to `vcl_error` by way of the `error` directive (notice we have to use a Fastly function to convert it from a STRING to a INTEGER):

```
if (req.restarts == 1) {
  if (req.http.X-Error) {
    error std.strtol(req.http.X-Error, 10);
  }
}
```

Once we reach `vcl_error` we can now construct our custom error page. In the example code I give I just use a single synthetic object to represent the error but in real-life you'd check the status code provided to `vcl_error` and then construct a _specific_ error page. Once we create the object we jump back over to `vcl_deliver`.

Now we're back at `vcl_deliver` for the second time you'll find the first check we had will no longer pass (avoiding an endless loop) but the second check _will_ pass:

```
if (req.restarts == 1) {
```

Once we know we're into the restarted request we can check the `req` object for the headers/information we were tracking earlier and use that information to update two response headers that otherwise are set by Fastly: `X-Cache` and `X-Cache-Hits`.

The reason we need to override these two headers is because otherwise Fastly records them as `X-Cache: MISS` and `X-Cache-Hits: 0` even after a secondary request for the same resource. This would be concerning to people, because it seems to suggest we went back to the origin again when we should have gotten a cache HIT (we in fact _do_ get a cache HIT but Fastly is incorrectly reporting that fact).

To understand what's going on with Fastly incorrectly reporting the state of the request, let's look at the request flow...

```
# Request 1

vcl_recv > vcl_hash > vcl_miss > vcl_fetch > vcl_deliver (restart) > vcl_recv > vcl_error > vcl_deliver.

# Request 2

vcl_recv > vcl_hash > vcl_hit > vcl_deliver (restart) > vcl_recv > vcl_error > vcl_deliver.
```

For request 1 we have a cold cache scenario whereby the requested resource (e.g. `/i-dont-exist`) isn't cached. So we go to the origin and the origin returns a `404 Not Found`. At that point `vcl_deliver` identifies the error and restarts the request. We construct a custom error page within `vcl_error` and finally serve the custom error page via `vcl_deliver`.

After request 1 the response headers would suggest (correctly) that we got a `X-Cache: MISS` and `X-Cache-Hits: 0`.

So what happens if we request `/i-dont-exist` again? Well, unless we fixed those headers, the cache miss and zero hits would be reported again even though we know that the second request got a cache HIT.

This happens because Fastly's internal logic for setting `X-Cache: MISS` and `X-Cache-Hits` looks a bit like this:

```
set resp.http.X-Cache = if(resp.http.X-Cache, resp.http.X-Cache ", ","") if(fastly_info.state ~ "HIT($|-)", "HIT", "MISS");

if(!resp.http.X-Cache-Hits) {
  set resp.http.X-Cache-Hits = obj.hits;
} else {
  set resp.http.X-Cache-Hits = resp.http.X-Cache-Hits ", " obj.hits;
}
```

Even though the second request gets a cache HIT for the origin's _original_ error response, the serving of the custom error page from the edge causes the Fastly internal state (i.e. `fastly_info.state`) to report as `ERROR` instead of a `HIT` (_that_ happens because we call `error` from within `vcl_recv` after restarting the request).

This is why when we get back into `vcl_deliver` we override those headers according to the values we tracked within the initial request flow.

<div id="9"></div>
## Conclusion

So there we have it, a run down of how some important aspects of Varnish and VCL work (and specifically for Fastly's implementation).

One thing I want to mention is that I am personally a HUGE fan of Fastly and the tools they provide. They are an amazing company and their software has helped BuzzFeed (and many other large organisations) to scale massively with ease.

I would also highly recommend watching this talk by Rogier Mulhuijzen (Senior Varnish Engineer - who currently works for Fastly) on "Advanced VCL": [vimeo.com/226067901](https://vimeo.com/226067901). It goes into great detail about some complex aspects of VCL and Varnish and really does a great job of elucidating them.

Also recommended is [this Fastly talk](https://vimeo.com/178057523) which details how 'clustering' works (see also [this fastly community support post](https://support.fastly.com/hc/en-us/community/posts/360040445272--Fastly-Force-Shield-AND-Fastly-No-Shield-Usage) that details how to utilize the request headers `Fastly-Force-Shield` and `Fastly-No-Shield`).

Lastly, there was [a recent article from an engineer working at the Financial Times](https://medium.com/@samparkinson_/making-a-request-to-the-financial-times-b2119a2f422d), detailing the complete request flow from DNS to Delivery. It's very interesting and covers a lot of information about Fastly. Highly recommended reading.
