// Ch8 §7：select 也能放 send case——双向多路复用。
// 主 goroutine 在每轮里「要么把一个新随机数推进 numbers，要么从 primes 里取一个素数」，
// 哪条先就绪走哪条；这样上游不会把 buffer 撑爆，下游也能及时被消费。
//
// 注意：select 进入时 **所有** case 的 channel 与发送侧表达式都会被 **求值一次**——
// `rand.Intn(...)` 即使本轮没有命中 send case 也会被调用，只是值被丢弃。
package main

import (
	"fmt"
	"math"
	"math/rand"
)

func primesOnly(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if isPrime(n) {
				out <- n
			}
		}
	}()
	return out
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	numbers := make(chan int)
	primes := primesOnly(numbers)

	for found := 0; found < 10; {
		select {
		case numbers <- rand.Intn(1_000_000_000) + 1:
			// 推一个新候选数到上游
		case p := <-primes:
			fmt.Println("found prime:", p)
			found++
		}
	}
}
