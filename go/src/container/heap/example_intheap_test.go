// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This example demonstrates an integer heap built using the heap interface.
package heap_test

import (
	"container/heap"
	"fmt"
)

// An IntHeap is a min-heap of ints.
// 初始化一个最小堆 示例， 需要实现 堆接口 Push 和 Pop
// 还需要实现继承的sort.Interface的三个接口 Len, Less, Swap
type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }      // 返回堆长度
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] } // 比较i 和j 的值， i < j 返回true
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	// Push 和Pop 使用指针接收， 因为它们会修改slices的长度，而不仅仅只是修改内容
	// append 表示从数组后面加入新元素
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
	// 先用一个临时变量old 指向数组 *h
	// 得到数组的长度 n
	// 得到数组最后一个元素 x
	// 修改 数组的内容， 将最后一个元素删除
	// 返回最后一个元素， 其实是堆顶元素， 因为heap.Pop() 在调用当前Pop时 先交换了堆顶和堆尾元素
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// This example inserts several ints into an IntHeap, checks the minimum,
// and removes them in order of priority.
func Example_intHeap() {
	h := &IntHeap{2, 1, 5}
	heap.Init(h)                         // 堆化， 构建最小堆
	heap.Push(h, 3)                      // 插入新值,这里是从尾部插入， 所以需要 up 向上比较父节点调整
	fmt.Printf("minimum: %d\n", (*h)[0]) // 得到最小值 在堆顶，也就是数组元素0
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h)) // 弹出堆顶元素
	}
	// Output:
	// minimum: 1
	// 1 2 3 5
}
