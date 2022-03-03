package main

import "fmt"

func main() {
	c := make(chan int)
	go func() {
		c <- 1 // send to channel
	}()

	x := <-c       // recv from channel
	fmt.Println(x) // 1
}
