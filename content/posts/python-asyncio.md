---
title: "Guide to Python 3.8 Asyncio"
date: 2019-11-30T13:35:48Z
categories:
  - "code"
  - "development"
  - "guide"
tags:
  - "concurrency"
  - "asyncio"
  - "python"
draft: false
---

This is a _quick_ guide to Python's `asyncio` module and is based on Python version 3.8.

- [Introduction](#introduction)
- [Event Loop](#event-loop)
- [Awaitables](#awaitables)
  - [Coroutines](#coroutines)
  - [Tasks](#tasks)
  - [Futures](#futures)
- [Running an asyncio program](#running-an-asyncio-program)
  - [Running Async Code in the REPL](#running-async-code-in-the-repl)
- [Concurrent Functions](#concurrent-functions)
- [Deprecated Functions](#deprecated-functions)
- [Examples](#examples)
  - [`gather`](#gather)
  - [`wait`](#wait)
  - [`wait_for`](#wait-for)
  - [`as_completed`](#as-completed)
  - [`create_task`](#create-task)
  - [Callbacks](#callbacks)
- [Pools](#pools)

## Introduction

> asyncio is a library to write concurrent code using the `async`/`await` syntax. -- [docs.python.org/3.8/library/asyncio.html](https://docs.python.org/3.8/library/asyncio.html)

The asyncio module provides both high-level and low-level APIs. Library and Framework developers will be expected to use the low-level APIs, while all other users are encouraged to use the high-level APIs.

## Event Loop

The core element of all asyncio applications is the 'event loop'. The event loop is what schedules and runs asynchronous tasks (it also handles network IO operations and the running of subprocesses).

<a href="../../images/event-loop.png">
    <img src="../../images/event-loop.png">
</a>

<div class="credit">
  <a href="https://eng.paxos.com/python-3s-killer-feature-asyncio">Image Credit</a>
</div>

> Note: for more API information on the event loop, please refer to [the documentation](https://docs.python.org/3.8/library/asyncio-eventloop.html).

## Awaitables

Something is _awaitable_ if it can be used in an `await` expression.

There are three main types of awaitables:

1. Coroutines
2. Tasks
3. Futures

> Note: Futures is a _low-level_ type and so you shouldn't need to worry about it too much if you're not a library/framework developer (as you should be using the higher-level abstraction APIs instead).

### Coroutines

There are two closely related terms used here:

- a _coroutine function_: an `async def` function.
- a _coroutine object_: an object returned by calling a coroutine function.

> Generator based coroutine functions (e.g. those defined by decorating a function with `@asyncio.coroutine`) are superseded by the `async`/`await` syntax, but will continue to be supported _until_ Python 3.10 -- [docs.python.org/3.8/library/asyncio-task.html](https://docs.python.org/3.8/library/asyncio-task.html#asyncio-generator-based-coro). 
> 
> Refer to my post "[iterators, generators, coroutines](/posts/python-generators/)" for more details about generator based coroutines and their asyncio history.

### Tasks

[Tasks](https://docs.python.org/3.8/library/asyncio-task.html#asyncio.Task) are used to schedule coroutines _concurrently_.

All asyncio applications will typically have (at least) a single 'main' entrypoint task that will be scheduled to run immediately on the event loop. This is done using the `asyncio.run` function (see '[Running an asyncio program](#running-an-asyncio-program)'). 

A coroutine function is expected to be passed to `asyncio.run`, while _internally_ asyncio will check this using the helper function `coroutines.iscoroutine` (see: [source code](https://github.com/python/cpython/blob/master/Lib/asyncio/runners.py#L8)). If not a coroutine, then an error is raised, otherwise the coroutine will be passed to `loop.run_until_complete` (see: [source code](https://github.com/python/cpython/blob/master/Lib/asyncio/base_events.py#L599)). 

The `run_until_complete` function expects a [Future](#futures) (see below section for what a Future is) and uses another helper function `futures.isfuture` to check the type provided. If not a Future, then the low-level API `ensure_future` is used to convert the coroutine into a Future (see [source code](https://github.com/python/cpython/blob/master/Lib/asyncio/tasks.py#L653)).

In older versions of Python, if you were going to manually create your own Future and schedule it onto the event loop, then you would have used `asyncio.ensure_future` (now considered to be a low-level API), but with Python 3.7+ this has been superseded by `asyncio.create_task`. 

Additionally with Python 3.7, the idea of interacting with the event loop directly (e.g. getting the event loop, creating a task with `create_task` and then passing it to the event loop) has been replaced with `asyncio.run`, which abstracts it all away for you (see '[Running an asyncio program](#running-an-asyncio-program)' to understand what that means).

The following APIs let you see the state of the tasks running on the event loop:

- `asyncio.current_task`
- `asyncio.all_tasks`

> Note: for other available methods on a Task object please refer to [the documentation](https://docs.python.org/3.8/library/asyncio-task.html#asyncio.Task).

### Futures

A Future is a low-level awaitable object that represents an eventual result of an asynchronous operation.

To use an analogy: it's like an empty postbox. At _some point_ in the future the postman will arrive and stick a letter into the postbox.

This API exists to enable callback-based code to be used with `async`/`await`, while [`loop.run_in_executor`](https://docs.python.org/3.8/library/asyncio-eventloop.html#asyncio.loop.run_in_executor) is an example of an asyncio low-level API function that returns a Future (see also some of the APIs listed in [Concurrent Functions](#concurrent-functions)).

> Note: for other available methods on a Future please refer to [the documentation](https://docs.python.org/3.8/library/asyncio-future.html#asyncio.Future).

## Running an asyncio program

The high-level API (as per Python 3.7+) is:

```python
import asyncio

async def foo():
    print("Foo!")

async def hello_world():
    await foo()  # waits for `foo()` to complete
    print("Hello World!")

asyncio.run(hello_world())
```

The `.run` function always creates a _new_ event loop and _closes_ it at the end. If you were using the lower-level APIs, then this would be something you'd have to handle manually (as demonstrated below).

```python
loop = asyncio.get_event_loop()
loop.run_until_complete(hello_world())
loop.close()
```

### Running Async Code in the REPL

Prior to Python 3.8 you couldn't execute async code within the standard Python REPL (it would have required you to use the IPython REPL instead). 

To do this with the latest version of Python you would run `python -m asyncio`. Once the REPL has started you don't need to use `asyncio.run()`, but just use the `await` statement directly.

```
asyncio REPL 3.8.0+ (heads/3.8:5f234538ab, Dec  1 2019, 11:05:25)

[Clang 10.0.1 (clang-1001.0.46.4)] on darwin

Use "await" directly instead of "asyncio.run()".
Type "help", "copyright", "credits" or "license" for more information.

>>> import asyncio
>>> async def foo():
...   await asyncio.sleep(5)
...   print("done")
...
>>> await foo()
done
```

> Notice the REPL automatically executes `import asyncio` when starting up so we're able to use any `asyncio` functions (such as the `.sleep` function) without having to manually type that import statement ourselves.

## Concurrent Functions

The following functions help to co-ordinate the running of functions concurrently, and offer varying degrees of control dependant on the needs of your application.

- `asyncio.gather`: takes a sequence of awaitables, returns an aggregate list of successfully awaited values.
- `asyncio.shield`: prevent an awaitable object from being cancelled.
- `asyncio.wait`: wait for a sequence of awaitables, until the given 'condition' is met.
- `asyncio.wait_for`: wait for a single awaitable, until the given 'timeout' is reached.
- `asyncio.as_completed`: similar to `gather` but returns Futures that are populated when results are ready.

> Note: `gather` has specific options for handling errors and cancellations. For example, if `return_exceptions: False` then the first exception raised by one of the awaitables is returned to the caller of `gather`, where as if set to `True` then the exceptions are aggregated in the list alonside successful results. If `gather()` is cancelled, all submitted awaitables (that have not completed yet) are also cancelled.

## Deprecated functions

- `@asyncio.coroutine`: removed in favour of `async def` in Python 3.10
- `asyncio.sleep`: the `loop` parameter will be removed in Python 3.10

> Note: you'll find in most of these APIs a `loop` argument can be provided to enable you to indicate the specific event loop you want to utilize). It seems Python has deprecated this argument in 3.8, and will remove it completely in 3.10.

## Examples

### `gather`

The following example demonstrates how to wait for multiple asynchronous tasks to complete.

```
import asyncio


async def foo(n):
    await asyncio.sleep(5)  # wait 5s before continuing
    print(f"n: {n}!")


async def main():
    tasks = [foo(1), foo(2), foo(3)]
    await asyncio.gather(*tasks)


asyncio.run(main())
```

### `wait`

The following example uses the `FIRST_COMPLETED` option, meaning whichever task finishes first is what will be returned.

```
import asyncio
from random import randrange


async def foo(n):
    s = randrange(5)
    print(f"{n} will sleep for: {s} seconds")
    await asyncio.sleep(s)
    print(f"n: {n}!")


async def main():
    tasks = [foo(1), foo(2), foo(3)]
    result = await asyncio.wait(tasks, return_when=asyncio.FIRST_COMPLETED)
    print(result)


asyncio.run(main())
```

An example output of this program would be:

```
1 will sleep for: 4 seconds
2 will sleep for: 2 seconds
3 will sleep for: 1 seconds

n: 3!

({<Task finished coro=<foo() done, defined at await.py:5> result=None>}, {<Task pending coro=<foo() running at await.py:8> wait_for=<Future pending cb=[<TaskWakeupMethWrapper object at 0x10322b468>()]>>, <Task pending coro=<foo() running at await.py:8> wait_for=<Future pending cb=[<TaskWakeupMethWrapper object at 0x10322b4c8>()]>>})
```

### `wait_for`

The following example demonstrates how we can utilize a timeout to prevent waiting endlessly for an asynchronous task to finish.

```
import asyncio


async def foo(n):
    await asyncio.sleep(10)
    print(f"n: {n}!")


async def main():
    try:
        await asyncio.wait_for(foo(1), timeout=5)
    except asyncio.TimeoutError:
        print("timeout!")


asyncio.run(main())
```

> Note: the `asyncio.TimeoutError` doesn't provide any extra information so there's no point in trying to use it in your output (e.g. `except asyncio.TimeoutError as err: print(err)`).

### `as_completed`

The following example demonstrates how `as_complete` will yield the first task to complete, followed by the next quickest, and the next until all tasks are completed.

```
import asyncio
from random import randrange


async def foo(n):
    s = randrange(10)
    print(f"{n} will sleep for: {s} seconds")
    await asyncio.sleep(s)
    return f"{n}!"


async def main():
    counter = 0
    tasks = [foo("a"), foo("b"), foo("c")]

    for future in asyncio.as_completed(tasks):
        n = "quickest" if counter == 0 else "next quickest"
        counter += 1
        result = await future
        print(f"the {n} result was: {result}")


asyncio.run(main())
```

An example output of this program would be:

```
c will sleep for: 9 seconds
a will sleep for: 1 seconds
b will sleep for: 0 seconds

the quickest result was: b!
the next quickest result was: a!
the next quickest result was: c!
```

### `create_task`

The following example demonstrates how to convert a coroutine into a Task and schedule it onto the event loop.

```
import asyncio


async def foo():
    await asyncio.sleep(10)
    print("Foo!")


async def hello_world():
    task = asyncio.create_task(foo())
    print(task)
    await asyncio.sleep(5)
    print("Hello World!")
    await asyncio.sleep(10)
    print(task)


asyncio.run(hello_world())
```

We can see from the above program that we use `create_task` to convert our coroutine function into a Task. This automatically schedules the Task to be run on the event loop at the next available tick. 

This is in contrast to the lower-level API `ensure_future` (which is the preferred way of creating new Tasks). The `ensure_future` function has specific logic branches that make it useful for more input types than `create_task` which only supports scheduling a coroutine onto the event loop and wrapping it inside a Task (see: [`ensure_future` source code](https://github.com/python/cpython/blob/master/Lib/asyncio/tasks.py#L653)).

The output of this program would be:

```
<Task pending coro=<foo() running at create_task.py:4>>
Hello World!
Foo!
<Task finished coro=<foo() done, defined at create_task.py:4> result=None>
```

Let's review the code and compare to the above output we can see...

We convert `foo()` into a Task and then print the returned Task immediately after it is created. So when we print the Task we can see that its status is shown as 'pending' (as it hasn't been executed yet). 

Next we'll sleep for five seconds, as this will cause the `foo` Task to now be run (as the current Task `hello_world` will be considered busy).

Within the `foo` Task we also sleep, but for a _longer_ period of time than `hello_world`, and so the event loop will now context switch _back_ to the `hello_world` Task, where upon the sleep will pass and we'll print the output string `Hello World`.

Finally, we sleep again for ten seconds. This is just so we can give the `foo` Task enough time to complete and print its own output. If we didn't do that then the `hello_world` task would finish and close down the event loop. The last line of `hello_world` is printing the `foo` Task, where we'll see the status of the `foo` Task will now show as  'finished'.

### Callbacks

When dealing with a Task, which really is a Future, then you have the ability to execute a 'callback' function once the Future has a value set on it.

The following example demonstrates this by modifying the previous [`create_task`](#create_task) example code:

```
import asyncio


async def foo():
    await asyncio.sleep(10)
    return "Foo!"


def got_result(future):
    print(f"got the result! {future.result()}")


async def hello_world():
    task = asyncio.create_task(foo())
    task.add_done_callback(got_result)
    print(task)
    await asyncio.sleep(5)
    print("Hello World!")
    await asyncio.sleep(10)
    print(task)


asyncio.run(hello_world())
```

Notice in the above program we add a new `got_result` function that expects to receive a Future type, and thus calls `.result()` on the Future.

Also notice that to get this function to be called, we pass it to `.add_done_callback()` which is called on the Task returned by `create_task`.

The output of this program is:

```
<Task pending coro=<foo() running at gather.py:4> cb=[got_result() at gather.py:9]>
Hello World!
got the result! Foo!
<Task finished coro=<foo() done, defined at gather.py:4> result='Foo!'>
```

## Pools

When dealing with lots of concurrent operations it might be wise to utilize a 'pool' of threads (or subprocesses) to prevent exhausting your application's host resources. Asyncio provides a concept referred to as a Executor to help with this (see: [Executor documentation](https://docs.python.org/3.8/library/concurrent.futures.html#concurrent.futures.Executor)).

There are two types of 'executors':

- [`ThreadPoolExecutor`](https://docs.python.org/3.8/library/concurrent.futures.html#threadpoolexecutor)
- [`ProcessPoolExecutor`](https://docs.python.org/3.8/library/concurrent.futures.html#processpoolexecutor)

In order to execute code within one of these executors, you need to call the event loop's `.run_in_executor()` function and pass in the executor type as the first argument. If `None` is provided, then the _default_ executor is used (which is the `ThreadPoolExecutor`).

The following example is copied verbatim from the [Python documentation](https://docs.python.org/3.8/library/asyncio-eventloop.html#executing-code-in-thread-or-process-pools):

```
import asyncio
import concurrent.futures


def blocking_io():
    # File operations (such as logging) can block the
    # event loop: run them in a thread pool.
    with open("/dev/urandom", "rb") as f:
        return f.read(100)


def cpu_bound():
    # CPU-bound operations will block the event loop:
    # in general it is preferable to run them in a
    # process pool.
    return sum(i * i for i in range(10 ** 7))


async def main():
    loop = asyncio.get_running_loop()

    # 1. Run in the default loop's executor:
    result = await loop.run_in_executor(None, blocking_io)
    print("default thread pool", result)

    # 2. Run in a custom thread pool:
    with concurrent.futures.ThreadPoolExecutor() as pool:
        result = await loop.run_in_executor(pool, blocking_io)
        print("custom thread pool", result)

    # 3. Run in a custom process pool:
    with concurrent.futures.ProcessPoolExecutor() as pool:
        result = await loop.run_in_executor(pool, cpu_bound)
        print("custom process pool", result)


asyncio.run(main())
```

There is also an alternative way of scheduling a task to be run concurrently within a pool, without having to acquire the current event loop and passing the pool into it (as the above example demonstrates). 

To do this we'll need to 'submit' a function to be run in the pool, as shown below:

```
import asyncio
import concurrent.futures
import time

THREAD_POOL = concurrent.futures.ThreadPoolExecutor(max_workers=5)


async def main():
    future = THREAD_POOL.submit(time.sleep, 5)

    for result in concurrent.futures.as_completed([future]):
        assert future.done() and not future.cancelled()
        print("all done!")


asyncio.run(main())
```

One thing worth noting here is that because we've not used the `with` statement (like we did in the earlier pool example) it means we're not shutting down the pool once it has finished its work, and so (depending on if your program continues running) you may discover resources aren't being cleaned up.

To solve that problem we can call the `.shutdown()` method which is exposed to both types of executors via its parent class `concurrent.futures.Executor`. Below is an updated example that does that:

```
import asyncio
import concurrent.futures
import time

THREAD_POOL = concurrent.futures.ThreadPoolExecutor(max_workers=5)


async def main():
    future = THREAD_POOL.submit(time.sleep, 5)

    THREAD_POOL.shutdown()

    assert future.done() and not future.cancelled()

    print("all done!")


asyncio.run(main())
```

Notice the placement of the call to `.shutdown()` is _before_ we've explictly waited for the scheduled task to complete, and yet when we assert if the returned future is `.done()` we find that it is? 

This works because the default behaviour for the shutdown method is `wait=True` which means it'll wait for all scheduled tasks to complete before shutting down the executor pool. This also means it's a blocking call. 

If we passed `.shutdown(wait=False)` instead, then the call to `future.done()` would indeed raise an exception as the scheduled task would still be running and so in that case we'd need to ensure that we use another mechanism for acquiring the results of the scheduled tasks (such as `concurrent.futures.as_completed` or `concurrent.futures.wait`).
