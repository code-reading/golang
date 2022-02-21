package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	if _, err := os.Open("non-existing"); err != nil {
		// errors.Is(err, xxx), 这里 最终调用了err.Is() 函数完成逻辑判断
		// 这里说明， 可以在给家的错误 定义Is 用来完成 业务错误判断
		if errors.Is(err, fs.ErrNotExist) { // 判断错误是不是fs.ErrNotExist
			fmt.Println("file does not exist")
		} else {
			fmt.Println(err)
		}
	}
}
