// Ch8 §9 + §10：fan-in（多源合并）+ 「关闭后置 nil 屏蔽 case」。
// 两个数据源各自结束时会 close 自己的 channel；如果不处理，select 会一直命中
// 已关 channel 上的 receive（永远 ready，读到零值），变成「忙转」。
// 标准修法：检测到 `ok=false` 就把对应变量置 nil——nil channel 的 case 永远不就绪，
// 相当于把这条路从 select 里摘掉。两条路都 nil → 外层 for 退出。
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func generateAmounts(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < n; i++ {
			out <- rand.Intn(100) + 1
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return out
}

func main() {
	sales := generateAmounts(50)
	expenses := generateAmounts(40)
	pnl := 0
	step := 0

	// 用 || 而不是 &&：只要还有任一条非 nil，就继续 select，避免提前退出丢数据。
	for sales != nil || expenses != nil {
		select {
		case sale, ok := <-sales:
			if !ok {
				sales = nil // 该路已关，置 nil 屏蔽 case
				continue
			}
			fmt.Println(step, "sale:", sale)
			pnl += sale
			step++
		case exp, ok := <-expenses:
			if !ok {
				expenses = nil
				continue
			}
			fmt.Println(step, "expense:", exp)
			pnl -= exp
			step++
		}
	}
	fmt.Println("end of day P&L:", pnl)
}
