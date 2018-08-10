---
title: "Thinking about Interfaces in Go"
date: 2018-07-21T13:00:29+01:00
categories:
  - "code"
  - "guides"
tags:
  - "dependencies"
  - "go"
  - "interfaces"
draft: false
---

- [Interfaces in Go](#interfaces-in-go)
- [Keep Interfaces Small](#keep-interfaces-small)
- [Standard Library Interfaces](#standard-library-interfaces)
- [Tight Coupling](#tight-coupling)
- [Dependency Injection](#dependency-injection)
- [Refactoring Considerations](#refactoring-considerations)
- [Testing](#testing)
- [More flexible solutions?](#more-flexible-solutions)
- [Conclusion](#conclusion)

This post is going to explain the importance of interfaces, and the concept of programming to abstractions (using the [Go](https://golang.org/) programming language), by way of a simple example. 

While treading what might seem like familiar ground to some readers, this is a fundamental skill to understand because it enables you to design more flexible and maintable services. 

## Interfaces in Go

An 'interface' in Go looks something like the following:

```
type Foo interface {
    Bar(s string) (string, error)
}
```

If an object in your code implements a `Bar` function, with the exact same signature (e.g. accepts a string and returns either a string or an error), then that object is said to _implement_ the `Foo` interface.

An example of this would be:

```
type thing struct{}

func (l *thing) Bar(s string) (string, error) {
  ...
}
```

Now you can define a function that will accept that object, as long as it fulfils the `Foo` interface, like so:

```
func doStuffWith(thing Foo)
```

This is different to other languages, where you have to _explicitly_ assign an interface type to an object, like with Java:

```
class testClass implements Foo
```

Because of this flexibility in how interfaces are 'applied', it also means that an object could end up implementing _multiple_ interfaces.

Imagine we have two interfaces:

```
type Foo interface {
  Bar(s string) (string, error)
}

type Beeper interface {
  Beep(s string) (string, error)
}
```

We can define an object that fulfils _both_ interfaces simply by implementing the functions they define:

```
type thing struct{}

func (l *thing) Bar(s string) (string, error) {
  ...
}

func (l *thing) Beep(s string) (string, error) {
  ...
}
```

## Keep Interfaces Small

You'll find in the [Go Proverbs](https://go-proverbs.github.io/), the following useful tip:

> The bigger the interface, the weaker the abstraction.

The reason for this is due to how interfaces are designed in Go and the fact that an object can potentially support multiple interfaces. 

By making an interface too big, we reduce an object's ability to support it. Consider the following example:

```
type FooBeeper interface {
  Bar(s string) (string, error)
  Beep(s string) (string, error)
}

type thing struct{}

func (l *thing) Bar(s string) (string, error) {
  ...
}

func (l *thing) Beep(s string) (string, error) {
  ...
}

type differentThing struct{}

func (l *differentThing) Bar(s string) (string, error) {
  ...
}

type anotherThing struct{}

func (l *anotherThing) Beep(s string) (string, error) {
  ...
}
```

In the above example we've defined a `FooBeeper` interface that requires two methods: `Bar` and `Beep`. Now if we look at the various objects we've defined `thing`, `differentThing` and `anotherThing` we'll find:

- `thing`: fulfils the `FooBeeper` interface
- `differentThing`: does _not_ fulfil the `FooBeeper` interface
- `anotherThing`: does _not_ fulfil the `FooBeeper` interface

Alternatively, if we were to break the `FooBeeper` interface up into separate smaller interfaces (like we demonstrated earlier), then in our above example, the `differentThing` and `anotherThing` would become more re-usable.

That's ultimately what this Go proverb is suggesting: smaller interfaces allow for greater code reuse.

## Standard Library Interfaces

Imagine we have a function `process`, whose responsibility is to make a HTTP request and do something with the response data:

```
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func process(n int) (string, error) {
	url := fmt.Sprintf("http://httpbin.org/links/%d/0", n)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("url get error: %s\n", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("body read error: %s\n", err)
		return "", err
	}

	return string(body), nil
}

func main() {
	data, err := process(5)
	if err != nil {
		fmt.Printf("\ndata processing error: %s\n", err)
		return
	}
	fmt.Printf("Success: %v", data)
}
```

We can see our `process` function accepts an integer, which is interpolated into the URL that is requested. We then use the `http.Get` function from the [net/http](https://golang.org/pkg/net/http/) package to request the URL.

The function then stringify's the response body and returns it. This is sufficient for a basic example, but in the real world this function would likely do lots more processing to the response data.

It may not be immediately obvious but there are already many instances where interfaces are being utilised. Let's break down the code and see what interfaces there are.

The `http.Get` function returns a pointer to a `http.Response` struct, and from within that struct we extract the `Body` field and pass it to `ioutil.ReadAll`. 

The `Body` field's 'type' is set to the [`io.ReadCloser`](https://golang.org/src/io/io.go?s=4977:5022#L116) interface. If we look at that interface we'll see it's made up of _nested_ interface types:

```
type ReadCloser interface {
    Reader
    Closer
}
```

If we now look at the [`io.Reader`](https://golang.org/src/io/io.go?s=3303:3363#L67) and [`io.Closer`](https://golang.org/src/io/io.go?s=4043:4083#L88) interfaces, we'll find:

```
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}
```

This means that for the response body object to be valid, it must support the `Read` and `Close` functions defined by these interfaces (the returned object will likely include other functions, but it needs `Read` and `Close` at a minimum). 

The next thing that happens in the code is that we pass `http.Response.Body` to an input/output function called `ioutil.ReadAll`. 

If we look at the signature of `ioutil.ReadAll` we'll see that it accepts a type of `io.Reader`, which we've seen already, and so this is another indication of why smaller interfaces enable re-usability.

What the `io.Reader` interface means for our code is that the input we provide to `ioutil.ReadAll` must support a `Read` function, and (because `http.Response.Body` implements the `io.ReadCloser` interface) we know it does implement that required function.

So already we've seen quite a few built-in interfaces being utilised to support the standard library code we're using. More importantly, you'll find the use of these interfaces (`io.ReadCloser`, `io.Reader`, `io.Closer` and others) are used _everywhere_ in the Go codebase (highlighting again how small interfaces enable greater code re-usability).

## Tight Coupling

Now there's an issue with the above code, specifically the `process` function, and that is we've tightly coupled the `net/http` package to the function.

What this means is that the `process` function has to intrinsically _know_ about HTTP and dealing with the various methods available to that package.

Also, if we want to test this function we're going to have a harder time because the `http.Get` call would need to be mocked somehow. We don't want our test suite to have to rely on a stable network connection or the fact that the endpoint being requested might be down for maintenance.

The solution to this problem is to invert the responsibility of the `process` function, also known as 'dependency injection'. This is the basis of one of the [S.O.L.I.D](https://en.wikipedia.org/wiki/SOLID) principles: 'inversion of control'.

## Dependency Injection

If we call a function, then it is our responsibility to provide it with all the things it needs in order to do its job.

In the case of our `process` function, it needs to be able to acquire data from somewhere (that could be a file, it could be a remote procedure call, it shouldn't matter). The most important aspect to consider is _how_ it acquires that data. 

The _how_ is not the responsibility of the `process` function, especially if we decide later on that we want to change the implementation from HTTP to GRPC or some other data source.

Meaning, we need to provide that functionality to the `process` function. Let's see what this might look like in practice (this is just a first iteration and so is actually not a great solution, but is _a_ solution):

```
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type dataSource interface {
	Get(url string) (*http.Response, error)
}

type httpbin struct{}

func (l *httpbin) Get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("url get error: %s\n", err)
		return &http.Response{}, err
	}

	return resp, nil
}

func process(n int, ds dataSource) (string, error) {
	url := fmt.Sprintf("http:/httpbin.org/links/%d/0", n)

	resp, err := ds.Get(url)
	if err != nil {
		fmt.Printf("data source get error: %s\n", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("body read error: %s\n", err)
		return "", err
	}

	return string(body), nil
}

func main() {
	data, err := process(5, &httpbin{})
	if err != nil {
		fmt.Printf("\ndata processing error: %s\n", err)
		return
	}
	fmt.Printf("\nSuccess: %v\n", data)
}
```

## Refactoring Considerations

Let's start by looking at the interface we've defined:

```
type dataSource interface {
	Get(url string) (*http.Response, error)
}
```

We've not been overly explicit when naming this interface `dataSource`. Its name is quite generic on purpose so as not to imply an underlying implementation bias.

Unfortunately the defined `Get` method is still too tightly coupled to a specific implementation (i.e. it specifies `http.Response` as a return type).

Meaning, that although the refactored code is _better_, it is far from perfect. 

Next we define our own object for handling the implementation of the `Get` method, which internally is going to use `http.Get` to acquire the data:

```
type httpbin struct{}

func (l *httpbin) Get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("url get error: %s\n", err)
		return &http.Response{}, err
	}

	return resp, nil
}
```

By using this interface as the accepted type in the `process` function signature, we're going to be able to decouple the function from having to acquire the data, and thus allow testing to become much easier (as we'll see shortly), but the `process` function is still fundamentally coupled to HTTP as the underlying transport mechanism.

The reason this is a problem is because the `process` function still _knows_ that the returned object is a `http.Response` because it has to reference the `Body` field of the response, which isn't defined on the object we've injected (meaning the function intrinsically _knows_ of its existence).

How far you take your interface design is up to you. You don't necessarily have to solve all possible concerns at once (unless there really is a need to do so). 

Meaning, this refactor _could_ be considered 'good enough' for your use cases. Alternatively your values and standards may differ, and so you need to consider your options for how you might what to design this solution in such a way that it would allow the code to not be so reliant on HTTP as the transport mechanism.

> Note: we'll revisit this code later and consider another refactor that will help clean up this first pass of code decoupling.

But first, let's look at how we might want to test this initial code refactor (as testing this code allows us to learn some interesting things when it comes to needing to mock interfaces).

## Testing

Below is a simple test suite that demonstrates how we're now able to construct our own object, with a stubbed response, and pass that to the `process` function:

```
package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

type fakeHTTPBin struct{}

func (l *fakeHTTPBin) Get(url string) (*http.Response, error) {
	body := "Hello World"

	resp := &http.Response{
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		StatusCode:    http.StatusOK,
		Request:       &http.Request{},
	}

	return resp, nil
}

func TestBasics(t *testing.T) {
	expect := "Hello World"
	actual, _ := process(5, &fakeHTTPBin{})

	if actual != expect {
		t.Errorf("expected %s, actual %s", expect, actual)
	}
}
```

Much like we do in the real implementation, we define a struct (in this case we've named it more explicitly) `fakeHTTPBin`.

The difference now, and what allows us to test our code is that we're manually creating a `http.Response` object with dummy data.

One part of this code that requires some extra explanation would be the value assigned to the response `Body` field:

```
ioutil.NopCloser(bytes.NewBufferString(body))
```

If we remember from earlier:

> The `Body` field's 'type' is set to the `io.ReadCloser` interface.

This means when mocking the `Body` value we need to return something that has both a `Read` and `Close` method. So we've used `ioutil.NopCloser` which, if we look at its signature, we see returns an `io.ReadCloser` interface:

```
func NopCloser(r io.Reader) io.ReadCloser
```

The `io.ReadCloser` interface is exactly what we need (as that interface indicates the returned concrete type will indeed implement the required `Read` and `Close` methods). 

But to use it we need to provide the `NopCloser` function something that supports the `io.Reader` interface.

If we were to provide a simple string like `"Hello World"`, then this wouldn't implement the required interface. So we wrap the string in a call to `bytes.NewBufferString`.

The reason we do this is because the returned type is something that supports the `io.Reader` interface we need.

But that might not be immediately obvious when looking at the signature for `bytes.NewBufferString`:

```
func NewBufferString(s string) *Buffer
```

So yes it accepts a string, but we want an `io.Reader` as the return type, where as this function returns a pointer to a [`Buffer`](https://golang.org/src/bytes/buffer.go?s=402:817#L7) type? 

If we look at the implementation of `Buffer` though, we will see that it does actually [implement](https://golang.org/src/bytes/buffer.go?s=9564:9614#L287) the required `Read` function necessary to support the `io.Reader` interface.

Great! Our test can now call the `process` function and process the mocked dependency and the code/test works as intended.

## More flexible solutions?

OK, so we've already explained why this implementation might not be the best we could do. Let's now consider an alternative implementation:

```
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type dataSource interface {
	Get(url string) ([]byte, error)
}

type httpbin struct{}

func (l *httpbin) Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("url get error: %s\n", err)
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("body read error: %s\n", err)
		return []byte{}, err
	}

	return body, nil
}

func process(n int, ds dataSource) (string, error) {
	url := fmt.Sprintf("http://httpbin.org/links/%d/0", n)

	resp, err := ds.Get(url)
	if err != nil {
		fmt.Printf("data source get error: %s\n", err)
		return "", err
	}

	return string(resp), nil
}

func main() {
	data, err := process(5, &httpbin{})
	if err != nil {
		fmt.Printf("\ndata processing error: %s\n", err)
		return
	}
	fmt.Printf("\nSuccess: %v\n", data)
}
```

All we've really done here is move more of the logic related to HTTP up into the `httpbin.Get` implementation of the `dataSource` interface. We've also changed the response type from `(*http.Response, error)` to `([]byte, error)` to account for these movements.

Now the `process` function has even _less_ responsibility as far as acquiring data is concerned. This also means our test suite benefits by having a much simpler implementation:

```
package main

import "testing"

type fakeHTTPBin struct{}

func (l *fakeHTTPBin) Get(url string) ([]byte, error) {
	return []byte("Hello World"), nil
}

func TestBasics(t *testing.T) {
	expect := "Hello World"
	actual, _ := process(5, &fakeHTTPBin{})

	if actual != expect {
		t.Errorf("expected %s, actual %s", expect, actual)
	}
}
```

Now our `fakeHTTPBin.Get` only has to return a byte array.

## Conclusion

Is there more we can do to improve this code's design? Sure. But we'll leave a new refactor iteration to another post. 

Hopefully this has given you a feeling for how interfaces are used in the Go standard library and how you might utilise custom interfaces yourself.