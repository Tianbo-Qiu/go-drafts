// Ch6 §1：用 Mutex + Cond 自拼一个可动态 Add 的 WaitGroup；计数归零时 Broadcast 唤醒所有 Wait 方。
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type WaitGrp struct {
	groupSize int
	cond      *sync.Cond
}

func NewWaitGrp() *WaitGrp {
	return &WaitGrp{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (wg *WaitGrp) Add(delta int) {
	wg.cond.L.Lock()
	wg.groupSize += delta
	wg.cond.L.Unlock()
}

func (wg *WaitGrp) Done() {
	wg.cond.L.Lock()
	wg.groupSize--
	if wg.groupSize == 0 {
		wg.cond.Broadcast()
	}
	wg.cond.L.Unlock()
}

func (wg *WaitGrp) Wait() {
	wg.cond.L.Lock()
	for wg.groupSize > 0 {
		wg.cond.Wait()
	}
	wg.cond.L.Unlock()
}

func doWork(id int, wg *WaitGrp) {
	defer wg.Done()
	i := rand.Intn(5)
	time.Sleep(time.Duration(i*100) * time.Millisecond)
	fmt.Println(id, "done working after", i*100, "ms")
}

func main() {
	wg := NewWaitGrp()
	for i := range 4 {
		wg.Add(1)
		go doWork(i, wg)
	}
	wg.Wait()
	fmt.Println("all completed")
}
