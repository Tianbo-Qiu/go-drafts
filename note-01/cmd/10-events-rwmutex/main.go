// 写少读多：追加事件用 Lock；大量只读快照用 RLock，读之间可并发。
package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type eventLog struct {
	mu     sync.RWMutex
	events []string
}

func (l *eventLog) appendEvent(s string) {
	l.mu.Lock()
	l.events = append(l.events, s)
	l.mu.Unlock()
}

func (l *eventLog) snapshot() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]string, len(l.events))
	copy(out, l.events)
	return out
}

func main() {
	log := &eventLog{}
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 50 {
			log.appendEvent("match " + strconv.Itoa(i))
			time.Sleep(5 * time.Millisecond)
		}
	}()

	const readers = 20
	wg.Add(readers)
	for range readers {
		go func() {
			defer wg.Done()
			for range 80 {
				_ = log.snapshot()
			}
		}()
	}

	wg.Wait()
	fmt.Println("events:", len(log.snapshot()))
}
