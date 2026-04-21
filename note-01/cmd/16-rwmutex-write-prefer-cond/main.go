// 用 sync.Cond + 计数器实现「写偏好」读写锁：有写者等待时，新读者在 ReadLock 里阻塞。
// 教学用；生产请用 sync.RWMutex 或经充分测试的实现。
package main

import (
	"fmt"
	"sync"
	"time"
)

type writePreferRW struct {
	readersCounter int
	writersWaiting int
	writerActive   bool
	cond           *sync.Cond
}

func newWritePreferRW() *writePreferRW {
	return &writePreferRW{cond: sync.NewCond(&sync.Mutex{})}
}

func (rw *writePreferRW) ReadLock() {
	rw.cond.L.Lock()
	for rw.writersWaiting > 0 || rw.writerActive {
		rw.cond.Wait()
	}
	rw.readersCounter++
	rw.cond.L.Unlock()
}

func (rw *writePreferRW) WriteLock() {
	rw.cond.L.Lock()
	rw.writersWaiting++
	for rw.readersCounter > 0 || rw.writerActive {
		rw.cond.Wait()
	}
	rw.writersWaiting--
	rw.writerActive = true
	rw.cond.L.Unlock()
}

func (rw *writePreferRW) ReadUnlock() {
	rw.cond.L.Lock()
	rw.readersCounter--
	if rw.readersCounter == 0 {
		rw.cond.Broadcast()
	}
	rw.cond.L.Unlock()
}

func (rw *writePreferRW) WriteUnlock() {
	rw.cond.L.Lock()
	rw.writerActive = false
	rw.cond.Broadcast()
	rw.cond.L.Unlock()
}

func main() {
	rw := newWritePreferRW()
	var wg sync.WaitGroup
	for i := range 3 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rw.ReadLock()
			fmt.Println("read", id)
			time.Sleep(50 * time.Millisecond)
			rw.ReadUnlock()
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		rw.WriteLock()
		fmt.Println("write exclusive")
		rw.WriteUnlock()
	}()
	wg.Wait()
	fmt.Println("done")
}
