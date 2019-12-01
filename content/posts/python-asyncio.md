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

> asyncio is a library to write concurrent code using the `async`/`await` syntax. -- [docs.python.org/3.8/library/asyncio.html](https://docs.python.org/3.8/library/asyncio.html)

The asyncio module provides both high-level and low-level APIs. Library and Framework developers will be expected to use the low-level APIs, while all other users are encouraged to use the high-level APIs.

## Event Loop

The core element of all asyncio applications is the 'event loop'. The event loop is what schedules and runs asynchronous tasks (it also handles network IO operations and the running of subprocesses).

<a href="../../images/event-loop.png">
    <img src="../../images/event-loop.png">
</a>

> Image Credit: https://eng.paxos.com/python-3s-killer-feature-asyncio

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

> Generator based coroutine functions (e.g. those defined by decorating a function with `@asyncio.coroutine`) are superseded by the `async`/`await` syntax, but will continue to be supported _until_ Python 3.10 -- [docs.python.org/3.8/library/asyncio-task.html](https://docs.python.org/3.8/library/asyncio-task.html#asyncio-generator-based-coro)

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

## Concurrent Functions

The following functions help to co-ordinate the running of functions concurrently, and offer varying degrees of control dependant on the needs of your application.

- `asyncio.gather`: takes a sequence of awaitables, returns an aggregate list of successfully awaited values.
- `asyncio.shield`: prevent an awaitable object from being cancelled.
- `asyncio.wait`: wait for a sequence of awaitables, until the given 'condition' is met.
- `asyncio.wait_for`: wait for a single awaitable, until the given 'timeout' is reached.
- `asyncio.as_completed`: similar to `gather` but returns Futures that are populated when results are ready.

> Note: `gather` has specific options for handling errors and cancellations. For example, if `return_exceptions: False` then the first exception raised by one of the awaitables is returned to the caller of `gather`, where as if set to `True` then the exceptions are aggregated in the list alonside successful results. If `gather()` is cancelled, all submitted awaitables (that have not completed yet) are also cancelled.

## Deprecated functions

- `@asyncio.coroutine`: removed in Python 3.10
- `asyncio.sleep`: removed in Python 3.10

> Note: you'll find in most of these APIs a `loop` argument can be provided to enable you to indicate the specific event loop you want to utilize). It seems Python has deprecated this argument in 3.8, and will remove it completely in 3.10.

## Examples

### Gather

The following example demonstrates how simple the `gather` behaviour is. 

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

### Wait

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

The output of this program would be:

```
1 will sleep for: 4 seconds
2 will sleep for: 2 seconds
3 will sleep for: 1 seconds

n: 3!

({<Task finished coro=<foo() done, defined at await.py:5> result=None>}, {<Task pending coro=<foo() running at await.py:8> wait_for=<Future pending cb=[<TaskWakeupMethWrapper object at 0x10322b468>()]>>, <Task pending coro=<foo() running at await.py:8> wait_for=<Future pending cb=[<TaskWakeupMethWrapper object at 0x10322b4c8>()]>>})
```
