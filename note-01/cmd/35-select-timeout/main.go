// Ch8 §6：用 `time.After(d)` 给单次 select 加超时。
// time.After 返回一个一次性 channel，d 后会发出一个 time.Time；select 哪条先就绪走哪条。
// 命令行参数指定超时秒数（默认 5）；消息固定 3 秒后到达，可观察 t<3 / t>=3 两种结果。
//
// 注意：time.After 创建的 timer 在 d 到期前不会被 GC，循环里高频用容易堆积；
// 真实代码里多用 `context.WithTimeout` 或 `time.NewTimer` + `Stop()`。
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	timeoutSec := 5
	if len(os.Args) > 1 {
		if v, err := strconv.Atoi(os.Args[1]); err == nil {
			timeoutSec = v
		}
	}
	d := time.Duration(timeoutSec) * time.Second

	msgs := sendMsgAfter(3 * time.Second)
	fmt.Printf("waiting up to %d seconds for message...\n", timeoutSec)

	select {
	case msg := <-msgs:
		fmt.Println("message received:", msg)
	case t := <-time.After(d):
		fmt.Println("timed out at:", t.Format("15:04:05"))
	}
}

func sendMsgAfter(d time.Duration) <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(d)
		ch <- "hello"
	}()
	return ch
}
