package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	s := strings.NewReader("ABC\nDEF\r\nGHI\nJKL")
	bs := bufio.NewScanner(s)
	for bs.Scan() {
		fmt.Printf("%s %v\n", bs.Bytes(), bs.Text())
	}
}

/*
output
ABC ABC
DEF DEF
GHI GHI
JKL JKL
*/
