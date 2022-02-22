package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Printf("bool align: %d\n", unsafe.Alignof(bool(true)))
	fmt.Printf("int32 align: %d\n", unsafe.Alignof(int32(0)))
	fmt.Printf("int8 align: %d\n", unsafe.Alignof(int8(0)))
	fmt.Printf("int64 align: %d\n", unsafe.Alignof(int64(0)))
	fmt.Printf("byte align: %d\n", unsafe.Alignof(byte(0)))
	fmt.Printf("string align: %d\n", unsafe.Alignof("EDDYCJY"))
	fmt.Printf("map align: %d\n", unsafe.Alignof(map[string]string{}))
}

/*
在 Go 中可以调用 unsafe.Alignof 来返回相应类型的对齐系数。
通过观察输出结果，可得知基本都是 2n，最大也不会超过 8。
这是因为我们的64位编译器默认对齐系数是 8，因此最大值不会超过这个数。

output
bool align: 1
int32 align: 4
int8 align: 1
int64 align: 8
byte align: 1
string align: 8
map align: 8
*/
