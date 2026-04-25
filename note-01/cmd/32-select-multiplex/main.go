// Ch8 §1：select 的最小用法——同时等多条 channel，谁先到就处理谁。
// 两条 channel 用不同节奏发消息，select 在循环里把先就绪的那条取走打印；
// 多路就绪时，规范规定 select **伪随机** 选一条。
package main

import (
	"fmt"
	"time"
)

func main() {
	chA := printAtInterval("A", 1*time.Second)
	chB := printAtInterval("B", 3*time.Second)

	for i := 0; i < 10; i++ {
		select {
		case msg := <-chA:
			fmt.Println(time.Now().Format("15:04:05"), msg)
		case msg := <-chB:
			fmt.Println(time.Now().Format("15:04:05"), msg)
		}
	}
}

func printAtInterval(msg string, period time.Duration) <-chan string {
	ch := make(chan string)
	go func() {
		for {
			time.Sleep(period)
			ch <- msg
		}
	}()
	return ch
}
