package slqueue

import (
	"sync"
	"testing"
)

func Test_SingleThreadQueue_Push_Pop(t *testing.T) {
	q := NewSingleThreadQueue(10)
	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			q.Push(i)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			q.Pop()
			wg.Done()
		}()
	}
	wg.Wait()
}
