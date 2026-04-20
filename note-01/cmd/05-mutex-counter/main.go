// 与 04 对照：用 mutex 保护共享计数器，消除数据竞争。
package main

import (
	"fmt"
	"sync"
)

func main() {
	const n = 5000
	var wg sync.WaitGroup
	var mu sync.Mutex
	var counter int
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("final:", counter)
}
