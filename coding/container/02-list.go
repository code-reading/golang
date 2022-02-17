package main

import (
	"container/list"
	"fmt"
)

func main() {
	var l = list.New()    // 初始化一个链表
	e0 := l.PushBack(10)  // 10
	e1 := l.PushFront(11) // 11 10
	e2 := l.PushBack(7)   // 11 10 7

	l.InsertBefore(3, e0)  // 11 3 10 7
	l.InsertAfter(196, e1) // 11 196 3 10 7
	l.InsertAfter(129, e2) // 11 196 3 10 7 129
	l.MoveToBack(e1)       // 196 3 10 7 129 11
	l.MoveToFront(e2)      // 7 196 3 10 129 11

	// 从链表前出元素, 停止条件是 元素e 不为空, 迭代时 e = e.Next()
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Printf("%v ", e.Value)
	}
}

// output 7 196 3 10 129 11
