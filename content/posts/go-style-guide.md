---
title: "Go Style Guide"
date: 2022-01-11T14:28:30Z
categories:
  - "code"
  - "development"
  - "guide"
tags:
  - "bash"
  - "go"
  - "style"
draft: false
---

This is my own personal style guide for Go.

# Table of Contents

- [Reference Materials](#reference-materials)
- [Naming](#naming)
- [Whitespace](#whitespace)
- [Quick note on Code Design](#quick-note-on-code-design)
- [Quick guide to Error wrapping](#quick-guide-to-error-wrapping)
- [Quick guide to `panic`](#quick-guide-to-panic)
- [Quick guide to slice 'gotchas'](#quick-guide-to-slice-gotchas)
- [Quick guide to pass-by-value vs pass-by-pointer](#quick-guide-to-pass-by-value-vs-pass-by-pointer)
- [Quick guide to functions with large signature](#quick-guide-to-functions-with-large-signature)

## Reference Materials

The following reference materials are my 'go to' whenever I'm unsure of something (they're mostly official resources).

- [Effective Go](https://golang.org/doc/effective_go)
- [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [What's in a name?](https://talks.golang.org/2014/names.slide)
- [Commit messages](https://github.com/golang/go/wiki/CommitMessage)
- [Comments](https://github.com/golang/go/wiki/Comments)
- [Slice Gotchas](https://blogtitle.github.io/go-slices-gotchas/)
- [Thinking about interfaces](https://www.integralist.co.uk/posts/go-interfaces/)

> **NOTE**: Refer to the [specification](https://golang.org/ref/spec) if ever confused about what the expected behaviour is. 

## Naming

The following is a summary of how to name things in Go, gleaned from either my own experiences over the years or from some of the above reference materials.

- Choose package names that lend meaning to the names they export.
- Where types are descriptive, name should be short (1 or 2 char name).
- If longer name required, consider refactoring into smaller functions.
- Commonly used names:
  - Prefer `i` to `index`.
  - Prefer `r` to `reader`.
  - Prefer `buf` to `buffer`.
  - Prefer `cfg` to `config`.
  - Prefer `dst, src` to `destination, source`.
  - Prefer `in, out` when referring to stdin/stdout.
  - Prefer `rx, tx` when dealing with channels.
    - i.e. receiver, transmitter.
  - Prefer `data` when referring to file content.
    - Regardless of it being a `string` or `[]byte`.
  - Use `ok` instead of longer alternatives.
- Errors:
  - Types: `<T>Error` (e.g. `type ExitError struct {...}`).
  - Values: `Err<T>` (e.g. `var ErrFoo = errors.New("bar: baz")`).
- Interfaces:
  - When an interface includes multiple methods, choose a name that accurately describes its purpose.
  - Interfaces that specify just one method are usually just that function name with 'er' appended to it.
    - Sometimes the result isn't correct English, that's OK.
    - Sometimes we use English to make it nicer.
- Return values on exported functions should only be named for documentation purposes.
  - Side effect is that the variable is initialised at start of function with zero value.
  - This can, in some cases, lead to a nice code design.
- `Set<T>` vs `Register<T>`
  - **Set**: use when flipping a bit (e.g. setting an int, string etc).
  - **Register**: use when operation is going _into_ something (e.g. registering a CLI flag inside a command).

> **NOTE**: Refer also to https://github.com/kettanaito/naming-cheatsheet

## Whitespace

The go standard library has no strong conventions or idioms for how to handle whitespace. So try and be concise without leaving the user with a wall of text to digest. Additionally, you can use block syntax `{...}` to help group related logic:

```go
// Simple code is fine to condense the whitespace.
if ... {
  foo
  for x := range y {
    ...
  }
  bar
}

// Complex code could benefit from some whitespace (also separate block syntax for grouping related logic).
if {
  ...

  {
    ...grouping of related logic...
  }

  ...
}
```

## Quick note on Code Design

Not always obvious but be wary of returning concrete types when building a package to be used as a library.

Here is an example of why this might be problematic: we had a library that defined a constructor that returned a struct of type `*T`. This struct had methods attached and inside of those methods were API calls. We built a separate CLI that consumed the package library and realised our CLI's test suite wasn't able to mock the type appropriately as some of the fields on the struct were private and would determine if an attached method would make an API call.

The solution was for us to return an interface. This made it simple to mock the behaviours we wanted (e.g. pretend there was an API error, how does our CLI handle it).

## Quick guide to Error wrapping

When you wrap errors your message **should include**:

- A pointer to where within your method the failure occurred.
- Values that will be useful during debugging (e.g ids).
- (sometimes) details about why the error occurred.
- Other relevant info the caller doesnt know.

And your message **should NOT include**:

- The name of your function
- Any of the arguments to your function
- Any other information that is already known to the caller

Here is a BAD example where the caller of a function that fails is seeing duplicate information:

```go
// Source
func MightFail(id string) error {
    err := sqlStatement()
    if err != nil {
        return fmt.Errorf("mightFail failed with id %v because of sql: %w", id, err
    }
    ...
    return nil
}

// Caller
func business(ids []string) error {
    for _, id := range ids {
        err := MightFail(id)
        if err != nil {
            return fmt.Errorf("business failed MightFail on id %v: %w", id, err)
        }
    }
}
```

The resolution to the above bad code is: only include information the caller doesn’t have. The caller is free to annotate your errors with information such as the name of your function, arguments they passed in, etc. There is no need for you to provide that information to them, as its obvious up front. If this same logic is applied consistently you'll end up with error messages that are high-signal and to-the-point.

## Quick guide to `panic`

- The use of `panic` is reserved for when an error is _unrecoverable_.
- What constitutes an "unrecoverable" error is contentious. Here are some definitions:
    - To indicate that something impossible has happened, such as exiting an infinite loop.
    - During initialization, if the library truly cannot set itself up, it might be reasonable to `panic`.
    - When something internally has fundamentally failed.
    - When a programmer gives something to a function which the function explicitly states is invalid.
- [`bytes.Truncate`](https://github.com/golang/go/blob/8ac6544/src/bytes/buffer.go#L88-L90) is an example of the last sub-point.
  - The above example could be considered _aggressive_. 
  - Instead the standard library could have returned an error so the caller could decide the appropriate action to take.
- The use (and conditions) of `panic` should be documented (example: [`bytes.Truncate`](https://github.com/golang/go/blob/8ac6544/src/bytes/buffer.go#L81))
- The use of `recover` is for when you disagree with the library authors.
- Wherever possible avoid `panic` and return an error for the caller to handle.

## Quick guide to slice 'gotchas'

When taking a slice of a slice you might stumble into behaviour which appears confusing at first. The `cap`, `len` and `data` fields might change, but the underlying array is not re-allocated, nor copied over and so modifications to the slice will modify the original backing array.

> Refer to the golang language specification section on ["full slice expressions"](https://golang.org/ref/spec#Slice_expressions) syntax (`[low : high : max]`) for controlling the capacity of a slice.

### Ghost update 1

The underlying array is modified after updating an element on the slice:

```go
a := []int{1, 2}
b := a[:1]     /* [1]     */
b[0] = 42      /* [42]    */
fmt.Println(a) /* [42, 2] */
```

### Ghost update 2

When data gets appended to `b` (a slice of the `a` slice), the underlying array has enough capacity to hold two more elements, so `append` will not re-allocate. This means that appending to `b` might not only change `a` but also `c` (a slice of the `a` slice).

```go
a := []int{1, 2, 3, 4}
b := a[:2] /* [1, 2] */
c := a[2:] /* [3, 4] */
b = append(b, 5)
fmt.Println(a) /* [1 2 5 4] */
fmt.Println(b) /* [1 2 5]   */
fmt.Println(c) /* [5 4]     */
```

The fix is `b := a[:2:2]` which sets the capacity of the `b` slice such that `append` will cause a new array to be allocated. This means `a` will not be modified, nor will the `c` slice of `a`.

> **NOTE**:  there are more examples/explanations in https://blogtitle.github.io/go-slices-gotchas/

## Quick guide to pass-by-value vs pass-by-pointer

> Reference articles: [goinbigdata.com](https://goinbigdata.com/golang-pass-by-pointer-vs-pass-by-value/) and [dave.cheney.net](https://dave.cheney.net/2017/04/29/there-is-no-pass-by-reference-in-go).

In essence when people say 'pass by reference', the point they're trying to get across is: "this _isn't_ a copy of the value being passed". Where as 'pass by reference' is a very _specific_ type of behaviour.

All primitive/basic types (int and its variants, float and its variants, boolean, string, array, and struct) in Go are passed by value.

Maps and slices are passed by pointer (sometimes incorrectly called pass-by-reference). This is where a new copy of the 'pointer' to the same memory address is created.

Go does not have pass-by-reference semantics because Go does not have 'reference variables' (which is something you'd find in C++). 

In C++ you can create `a = 10` and then _alias_ `b` to `a` (`&b = a`) such that updating `b` would _affect_ `a`. Go doesn't have this behaviour. Every variable is stored in its own memory space. Meaning if we had `b := &a` and updated `b` then we wouldn't cause any change to `a`.

When we define a function that accepts a pointer (e.g. `changeName(p *Person)`) and we pass a pointer to it (e.g. `changeName(&person)`) the variable person is modified inside the `changeName` function. This happens because `&person` and `p` are two _different_ pointers to the _same_ struct which is stored at the same memory address. This is quite different to C++'s reference variables.

## Quick guide to functions with large signature

Your functions should have concise/relevant arguments passed in.

Don't, for example, pass in an argument whose type is a large object and which the function then has to know how that object is structured as that's violating the Law of Demeter. Instead choose a field from the object to pass in as it'll likely have a simpler type (like a `string` or `int`).

Three approaches to dealing with functions that potentially could have a large number of arguments...

1. Make multiple functions to help reduce the number of arguments.
2. Pass in a `<T>Options` struct.
3. Variadic arguments that accept a func type.

I would say go with option 1 whenever possible, and almost never choose option 2 over option 3 as the latter is much more flexible.

The problem with option 2 is that it can become quite cumbersome to construct an object with lots of fields, and more importantly it can be hard to know which fields are _required_ and which are _optional_. Yes it's nice that you can easily omit optional fields easily, but then option 3 also provides that benefit while also solving the problem of knowing what arguments are required vs optional.

Using option 3 can be helpful when you want to make the function signature clear, by accepting a couple of concrete arguments that are _required_ for the function to work, while shifting _optional_ arguments into separate functions, as demonstrated below...

```go
type Client struct {
  host, proxy string
  port int
}

type Option func(*Client) // call this function to apply the option

func WithPort(port int) Option {
  return func(c *Client) { c.port = port }
}

func WithProxy(proxy string) Option {
  return func(c *Client) { c.proxy = proxy }
}

func NewClient(host string, options ...Option) *Client {
  c := &Client{host: host, port: 80} // default values
  for _, option := range options {
    option(c) // apply the options by calling each one of them
  }
  return c
}
```
