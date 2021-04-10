package slqueue

import (
	"sync"
)

type SizeLimitedQueue struct {
	// nonFullCond is a cond var to notify that the queue length changed to
	// non-full from full.
	// This is intended to be used in waiting the queue length change in Push()
	nonFullCond *sync.Cond

	// nonEmptyCond is a cond var to notify that the queue length changed to
	// non-empty from empty.
	// This is intended to be used in waiting the queue length change in Pop()
	nonEmptyCond *sync.Cond
	capacity     int
	queue        []int
	mu           *sync.Mutex
}

func New(capacity int) *SizeLimitedQueue {
	mu := &sync.Mutex{}
	return &SizeLimitedQueue{
		nonFullCond:  sync.NewCond(mu),
		nonEmptyCond: sync.NewCond(mu),
		capacity:     capacity,
		queue:        []int{},
		mu:           mu,
	}
}

func (s *SizeLimitedQueue) Push(i int) {
	s.nonFullCond.L.Lock()
	for len(s.queue) == s.capacity {
		// wait for the cond var gets notified when the queue becomes non-full
		// because Push() can push an element only when the queue is non-full
		// This is notified when the queue gets non-full (= an existing element is popped in Pop())
		s.nonFullCond.Wait()
	}

	wasEmpty := len(s.queue) == 0
	s.queue = append(s.queue, i)

	// if the queue was empty but now non-empty,
	// it should be notified to the goroutines which are waiting for Pop()
	// because they can start running only when the queue is non-empty
	if wasEmpty {
		s.nonEmptyCond.Signal()
	}
	s.nonFullCond.L.Unlock()
}

func (s *SizeLimitedQueue) Pop() int {
	s.nonEmptyCond.L.Lock()
	for len(s.queue) == 0 {
		// wait for the cond var gets notified when the queue becomes non-empty
		// because Pop() can pop an element only when the queue is non-empty
		// This is notified when the queue gets non-empty (= a new element is pushed in Push())
		s.nonEmptyCond.Wait()
	}

	wasFull := len(s.queue) == s.capacity
	ret := s.queue[0]
	s.queue = s.queue[1:]

	if wasFull {
		// if the queue was full but now non-full,
		// it should be notified to the goroutines which are waiting for Push()
		// because they can start running only when the queue is non-full
		s.nonFullCond.Signal()
	}
	s.nonEmptyCond.L.Unlock()

	return ret
}
