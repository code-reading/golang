package main

import (
	"fmt"
	"time"
)

type user struct {
	name string
	age  int8
}

var u = user{name: "Ankur", age: 20}
var g = &u

func modifyUser(pu *user) {
	fmt.Println("modifyUser Received Value", pu)
	pu.name = "Anand"
}

func printUser(u <-chan *user) {
	time.Sleep(2 * time.Second)
	fmt.Println("printUser goroutine called", <-u)
}

func main() {
	c := make(chan *user, 5)
	c <- g // 值拷贝
	fmt.Println(g)
	// modify g
	g = &user{name: "Ankur Anand", age: 100}
	go printUser(c) // 打印 channel 发现 里数据是 g 指向 的 &u 的拷贝， 并不是g的地址
	go modifyUser(g)
	time.Sleep(5 * time.Second)
	fmt.Println(g)
}

/*
&{Ankur 20}
modifyUser Received Value &{Ankur Anand 100}
printUser goroutine called &{Ankur 20}
&{Anand 100}
*/
