
## Container  容器数据类型

Container 主要实现了三个数据结构: 堆， 链表， 环 

[分析示例:coding/container](../coding/container)

[源码位置:/src/container](../go/src/container)

### Heap 

Heap 包定义了堆的实现接口，提供了 堆接口Push, Pop 并继承sort.Interface 的 Len, Less, Swap接口 ， 所以实现一个最小堆， 需要实现上面五个接口； 

需要注意， Pop 时， heap 接口会先将堆顶堆尾数据交换，所以实现 Pop 接口时，读取堆顶元素，实际是读取数组的最后一个元素 

Heap 主要提供了以下方法列表 

h 是自定义数据结构的指针

- heap.Init(h) 初始化最小堆 

- heap.Push(h) 向堆中添加数据 

- heap.Pop(h) 删除并返回堆顶数据

- heap.Remove(h, i) 删除第i个数据元素

- heap.Fix(h, i) 在外部修改了i元素值后， Fix重新堆化, 调整成最小堆

### 堆的使用场景

- 构建优先队列

- 支持堆排序

- 快速找出一个集合中的最小值（或者最大值）


### List 

List 提供了对链表的实现,  list.go 中提供两个元素, List 和 Element , 其中 List 实现了一个双向链表, Element表示链表中元素的结构

> <font color='red'>注意: container/list默认不是线程安全的，要保证数据安全，那么可以使用Lock锁解决。</font>

### 结构定义 

```go 
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

type List struct {
	// root 链表的根元素, 用来判断链表是否已经到了链尾了
	root Element // sentinel list element, only &root, root.prev, and root.next are used
	// 链表长度
	len int // current list length excluding (this) sentinel element
}
```

### 方法集

```go 
type Element
    func (e *Element) Next() *Element                                   // 返回该元素的下一个元素，如果没有下一个元素则返回 nil
    func (e *Element) Prev() *Element                                   // 返回该元素的前一个元素，如果没有前一个元素则返回nil

type List     
    // New, Init, 读取第一个， 最后一个元素， 获取链表长度， 删除一个元素                           
    func New() *List                                                    // 返回一个初始化的list
    func (l *List) Init() *List                                         // list l 初始化或者清除 list l
    func (l *List) Back() *Element                                      // 获取list l的最后一个元素
    func (l *List) Front() *Element                                     // 获取list l的最后一个元素
    func (l *List) Len() int                                            // 获取 list l 的长度
    func (l *List) Remove(e *Element) interface{}                       // 如果元素 e 属于list l，将其从 list 中删除，并返回元素 e 的值

    // 插入一个元素， 插入到链表头部或者末尾 
    func (l *List) PushFront(v interface{}) *Element                    // 在 list l 的首部插入值为 v 的元素，并返回该元素              
    func (l *List) PushBack(v interface{}) *Element                     // 在 list l 的末尾插入值为 v 的元素，并返回该元素              

    // 在指定元素之前/之后插入一个新值v的元素
    func (l *List) InsertAfter(v interface{}, mark *Element) *Element   // 在 list l 中元素 mark 之后插入一个值为 v 的元素，并返回该元素，如果 mark 不是list中元素，则 list 不改变
    func (l *List) InsertBefore(v interface{}, mark *Element) *Element  // 在 list l 中元素 mark 之前插入一个值为 v 的元素，并返回该元素，如果 mark 不是list中元素，则 list 不改变
    
    // 移动一个已经存在的元素到指定元素之前/之后
    func (l *List) MoveAfter(e, mark *Element)                          // 将元素 e 移动到元素 mark 之后，如果元素e 或者 mark 不属于 list l，或者 e==mark，则 list l 不改变
    func (l *List) MoveBefore(e, mark *Element)                         // 将元素 e 移动到元素 mark 之前，如果元素e 或者 mark 不属于 list l，或者 e==mark，则 list l 不改变
    
    // 移动一个存在的元素到链表头部/末尾
    func (l *List) MoveToBack(e *Element)                               // 将元素 e 移动到 list l 的末尾，如果 e 不属于list l，则list不改变             
    func (l *List) MoveToFront(e *Element)                              // 将元素 e 移动到 list l 的首部，如果 e 不属于list l，则list不改变             
    
    // 在链表头部/末尾 插入另外一个链表
    func (l *List) PushBackList(other *List)                            // 在 list l 的尾部插入另外一个 list，其中l 和 other 可以相等               
    func (l *List) PushFrontList(other *List)                           // 在 list l 的首部插入另外一个 list，其中 l 和 other 可以相等              
```

### var l list.List 声明的变量 l 可以直接用吗？ 值将会是什么呢？

- 可以的。这被称为 开箱即用。 这种通过语句 var l list.List 声明的链表 l 可以直接使用的原因就是在于它的 延迟初始化 机制。

- l将会是一个长度为 0 的链表，这个链表持有的根元素也将会是一个空壳，其中只会包含缺省的内容。
