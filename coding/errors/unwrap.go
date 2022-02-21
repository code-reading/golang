package main

import (
	"errors"
	"fmt"
)

func main() {
	e := Unwrap{
		msg: "msg",
		err: errors.New("error for unwrap"),
	}
	fmt.Println(errors.Unwrap(e))
	// prefix:abc, error:error for unwrap
	// 可以看到这里返回的是Unwrap 自定义的Unwrap 函数的修改的错误值
	// 这种修改 可以通过errors.Unwrap 无缝调用
}

type Unwrap struct {
	msg string
	err error
}

func (e Unwrap) Error() string {
	return e.msg
}

// 可以自定义错误， 也可以就自定义错误结果Unwrap 返回原始错误err
func (e Unwrap) Unwrap() error {
	// return e.err
	return fmt.Errorf("prefix:abc, error:%v", e.err)
}
