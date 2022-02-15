package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// A StringReader delivers its data one string segment at a time via Read.
type StringReader struct {
	data []string
	step int
}

// 在将reader数据读取到缓存buf时 会用到这个read
// 这个read step 默认为1 表示每次读取step个字符， 起始值为0
func (r *StringReader) Read(p []byte) (n int, err error) {
	if r.step < len(r.data) {
		s := r.data[r.step]
		n = copy(p, s)
		r.step++
	} else {
		err = io.EOF // 读取结束
	}
	return
}

func readRuneSegments(segments []string) {
	got := ""
	want := strings.Join(segments, "")
	r := bufio.NewReader(&StringReader{data: segments})
	for {
		r, _, err := r.ReadRune()
		if err != nil {
			if err != io.EOF {
				return
			}
			break
		}
		got += string(r)
	}
	if got != want {
		fmt.Errorf("segments=%v got=%s want=%s", segments, got, want)
	}
}

var segmentList = [][]string{
	{},
	{""},
	{"日", "本語"},
	{"\u65e5", "\u672c", "\u8a9e"},
	{"\U000065e5", "\U0000672c", "\U00008a9e"},
	{"\xe6", "\x97\xa5\xe6", "\x9c\xac\xe8\xaa\x9e"},
	{"Hello", ", ", "World", "!"},
	{"Hello", ", ", "", "World", "!"},
}

func main() {
	for _, s := range segmentList {
		readRuneSegments(s)
	}
}
