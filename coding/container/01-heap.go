package main

import (
	"container/heap"
	"fmt"
)

// 借助container/heap 接口 实现最小堆

type IntHeap []int

// 实现五个接口

func (p IntHeap) Len() int {
	return len(p)
}

func (p IntHeap) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p IntHeap) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *IntHeap) Push(data interface{}) {
	(*p) = append(*p, data.(int))
}

func (p *IntHeap) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	*p = old[:n-1]
	return item
}

func main() {
	ih := &IntHeap{2, 1, 5}
	heap.Init(ih)
	heap.Push(ih, 3)
	fmt.Printf("minimum: %d\n", (*ih)[0])
	for ih.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(ih))
	}
}

/*
output
minimum: 1
1 2 3 5
*/
