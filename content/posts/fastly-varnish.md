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
- [Varnish Default VCL](#2)
- [Fastly Default VCL](#3)
- [Custom VCL](#3.0)
- [Fastly TTLs](#3.1)
- [Fastly Default Cached Status Codes](#3.2)
- [Fastly Request Flow Diagram](#4)
- [Error Handling](#4.0)
- [State Variables](#4.1)
- [Persisting State](#4.2) (inc. clustering architecture)
- [Hit for Pass](#5)
- [Serving Stale](#6) (inc. caveats of Fastly’s Shielding)
- [Disable Caching](#7)
- [Logging](#8)
- [Conclusion](#9)

<div id="1"></div>
## Introduction

[Varnish](https://varnish-cache.org/) is an open-source HTTP accelerator.  
More concretely it is a web application that acts like a [HTTP reverse-proxy](https://en.wikipedia.org/wiki/Reverse_proxy). 

You place Varnish in front of your application servers (those that are serving HTTP content) and it will cache that content for you. If you want more information on what Varnish cache can do for you, then I recommend reading through their [introduction article](https://varnish-cache.org/intro/index.html) (and watching the video linked there as well).

[Fastly](https://www.fastly.com/) is many things, but for most people they are a CDN provider who utilise a highly customised version of Varnish. This post is about Varnish and explaining a couple of specific features (such as hit-for-pass and serving stale) and how they work in relation to Fastly's implementation of Varnish.

One stumbling block for Varnish is the fact that it only accelerates HTTP, not HTTPS. In order to handle HTTPS you would need a TLS/SSL termination process sitting in front of Varnish to convert HTTPS to HTTP. Alternatively you could use a termination process (such as nginx) _behind_ Varnish to fetch the content from your origins over HTTPS and to return it as HTTP for Varnish to then process and cache.

> Note: Fastly helps both with the HTTPS problem, and also with scaling Varnish in general.

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

> Note: `vcl_hash` is the only exception because it's not a _state_ per se, so you don't execute `return(hash)` but `return(lookup)` as this helps distinguish that we're performing an action and not a state change (i.e. we're going to _lookup_ in the cache).

The reason for this post is because when dealing with Varnish and VCL it gets very confusing having to jump between official documentation for VCL and Fastly's specific implementation of it. Even more so because the version of Varnish Fastly are using is now quite old and yet they've also implemented some features from more recent Varnish versions. Meaning you end up getting in a muddle about what should and should not be the expected behaviour (especially around the general request flow cycle).

Ultimately this is not a "VCL 101". If you need help understanding anything mentioned in this post, then I recommend reading:

- [Varnish Book](http://book.varnish-software.com/4.0/)
- [Varnish Blog](https://info.varnish-software.com/blog)
- [Fastly Blog](https://www.fastly.com/blog)

> Fastly has a couple of _excellent_ articles on utilising the `Vary` HTTP header (highly recommended reading).

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

> Note: in VCL you can allow _any_ response status code to be cached by executing `set beresp.cacheable = true;` within `vcl_fetch`.

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

Now, I wanted to talk briefly about error handling because there are situations where an error can occur, and it can cause Varnish to change to an _unexpected_ state. I'll give a real-life example of this...

We noticed that we were getting a raw `503 Backend Unavailable` error from Varnish displayed to our customers. This is odd? We have VCL code in `vcl_fetch` (the state that you move to once the response from the origin has been received by Fastly/Varnish) which checks the response status code for a 5xx and handles the error there. Why didn't that code run?

Well, it turns out that `vcl_fetch` is only executed if the backend/origin was considered 'available' (i.e. Fastly could make a request to it). In this scenario what happened was that our backend _was_ available but there was a network issue with one of Fastly's POPs which meant it was unable to route certain traffic, resulting in the backend appearing as 'unavailable'.

So what happens in those scenarios? In this case Varnish won't execute `vcl_fetch` because of course no request was ever made (how could Varnish make a request if it thinks the backend is unavailable), so instead Varnish jumps from `vcl_miss` (where the request to the backend would be initiated from) to `vcl_error`.

This means in order to handle that very specific error scenario, we'd need to have similar code for checking the status code (and trying to serve stale, see later in this article for more information on that) within `vcl_error`.

<div id="4.1"></div>
## State Variables

Each Varnish 'state' has a set of built-in variables you can use.

Below is a list of available variables and which states they're available to:

> Based on Varnish 3.0 (which is the only explicit documentation I could find on this), although you can see in various request flow diagrams for different Varnish versions the variables listed next to each state. But [this](http://book.varnish-software.com/3.0/VCL_functions.html#variable-availability-in-vcl) was the first explicit list I found.

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

We can see in [Fastly's documentation](https://docs.fastly.com/guides/performance-tuning/request-collapsing), certain VCL subroutines run on the edge and some on the shield (i.e. what we're calling "cluster"):

- Edge: 
  - `vcl_recv`, `vcl_hash` †, `vcl_deliver`, `vcl_log`, `vcl_error`  
- Shield (cluster): 
  - `vcl_miss`, `vcl_hit`, `vcl_pass`, `vcl_fetch`, `vcl_error`

> † not documented, but Fastly support suggested it would execute at the edge.

### Clustering

I've used the terminology "cluster node" to describe these cache server nodes handling the alternative states (i.e. those nodes _NOT_ handling recv/hash/deliver), but fastly uses the term "shield" node while describing the concept of having different cache nodes handle different states as "clustering". I've avoided this terminology because it overlaps with _another_ fastly feature called [Shielding](https://docs.fastly.com/guides/performance-tuning/shielding) and so I didn't want those two concepts to get confused.

The following diagram visualizes the approach for how a request inside of a POP will reach a specific cache node...

<a href="../../images/fastly-pop.png">
    <img src="../../images/fastly-pop.png">
</a>

The benefit of clustering (summarized in [this fastly community support post](https://support.fastly.com/hc/en-us/community/posts/360046680252-What-is-Clustering-)) is that your request only ever goes through (at _most_) two cache server nodes (the edge node and a cluster/shield node). If clustering was disabled, then an edge node would handle the complete request 'state' life cycle (e.g. recv/hash/fetch/deliver) and thus all the nodes within a POP (at the time of writing: 64 of them) would have to go through to origin in order to request your content.

With clustering enabled, the resource cache key/hash is used to identify a specific cluster/shield node (this is referred to as the "primary" node, or _fetching_ node). This clustering approach means multiple requests to different edge nodes would all go to the same "primary" cache node to _fetch_ the content from the origin (and _request collapsing_ would help protect the origin from traffic overload).

The primary cache node will store the content on-disk, while the edge cache node will store the content in-memory. Finally, any node within the cluster can be selected as the 'edge' node for an incoming request (its selected at random), hence the in-memory copy of the cached content could exist there and you only hit one cache server before a response is served back to the user.

When we `return(restart)` a request, we _break_ 'clustering' which means the request will go back to an edge server and that edge server will handle the full request cycle. This is where a request header such as `Fastly-Force-Shield: 1` will re-enable clustering.

OK, so two more _really_ important things to be aware of at this point:

1. Data added to the `req` object _cannot_ persist across boundaries (except for when initially moving from the edge to the cluster).
2. Data added to the `req` object _can_ persist a restart, but _not_ when they are added from the cluster environment.

For number 1. that means: `req` data you set in `vcl_recv` and `vcl_hash` will be available in states like `vcl_pass` and `vcl_miss`.

For number 2. that means: if you were in `vcl_deliver` and you set a value on `req` and then triggered a restart, the value would be available in `vcl_recv`. Yet, if you were in `vcl_miss` for example and you set `req.http.X-Foo` and let's say in `vcl_fetch` you look at the response from the origin and see the origin sent you back a 5xx status, you might decide you want to restart the request and try again. But if you were expecting `X-Foo` to be set on the `req` object when the code in `vcl_recv` was re-executed, you'd be wrong. That's because the header was set on the `req` object while it was in a state that is executed on a cluster node; and so the `req` data set there doesn't persist a restart.

> SUMMARY: MODIFICATIONS ON THE 'CLUSTER' DON’T PERSIST TO 'EDGE'

If you're starting to think: "this makes things tricky", you'd be right :-)

### Breadcrumb Trail

Let's now revisit our requirement, which was to create a breadcrumb trail using a HTTP header (this is where all this context becomes important).

The first thing we have to do (in `vcl_recv`) is:

```
if (req.restarts == 0) {
  set req.http.X-VCL-Route = "VCL_RECV";
} else {
  set req.http.X-VCL-Route = req.http.X-VCL-Route ",VCL_RECV";
}
```

> Note: we check `req.restarts` to make sure we don't include a leading `,` unnecessarily.

So we know from the "[State Variables](#4.1)" section earlier, that the `req` object is available for reading and writing.

Varnish pretty much always calls `vcl_hash` at the end of `vcl_recv` so we add the following into `vcl_hash`:

```
set req.http.X-VCL-Route = req.http.X-VCL-Route ",VCL_HASH(host: " req.http.host ", url: " req.url ")";
```

You can see we're not setting the value anew on the header, but am _appending_ to the header.

> The idea of appending to the header, again makes things tricky (as we'll see) when we come to trying to persist data across not only the cluster but the caching of an object as well.

With the above 'note' fresh in your mind, let's look at `vcl_miss` and what we do there (remember `vcl_miss` is executing on a cluster node so anything we set on `req` won't persist a restart):

```
set req.http.X-PreFetch-Miss = ",VCL_MISS(" bereq.http.host bereq.url ")";
```

So we still set a value on the `req` object, but we don't append to `X-VCL-Route`, we instead create a new header `X-PreFetch-Miss`.

> The eagle eyed amongst you may notice we took the value we assigned to the new header from `bereq`. I do this for semantic reasons rather than any real _need_. When we move to `vcl_miss` the `req` object is copied to `bereq`. So to make the distinction that `req` (once outside of `vcl_recv`) is only useful for tracking information, I use `bereq` as the value source. But I could have just used `req.http.host` and `req.url` as the value assigned to the header.

Similarly in `vcl_pass` (that also executes on a cluster node), we create a new header again and don't append to `X-VCL-Route`:

```
set req.http.X-PreFetch-Pass = ",VCL_PASS";
```

Now at this point in the request cycle we know that `vcl_pass` and `vcl_miss` are both going to make a request to origin for content and subsequently end up in `vcl_fetch`. But as we'll see in a moment, in `vcl_fetch` we don't assign data to the `req` object this time, instead we assign to `beresp`:

```
set beresp.http.X-VCL-Route = req.http.X-VCL-Route;
set beresp.http.X-PreFetch-Pass = req.http.X-PreFetch-Pass;
set beresp.http.X-PreFetch-Miss = req.http.X-PreFetch-Miss;
set beresp.http.X-PostFetch = ",VCL_FETCH(status: " beresp.status ", url: " req.url ")";
```

So there are a few things happening here:

- We take the `X-VCL-Route` and we assign it to a header of the same name, but on the `beresp` object.
- We take the `X-PreFetch-Pass` and `X-PreFetch-Miss` headers and also assign those to the `beresp` object.
- Finally we create a new header `X-PostFetch` and give it a fresh value that indicates we're in the fetch state.

Why do we do this?

Firstly, when we move from `vcl_pass` or `vcl_miss` to `vcl_fetch` the content we fetched from origin is assigned to the object `beresp`. When we leave `vcl_fetch` the object `beresp` will be stored in the cache. So any header we set on that object will exist when we pull the object from the cache at a later time (e.g. when we lookup content in the cache and we move to `vcl_hit` that subroutine will have access to an `obj` object, which is the `beresp` object pulled from the cache).

Also when we leave `vcl_fetch` and move to `vcl_deliver`, the `beresp` object is copied into a new object (available in `vcl_deliver`) called `resp`. You'll find when `vcl_hit` moves to `vcl_deliver` the `obj` object which was pulled from the cache is also copied over to `resp` in `vcl_deliver`.

Now the reason why we set data onto `beresp` in `vcl_fetch` is because we're ultimately about to cross the boundary of a cluster node (`vcl_fetch`) to an edge node (`vcl_deliver`) and so if we were to continue setting data onto `req` (like we did in `vcl_pass` and `vcl_miss`), then that data would be lost by the time we changed state from `vcl_fetch` to `vcl_deliver`.

The reason we set separate headers for the `vcl_pass`, `vcl_miss` and `vcl_fetch` states is because we wanted to ensure the `beresp` object stored in the cache had a clean request history at the point in time when it was cached. Otherwise we would have issues later on when pulling the object from the cache and trying to append values in `vcl_deliver` (we could end up with large chunks of the breadcrumb trail duplicated).

So for this reason, we separate the baseline routing (i.e. `vcl_recv` and `vcl_hash`) from all the other states (e.g. `vcl_pass`, `vcl_miss`, `vcl_fetch`) and then when we arrive at `vcl_deliver` we grab the baseline `X-VCL-Route` header and append to it the values from `X-PreFetch-Pass`, `X-PreFetch-Miss` and `X-PostFetch` only once we identify (via other inputs - which we'll see shortly) the actual route taken.

Let's look at `vcl_deliver` and see what it is we do there to collate everything together:

```
if (resp.http.X-VCL-Route) {
  set req.http.X-VCL-Route = resp.http.X-VCL-Route;
}

if (fastly_info.state ~ "^HITPASS") {
  set req.http.X-VCL-Route = req.http.X-VCL-Route ",VCL_HIT(object: uncacheable, return: pass)";
}
elseif (fastly_info.state ~ "^HIT") {
  set req.http.X-VCL-Route = req.http.X-VCL-Route ",VCL_HIT(" req.http.host req.url ")";
}
else {
  if (resp.http.X-PreFetch-Pass) {
    set req.http.X-VCL-Route = req.http.X-VCL-Route resp.http.X-PreFetch-Pass;
  }

  if (resp.http.X-PreFetch-Miss) {
    set req.http.X-VCL-Route = req.http.X-VCL-Route resp.http.X-PreFetch-Miss;
  }

  if (resp.http.X-PostFetch) {
    set req.http.X-VCL-Route = req.http.X-VCL-Route resp.http.X-PostFetch;
  }
}

set req.http.X-VCL-Route = req.http.X-VCL-Route ",VCL_DELIVER";
```

So we start by checking if there is a `X-VCL-Route` header available on the incoming object, and if so we overwrite the existing `req.http.X-VCL-Route` header with that `resp` object's value. This is important because we could have arrived at `vcl_deliver` from the edge node (or never even left the edge node) depending on specific scenarios such as going from `vcl_recv` straight to `vcl_error`.

This point about `vcl_error` is _very_ important, we'll skip that discussion for moment and come back to it.

Next in `vcl_deliver`, once we have reset the `X-VCL-Route` header on the `req` object, we look at `fastly_info.state` which is Fastly's own internal system for tracking the current state of Varnish. We first look to see if we've had a 'hit-for-pass' (see the [next section](#5) for details on that), and if so we append the relevant information to the header. Next we check we had a hit from an earlier cache lookup, again, if we have then we append the relevant information.

> If you want more information on `fastly_info.state` see [this community comment](https://community.fastly.com/t/useful-variables-to-log/303/3).

After that we check if the `X-PreFetch-Pass`, `X-PreFetch-Miss` or `X-PostFetch` headers exist, and if so we append the relevant details to the `X-VCL-Route` header. Finally leaving us with appending the _current_ state (i.e. we're in `vcl_deliver`) to the header.

Right, OK so now let's go back and revisit `vcl_error`...

The `vcl_error` subroutine is a tricky one because it exists in _both_ the edge and the cluster environments. 

Meaning if you execute `error 401` from `vcl_recv`, then `vcl_error` will execute in the context of the edge node; whereas if you executed an `error` from a cluster environment like `vcl_fetch`, then `vcl_error` would execute in the context of the cluster node.

Meaning, how you transition information between `vcl_error` and `vcl_deliver` could depend on whether you're on a edge or cluster node.

To help explain this I'm going to give another _real_ example, where I wanted to lookup some content in our cache and if it didn't exist I wanted to restart the request and use a different origin server to serve the content.

To do this I expected the route to go from `vcl_recv`, to `vcl_hash` and the lookup to fail so we would end up in `vcl_miss`. Now from `vcl_miss` I could have triggered a restart, but anything I set on the `req` object at that point (such as any breadcrumb data appended to `X-VCL-Route`) would have been lost as we transitioned from the cluster back to the edge (where `vcl_recv` is).

I needed a way to persist the "miss" breadcrumb, so instead of returning a restart from `vcl_miss` I would trigger a custom error such as `error 901` and inside of `vcl_error` I have the following logic:

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

> † which would include `vcl_recv`, `vcl_hash` and `vcl_miss`. Remember the `req` object _does_ persist across the edge/cluster boundaries, but _only_ when going from `vcl_recv`. After that, anything set on `req` is lost when crossing boundaries.

Now we could have arrived at `vcl_error` from `vcl_recv` (e.g. if in our `vcl_recv` we had logic for checking Basic Authentication and none was found on the incoming request we could decide from `vcl_recv` to execute `error 401`) or we could have arrived at `vcl_error` from `vcl_miss` (as per our earlier example). So we need to check the internal Fastly state to identify this, hence checking `fastly_info.state ~ "^MISS"`.

After that we append to the `obj` object's `X-VCL-Route` header our current state (i.e. so we know we came into `vcl_error`). Finally we look at the status on the `obj` object and see it's a 901 custom status code and so so we append that information so we know what happened.

But you'll notice we don't restart the request from `vcl_error`, because if we did come from `vcl_miss` the data in `obj` would be lost because ultimately it was set in a cluster environment (as `vcl_error` would be running in a cluster environment when coming from `vcl_miss`).

Instead we `return(deliver)`, because all that data assigned to `obj` is guaranteed to be copied into `resp` for us to reference when transitioning to `vcl_deliver` at the edge.

Once we're at `vcl_deliver` we continue to set breadcrumb tracking onto `req.http.X-VCL-Route` as we know that will persist a restart from the edge.

Phew! Well that was easy wasn't it...

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

### Caveats of Fastly's Shielding

Be careful with changes you make to a request as they could result in the lookup hash to change between the edge and shield nodes. 

Also, be aware that the "backend" will change when shielding is enabled. Traditionally (i.e. without shielding) you defined your backend with a specific value (e.g. an S3 bucket or a domain such as `https://app.domain.com`) and it would stay set to that value unless you yourself implemented custom vcl logic to change its value. 

But with shielding enabled, the 'edge' cache node will dynamically change the backend to be a shield node value (as it's effectively _always_ going to pass through that node if there is no cached content found). Once on the shield node, _its_ "backend" value is set to whatever your actual origin is (e.g. an S3 bucket).

You can utilize the existence of the header `Fastly-FF` to indicate if your code is currently running on a shield node. This header doesn't exist on an edge node, and its presence indicates that the request has already come from a fastly cache server. 

Alternatively you could check `req.backend.is_shield` which (if available/set) would indicate your code was executing on a non-shield cache node (i.e. the edge cache node). You might prefer to use the latter because if your client uses Fastly and they put your service behind their Fastly account, then `Fastly-FF` would be set as the request travels through the system and so your check for `Fastly-FF` on your shield node might execute when it shouldn't.

It's probably best to only modify your backends dynamically whilst your VCL is executing on the shield (e.g. `if (!req.backend.is_shield)`, maybe abstract in a variable `declare local var.shield_node BOOL;`) and to also only `restart` a request in vcl_deliver when executing on the shield node. 

You might also need to modify vcl_hash so that the generated hash is consistent with the 'edge' cache node if your shield modifies the request! Remember that modifying either the host or the path will cause a different cache key to be generated and so modifying that in either the edge _or_ the shield means modifying the relevant vcl_hash subroutine so the hashes are _consistent_ between the edge and shield.

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

Lastly, when enabling shielding, make sure to deploy your VCL code changes first _before_ enabling shielding. This way you avoid a race condition whereby a shield has old VCL (i.e. no conditional checks for either `Fastly-FF` or `req.backend.is_shield`) and thus tries to do something that should only happen on the edge cache node.

When using `Fastly-Debug:1` to inspect debug response headers, and we look at `fastly-state`, `fastly-debug-path` and `fastly-debug-ttl` we might see...

```
< fastly-state: HIT-STALE
< fastly-debug-path: (D cache-lhr6346-LHR 1563794040) (F cache-lhr6324-LHR 1563794019)
< fastly-debug-ttl: (H cache-lhr6346-LHR -10.999 31536000.000 20)
```

The `fastly-debug-path` suggests we delivered from the edge node `lhr6346` while we fetched from the cluster/shield node `lhr6324`, and yet the `fastly-debug-ttl` header suggests we got a HIT (`H`) from the edge node `lhr6346`. This is just a side-effect of the stale/cached content (coming back from the fetching cluster/shield node) being stored in-memory on the edge node and so it's indicated as a HIT from the edge when really it came from the cluster/shield node (the `fastly-debug-ttl` header is set on the edge node, which re-enforces this understanding).

What makes it confusing is that you don't necessarily know if the request went to the cluster cache node (i.e. the fetching cache node) or whether the stale content actually came from the edge node's in-memory cache. The only way to be sure is to check the `fastly-state` response header and see if you got back `HIT-STALE` or `HIT-STALE-CLUSTER`.

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

<div id="9"></div>
## Conclusion

So there we have it, a quick run down of how some important aspects of Varnish and VCL work (and specifically for Fastly's implementation).

One thing I want to mention is that I am personally a HUGE fan of Fastly and the tools they provide. They are an amazing company and their software has helped BuzzFeed (and many other large organisations) to scale massively with ease.

I would also highly recommend watching this talk by Rogier Mulhuijzen (Senior Varnish Engineer - who currently works for Fastly) on "Advanced VCL": [vimeo.com/226067901](https://vimeo.com/226067901). It goes into great detail about some complex aspects of VCL and Varnish and really does a great job of elucidating them.

Also recommended is [this Fastly talk](https://vimeo.com/178057523) which details how 'clustering' works (see also [this fastly community support post](https://support.fastly.com/hc/en-us/community/posts/360040445272--Fastly-Force-Shield-AND-Fastly-No-Shield-Usage) that details how to utilize the request headers `Fastly-Force-Shield` and `Fastly-No-Shield`).

Lastly, there was [a recent article from an engineer working at the Financial Times](https://medium.com/@samparkinson_/making-a-request-to-the-financial-times-b2119a2f422d), detailing the complete request flow from DNS to Delivery. It's very interesting and covers a lot of information about Fastly. Highly recommended reading.
