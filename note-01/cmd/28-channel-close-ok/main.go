// Ch7 §5：close 与 `msg, ok := <-ch`。
// close 后不能再发送；但接收仍然合法：一旦 buffer 排空，再读立即返回 零值 + ok=false。
// 用 ok 标志比哨兵值更稳：哨兵值有时与业务值冲突，而 close 是 channel 自己的状态变化。
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	go receiver(ch)

	for i := 1; i <= 3; i++ {
		fmt.Println(time.Now().Format("15:04:05"), "sending:", i)
		ch <- i
		time.Sleep(1 * time.Second)
	}
	close(ch)
	time.Sleep(3 * time.Second)
}

func receiver(ch <-chan int) {
	for {
		msg, ok := <-ch
		fmt.Println(time.Now().Format("15:04:05"), "received:", msg, "ok:", ok)
		if !ok {
			return // channel 已关且排空，退出循环
		}
		time.Sleep(1 * time.Second)
	}
}
