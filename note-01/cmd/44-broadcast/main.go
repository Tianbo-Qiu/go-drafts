// Ch9 §7：Broadcast / Tee——一对多复制。
// 与 fan-out 的区别：
//   fan-out  把同一条 channel 交给 N 个 worker，**每条消息只去一个 worker**（负载均衡）。
//   broadcast 给每个下游一条 **独立** 的 channel，**每条消息抄送给所有下游**。
//
// 实现：起一个 dispatch goroutine，按源消息顺序把同一份值依次写到每个 output channel。
// 注意：dispatch 是 **逐条写每个下游**——任何一个下游卡住，整条 broadcast 都卡。
// 工程上要么给每个 output 加 buffer，要么消费侧用 goroutine 异步处理。
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

// Broadcast 把 input 复制到 n 条独立的 output channel；input 关 / quit 都会让所有 output 一并关闭。
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
	// 收窄方向供下游使用
	ros := make([]<-chan K, n)
	for i, c := range outs {
		ros[i] = c
	}
	return ros
}

// 聚合型下游 1：取最长的 10 个不重复词。
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

// 聚合型下游 2：取出现频率最高的 10 个词。
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
	quit := make(chan struct{})
	defer close(quit)

	urls := generateUrls(quit)
	pages := make([]<-chan string, downloaders)
	for i := 0; i < downloaders; i++ {
		pages[i] = downloadPages(quit, urls)
	}
	words := extractWords(quit, FanIn(quit, pages...))

	streams := Broadcast(quit, words, 2) // 一份输入复制给两个下游
	longest := longestWords(quit, streams[0])
	frequent := frequentWords(quit, streams[1])

	fmt.Println("longest words: ", <-longest)
	fmt.Println("frequent words:", <-frequent)
}
