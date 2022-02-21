package main

import (
	"context"
	"fmt"
)

/*
	需求，构建一个数字生成器;
	调用方控制生成数字个数，结束时候，调用方告知生成器停止生成数字
*/

func main() {
	// 使用闭包构建一个生成器
	// 返回一个只读的int类型的channel
	gen := func(ctx context.Context) <-chan int {
		dst := make(chan int)
		n := 1
		go func() {
			for {
				select {
				case <-ctx.Done():
					return // 返回，防止goroutine泄露
				default:
					dst <- n
					n++
				}
			}
		}()
		return dst
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := range gen(ctx) {
		fmt.Println(i)
		if i == 5 {
			break
		}
	}
}

/*
output:
1
2
3
4
5
*/
