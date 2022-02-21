package main

import (
	"context"
	"fmt"
)

/*
	需求: 在方法之间，goroutine之间，传递通用的token, trace_id 等通用参数
	建议将ctx 放在函数的第一位, 并且 函数参数不要放到ctx中
	也不要将ctx放到函数的结构体参数中
*/

// alias
type favContextKey string

func main() {
	// 闭包
	f := func(ctx context.Context, key favContextKey) {
		// ctx.Value 取值 只返回一个值 没有返回两个 v,ok
		if v := ctx.Value(key); v != nil {
			fmt.Println("found the key: ", v)
			return
		}
		fmt.Println("found not the key: ", key)
	}
	k := favContextKey("color")
	// context.WithValue 只返回一个ctx 没有cancel， 不需要cancel
	ctx := context.WithValue(context.Background(), k, "red")
	f(ctx, k)
	f(ctx, favContextKey("language"))
}

/*
found the key:  red
found not the key:  language
*/
