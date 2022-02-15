package main

import (
	"bufio"
	"bytes"
	"fmt"
)

func main() {
	b := bytes.NewBuffer(make([]byte, 0))
	bw := bufio.NewWriter(b)
	// s := strings.NewReader("123")
	br := bufio.NewReader(b) // 将buffer 同时挂载到Writer 和Reader 上 通过NewReadWriter联通
	rw := bufio.NewReadWriter(br, bw)
	// p, isPrefix, _ := rw.ReadLine() 建议用ReadBytes("\n") 或者 ReadString("\n")代替
	// fmt.Printf("line:%s, isPrefix:%v\n", p, isPrefix) //123
	p, _ := rw.ReadString('\n')
	fmt.Println(p)
	rw.WriteString("asdf")
	rw.Flush()
	fmt.Println(b)
	pp, _ := rw.ReadBytes('\n')
	fmt.Println(string(pp)) // output asdf
}
