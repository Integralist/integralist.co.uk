---
title: "Rust Smart Pointers"
date: 2022-06-20T09:56:35+01:00
categories:
  - "code"
  - "guide"
tags:
  - "rust"
  - "rustlang"
  - "memory"
  - "pointers"
  - "safety"
draft: false
---

## Summary

- `Box<T>`: A pointer type for heap allocation.
  - Enables recursion where a type with a fixed size is required.
- `Rc<T>`: A single-threaded reference-counting pointer.
  - Enables multiple owners (READ-ONLY).
  - Supports mutability by wrapping value in another type.
    - `Cell<T>` or `RefCell<T>`.
- `Arc<T>`: A thread-safe reference-counting pointer.
  - Enables multiple owners (READ-ONLY).
  - Thread safety means there is an additional performance overhead.
  - Supports mutability by wrapping value in another type.
    - `Mutex<T>`, `RwLock<T>` or one of the `Atomic*` types.
- `Cell<T>`: A single-threaded mutable memory location for 'values'.
  - Behaves like an exclusive borrow, aka a `&mut T`.
  - The contained value is _moved_ in and out of the cell.
- `RefCell<T>`: A single-threaded mutable memory location for 'references'. 
  - Removes the compile-time borrow-checks that `Cell<T>` has.
  - Rust's borrow rules are dynamically checked at _runtime_.
  - More flexible than `Cell<T>` but with possible 'runtime panic' cost.

> **NOTE**: Not discussed here are `Mutex<T>` and `RwLock<T>`, which provide mutual-exclusion.

## IMPORTANT

I'm not the author of the _following_ content. 

In order to better understand the topic of smart pointers in Rust, I read through the official Rust documentation and simply cherry-picked useful subsets of that information and grouped it in a way that helps make the topic easier to understand. 

This means I take no credit for the following content.  
It was written by many people much smarter than me.

## Context

A _pointer_ is a general concept for a variable that contains an address in memory. This address refers to, or “points at,” some other data. The most common kind of pointer in Rust is a reference. References are indicated by the & symbol and borrow the value they point to. They don’t have any special capabilities other than referring to data, and have no overhead.

_Smart pointers_, on the other hand, are data structures that act like a pointer but also have additional metadata and capabilities.

> **NOTE**: Both `String` and `Vec<T>` types count as smart pointers because they own some memory and allow you to manipulate it. They also have metadata and extra capabilities or guarantees. 

Smart pointers are usually implemented using structs. Unlike an ordinary struct, smart pointers implement the `Deref` and `Drop` traits. The `Deref` trait allows an instance of the smart pointer struct to behave like a reference so you can write your code to work with either references or smart pointers. The `Drop` trait allows you to customize the code that's run when an instance of the smart pointer goes out of scope.

## `Box<T>`

Boxes allow you to store data on the heap rather than the stack. What remains on the stack is the pointer to the heap data.

Usage:

- When you have a type whose size can’t be known at compile time and you want to use a value of that type in a context that requires an exact size (e.g. recursive data types).
- When you have a large amount of data and you want to transfer ownership but ensure the data won’t be copied when you do so (e.g. only the small amount of pointer data is copied around on the stack, while the data it references stays in one place on the heap).
- When you want to own a value and you care only that it’s a type that implements a particular trait rather than being of a specific type (i.e. _trait objects_ used for dynamic dispatch).

## `Rc<T>`

This type is an abbreviation for _reference counting_ and it enables multiple ownership. It lets us have multiple “owning” pointers to the same data, and the data will be freed (destructors will be run) when all pointers are out of scope.

We use the `Rc<T>` type when we want to allocate some data on the heap for multiple parts of our program to read and we can’t determine at compile time which part will finish using the data last. 

> **NOTE**: `Rc<T>` is only for use in single-threaded scenarios. When shared ownership between threads is needed, `Arc<T>` (Atomic Reference Counted) can be used.

## `Arc<T>`

The same as `Rc<T>` but thread-safe (it has the same API as `Rc<T>` to make them interchangeable). But thread-safety comes with a performance cost so if you don't need thread-safety, then definitely opt for `Rc<T>` instead.

## `Cell<T>`/`RefCell<T>`

Rust memory safety is based on this rule: Given an object `T`, it is only possible to have one of the following:

- Having several immutable references (`&T`) to the object.
- Having one mutable reference (`&mut T`) to the object.

This rule can be bent using `Cell<T>` and is referred to as _interior mutability_.

`Cell<T>` implements interior mutability by moving values in and out of the `Cell<T>`. To use references instead of values, use the `RefCell<T>` type, and acquire a write lock before mutating.

Borrows for `RefCell<T>`s are tracked _at runtime_, unlike Rust’s native reference types which are entirely tracked statically, at compile time. 

Because `RefCell<T>` borrows are _dynamic_ it is possible to attempt to borrow a value that is already mutably borrowed; when this happens it results in thread panic.

> **NOTE**: Neither `Cell<T>` nor `RefCell<T>` are thread-safe.  
> The `Sync` marker trait is not implemented.

Many shared smart pointer types, including `Rc<T>` and `Arc<T>`, provide containers that can be cloned and shared between multiple parties. The contained values can only be borrowed with `&`, not `&mut`. Without cells it would be impossible to mutate data inside of these smart pointers at all.

It’s very common then to put a `RefCell<T>` inside shared pointer types to reintroduce mutability. But as `RefCell<T>`s are for single-threaded scenarios, consider using `RwLock<T>` or `Mutex<T>` instead of `RefCell<T>` if you need shared mutability in a multi-threaded situation.
