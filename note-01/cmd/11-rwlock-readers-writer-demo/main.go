// RWMutex 演示：多个读协程可同时 RLock；写协程 Lock 时独占，需等读释放。
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var count int
	var rwLock sync.RWMutex
	var wg sync.WaitGroup

	// 启动 3 个读协程
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			rwLock.RLock()
			defer rwLock.RUnlock()

			fmt.Printf("读协程 %d 开始读取，count = %d\n", readerID, count)
			time.Sleep(1 * time.Second)
			fmt.Printf("读协程 %d 读取结束\n", readerID)
		}(i)
	}

	time.Sleep(200 * time.Millisecond)

	// 启动 1 个写协程
	wg.Add(1)
	go func() {
		defer wg.Done()

		rwLock.Lock()
		defer rwLock.Unlock()

		fmt.Println("=== 写协程开始写入 ===")
		count = 100
		time.Sleep(1 * time.Second)
		fmt.Println("=== 写协程写入完成 ===")
	}()

	// 再启动 2 个读协程（在写持锁时会被阻塞）
	for i := 4; i <= 5; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			rwLock.RLock()
			defer rwLock.RUnlock()

			fmt.Printf("读协程 %d 开始读取，count = %d\n", readerID, count)
			time.Sleep(1 * time.Second)
			fmt.Printf("读协程 %d 读取结束\n", readerID)
		}(i)
	}

	wg.Wait()
	fmt.Println("所有协程执行完毕")
}
