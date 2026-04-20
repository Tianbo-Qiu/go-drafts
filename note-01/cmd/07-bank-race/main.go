// 两 goroutine 同时读改写同一 balance → data race。建议：go run -race ./cmd/07-bank-race
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	balance := 1000
	const iters = 50_000
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range iters {
			balance += 10
		}
	}()
	go func() {
		defer wg.Done()
		for range iters {
			balance -= 10
		}
	}()
	wg.Wait()
	fmt.Println("final balance (undefined if raced):", balance)
}
