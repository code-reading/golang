package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func main() {
	txt := "reflect"
	txtBytes := string2bytes(txt)
	ctxt := bytes2string(txtBytes)
	fmt.Println("raw:", txt, " txtBytes:", string(txtBytes), " ctxt:", ctxt)
}

// output raw: reflect  txtBytes: reflect  ctxt: reflect

func string2bytes(s string) []byte {
	stringHeander := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: stringHeander.Data,
		Len:  stringHeander.Len,
		Cap:  stringHeander.Len,
	}
	// 将 bh地址转换成unsafe.Pointer
	// 再将unsafe.Pointer 转换成*[]byte指针
	// 在取指针下面的内容 *(*[]byte)()
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func bytes2string(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&sh))
}
