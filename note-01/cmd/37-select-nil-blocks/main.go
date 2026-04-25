// Ch8 §8：nil channel 上发送或接收 **永远阻塞**，runtime 会报死锁。
// 这不是 bug，是规范定义的语义；正因为「永远阻塞」，nil channel 在 select 里可用来
// **动态禁用某个 case**——见 cmd/38-select-fan-in-disable。
//
// 运行预期：
//   fatal error: all goroutines are asleep - deadlock!
package main

import "fmt"

func main() {
	var ch chan string // 零值是 nil
	ch <- "message"    // 永久阻塞 → 整个程序仅此一个 goroutine → runtime 检测到死锁
	fmt.Println("never printed")
}
