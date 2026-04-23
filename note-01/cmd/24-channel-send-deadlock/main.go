// Ch7 §2：同步 channel 反例——发送方等不到接收者。
// 接收 goroutine 先睡 5 秒后直接返回，永远不会读 channel；
// 主 goroutine 的 `msgs <- "HELLO"` 因此永久阻塞，运行时最终报：
//   fatal error: all goroutines are asleep - deadlock!
// 这是一个有意的反例，演示 channel 默认同步语义下「一边缺席」就会死锁。
package main

import (
	"fmt"
	"time"
)

func main() {
	msgs := make(chan string)
	go receiver(msgs)

	msgs <- "HELLO"
	fmt.Println("main exit") // 不会打印
}

func receiver(msgs chan string) {
	time.Sleep(5 * time.Second)
	fmt.Println("receiver slept for 5 seconds and returns without reading")
}
