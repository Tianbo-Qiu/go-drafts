// Ch9 §2：quit channel 的最小模板——
// 一条专用 channel 只用来传「停止」信号；生产方在 select 里同时盯着「能发就发」与「<-quit」，
// 收到 quit（通常是被 close）就 return。close 一对多广播让所有监听者一次性退出（见 Ch7 §5.4）。
//
// 工程里多用 chan struct{}（零内存）甚至 context.Context.Done() 取代裸 chan int；
// 这里用 chan struct{} 把「只传信号、不传值」表达得更清楚。
package main

import (
	"fmt"
	"sync"
)

// 生产方：在循环里持续把累加值推到 numbers，直到 quit 被关。
func produceSums(numbers chan<- int, quit <-chan struct{}, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		next := 0
		for i := 1; ; i++ {
			next += i
			select {
			case numbers <- next:
			case <-quit:
				fmt.Println("producer: quit signaled, exiting")
				return
			}
		}
	}()
}

func main() {
	numbers := make(chan int)
	quit := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	produceSums(numbers, quit, &wg)

	for i := 0; i < 10; i++ {
		fmt.Println(<-numbers)
	}
	close(quit) // 广播停止：生产 goroutine 在下一轮 select 里命中 <-quit 后 return
	wg.Wait()
}
