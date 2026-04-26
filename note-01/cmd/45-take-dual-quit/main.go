// Ch9 §8：Take + 双 quit channel——「早停」的边界处理。
//
// 单 quit 的问题：Take 取够 N 条后想关掉 **上游**（停止下载、解析），但如果它复用同一条 quit，
// 那么 **下游** 的聚合 stage（longestWords / frequentWords）也会被同一信号打断、来不及输出最终结果。
//
// 解决：把信号一分为二
//   quitWords：上游产线开关——Take 取够数后由 Take 自己 close
//   quit     ：下游收尾开关——主流程 defer close，等聚合结果都拿到再触发
//
//   [generateUrls(quitWords)]
//          ↓
//   [downloadPages × N (quitWords)] → FanIn (quitWords)
//          ↓
//   [extractWords (quitWords)]
//          ↓
//   [Take(quitWords, N)]              ← 取够 N 后 close(quitWords)，上游全员退出
//          ↓
//   [Broadcast(quit)] → [longestWords(quit), frequentWords(quit)]
//                                       ← 用 quit，等主流程收完结果再 close
//
// 决策：把 stage 当 **生产方** 还是 **消费方** 看；上游链路用 quitWords，下游聚合 / 输出用 quit。
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

const (
	downloaders = 20
	takeN       = 10000
)

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

// Take 转发 input 中的前 n 条到 output；取够后 close(quit) 通知 **上游** 整体收手。
// 注意：这里的 quit 是上游产线的 quitWords，不是下游 sink 的 quit。
func Take[K any](quit chan struct{}, n int, input <-chan K) <-chan K {
	out := make(chan K)
	go func() {
		defer close(out)
		var once sync.Once // 防止 quit 被双 close
		for n > 0 {
			select {
			case msg, ok := <-input:
				if !ok {
					return
				}
				select {
				case out <- msg:
					n--
				case <-quit:
					return
				}
			case <-quit:
				return
			}
		}
		once.Do(func() { close(quit) }) // 取够了：通知上游停产
	}()
	return out
}

func Broadcast[K any](quit <-chan struct{}, input <-chan K, n int) []<-chan K {
	outs := make([]chan K, n)
	for i := range outs {
		outs[i] = make(chan K)
	}
	go func() {
		defer func() {
			for _, c := range outs {
				close(c)
			}
		}()
		for {
			select {
			case msg, ok := <-input:
				if !ok {
					return
				}
				for _, c := range outs {
					select {
					case c <- msg:
					case <-quit:
						return
					}
				}
			case <-quit:
				return
			}
		}
	}()
	ros := make([]<-chan K, n)
	for i, c := range outs {
		ros[i] = c
	}
	return ros
}

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
					break collect
				}
				if !seen[w] {
					seen[w] = true
					uniq = append(uniq, w)
				}
			case <-quit:
				return
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

func frequentWords(quit <-chan struct{}, words <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		freq := make(map[string]int)
		list := make([]string, 0)
	collect:
		for {
			select {
			case w, ok := <-words:
				if !ok {
					break collect
				}
				if freq[w] == 0 {
					list = append(list, w)
				}
				freq[w]++
			case <-quit:
				return
			}
		}
		sort.Slice(list, func(a, b int) bool { return freq[list[a]] > freq[list[b]] })
		k := 10
		if len(list) < k {
			k = len(list)
		}
		select {
		case out <- strings.Join(list[:k], ", "):
		case <-quit:
		}
	}()
	return out
}

func main() {
	quitWords := make(chan struct{}) // 上游产线开关：Take 取够后自己关
	quit := make(chan struct{})      // 下游收尾开关：等聚合结果都拿到才关
	defer close(quit)

	urls := generateUrls(quitWords)
	pages := make([]<-chan string, downloaders)
	for i := 0; i < downloaders; i++ {
		pages[i] = downloadPages(quitWords, urls)
	}

	words := Take(quitWords, takeN, extractWords(quitWords, FanIn(quitWords, pages...)))

	streams := Broadcast(quit, words, 2)
	longest := longestWords(quit, streams[0])
	frequent := frequentWords(quit, streams[1])

	fmt.Println("longest words: ", <-longest)
	fmt.Println("frequent words:", <-frequent)
}
