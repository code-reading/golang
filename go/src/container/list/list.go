// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package list implements a doubly linked list.
//
// To iterate over a list (where l is a *List):
//	for e := l.Front(); e != nil; e = e.Next() {
//		// do something with e.Value
//	}
//
// list 实现了一个双链表
/*
链表迭代
for e := l.Front(); e !=nil e = e.Next(){
	// do something with e.Value
}
*/
package list

// Element is an element of a linked list.
// 一个链表元素结构
type Element struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element // 上一个和下一个元素的指针

	// The list to which this element belongs.
	list *List // 元素所在的链表

	// The value stored with this element.
	Value interface{} // 元素值
}

// Next returns the next list element or nil.
// 返回该元素的下一个元素，如果没有下一个元素则返回 nil
// 怎么判断没有下一个元素
// 首先要判断当前链表存在， e.list !=nil
// 其次要判断取出的链表节点 不等于根节点， p != &e.list.root
// 因为初始化链表时 根节点的首位都是赋值给root 根节点了
func (e *Element) Next() *Element {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
// 返回该元素的前一个元素，如果没有前一个元素则返回nil
func (e *Element) Prev() *Element {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
// List 是一个双链表
type List struct {
	// root 链表的根元素, 用来判断链表是否已经到了链尾了
	root Element // sentinel list element, only &root, root.prev, and root.next are used
	// 链表长度
	len int // current list length excluding (this) sentinel element
}

// Init initializes or clears list l.
// 初始化或者清空一个链表
// 初始化一个链表时, root.next 和 root.prev 都等于root
// 当root.next == root.prev时 表示链表已经到链尾了
func (l *List) Init() *List {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New returns an initialized list.
// 返回一个初始化的list
func New() *List { return new(List).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
// 获取 list l 的长度
func (l *List) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
// 返回链表中的第一个元素，如果为空, 就是链表长度 l.len == 0  则返回nil
func (l *List) Front() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
// 返回最后一个元素， 如果为空 则返回 nil
func (l *List) Back() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List value.
// 延迟初始化
func (l *List) lazyInit() {
	// 如果是空链表，则初始化
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
// 在at 后面插入一个元素e, 并更新链表长度， 并返回e
func (l *List) insert(e, at *Element) *Element {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *List) insertValue(v interface{}, at *Element) *Element {
	return l.insert(&Element{Value: v}, at)
}

// remove removes e from its list, decrements l.len, and returns e.
// 移除元素e
// 这里默认l 不是空链表
func (l *List) remove(e *Element) *Element {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
	return e
}

// move moves e to next to at and returns e.
// 将e 移动到at 后面
func (l *List) move(e, at *Element) *Element {
	// 如果e 就是at 就不操作
	if e == at {
		return e
	}
	// 将e 剥离出来
	e.prev.next = e.next
	e.next.prev = e.prev

	// 挂载到at 后面
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e

	return e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
// 在链表l 中有e 元素 则移除它, 不管是否移除成功， 都返回e元素的值
func (l *List) Remove(e *Element) interface{} {
	if e.list == l { // 只有e元素在链表中才移除
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		// 如果e 是一个空元素 而且 链表也是空的， 则会crash掉
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
// 在 list l 的首部插入值为 v 的元素，并返回该元素
func (l *List) PushFront(v interface{}) *Element {
	l.lazyInit() // 判断如果为空链表则先初始化
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
// 在 list l 的末尾插入值为 v 的元素，并返回该元素
func (l *List) PushBack(v interface{}) *Element {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
// 在 list l 中元素 mark 之前插入一个值为 v 的元素，并返回该元素，如果 mark 不是list中元素，则 list 不改变
func (l *List) InsertBefore(v interface{}, mark *Element) *Element {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
// 在mark 元素之后 插入一个值为v的元素, 如果mark 不是list中的元素， 则List不改变
func (l *List) InsertAfter(v interface{}, mark *Element) *Element {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
// 将元素 e 移动到 list l 的首部，如果 e 不属于list l，则list不改变
func (l *List) MoveToFront(e *Element) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
// 将元素 e 移动到 list l 的末尾，如果 e 不属于list l，则list不改变
func (l *List) MoveToBack(e *Element) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
// 将元素 e 移动到元素 mark 之前，如果元素e 或者 mark 不属于 list l，或者 e==mark，则 list l 不改变
func (l *List) MoveBefore(e, mark *Element) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
// 将元素 e 移动到元素 mark 之后，如果元素e 或者 mark 不属于 list l，或者 e==mark，则 list l 不改变
func (l *List) MoveAfter(e, mark *Element) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
// 在 list l 的尾部插入另外一个 list，其中l 和 other 可以相等, 但是不能为空
func (l *List) PushBackList(other *List) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
// 在 list l 的首部插入另外一个 list，其中 l 和 other 可以相等
func (l *List) PushFrontList(other *List) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
