package main

import (
	"container/ring"
	"fmt"
)

func main() {
	// 一个空Ring
	var rr ring.Ring
	fmt.Printf("rr: %+v\n", rr)
	// rr: {next:<nil> prev:<nil> Value:<nil>}

	var rrr ring.Ring = ring.Ring{}
	fmt.Printf("rrr: %+v\n", rrr)
	// rrr: {next:<nil> prev:<nil> Value:<nil>}

	// 可见空ring 其r.next == nil , 所以源码中多处比较r.next == nil 调用 init() 方法
	// 是用于判断ring为空 然后初始化这个ring并返回

	// 创建一个环, 包含三个元素
	r := ring.New(3)
	fmt.Printf("ring: %+v\n", *r)
	// ring: {next:0xc000134020 prev:0xc000134040 Value:<nil>}

	// 初始化
	for i := 1; i <= 3; i++ {
		r.Value = i
		r = r.Next()
	}
	fmt.Printf("init ring: %+v\n", *r)
	// init ring: {next:0xc000056040 prev:0xc000056060 Value:1}

	// sum
	s := 0
	r.Do(func(i interface{}) {
		fmt.Printf("%v ", i)
		s += i.(int)
	})
	fmt.Printf("sum ring: %d\n", s)
	// 1 2 3 sum ring: 6
}

// ring 方法的好处是 可以方便的通过ring.Do(func(interface{}){}) 将ring中的元素依次给func 方法调用
