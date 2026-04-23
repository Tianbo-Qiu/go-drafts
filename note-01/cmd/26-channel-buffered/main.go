// Ch7 §3：带缓冲 channel——`make(chan T, N)`。
// 只要 buffer 未满，发送方就不会阻塞；`len(ch)` 可以观察当前 buffer 里还有多少消息。
// 这里 buffer=3，主 goroutine 一口气塞入 1..6 与哨兵 -1；接收者每秒读一条，
// 输出可以看到发送方在 buffer 满时被压回、接收者拿走一条后再继续发送。
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	msgs := make(chan int, 3)
	var wg sync.WaitGroup
	wg.Add(1)
	go receiver(msgs, &wg)

	for i := 1; i <= 6; i++ {
		fmt.Printf("[%s] sending %d, buffer size before send: %d\n",
			time.Now().Format("15:04:05"), i, len(msgs))
		msgs <- i
	}
	msgs <- -1 // 哨兵：约定接收者看到 -1 就退出
	wg.Wait()
}

func receiver(msgs chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := 0; msg != -1; {
		time.Sleep(1 * time.Second)
		msg = <-msgs
		fmt.Printf("[%s] received: %d\n", time.Now().Format("15:04:05"), msg)
	}
}
