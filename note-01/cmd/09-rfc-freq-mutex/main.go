// 多 goroutine 拉 RFC 文本并累加字母频次：锁只包住对共享 freq 的写循环，不把 HTTP/ReadAll 放在锁里。
// 需能访问 rfc-editor.org。
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	freq := make([]int, 26)
	client := &http.Client{Timeout: 15 * time.Second}

	for i := 1000; i <= 1008; i++ {
		url := fmt.Sprintf("https://www.rfc-editor.org/rfc/rfc%d.txt", i)
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if err := countLetters(client, u, freq, &mu); err != nil {
				fmt.Println("skip:", u, err)
			}
		}(url)
	}
	wg.Wait()

	mu.Lock()
	for i, c := range alphabet {
		if freq[i] > 0 {
			fmt.Printf("'%c': %d\n", c, freq[i])
		}
	}
	mu.Unlock()
}

func countLetters(client *http.Client, url string, freq []int, mu *sync.Mutex) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	mu.Lock()
	for _, b := range body {
		ch := strings.ToLower(string(b))
		idx := strings.Index(alphabet, ch)
		if idx != -1 {
			freq[idx]++
		}
	}
	mu.Unlock()
	fmt.Println("completed:", url)
	return nil
}
