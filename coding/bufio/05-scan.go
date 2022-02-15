package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	file, err := os.Create("scanner.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString("http://studygolang.com.\nIt is the home of gophers.\nIf you are studying golang, welcome you!")
	// 将文件 offset 设置到文件开头
	file.Seek(0, os.SEEK_SET)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

/*
output:
http://studygolang.com.
It is the home of gophers.
If you are studying golang, welcome you!
*/
