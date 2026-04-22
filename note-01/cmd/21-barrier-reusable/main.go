// Ch6 §3：可复用的 Barrier（带 generation 分代），多 goroutine 在卡点汇合，同轮一起放行。
// 运行期较长，观察到两三轮后可直接 Ctrl+C。
package main

import (
	"fmt"
	"sync"
	"time"
)

type Barrier struct {
	size       int
	waitCount  int
	generation int // 关键：区分不同轮次，避免跨轮唤醒串台
	cond       *sync.Cond
}

func NewBarrier(size int) *Barrier {
	return &Barrier{
		size: size,
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (b *Barrier) Wait() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	myGen := b.generation
	b.waitCount++

	if b.waitCount == b.size {
		b.waitCount = 0
		b.generation++
		b.cond.Broadcast()
		return
	}
	for myGen == b.generation {
		b.cond.Wait()
	}
}

func workAndWait(name string, timeToWork int, barrier *Barrier) {
	ts := time.Now()
	for i := 0; ; i++ {
		fmt.Printf("[%v] %v: %s running\n", time.Since(ts), i, name)
		time.Sleep(time.Duration(timeToWork) * time.Second)
		fmt.Printf("[%v] %v: %s waiting on barrier\n", time.Since(ts), i, name)
		barrier.Wait()
	}
}

func main() {
	barrier := NewBarrier(2)
	go workAndWait("A", 1, barrier)
	go workAndWait("B", 3, barrier)

	time.Sleep(10 * time.Second)
	fmt.Println("demo end")
}
