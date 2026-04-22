// Ch6 §1：递归文件搜索——展示在已有 goroutine 内部动态 Add，再让 WaitGroup 等所有分叉完成。
// 用法：go run ./cmd/19-waitgroup-file-search <dir> <substring>
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: go run ./cmd/19-waitgroup-file-search <dir> <substring>")
		os.Exit(1)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go fileSearch(os.Args[1], os.Args[2], &wg)
	wg.Wait()
}

func fileSearch(dir string, pattern string, wg *sync.WaitGroup) {
	defer wg.Done()
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("read dir:", err)
		return
	}
	for _, file := range files {
		fpath := filepath.Join(dir, file.Name())
		if strings.Contains(file.Name(), pattern) {
			fmt.Println("found", fpath)
		}
		if file.IsDir() {
			wg.Add(1)
			go fileSearch(fpath, pattern, wg)
		}
	}
}
