// Ch7 §5：`for msg := range ch` 是 close + `msg, ok := <-ch` 的语法糖。
// 当 channel 被 close 且 buffer 排空后，range 自动退出循环；比手写 ok 判断更简洁，
// 也避免忘记处理 ok=false 导致的无限循环读零值。
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
	for msg := range ch {
		fmt.Println(time.Now().Format("15:04:05"), "received:", msg)
		time.Sleep(1 * time.Second)
	}
	fmt.Println(time.Now().Format("15:04:05"), "receiver finished: channel closed")
}
