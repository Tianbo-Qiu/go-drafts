// 读完 Ch2 入门：五个 goroutine 并发跑「假工作」。
// 下面用 time.Sleep 主线程是反例——魔法数字、可能早退或白等，仅用于和 03 对比。
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	for i := range 5 {
		go doWork(i)
	}
	time.Sleep(3 * time.Second)
	fmt.Printf("naive sleep exit after %v (do not use this pattern in real code)\n", time.Since(start))
}

func doWork(id int) {
	fmt.Printf("work %d started at %s\n", id, time.Now().Format("15:04:05"))
	time.Sleep(1 * time.Second)
	fmt.Printf("work %d ended at %s\n", id, time.Now().Format("15:04:05"))
}
