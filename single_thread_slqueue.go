package slqueue

// SingleThreadQueue is a size-limited queue which works correctly
// under the single thread (goroutine) environment.
// If it's accessed from multiple goroutines, it can work correctly, but potentially
//   * More items than capacity might be stored
//   * Panic
// can happen.
type SingleThreadQueue struct {
	capacity int
	queue    []int
}

func NewSingleThreadQueue(capacity int) *SingleThreadQueue {
	return &SingleThreadQueue{
		capacity: capacity,
		queue:    []int{},
	}
}

// Push pushes the given value to the queue.
// When the queue capacity is full, it blocks until it has
// enough space to save the valud.
// Because this queue is for single-threaded environment,
// if you call the method concurrently, it can work wrong.
// Specifically, it can exceed the capacity.
func (s *SingleThreadQueue) Push(i int) {
	for len(s.queue) == s.capacity {
		// busy loop
	}

	s.queue = append(s.queue, i)
}

// Pop pops a value from the queue.
// When the queue contains nothing, it blocks until it has
// something to pop.
// Because this queue is for single-threaded environment,
// if you call the method concurrently, it can work wrong.
// Specifically, it can panic because it references the empty
// slice.
func (s *SingleThreadQueue) Pop() int {
	for len(s.queue) == 0 {
		// busy loop
	}

	ret := s.queue[0]
	s.queue = s.queue[1:]

	return ret
}
