// Ch8 §11：扇出 + 收集——「不用锁」的并行汇总。
// 每个 URL 起一个 goroutine 数它自己的字母频率，结果各自走一条 channel；
// 主流程顺序 `<-` 收齐再相加。和 cmd/09-rfc-freq-mutex 对比：
//   - 那里用一把 mutex 保护共享的 freq map（共享内存 + 锁）；
//   - 这里每个 worker **独占** 自己的 frequency 切片，结果走 channel 传出（CSP）。
// 没有共享可变状态，自然没有 data race；外网失败时单条丢弃，不影响其它路径。
//
// 注意：示例需要外网访问；用 `go run` 时若本机没法访问 rfc-editor.org 会得到空结果。
package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const allLetters = "abcdefghijklmnopqrstuvwxyz"

// countLetters 返回一个 channel：成功时发出长度 26 的频率切片，失败时直接 close。
func countLetters(url string) <-chan []int {
	out := make(chan []int, 1) // buffer=1：worker 写完即可退出，避免泄漏
	go func() {
		defer close(out)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("get %s: %v\n", url, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Printf("get %s: status %s\n", url, resp.Status)
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("read %s: %v\n", url, err)
			return
		}

		freq := make([]int, 26)
		for _, b := range body {
			c := strings.ToLower(string(b))
			if i := strings.Index(allLetters, c); i >= 0 {
				freq[i]++
			}
		}
		fmt.Println("completed:", url)
		out <- freq
	}()
	return out
}

func main() {
	results := make([]<-chan []int, 0, 31)
	for i := 1000; i <= 1030; i++ {
		url := fmt.Sprintf("https://rfc-editor.org/rfc/rfc%d.txt", i)
		results = append(results, countLetters(url))
	}

	total := make([]int, 26)
	for _, ch := range results {
		freq, ok := <-ch
		if !ok {
			continue // 该路失败：channel 已关、无数据
		}
		for i, c := range freq {
			total[i] += c
		}
		_ = freq
	}

	for i, c := range allLetters {
		fmt.Printf("%c-%d\n", c, total[i])
	}
}
