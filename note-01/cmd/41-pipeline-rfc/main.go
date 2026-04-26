// Ch9 §3：3-stage Pipeline——generateUrls → downloadPages → extractWords。
// 每个 stage 形如 `func(quit, in) <-chan T`：内部 goroutine + `defer close(out)` + for-select。
// 上游 close 后下游 range 自然退出；`quit` 用来在中途广播取消，所有 stage 都通过同一信号统一退出。
//
// 这是 **顺序** 下载版（一个一个抓）。下一节 cmd/42 会把 downloadPages 扇出成 N 个 worker。
//
// 注意：需要外网访问 rfc-editor.org；URL 范围可改小做快测。
package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// generateUrls：产生 RFC 1100..1130 的链接。
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

// downloadPages：把 url 拉下来变成 page 文本；网络错误就跳过，不让一条坏 URL 拖死整条管道。
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

// extractWords：把 page 切成小写单词，逐个发出。
func extractWords(quit <-chan struct{}, pages <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		re := regexp.MustCompile(`[a-zA-Z]+`)
		for {
			select {
			case page, ok := <-pages:
				if !ok {
					return
				}
				for _, w := range re.FindAllString(page, -1) {
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

func main() {
	quit := make(chan struct{})
	defer close(quit)

	words := extractWords(quit, downloadPages(quit, generateUrls(quit)))

	count := 0
	for w := range words {
		count++
		if count <= 20 {
			fmt.Println(w)
		}
	}
	fmt.Println("total words:", count)
}
