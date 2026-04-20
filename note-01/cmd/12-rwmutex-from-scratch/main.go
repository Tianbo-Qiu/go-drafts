// 教学用：用两把 Mutex 拼出「多读单写」语义，便于对照标准库 sync.RWMutex。
// 注意：未处理 writer 饥饿、panic 配对等问题，勿当生产实现。
package main

import (
	"fmt"
	"sync"
	"time"
)

// ReadWriteMutex：第一个读者抢 global，最后一个读者放 global；写者只操作 global。
type ReadWriteMutex struct {
	readersCount int
	readersLock  sync.Mutex
	globalLock   sync.Mutex
}

func (rw *ReadWriteMutex) ReadLock() {
	rw.readersLock.Lock()
	rw.readersCount++
	if rw.readersCount == 1 {
		rw.globalLock.Lock()
	}
	rw.readersLock.Unlock()
}

func (rw *ReadWriteMutex) ReadUnlock() {
	rw.readersLock.Lock()
	rw.readersCount--
	if rw.readersCount == 0 {
		rw.globalLock.Unlock()
	}
	rw.readersLock.Unlock()
}

func (rw *ReadWriteMutex) WriteLock() {
	rw.globalLock.Lock()
}

func (rw *ReadWriteMutex) WriteUnlock() {
	rw.globalLock.Unlock()
}

func main() {
	var rw ReadWriteMutex
	var wg sync.WaitGroup

	for id := 1; id <= 3; id++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			rw.ReadLock()
			defer rw.ReadUnlock()
			fmt.Printf("reader %d: enter (shared read)\n", readerID)
			time.Sleep(300 * time.Millisecond)
			fmt.Printf("reader %d: leave\n", readerID)
		}(id)
	}

	time.Sleep(50 * time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()
		rw.WriteLock()
		defer rw.WriteUnlock()
		fmt.Println("writer: enter (exclusive)")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("writer: leave")
	}()

	wg.Wait()
	fmt.Println("done")
}
