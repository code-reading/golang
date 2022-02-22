package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func main() {
	var a int32
	var b = &a
	fmt.Println(reflect.TypeOf(unsafe.Sizeof(a))) // uintptr
	fmt.Println(unsafe.Sizeof(a))                 // 4
	fmt.Println(reflect.TypeOf(b).Kind())         // ptr
	fmt.Println(unsafe.Sizeof(b))                 // 8
}

/*
output
uintptr
4
ptr
8
*/
// 对于 a来说，它是int32类型，在内存中占4个字节，而对于b来说，是*int32类型，即底层为ptr指针类型，在64位机下占8字节。
