package slqueue

import (
	"sync"
)

type SizeLimitedQueue struct {
	cond     *sync.Cond
	capacity int
	queue    []int
}

func New(capacity int) *SizeLimitedQueue {
	return &SizeLimitedQueue{
		cond:     sync.NewCond(&sync.Mutex{}),
		capacity: capacity,
		queue:    []int{},
	}
}

func (s *SizeLimitedQueue) Push(i int) {
	// Acquire lock before entering the critical section
	s.cond.L.Lock()
	for len(s.queue) == s.capacity {
		// Wait for a signal sent by Broadcast()
		// When receives a signal, it goes to the head of the loop
		// then checks the condition again
		s.cond.Wait()
	}

	s.queue = append(s.queue, i)

	// Because condition (= length of s.queue) is changed,
	// it sends a signal to all the goroutines
	// Because they wait for the signal, it doesn't enter busy-loop,
	// so it is more efficient.
	s.cond.Broadcast()
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
