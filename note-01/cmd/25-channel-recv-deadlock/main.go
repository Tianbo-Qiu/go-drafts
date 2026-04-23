// Ch7 §2：同步 channel 反例——接收方等不到发送者。
// sender goroutine 睡 5 秒后直接返回，永远不会写 channel；
// 主 goroutine 的 `<-msgs` 因此永久阻塞，运行时最终报：
//   fatal error: all goroutines are asleep - deadlock!
// 与 24-channel-send-deadlock 对称：哪一侧缺席都会卡死。
package main

import (
	"fmt"
	"time"
)

func main() {
	msgs := make(chan string)
	go sender(msgs)

	fmt.Println("reading message from channel...")
	msg := <-msgs
	fmt.Println("received:", msg) // 不会打印
}

func sender(msgs chan string) {
	time.Sleep(5 * time.Second)
	fmt.Println("sender slept for 5 seconds and returns without sending")
}
