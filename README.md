## Overview

This repo is not a library. It is a reference implementation of "hint" for the programmers who are trying to understand how sync.Cond works, how it is used in the real world.

I'll leave a brief explanation in this README, but I would strongly recommend you to read **[my full article](https://dtyler.io/articles/2021/04/13/sync_cond/) to get better understanding** instead of just looking at this repository.

This repository contains a working code which implements a size-limited-queue. First, let me describe the spec of it:

* The queue can contain only int values.
* The queue supports `Push` and  `Pop`. Like the common queue data structure, the order is FIFO.
* The queue supports **size capacity feature**. The queue can contain elements up to the capacity.
* When trying to push an element to the queue when the queue is full, it blocks until it gets at least a space.
* When trying to pop an element from the queue when the queue is empty, it blocks until it gets at least an element.

There are three implementations described below:

* [single_thread_slqueue.go](https://github.com/dty1er/size-limited-queue/blob/main/single_thread_slqueue.go)
* [mutex_slqueue.go](https://github.com/dty1er/size-limited-queue/blob/main/mutex_slqueue.go)
* [slqueue.go](https://github.com/dty1er/size-limited-queue/blob/main/slqueue.go)

`slqueue.go` has simpler implementation at the old revision. [176eb78](https://github.com/dty1er/size-limited-queue/blob/176eb788c5be9f7c9fb98b57f39d9953a56204c1/slqueue.go)

### single_thread_slqueue.go

[single_thread_slqueue.go](https://github.com/dty1er/size-limited-queue/blob/main/single_thread_slqueue.go)

It works basically, but doesn't work correctly under multithread environment. The length check in `Push` / `Pop` and queue manipulation (`append` / moving the queue head to pop) must be atomic, but this implementation does not consider it. As a result, 

* Queue capacity is 10, now the Queue length is 9
* Goroutine A and B tries to push an value
  * A checks the queue capacity, then it is 9
  * At almost the same time, B checks the queue capacity, then it is 9 because A still have not finished the queue manipulation (`append`)
  * A appends an element, now the queue length is 10
  * B appends an element, now the queue length is 11 <- violates the queue spec!

I implemented single_thread_slqueue.go to compare it with upcoming `mutex_slqueue.go`.

### mutex_slqueue.go

[mutex_slqueue.go](https://github.com/dty1er/size-limited-queue/blob/main/mutex_slqueue.go)

This uses `sync.Mutex` to make the queue length check and queue manipulation atomic.
This actually works following all the spec, but the problem is its inefficiency.

```go
func (s *MutexQueue) Push(i int) {
	s.mu.Lock()
	for len(s.queue) == s.capacity { // 1
		s.mu.Unlock()
		runtime.Gosched()
		s.mu.Lock()
	}

	s.queue = append(s.queue, i)
	s.mu.Unlock()
}
```

See `1` in the code. Because of the spec `When trying to push an element to the queue when the queue is full, it blocks until it gets at least a space`, it needs spin (for-loop) to achieve it. However, spin is usually inefficient. Using `sync.Cond`, it can be improved to get more efficient.

### slqueue.go - simple

[slqueue.go](https://github.com/dty1er/size-limited-queue/blob/176eb788c5be9f7c9fb98b57f39d9953a56204c1/slqueue.go)

This uses `sync.Cond` to make mutex implementation better. See the implementation:

```go
func (s *SizeLimitedQueue) Push(i int) {
	s.cond.L.Lock()
	for len(s.queue) == s.capacity { // 1
		s.cond.Wait() // 2
	}

	s.queue = append(s.queue, i)

	s.cond.Broadcast() // 3
	s.cond.L.Unlock()
}

func (s *SizeLimitedQueue) Pop() int {
	s.cond.L.Lock()
	for len(s.queue) == 0 {
		s.cond.Wait()
	}

	ret := s.queue[0]
	s.queue = s.queue[1:]
	s.cond.Broadcast()
	s.cond.L.Unlock()

	return ret
}
```

When trying to `Push`, it first check the length (at `1`) then if the condition is not met, call `cond.Wait()` (at `2`) . In the wait, runtime will suspend the waiting goroutine and waits for the "notification" on the cond. When a goroutine calls `cond.Broadcast()`, the cond is notified - waiting goroutines are waken up then goes to the top of for-loop.

This has an advantage because it does not require spins.

### slqueue.go - improved

[slqueue.go](https://github.com/dty1er/size-limited-queue/blob/main/slqueue.go)

On the above `slqueue.go`, some optimizations are applied. They are described more specifically on my [article](https://dtyler.io/articles/2021/04/13/sync_cond/).

## For readers

I'm quite sure this README is not enough to understand `sync.Cond`. This is just a brief "summary".

To understand `sync.Cond` better, I'd say you have to understand a synchronization primitive "Condition Variable" in POSIX. `sync.Cond` is actually just a Go version of it.

I described what "Condition Variable" is, and the detailed description of above code, also some supplementary information about it in my [article](https://dtyler.io/articles/2021/04/13/sync_cond/). I would recommend reading it with the actual source code in this repository.

If you find this repo helpful to learn `sync.Cond`, please leave a star!
