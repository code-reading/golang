package main

import (
	"errors"
	"fmt"
)

// 对外的错误定义
var ErrC1 = errors.New("errors custom 1")
var ErrC2 = errors.New("errors custom 2")

// 定义错误
type errorNR string

// 内部业务错误定义
var ERR1 errorNR = "error 1"
var ERR2 errorNR = "error 2"

// 可以通过自定义错误Is 函数 来为外部提供错误比较对
// 而外部调用的是errors.Is() 方法
func (e errorNR) Is(target error) bool {
	switch target {
	case ErrC1:
		return e == ERR1
	case ErrC2:
		return e == ERR2
	default:
		return false
	}
}

// 实现error 接口
func (e errorNR) Error() string {
	return string(e)
}

func Error01() error {
	return ERR1
}
func Error02() error {
	return ERR2
}

func main() {
	e01 := Error01()
	if errors.Is(e01, ErrC1) {
		fmt.Println("e01:", e01)
	}
	e02 := Error02()
	if errors.Is(e02, ErrC2) {
		fmt.Println("e02:", e02)
	}
}

/*
output
e01: error 1
e02: error 2
*/
