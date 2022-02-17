package main

import (
	"container/ring"
	"fmt"
)

func main() {
	// 初始化ring 长度为6
	// 注意 ring 长度初始化之后就不能修改了， 扩容可以通过Link 另外一个ring 来达到目的
	r := ring.New(6)

	n := r.Len() // 获取ring的长度

	//通过遍历初始化 ring
	for i := 0; i < n; i++ {
		r.Value = i
		r = r.Next() // 通过Next() 移动ring
	}
	// 打印ring 查看当前的引用指针位置
	r.Do(func(p interface{}) {
		fmt.Printf("%v ", p)
	})
	// 上面初始化的长度位n , 然后ring的长度也是n
	// 所以初始化结束之后 当前r的引用指针位置指向0
	// output 0 1 2 3 4 5

	fmt.Println()
	// 删除元素 通过Unlink进行, 是从当前r 的下一个开始计算, 即从r.Next()开始
	// Unlink three elements from r, starting from r.Next()
	// 删除元素是从当前执行的下一个开始, 当前指针指向0 , r.Next() 是从1 开始 删除三个元素
	// 就是把 1,2,3 删除了
	r.Unlink(3)

	// Iterate through the remaining ring and print its contents
	r.Do(func(p interface{}) {
		fmt.Printf("%v ", p)
	})
	// output 0 4 5
}
