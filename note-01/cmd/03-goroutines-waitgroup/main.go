// Ch1 提到的共享内存同步：用 WaitGroup 等待一组 goroutine 结束（替代魔法 Sleep）。
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			doWork(id)
		}(i)
	}
	wg.Wait()
	fmt.Printf("parallel with WaitGroup: all done in %v\n", time.Since(start))
}

func doWork(id int) {
	fmt.Printf("work %d started at %s\n", id, time.Now().Format("15:04:05"))
	time.Sleep(1 * time.Second)
	fmt.Printf("work %d ended at %s\n", id, time.Now().Format("15:04:05"))
}
