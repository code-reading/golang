package main

import (
	"fmt"
	"unsafe"
)

type Programmer struct {
	name     string
	language string
	age      int
}

type Modified struct {
	age      int
	name     string
	language string
}

func main() {
	m := Modified{10, "stefno", "go"}
	fmt.Println(m)
	fmt.Println(unsafe.Offsetof(m.age))      // 0 age在结构体user中的偏移量,也是结构体的地址
	fmt.Println(unsafe.Offsetof(m.name))     // 8
	fmt.Println(unsafe.Offsetof(m.language)) // 24

	p := Programmer{"stefno", "go", 10}
	fmt.Println(p)
	fmt.Println(unsafe.Offsetof(p.name))     // 0  name在结构体user中的偏移量,也是结构体的地址
	fmt.Println(unsafe.Offsetof(p.language)) // 16
	fmt.Println(unsafe.Offsetof(p.age))      // 32

	// 获取结构体的第一个字段地址
	name := (*string)(unsafe.Pointer(&p))
	*name = "update name " // 更新地址下的内容
	// 获取结构体第二个字段地址 这里需要用到偏移
	// 注意，这里先通过unsafe.Pointer(&p) 获取结构体第一个字段地址
	// 但是 Pointer 不能计算， 所以需要先转换为 uintptr
	// 再加上  unsafe.Offsetof(p.language) 指定字段的偏移
	// unsafe.Pointer 可以理解返回的是一个*void 空指针
	// golang 是强类型语言, 使用时必须确定类型
	// 那么就需要进行类型强转， 因为结构体字段是string
	// 所以强转指针 *string
	lang := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.language)))
	// 注意 lang 是指针
	// *lang  才是指针所指地址赋值
	*lang = "Golang"
	fmt.Println(p)
	//下面修改 age 的值
	// age的偏移， 是要计算上前面的字段的地址长度的
	// 第一种是继续使用 unsafe.Offsetof 算偏移, 这个是直接算当前字段离结构体起始位置的偏移，已经包含了
	// 前面字段的长度， 不需要额外添加了
	// 所以age的起始地址不是
	// uintptr(unsafe.Pointer(&p)) +unsafe.Offsetof(p.language) + unsafe.Offsetof(p.age)
	// 而是 unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.age)
	age := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.age)))
	*age = 18
	fmt.Println(p)
	// 先后端结构体的地址，
	// unsafe.Sizeof(p.name) + unsafe.Sizeof(language) 得到name 和 languange 占用的字节数
	// 结构体地址+ 占用的字节数就得到了 p.age 的地址， 转成int 指针
	pAge := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Sizeof(p.language) + unsafe.Sizeof(p.name)))
	*pAge = 20
	fmt.Println(p)
	// 第三种也是Sizeof 算大小, 但是传入的是相应字段类型的零值
	sage := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) +
		unsafe.Sizeof(string("")) + // 这里p.name 的占字节数
		unsafe.Sizeof(string("")))) // 这里是 p.language 的占字节数
	*sage = 30
	fmt.Println(p)
}

/*
output:
{10 stefno go}
0
8
24
{stefno go 10}
0
16
32
{update name  Golang 10}
{update name  Golang 18}
{update name  Golang 20}
{update name  Golang 30}
*/
