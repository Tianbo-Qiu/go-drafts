// 共享内存 + Sleep 轮询：仅作「看见并发」的演示，生产应使用 channel/WaitGroup 等。
package main

import (
	"fmt"
	"time"
)

func main() {
	count := 5
	go countdown(&count)
	for count > 0 {
		time.Sleep(500 * time.Millisecond)
		fmt.Println(count)
	}
}

func countdown(count *int) {
	for *count > 0 {
		time.Sleep(1 * time.Second)
		*count--
	}
}
