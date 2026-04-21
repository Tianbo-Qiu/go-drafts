// sync.Cond：在 mutex 保护下等待「余额足够」再消费；配合 Signal 唤醒。
// 迭代次数已缩小，便于本地跑通。
package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	money := 100
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)
	go stingy(&money, cond)
	go spendy(&money, cond)
	time.Sleep(2 * time.Second)
	mu.Lock()
	fmt.Println("money in bank account:", money)
	mu.Unlock()
}

func stingy(money *int, cond *sync.Cond) {
	for range 100_000 {
		cond.L.Lock()
		*money += 10
		cond.Signal()
		cond.L.Unlock()
	}
	fmt.Println("stingy done")
}

func spendy(money *int, cond *sync.Cond) {
	for range 20_000 {
		cond.L.Lock()
		for *money < 50 {
			cond.Wait()
		}
		*money -= 50
		if *money < 0 {
			fmt.Println("money is negative!")
			os.Exit(1)
		}
		cond.L.Unlock()
	}
	fmt.Println("spendy done")
}
