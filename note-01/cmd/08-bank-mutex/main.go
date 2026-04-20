// 与 07 对照：mutex 保护 balance，临界区仅包住读改写。
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	balance := 1000
	const iters = 50_000
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range iters {
			mu.Lock()
			balance += 10
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		for range iters {
			mu.Lock()
			balance -= 10
			mu.Unlock()
		}
	}()
	wg.Wait()
	mu.Lock()
	fmt.Println("final balance:", balance)
	mu.Unlock()
}
