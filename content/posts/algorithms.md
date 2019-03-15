---
title: "Algorithms in Python"
date: 2019-03-13T09:15:59Z
categories:
  - "code"
  - "guide"
tags:
  - "algorithms"
  - "data-structures"
  - "design"
  - "python"
  - "searching"
  - "sorting"
draft: false
---

In this post I'll be demonstrating a few common algorithms using the Python language. I'm only covering a very small subset of popular algorithms because otherwise this would become a long and diluted list. 

Instead I'm going to focus specifically on algorithms that I find useful and are important to know and understand:

- [Binary Search](#binary-search)
- [Merge Sort](#merge-sort)
- [Quick Sort](#quick-sort)

We'll then wrap up by considering the [differences between a merge sort and a quick sort](#difference-between-merge-and-quick-sort).

## Binary Search

The most popular algorithm (by far) for searching a value in a _sorted_ list is 'binary search'. The reason for its popularity is that it provides [logarithmic performance](/posts/algorithmic-complexity-in-python/#logarithmic-time) (on average) for access, search, insertion and deletion operations.

- Average: `O(log(n))`
- Worst: `O(n)`

The algorithm can be broken down into the following pseudo-steps:

- define 'start' and 'end' positions (usually length of list).
- locate middle of list.
- stop searching if correct value found.
- if value is greater: change end position to the middle index.
- if value is smaller: change start position to the middle index.
- repeat above steps until value is found.

What this ultimately achieves is shortening the search 'window' of items by half each time (that's the logarithmic part). So if you have a list of 1000 elements, then we can say it'll take a maximum of ten operations to find the number you're looking for (that's: `log 2(10)` == `2^10` == `1024`).

That's outstanding.

Below is an implementation of this popular search algorithm:

```
def binary_search(collection, item):
    start = 0
    stop = len(collection) - 1

    while start <= stop:
        middle = round((start + stop) / 2)
        guess = collection[middle]

        if guess == item:
            return middle
        if guess > item:
            stop = middle - 1
        else:
            start = middle + 1

    return None

collection = [1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 20, 21]

result = binary_search(collection, 9)

print(f'the value was found at index {result}\n\n')  # index 4
```

> Note: you could swap `round((start + stop) / 2)` for `(start + stop) // 2` which uses Python's `//` floor division operator, but I typically opt for the clarity of using the explicit `round` function.

## Merge Sort

Given an unsorted list of integers, the merge sort algorithm enables us to sort the list. 

- Best: `Ω(n log(n))`
- Average: `O(n log(n))`
- Worst: `O(n log(n))`

The algorithm can be broken down into the following pseudo-steps:

- Recursively split given list into two partitions.
- Divide the list until reaching its smallest partition.
- Recursively merge each pair of partitions (repeat till reconstructed list).

That short list actually defines two distinct processes: a 'break up into partitions' step followed by a 'merge' step. The merge pseudo-steps could be further broken down to look something like the following:

- To merge two partitions we iterate over each of them.
- Compare elements from both partitions (left and right).
- Append smaller element to new 'results' (i.e. sorted) list.
- After comparing the elements, increment index for winning partition.

The above merge steps process is carried out against each recursively sorted set of partitions, and is done until we reach a final 'sorted' list.

Below is an implementation of this sorting algorithm:

```
def merge(left, right):
    left_index, right_index = 0, 0
    result = []

    while left_index < len(left) and right_index < len(right):
        if left[left_index] < right[right_index]:
            result.append(left[left_index])
            left_index += 1
        else:
            result.append(right[right_index])
            right_index += 1

    result += left[left_index:]
    result += right[right_index:]

    return result


def merge_sort(collection):
    if len(collection) <= 1:
        print('collection is <= 1\n')
        return collection

    middle = len(collection) // 2
    left = collection[:middle]
    right = collection[middle:]

    left = merge_sort(left)
    right = merge_sort(right)

    return merge(left, right)

collection = [10, 5, 2, 3, 7, 0, 9, 12]

result = merge_sort(collection)

print(f'merge sort of {collection} result: {result}\n\n')
```

## Quick Sort

Given an unsorted list of integers, the merge sort algorithm enables us to sort the list. 

- Best: `Ω(n log(n))`
- Average: `O(n log(n))`
- Worst: `O(n^2)`

The algorithm can be broken down into the following pseudo-steps:

- Pick a 'pivot' (e.g. random index from the list).
- Iterate the collection twice, creating two new lists.
- Create a list consisting of elements less than pivot.
- Create a list consisting of elements greater than pivot.
- function return value is `fn(less) + pivot + fn(greater)`.

> Note: notice the 'less' and 'greater' lists are passed recursively back through the quick sort function.

Below is an implementation of this sorting algorithm:

```
from random import randrange

collection = [10, 5, 2, 3, 7, 0, 9, 12]

def quicksort(collection):
    if len(collection) < 2:
        return collection
    else:
        random = randrange(0, len(collection))
        pivot = collection.pop(random)

        less = [i for i in collection if i <= pivot]
        greater = [i for i in collection if i > pivot]

        return quicksort(less) + [pivot] + quicksort(greater)

cloned_collection = collection.copy()  # avoid mutation

result = quicksort(cloned_collection)

print(f'quick sort of {collection} result: {result}\n\n')
```

## Difference between Merge and Quick sort

Both merge sort and quick sort are 'divide and conquer' algorithms, e.g. they divide a problem up into two and then process the individual partitions, and will keep dividing up the problem for as a long as it can. 

So when considering which you should use (i.e. which is better?) you'll find merge sort is more performant because its particular implementation is more efficient with the operations it carries out.

Specifically, the difference between merge sort and quick sort is that with quick sort you'll loop over the collection twice in order to create two partitions, but remember that doing so doesn't mean the partitions are sorted. 

These quick sort partitions are not sorted until the recursive calls end up with a small enough partition (length of 1) where the left and right partitions can be analyzed, joined and therefore considered 'sorted'. 

Compare that to merge sort which recursively creates multiple partitions (all the way down to the smallest possible partition) _before_ it then recursively rolls back up the execution stack and attempts to sort each of the partitions along the way.

Additionally, with a merge sort it's possible to parallelize the data over multiple processes, while quick sort requires data to be processed within a single process.

This means that quick sort has potentially more operations to be carried out than merge sort, and therefore has a greater time complexity. Although quick sort can offer better 'space' complexity (worst case: `O(log(n))`) compared to merge sort (worst case: `O(n)`).

All that said, the implementation of the algorithms can be tweaked as needed to produce better or worst performance. For example, quick sort could be modified to use another algorithm called [intro sort](https://en.wikipedia.org/wiki/Introsort) which is a mix of quick sort, insertion sort, and heapsort, that's worst-case `O(n log(n))` but retains the speed of quick sort in most cases.
