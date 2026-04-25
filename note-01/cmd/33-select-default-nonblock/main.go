// Ch8 §3：select + default = 非阻塞接收。
// 没有任何 case 就绪时走 default，可用来做轮询 / 心跳；
// 注意 `break` 只跳出 select，不跳外层 for——本例用 `return`，常见替代写法见底部注释。
package main

import (
	"fmt"
	"time"
)

func main() {
	msgs := sendAfter("ok", 3*time.Second)
	for {
		select {
		case msg := <-msgs:
			fmt.Println("received:", msg)
			return // 收到消息后整体退出；不能写 break，否则只跳 select
		default:
			fmt.Println("no message yet")
			time.Sleep(1 * time.Second)
		}
	}
}

func sendAfter(msg string, after time.Duration) <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(after)
		ch <- msg
	}()
	return ch
}

// 退出外层 for 的另外两种等价写法（任选其一即可）：
//
//   1) done 标志：
//       done := false
//       for !done {
//           select {
//           case msg := <-msgs:
//               fmt.Println("received:", msg)
//               done = true
//           default:
//               ...
//           }
//       }
//
//   2) labeled break：
//       loop:
//       for {
//           select {
//           case msg := <-msgs:
//               fmt.Println("received:", msg)
//               break loop // 关键字 break 后跟标签，跳出标签所在的 for
//           default:
//               ...
//           }
//       }
//       fmt.Println("done")
