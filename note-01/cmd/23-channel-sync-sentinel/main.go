// Ch7 §2：默认同步 channel 的最小示例——
// 发送方写入 channel 会阻塞到有接收者；约定 "STOP" 作为哨兵（poison pill）让接收者退出循环。
// 注意：主 goroutine 发完最后一条就 return，接收者打印 "STOP" 那一行可能来不及出现，
// 正式代码里应改用 close(ch) + range 或者 WaitGroup 收尾。
package main

import "fmt"

func main() {
	msgs := make(chan string)
	go receiver(msgs)

	msgs <- "HELLO"
	msgs <- "WORLD"
	msgs <- "STOP"
}

func receiver(msgs chan string) {
	for msg := ""; msg != "STOP"; {
		msg = <-msgs
		fmt.Println("received:", msg)
	}
}
