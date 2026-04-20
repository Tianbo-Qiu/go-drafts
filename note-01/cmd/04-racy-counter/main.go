// 故意制造数据竞争：用于演示 go run -race。
// 期望输出常小于 5000；无 -race 时也可能「碰巧」对，不能当正确。
package main

import (
	"fmt"
	"sync"
)

func main() {
	const n = 5000
	var wg sync.WaitGroup
	var counter int
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			counter++ // 读改写，非原子；多 goroutine 并发 → data race
		}()
	}
	wg.Wait()
	fmt.Println("final (undefined under Go memory model if raced):", counter)
}
