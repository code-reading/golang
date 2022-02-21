package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// errors.As 可以非常方便抽取结构化中错误值部分

/*
比如，有一个业务错误信息,  path: xxxx, Os: drawin, error:path not exist
这个错误信息很难进行字段提取
可以通过自定义错误方法， 提出这个错误信息

// 自定义错误类型

type PathError struct {
	path string
	Os string
	err error
}
// 实现错误接口
func (e PathError)Error()string {
	return fmt.Sprintf("xxx, xx ,x ", e.path, e.os, e.err)
}

业务函数
func A() error {
	return PathError{path:xxx, os:xxx, err:errors.New("xxxx")}
}
if e = A() ; e ！=nil {
	var pe *PathError
	注意: 这里明确知道返回的e 的类型就是 PathError, 且pe 传递为地址传递 ，否则会panic
	当然 如果要进一步提取PathError, 比如脱敏等， 可以在PathError 在实现一个As(error, interface{})方法
	if errors.As(e, &pe) {
		// dosomething with pe
		fmt.Println("path:", pe.path)
	}
}
*/

func main() {
	if _, err := os.Open("non-existing"); err != nil {
		// os.Open 返回的错误类型 是  *fs.PathError
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			fmt.Println("Failed at path:", pathError.Path)
		} else {
			fmt.Println(err)
		}
	}

	// Output:
	// Failed at path: non-existing
}
