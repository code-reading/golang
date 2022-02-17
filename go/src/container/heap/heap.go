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
// 堆是实现优先级队列的常用方法
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
/*
根据上面interface的定义，这个堆结构继承自sort.Interface，
而sort.Interface，须要实现三个方法：Len()， Less() ， Swap()

此外还需实现堆接口定义的两个方法：Push(x interface{})   /  Pop() interface{}，
因此想使用heap定义一个堆， 只须要定义实现了这五个方法结构就能够了

任何实现了本接口的类型均可以用于构建最小堆。最小堆能够经过heap.Init创建，数
据是递增顺序或者空的话也是最小堆。
最小堆的约束条件
!h.Less(j, i) for 0 <= i < h.Len() and 2*i+1 <= j <= 2*i+2 and j < h.Len()
*/
type Interface interface {
	sort.Interface
	Push(x interface{}) // add x as element Len()
	Pop() interface{}   // remove and return element Len() - 1.
}

// Init establishes the heap invariants required by the other routines in this package.
// Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// The complexity is O(n) where n = h.Len().
// 时间复杂度 O(n)
// 构建一个最小堆
// 初始化一个堆。一个堆在使用任何堆操做以前应先初始化。
// Init函数对于堆的约束性是幂等的（屡次执行无心义），并可能在任什么时候候堆的约束性被破坏时被调用。
// 本函数复杂度为O(n)，其中n等于h.Len()。
func Init(h Interface) {
	// heapify
	n := h.Len()
	// i 从 n/2 -1 开始， 然后i-- 到 i>=0
	for i := n/2 - 1; i >= 0; i-- {
		down(h, i, n)
	}
}

// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
// 加一个数据到堆中， 从尾部加入， 然后向上调整 达到最小堆
func Push(h Interface, x interface{}) {
	h.Push(x)
	up(h, h.Len()-1)
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
// 返回顶部的最小元素， 并删除它
// 然后将最后一个元素挪到顶部，以保证为完全二叉树
// 然后向下比较左右子节点，调整为满足最小堆
// //删除并返回堆h中的最小元素（不影响约束性）。复杂度O(log(n))，其中n等于h.Len()。该函数等价于Remove(h, 0)。
func Pop(h Interface) interface{} {
	n := h.Len() - 1
	h.Swap(0, n)   // 交换了堆顶和堆尾元素
	down(h, 0, n)  // 调整最小堆
	return h.Pop() // 是的这里返回的数组最后一个元素其实就是堆顶元素
}

// Remove removes and returns the element at index i from the heap.
// The complexity is O(log n) where n = h.Len().
// 删除第i个元素 操作复杂度 O(log n)
func Remove(h Interface, i int) interface{} {
	n := h.Len() - 1 // 堆最后一个元素的下标
	if n != i {      //如果要移除的元素不是最后一个
		h.Swap(i, n)        // 交换移除的元素和最后一个元素
		if !down(h, i, n) { // 从交换之后的元素到堆最后进行堆化调整，如果不需要调整
			up(h, i) // 则向上调整
			// 这里要注意， 如果down 向下调整了， 说明这个i元素一定大于其左右子节点元素
			// 那么也大于其父节点元素， 因为是最小堆，所以只需要向下down即可；
			// 如果i 被它左右节点都小，那么不需要down, 有可能比它的父节点也小，所以需要up调整
		}
	}
	// pop最后元素, 堆顶元素
	return h.Pop()
}

// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling Remove(h, i) followed by a Push of the new value.
// The complexity is O(log n) where n = h.Len().
// Fix 表示在修改第i节点值后 调整最小堆
// 在修改第i个元素后，调用本函数修复堆，比删除第i个元素后插入新元素更有效率。复杂度O(log(n))，其中n等于h.Len()。
func Fix(h Interface, i int) {
	if !down(h, i, h.Len()) {
		up(h, i)
	}
}

func up(h Interface, j int) {
	for {
		i := (j - 1) / 2 // parent
		// 如果父节点与插入节点相同，表示时第一个节点
		// 因为golang 是向0取整  所以 (0-1)/2 = 0
		// h.Less(j,i) 表示插入节点比父节点小
		// !h.Less(j, i) 表示父节点比插入节点小， 那么符合最小堆，则不需要调整
		if i == j || !h.Less(j, i) {
			break
		}
		// 插入值比父节点值小， 则需要交换，是的父节点值最小 才满足最小堆
		h.Swap(i, j)
		// 更新插入值待插入的索引位置， 为其当前的父节点位置
		j = i
	}
}

// 构建最小堆时， 待插入位置是堆顶，所以是与其左右子节点比较, 即向下调整堆 down
func down(h Interface, i0, n int) bool {
	i := i0 // 父节点
	for {
		j1 := 2*i + 1 // 左节点
		// 左节点 >= 堆大小，表示当前节点已经是叶子节点了
		// j1 考虑到溢出异常 < 0
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		// j2 = j1 + 1 右节点
		// j2 < n 表示右节点存在
		// h.Less(j2, j1) 为true 表示 j2 < j1
		// 取最小值 用于构建最小堆
		if j2 := j1 + 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		// h.Less(j, i) 表示子节点值小于父节点值 j < i
		// !h.Less(j, i) 表示父节点值小于子节点值 i < j
		// 最小堆构建时， 如果父节点值小于子节点值 则直接退出
		if !h.Less(j, i) {
			break
		}
		// 否则交换节点值
		h.Swap(i, j)
		// 更换新的父节点索引位置， 继续与其左右子节点比较
		i = j
	}
	// i > i0 表示 发生了down 堆化交换
	return i > i0
}
