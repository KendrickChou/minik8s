package main

import "fmt"

func main() {
	ch := make(chan []byte, 2)
	ch <- []byte("111")
	//ch <- []byte("222")
	by := <-ch
	fmt.Printf("%s", by)
}
