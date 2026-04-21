// 父 goroutine 持锁依次等三个子任务：展示 Wait 会暂时释放锁、被 Signal 唤醒后需用 for 复查条件。
package main

import (
	"fmt"
	"sync"
)

func doWork(cond *sync.Cond) {
	cond.L.Lock()
	fmt.Println("work started, hold lock")
	cond.Signal()
	fmt.Println("work done, sent signal, releasing lock")
	cond.L.Unlock()
}

func main() {
	cond := sync.NewCond(&sync.Mutex{})
	cond.L.Lock()
	for i := range 3 {
		fmt.Println("=============== LOOP", i)
		go doWork(cond)
		fmt.Println("waiting for child goroutine")
		cond.Wait()
		fmt.Println("child goroutine finished")
	}
	cond.L.Unlock()
}
