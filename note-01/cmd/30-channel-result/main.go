// Ch7 §6：用 channel 接收函数返回值——最小的「goroutine 结果回收」模式。
// 主 goroutine 先自己算一份，同时启动一个子 goroutine 把另一份结果写进 channel；
// 主流程需要时再 `<-resCh` 把结果取出来。相比「共享变量 + 锁」更贴近 CSP 直觉：
//   通过通信共享数据，而不是通过共享数据来通信。
package main

import "fmt"

func findFactors(n int) []int {
	result := make([]int, 0)
	for i := 1; i <= n; i++ {
		if n%i == 0 {
			result = append(result, i)
		}
	}
	return result
}

func main() {
	resCh := make(chan []int)
	go func() {
		resCh <- findFactors(12345)
	}()

	fmt.Println(findFactors(54321)) // 主 goroutine 自己算一份
	fmt.Println(<-resCh)            // 收上子 goroutine 的结果
}
