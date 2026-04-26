// Ch9 §4-5：Fan-out + Fan-in。
//   fan-out: 同一条 input channel，启动 N 个 worker goroutine 并发消费——天然「负载均衡」，
//            每条消息只会被某一个 worker 拿走（receive 是互斥的）。
//   fan-in : 把 N 条 output channel 合成一条；用 sync.WaitGroup 等所有上游结束后，单独一个
//            closer goroutine 关闭合并后的 channel——保证下游 `range` 能正常退出。
//
// 关键点：FanIn 的「合并」要用 select 兼听 `<-quit`，以免 quit 后还卡在 `output <- msg`。
package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const downloaders = 20

func generateUrls(quit <-chan struct{}) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for i := 1100; i <= 1130; i++ {
			url := fmt.Sprintf("https://rfc-editor.org/rfc/rfc%d.txt", i)
			select {
			case out <- url:
			case <-quit:
				return
			}
		}
	}()
	return out
}

func downloadPages(quit <-chan struct{}, urls <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case url, ok := <-urls:
				if !ok {
					return
				}
				body, err := fetch(url)
				if err != nil {
					fmt.Println("skip:", url, err)
					continue
				}
				select {
				case out <- body:
				case <-quit:
					return
				}
			case <-quit:
				return
			}
		}
	}()
	return out
}

func fetch(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func extractWords(quit <-chan struct{}, pages <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		re := regexp.MustCompile(`[a-zA-Z]+`)
		for {
			select {
			case pg, ok := <-pages:
				if !ok {
					return
				}
				for _, w := range re.FindAllString(pg, -1) {
					select {
					case out <- strings.ToLower(w):
					case <-quit:
						return
					}
				}
			case <-quit:
				return
			}
		}
	}()
	return out
}

// FanIn 把 N 条 input 合成一条 output：
//   - 每个 input 一条转发 goroutine；转发用 select 同时听 quit；
//   - 一个独立的 closer goroutine 等所有转发 goroutine wg.Done 之后再 close(output)，
//     保证下游 `range output` 能在所有 input 真正排空后退出。
func FanIn[K any](quit <-chan struct{}, inputs ...<-chan K) <-chan K {
	output := make(chan K)
	var wg sync.WaitGroup
	wg.Add(len(inputs))
	for _, in := range inputs {
		go func(ch <-chan K) {
			defer wg.Done()
			for msg := range ch {
				select {
				case output <- msg:
				case <-quit:
					return
				}
			}
		}(in)
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}

func main() {
	quit := make(chan struct{})
	defer close(quit)

	urls := generateUrls(quit)

	// fan-out：N 个 worker 并发消费同一条 urls；每个 url 只会被某一个 worker 拿走。
	pages := make([]<-chan string, downloaders)
	for i := 0; i < downloaders; i++ {
		pages[i] = downloadPages(quit, urls)
	}

	// fan-in：把 N 条 page 流合并到一条
	merged := FanIn(quit, pages...)
	words := extractWords(quit, merged)

	start := time.Now()
	count := 0
	for range words {
		count++
	}
	fmt.Printf("total words: %d, elapsed: %v\n", count, time.Since(start))
}
