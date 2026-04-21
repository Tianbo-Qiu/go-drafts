// Broadcast：最后一个到达的 goroutine 唤醒所有仍在 Wait 的同伴。
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	cond := sync.NewCond(&sync.Mutex{})
	playersRemaining := 4
	for playerID := range playersRemaining {
		go playerHandler(cond, &playersRemaining, playerID)
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(2 * time.Second)
}

func playerHandler(cond *sync.Cond, playersRemaining *int, playerID int) {
	cond.L.Lock()
	fmt.Println("player", playerID, "connected")
	*playersRemaining--
	if *playersRemaining == 0 {
		cond.Broadcast()
	}
	for *playersRemaining > 0 {
		fmt.Println("player", playerID, "waiting for more players...")
		cond.Wait()
	}
	cond.L.Unlock()
	fmt.Println("all players connected. player", playerID, "ready")
}
