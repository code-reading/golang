package main

import (
	"context"
	"fmt"
	"time"
)

/*
	超时取消任务并返回
	能基于某个时间点计时 超时timeout 就快速取消任务并返回
*/
const shortDuration = 1 * time.Millisecond // a reasonable duration to block in an example

func main() {
	// WithDeadline 是基于某个时间点开始计时 超时
	d := time.Now().Add(shortDuration)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	// 建议手动再次cancel
	defer cancel()
	select {
	case <-time.After(1 * time.Second):
		fmt.Println("finsihed with timeout")
	case <-ctx.Done():
		fmt.Println("finished with deadline expired")
		// deadline 会有错误说明
		fmt.Println(ctx.Err())
	}
}

//finished with deadline expired
// context deadline exceeded
