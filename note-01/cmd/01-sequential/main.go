// 读完 Ch1：同一组「假工作」在单线程里串行执行，总耗时约为各任务之和。
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	for i := range 5 {
		doWork(i)
	}
	fmt.Printf("sequential: all done in %v\n", time.Since(start))
}

func doWork(id int) {
	fmt.Printf("work %d started at %s\n", id, time.Now().Format("15:04:05"))
	time.Sleep(1 * time.Second)
	fmt.Printf("work %d ended at %s\n", id, time.Now().Format("15:04:05"))
}
