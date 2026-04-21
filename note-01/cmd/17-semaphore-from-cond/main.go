// 用 Cond 实现计数信号量；permits=0 时可在「先 Release 后 Acquire」场景里传递完成信号。
package main

import (
	"fmt"
	"sync"
)

type Semaphore struct {
	permits int
	cond    *sync.Cond
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		permits: n,
		cond:    sync.NewCond(&sync.Mutex{}),
	}
}

func (s *Semaphore) Acquire() {
	s.cond.L.Lock()
	for s.permits <= 0 {
		s.cond.Wait()
	}
	s.permits--
	s.cond.L.Unlock()
}

func (s *Semaphore) Release() {
	s.cond.L.Lock()
	s.permits++
	s.cond.Signal()
	s.cond.L.Unlock()
}

func main() {
	sem := NewSemaphore(0)
	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(id int) {
			defer wg.Done()
			fmt.Println(id, "work")
			sem.Release()
		}(i)
	}
	for i := range n {
		sem.Acquire()
		fmt.Println(i, "observed completion")
	}
	wg.Wait()
	fmt.Println("all", n, "done")
}
