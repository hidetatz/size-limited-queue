package slqueue

import (
	"runtime"
	"sync"
)

type MutexQueue struct {
	mu       sync.Mutex
	capacity int
	queue    []int
}

func NewMutexQueue(capacity int) *MutexQueue {
	return &MutexQueue{
		capacity: capacity,
		queue:    []int{},
	}
}

func (s *MutexQueue) Push(i int) {
	// acquire lock first to read len(s.queue) atomically
	s.mu.Lock()
	for len(s.queue) == s.capacity {

		// This unlock is necessary to prevent deadlock.
		// One possible deadlock scenario (which can heppen when these Unlock/Lock doesn't exist in the loop):
		// 1. call Push() when the queue is full
		// 2. enter the loop holding the mutex lock
		// 3. another goroutine calls Pop() <- deadlock!
		s.mu.Unlock()

		// yield the processor allowing other goroutines to run
		runtime.Gosched()

		// Then, acquire lock again to restart the loop, before entering the critical section
		s.mu.Lock()
	}

	s.queue = append(s.queue, i)
	s.mu.Unlock()
}

func (s *MutexQueue) Pop() int {
	s.mu.Lock()
	for len(s.queue) == 0 {
		s.mu.Unlock()
		runtime.Gosched()
		s.mu.Lock()
	}

	ret := s.queue[0]
	s.queue = s.queue[1:]
	s.mu.Unlock()

	return ret
}
