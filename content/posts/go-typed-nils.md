---
title: "Go's typed nils are gonna get yer"
date: 2024-06-12T06:48:56+01:00
draft: false
categories:
  - "code"
  - "guides"
tags:
  - "interfaces"
  - "nil"
  - "go"
---

The following code doesn't do what you might expect:

```go
package main

import "fmt"

func main() {
	var i *impl

	fmt.Println("i == nil:", i == nil)
	what(i)
}

type impl struct{}

func (i *impl) do() {}

func what(i interface{ do() }) {
	fmt.Println("i == nil:", i == nil)
}
```

If you expected the `what` function to print `i == nil: true`, then keep
reading...

## Typed Nils

The behavior observed is due to the way interfaces and nil values interact in
Go. To understand why the `what` function is able to see the `i` argument as
non-nil, we need to dig into the details of how Go handles interface values.

1. Interface Values: In Go, an interface value is a tuple of a type and a value.
   An interface value is `nil` only if both the type and the value are `nil`.
2. Concrete vs Interface nil: When you assign a concrete type value (which
   happens to be `nil`) to an interface, the interface itself is not `nil`. This
is because the interface value now contains a type (the concrete type) and a
value (`nil`).

## Explanation

1. Declaring `i` as `*impl` and initializing it to `nil`:
    ```go
    var i *impl
    ```
    Here, `i` is a pointer to `impl` and is initialized to `nil`.
    <p></p>
2. Printing `i == nil` in `main`:
    ```go
    fmt.Println("i == nil:", i == nil)
    ```
    This prints `true` because `i` is a `nil` pointer to `impl`.
    <p></p>
3. Calling `what(i)`:
    ```go
    what(i)
    ```
    The function `what` takes an argument of type `interface{ do() }`.
    <p></p>
4. Inside `what` function:
    ```go
    func what(i interface{ do() }) {
        fmt.Println("i == nil:", i == nil)
    }
    ```
    When `i` (which is `nil`) is passed to `what`, it is assigned to the parameter `i` of type `interface{ do() }`.
    <p></p>
5. Interface value construction:\
    The value of `i` inside `what` is now an interface that holds:
    - Type: `*impl` (the concrete type of the value passed in)
    - Value: `nil` (the value of the concrete type)
    <p></p>
6. Checking `i == nil`:
    ```go
    fmt.Println("i == nil:", i == nil)
    ```
    This prints `false` because the interface `i` is not `nil`:
    - The type part of the interface is `*impl`.
    - The value part of the interface is `nil`.
    <p></p>

## Summary

The `what` function sees the `i` argument as non-nil because, in Go, an
interface holding a `nil` pointer is not itself `nil`. The interface contains
type information (`*impl`) and a `nil` value. Therefore, when the code checks if
`i` is `nil`, it returns `false` since the type information (`*impl`) is
present.

## Reference material

- [Go FAQ][1]
- [Dave Cheney][2]

[1]: https://go.dev/doc/faq#nil_error
[2]: https://dave.cheney.net/2017/08/09/typed-nils-in-go-2
