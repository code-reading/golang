// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This example demonstrates a priority queue built using the heap interface.
// 通过堆接口 实现优先级队列的示例
package heap_test

import (
	"container/heap"
	"fmt"
)

// An Item is something we manage in a priority queue.
// 优先级队列中的数据项Item
type Item struct {
	value    string // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap. // Item 在heap中的索引
}

// A PriorityQueue implements heap.Interface and holds Items.
// 优先级队列
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	// 需要返回最大的优先级， 所以这里使用>
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i // 保存数组索引
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n // 保存数组索引
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak // 手动赋值为Nil, 避免内存泄露
	item.index = -1 // for safety //设置索引为-1
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
// 修改优先级队列项的优先级和值,并且调整最小堆（因为返回优先级高的项， 应该算是最大堆)
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func Example_priorityQueue() {
	// Some items and their priorities.
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pq[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq) // 队列构建堆化

	// Insert a new item and then modify its priority.
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&pq, item)           // 插入一个新的堆
	pq.update(item, item.value, 5) // 修改这个item的优先级为5

	// Take the items out; they arrive in decreasing priority order.
	// 降序的优先级次序
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("%.2d:%s ", item.priority, item.value)
	}
	// Output:
	// 05:orange 04:pear 03:banana 02:apple
}
