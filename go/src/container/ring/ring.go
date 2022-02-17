// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ring implements operations on circular lists.
// ring 实现了一个循环链表
package ring

// A Ring is an element of a circular list, or ring.
// Rings do not have a beginning or end; a pointer to any ring element
// serves as reference to the entire ring. Empty rings are represented
// as nil Ring pointers. The zero value for a Ring is a one-element
// ring with a nil Value.
// Ring 是循环链表的元素结构, 可以称之为环
// Rings 没有开始或结束位置
// 执行任何一个ring元素的指针,都可以作为整个环的引用
// Empty rings 表示一个空指针环
// Ring的零值 表示只有一个nil元素的ring
type Ring struct {
	next, prev *Ring // 前后环指针
	// 值，这个值在ring包中不会被处理
	Value interface{} // for use by client; untouched by this library
}

// 初始化一个ring 其前后指针都执行ring本身, 表示一个空ring
func (r *Ring) init() *Ring {
	r.next = r
	r.prev = r
	return r
}

// Next returns the next ring element. r must not be empty.
// 返回下一个ring 元素， r不能为Nil
// 如果没有下一个元素, 则返回一个初始化的ring
// r.next == nil 表示没有初始化，是一个空ring 此时
func (r *Ring) Next() *Ring {
	if r.next == nil {
		return r.init()
	}
	return r.next
}

// Prev returns the previous ring element. r must not be empty.
// 返回前一个ring元素
// 如果为空， 则直接初始化一个ring 返回
func (r *Ring) Prev() *Ring {
	if r.next == nil {
		return r.init()
	}
	return r.prev
}

// Move moves n % r.Len() elements backward (n < 0) or forward (n >= 0)
// in the ring and returns that ring element. r must not be empty.
// 指针从当前元素开始向后移动或者向前(n 可以为负数)
// n < 0 向前移动
// n > 0 向后移动
// n == 0 什么都不做
// r.next == nil 直接初始化这个ring 并返回
// 否则返回移动之后的r指针
func (r *Ring) Move(n int) *Ring {
	if r.next == nil {
		return r.init()
	}
	switch {
	case n < 0:
		for ; n < 0; n++ {
			r = r.prev
		}
	case n > 0:
		for ; n > 0; n-- {
			r = r.next
		}
	}
	return r
}

// New creates a ring of n elements.
// 用于创建一个新的 Ring, 接收一个整形参数，用于初始化 Ring 的长度
func New(n int) *Ring {
	if n <= 0 {
		return nil
	}
	r := new(Ring)
	p := r
	for i := 1; i < n; i++ {
		p.next = &Ring{prev: p}
		p = p.next
	}
	// 最后一个p 的后继执行根元素r
	// 根元素的上一个指针prev 指向最后一个元素
	p.next = r
	r.prev = p
	return r
}

// Link connects ring r with ring s such that r.Next()
// becomes s and returns the original value for r.Next().
// r must not be empty.
//
// If r and s point to the same ring, linking
// them removes the elements between r and s from the ring.
// The removed elements form a subring and the result is a
// reference to that subring (if no elements were removed,
// the result is still the original value for r.Next(),
// and not nil).
//
// If r and s point to different rings, linking
// them creates a single ring with the elements of s inserted
// after r. The result points to the element following the
// last element of s after insertion.
// 将两个 ring 连接到一起 (r 不能为空)
func (r *Ring) Link(s *Ring) *Ring {
	n := r.Next()
	if s != nil {
		p := s.Prev()
		// Note: Cannot use multiple assignment because
		// evaluation order of LHS is not specified.
		r.next = s
		s.prev = r
		n.prev = p
		p.next = n
	}
	return n
}

// Unlink removes n % r.Len() elements from the ring r, starting
// at r.Next(). If n % r.Len() == 0, r remains unchanged.
// The result is the removed subring. r must not be empty.
// 从当前元素开始，删除 n 个元素
// r.Move(n + 1) 表示在当前ring 位置向前移动n+1 位
//然后在和r.Link() 连接， 那么两个环已连接，就相当于把刚才走过的n个ring 删除了
func (r *Ring) Unlink(n int) *Ring {
	if n <= 0 {
		return nil
	}
	return r.Link(r.Move(n + 1))
}

// Len computes the number of elements in ring r.
// It executes in time proportional to the number of elements.
// 返回ring的长度
func (r *Ring) Len() int {
	n := 0
	// 只要r 不是nil 其长度就至少为1,
	// init()返回的ring 是r !=nil 所以初始化的r 长度位1 但是value 是nil
	if r != nil {
		n = 1
		for p := r.Next(); p != r; p = p.next {
			n++
		}
	}
	return n
}

// Do calls function f on each element of the ring, in forward order.
// The behavior of Do is undefined if f changes *r.
// Do 会依次将每个节点的 Value 当作参数调用这个函数 f,
// 实际上这是策略方法的引用，通过传递不同的函数以在同一个 ring 上实现多种不同的操作。
func (r *Ring) Do(f func(interface{})) {
	if r != nil {
		f(r.Value)
		for p := r.Next(); p != r; p = p.next {
			f(p.Value)
		}
	}
}
