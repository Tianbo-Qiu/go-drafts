// Ch6 §1：sync.WaitGroup 的基本用法——在启动 goroutine 前 Add，每个 goroutine 末尾 Done，主流程 Wait 到计数归零。
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(4)
	for i := range 4 {
		go doWork(i, &wg)
	}
	wg.Wait()
	fmt.Println("all completed")
}

func doWork(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	i := rand.Intn(5)
	time.Sleep(time.Duration(i*100) * time.Millisecond)
	fmt.Println(id, "done working after", i*100, "ms")
}
