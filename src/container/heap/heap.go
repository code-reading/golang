// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package heap provides heap operations for any type that implements
// heap.Interface. A heap is a tree with the property that each node is the
// minimum-valued node in its subtree.
//
// The minimum element in the tree is the root, at index 0.
//
// A heap is a common way to implement a priority queue. To build a priority
// queue, implement the Heap interface with the (negative) priority as the
// ordering for the Less method, so Push adds items while Pop removes the
// highest-priority item from the queue. The Examples include such an
// implementation; the file example_pq_test.go has the complete source.
//
package heap

import "sort"

// The Interface type describes the requirements
// for a type using the routines in this package.
// Any type that implements it may be used as a
// min-heap with the following invariants (established after
// Init has been called or if the data is empty or sorted):
//
//	!h.Less(j, i) for 0 <= i < h.Len() and 2*i+1 <= j <= 2*i+2 and j < h.Len()
//
// Note that Push and Pop in this interface are for package heap's
// implementation to call. To add and remove things from the heap,
// use heap.Push and heap.Pop.
type Interface interface {
	sort.Interface
	Push(x interface{}) // add x as element Len()
	Pop() interface{}   // remove and return element Len() - 1.
}

// Init establishes the heap invariants required by the other routines in this package.
// Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) where n = h.Len().
func Init(h Interface) {
	// heapify
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		down(h, i, n)
	}
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
//
func Push(h Interface, x interface{}) {
	h.Push(x)
	up(h, h.Len()-1)
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// It is equivalent to Remove(h, 0).
//
func Pop(h Interface) interface{} {
	n := h.Len() - 1
	h.Swap(0, n)
	down(h, 0, n)
	return h.Pop()
}

// Remove removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
//
func Remove(h Interface, i int) interface{} {
	n := h.Len() - 1
	// 如果要弹出的元素不是最顶部的元素
	if n != i {
		// 先将要弹出的元素置换到顶部
		h.Swap(i, n)
		// 然后从i位置 向下调整, 如果没有调整
		// 向下调整是因为当前i值一定大于它的左右节点, 所以要向下调整
		// 那么调整之后则不需要向上调整了
		if !down(h, i, n) {
			// 如果向下没有调整, 则表明插入或修改的值可能比上面的父节点要小
			// 所以要向上调整一下
			up(h, i)
		}
	}
	return h.Pop()
}

// 当i 位置的值修改之后， 重新调整 使得当前堆继续符合最小堆
// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling Remove(h, i) followed by a Push of the new value.
// The complexity is O(log(n)) where n = h.Len().
func Fix(h Interface, i int) {
	// 如果向下没有调整
	if !down(h, i, h.Len()) {
		// 则从当前位置向上调整
		up(h, i)
	}
}

func up(h Interface, j int) {
	for {
		// i 为头节点
		i := (j - 1) / 2 // parent
		// 如果头节点等于当前节点 或者 j > i(就是说当前节点比它的头节点大,已经符合最小堆) 则退出循环
		if i == j || !h.Less(j, i) {
			break
		}
		// 如果当前节点比父节点小，则交换位置
		h.Swap(i, j)
		// 继续调整父节点
		j = i
	}
}

func down(h Interface, i0, n int) bool {
	i := i0
	for {
		// i 节点的左节点
		j1 := 2*i + 1
		// 如果 左节点大于等于最后一个节点 或者 数值已经溢出 则跳出调整的循环
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		// j 为 左节点
		j := j1 // left child
		// j2 为右节点, 如果右节点下于最后一个节点 并且
		// 右节点值比 左节点小 则 将 右节点值赋值给 j
		// j 为左右节点中最小值的那个
		if j2 := j1 + 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		// 比较左右节点中最小那个和当前节点i的大小
		// 如果 当前节点值小于 j这个最新值 则直接退出循环不需要调整了
		// 如果大于 则 !h.Less(j,i ) = false
		if !h.Less(j, i) {
			break
		}
		// 交换 i 和j的值
		h.Swap(i, j)
		// 然后继续调整 j 这个最小堆
		i = j
	}
	// 如果上面的for调整了堆, 则i 一定要大于 i0
	// 因为调整一次 i 的值就会换成  它的左节点(2i+1) 或者 右节点(2i+2)
	// 所以 i > i0 就表示一定进行了调整
	return i > i0
}
