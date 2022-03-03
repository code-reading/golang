package main

func main() {
	var b chan int
	close(b) // panic: close of nil channel
}
