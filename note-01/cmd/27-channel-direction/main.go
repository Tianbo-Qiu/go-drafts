// Ch7 §4：channel 方向。
// 函数形参里用 `chan<- T`（只写）和 `<-chan T`（只读）声明方向，
// 编译期就能挡住「本该只读却误写」「本该只写却误读」这类错误。
// 底层仍是同一个双向 channel，只是在函数边界上把能力收窄。
package main

import (
	"fmt"
	"time"
)

func main() {
	msgs := make(chan int)
	go receiver(msgs)
	go sender(msgs)
	time.Sleep(5 * time.Second)
}

func receiver(msgs <-chan int) { // 只读 channel
	for {
		msg := <-msgs
		fmt.Println(time.Now().Format("15:04:05"), "received:", msg)
	}
}

func sender(msgs chan<- int) { // 只写 channel
	for i := 1; ; i++ {
		fmt.Println(time.Now().Format("15:04:05"), "sending:", i)
		msgs <- i
		time.Sleep(1 * time.Second)
	}
}
