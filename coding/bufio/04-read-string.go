package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func main() {
	s := "a\nb\nc"
	reader := bufio.NewReader(strings.NewReader(s))
	for {
		line, err := reader.ReadString('\n')
		// if err != nil {
		// 	if err == io.EOF {
		// 		// 处理当数据末尾没有\n的情况
		// 		fmt.Printf("io.EOF, %#v\n", line)
		// 		break
		// 	}
		// 	panic(err)
		// }
		// fmt.Printf("%#v\n", line)
		if err == nil || err == io.EOF {
			line = strings.TrimSpace(line)
			if len(line) != 0 {
				fmt.Println(line)
			}
		}
	}
}

/*
末尾没有分割符，会丢失数据
"a\n"
"b\n"
io.EOF, "c"
*/
