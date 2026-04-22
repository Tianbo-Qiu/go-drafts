// Ch6 §3：用可复用 Barrier 做多轮矩阵乘法——
// 主 goroutine 装载输入 → 第一道 Barrier 放行工人算各行 → 第二道 Barrier 等所有行算完 → 打印并进入下一轮。
package main

import (
	"fmt"
	"math/rand"
	"sync"
)

const n = 10

type Barrier struct {
	size       int
	waitCount  int
	generation int
	cond       *sync.Cond
}

func NewBarrier(size int) *Barrier {
	return &Barrier{
		size: size,
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (b *Barrier) Wait() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	myGen := b.generation
	b.waitCount++

	if b.waitCount == b.size {
		b.waitCount = 0
		b.generation++
		b.cond.Broadcast()
		return
	}
	for myGen == b.generation {
		b.cond.Wait()
	}
}

func genMatrix(m *[n][n]int) {
	for row := 0; row < n; row++ {
		for col := 0; col < n; col++ {
			m[row][col] = rand.Intn(10) - 5
		}
	}
}

func rowMultiply(a, b, result *[n][n]int, row int, barrier *Barrier) {
	for {
		barrier.Wait() // 等主 goroutine 装好矩阵
		for col := 0; col < n; col++ {
			sum := 0
			for i := 0; i < n; i++ {
				sum += a[row][i] * b[i][col]
			}
			result[row][col] = sum
		}
		barrier.Wait() // 本行算完；等其它行一起结束
	}
}

func main() {
	var matrixA, matrixB, result [n][n]int
	barrier := NewBarrier(n + 1) // n 个工人 + 主 goroutine

	for row := 0; row < n; row++ {
		go rowMultiply(&matrixA, &matrixB, &result, row, barrier)
	}

	for i := range 4 {
		genMatrix(&matrixA)
		genMatrix(&matrixB)

		barrier.Wait() // 放行：工人开始并行计算
		barrier.Wait() // 汇合：所有行算完才打印

		fmt.Println("========", i, "========")
		fmt.Println(matrixA)
		fmt.Println("x")
		fmt.Println(matrixB)
		fmt.Println("=")
		fmt.Println(result)
	}
}
