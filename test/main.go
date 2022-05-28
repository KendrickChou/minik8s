package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println(1)
			default:
				fmt.Println(3)
				time.Sleep(time.Second * 3)
			}
		}
	}()
	time.Sleep(time.Second)
	cancel()
	fmt.Println(2)
	time.Sleep(time.Second * 3)
}
