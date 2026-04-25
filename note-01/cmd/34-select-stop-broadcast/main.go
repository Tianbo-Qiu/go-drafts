// Ch8 §5：close 作扇出取消信号。
// 把 keyspace 切成多段，每段一个 worker goroutine 暴力试密码；任何一个 worker 找到答案后
// `close(stop)`，所有别的 worker 在下一轮 select 里命中 `case <-stop:` 立即退出。
// 这是 Ch7 §5.4 「close 一对多广播」的典型应用：与 worker 数量无关，一次 close 全员收到。
package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	password = "go far"
	alphabet = " abcdefghijklmnopqrstuvwxyz" // 27 个字符；下标 0 是空格
	keyspace = 387_420_488                   // 27^6
	chunk    = 1_000_000
)

func toBase27(n int) string {
	if n == 0 {
		return ""
	}
	s := ""
	for n > 0 {
		s = string(alphabet[n%27]) + s
		n /= 27
	}
	return s
}

// stop 故意用双向类型：发现答案的那个 worker 自己负责 close，配 sync.Once 保证只关一次。
func guess(from, upto int, stop chan struct{}, result chan<- string, once *sync.Once) {
	for n := from; n < upto; n++ {
		select {
		case <-stop:
			fmt.Printf("worker [%d, %d) stopped at %d\n", from, upto, n)
			return
		default:
			if toBase27(n) == password {
				result <- toBase27(n)
				once.Do(func() { close(stop) })
				return
			}
		}
	}
}

func main() {
	stop := make(chan struct{})
	result := make(chan string)
	var once sync.Once

	for i := 1; i < keyspace; i += chunk {
		end := i + chunk
		if end > keyspace {
			end = keyspace
		}
		go guess(i, end, stop, result, &once)
	}

	fmt.Println("password found:", <-result)
	time.Sleep(500 * time.Millisecond) // 留点时间让其它 worker 打印 stopped 信息
}
