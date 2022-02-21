package main

import (
	"context"
	"fmt"
	"time"
)

/*
	需求: 基于现在计时，超时timeout 就快速取消任务并返回
	如果是基于现在计时，则建议优先使用WithTimeout
	因为WithTimeout 源码也是调用的WithDeadline
*/
const shortDuration = 1 * time.Millisecond // a reasonable duration to block in an example

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("finished with expired")
	case <-ctx.Done():
		fmt.Println("finished with timeout")
		// timeout 之后 会有err 说明
		// 这个err 和WithDeadline 的err一样
		// 因为WithTimeout底层调用的就是WithDeadline
		fmt.Println(ctx.Err())
	}
}

// finished with timeout
// context deadline exceeded
