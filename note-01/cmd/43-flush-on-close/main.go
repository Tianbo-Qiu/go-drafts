// Ch9 §6：聚合型 stage——「先攒齐、再一次性输出」。
// 普通 stream stage 是「来一个处理一个」；longestWords 不一样——它要看到 **全部** 输入才能
// 决定哪些是最长的。所以模板从「for-select 边读边发」变成：
//   1) for-select 把 input drain 到本地状态；
//   2) input 关闭后，做一次性聚合（这里是排序 + 取前 10）；
//   3) 输出最终结果再 close 自己的 output。
//
// 注意：仍然要在循环里听 quit——否则中途取消时，stage 会被 input 永远阻塞住，且无法响应取消。
package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
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

// longestWords 是聚合 stage：drain 完后排序，取最长 10 个一次发出。
func longestWords(quit <-chan struct{}, words <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		seen := make(map[string]bool)
		uniq := make([]string, 0)
	collect:
		for {
			select {
			case w, ok := <-words:
				if !ok {
					break collect // input 关闭，跳到聚合阶段
				}
				if !seen[w] {
					seen[w] = true
					uniq = append(uniq, w)
				}
			case <-quit:
				return // 中途取消：不输出聚合结果
			}
		}
		sort.Slice(uniq, func(a, b int) bool { return len(uniq[a]) > len(uniq[b]) })
		k := 10
		if len(uniq) < k {
			k = len(uniq)
		}
		select {
		case out <- strings.Join(uniq[:k], ", "):
		case <-quit:
		}
	}()
	return out
}

func main() {
	quit := make(chan struct{})
	defer close(quit)

	urls := generateUrls(quit)
	pages := make([]<-chan string, downloaders)
	for i := 0; i < downloaders; i++ {
		pages[i] = downloadPages(quit, urls)
	}

	results := longestWords(quit, extractWords(quit, FanIn(quit, pages...)))
	fmt.Println("longest words:", <-results)
}
