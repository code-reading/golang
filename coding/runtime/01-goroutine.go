package main

import (
	"fmt"
	"time"
)

func goroutineA(a <-chan int) {
	fmt.Println("G1 received data: ", <-a)
}

func goroutineB(b <-chan int) {
	fmt.Println("G2 received data: ", <-b)
}

func main() {
	ch := make(chan int)
	go goroutineA(ch)
	go goroutineB(ch)
	ch <- 3
	time.Sleep(time.Second)
}
