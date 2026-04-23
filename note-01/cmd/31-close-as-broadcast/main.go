// Ch7 §5.4：close 天然是「一对多广播」——
// 向 channel 发送一条值只能唤醒一个接收者；但 close 后，**所有** 正在 `<-ch` 的接收者
// 都会同时被解除阻塞（读到零值 + ok=false）。这也是 context.Context 取消信号的底层形态：
// 取消 == close 一个内部的 `chan struct{}`，所有监听者一次性收到通知。
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	done := make(chan struct{}) // 常用做法：chan struct{} 不占额外内存，只传信号
	var wg sync.WaitGroup

	for i := range 4 {
		wg.Add(1)
		go worker(i, done, &wg)
	}

	time.Sleep(1 * time.Second)
	fmt.Println(time.Now().Format("15:04:05"), "main: broadcasting cancel via close(done)")
	close(done) // 一次 close → 所有 worker 同时解除阻塞

	wg.Wait()
	fmt.Println(time.Now().Format("15:04:05"), "main: all workers stopped")
}

func worker(id int, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println(time.Now().Format("15:04:05"), "worker", id, "started, waiting for cancel")
	<-done // 阻塞直到 done 被关闭
	fmt.Println(time.Now().Format("15:04:05"), "worker", id, "received cancel, exiting")
}
