
## Container  容器数据类型

Container 主要实现了三个数据结构: 堆， 链表， 环 

[分析示例:coding/container](../coding/container)

[源码位置:/src/container](../go/src/container)

### Heap  最小堆

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


### List 双链表

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


### Ring  循环链表， 环 

container/ring 包中的 Ring 类型实现的是一个循环链表，俗称的环。其实 List 在内部就是一个循环链表。它的根元素永远不会持有任何实际的元素值，而该元素的存在就是为了连接这个循环链表的首尾两端。

所以，也可以说：List 的零值是一个只包含了根元素，但不包含任何实际元素值的空链表。那么，既然 Ring 和 List 的本质上都是循环链表，它们到底有什么不同呢？

Ring 和 List 的不同有以下几种：

- Ring 类型的数据结构仅由它自身即可代表，而 List 类型则需要由它以及 Element 类型联合表示。这是表示方式上的不同，也是结构复杂度上的不同。

- 一个 Ring 类型的值严格来讲，只代表了其所属的循环链表中的一个元素，而一个 List 类型的值则代表了一个完整的链表。这是表示维度上的不同。

- 在创建并初始化一个 Ring 值得时候，我们可以指定它包含的元素数量，但是对于一个 List 值来说却不能这样做(也没必要这样做)。循环链表一旦被创建，其长度是不可变的。这是两个代码包中 New 函数在功能上的不同，也是两个类型在初始化值方面的第一个不同

- 仅通过 var r ring.Ring 语句声明的 r 将会是一个长度为 1 的循环链表，而 List 类型的零值则是一个长度为 0 的链表。别忘了，List 中的根元素不会持有实际元素的值，因此计算长度时不会包含它。这是两个类型在初始化值方面的第二个不同。

- Ring 值得 Len 方法的算法复杂度是 O(N) 的，而 List 值得 Len 方法的算法复杂度是 O(1)的。这是两者在性能方面最显而易见的差别。

### 方法集

```go
type Ring struct {
	next, prev *Ring // 前后环指针
	// 值，这个值在ring包中不会被处理
	Value interface{} // for use by client; untouched by this library
}


type Ring
    func New(n int) *Ring               // 用于创建一个新的 Ring, 接收一个整形参数，用于初始化 Ring 的长度  
    func (r *Ring) Len() int            // 环长度
    
    func (r *Ring) Next() *Ring         // 返回当前元素的下个元素
    func (r *Ring) Prev() *Ring         // 返回当前元素的上个元素
    func (r *Ring) Move(n int) *Ring    // 指针从当前元素开始向后移动或者向前(n 可以为负数)

    // Link & Unlink 组合起来可以对多个链表进行管理
    func (r *Ring) Link(s *Ring) *Ring  // 将两个 ring 连接到一起 (r 不能为空)
    func (r *Ring) Unlink(n int) *Ring  // 从当前元素开始，删除 n 个元素

    func (r *Ring) Do(f func(interface{}))  // Do 会依次将每个节点的 Value 当作参数调用这个函数 f, 实际上这是策略方法的引用，通过传递不同的函数以在同一个 ring 上实现多种不同的操作。
```

### 遍历ring 

```go 
// 方式一
p := ring.Next()
// do something with the first element

for p != ring {
    // do something with current element
    p = p.Next()
}

// 方式二 
ring.Do(func(i interface{}){
    // do something with current element 
})

// 方式三,四，五 ... 参考ring包提供的方式 
```

可见通过ring.Do() 可以非常方便的将一组参数/规则 应用到函数中;

ring 包提供了很多使用示例可供参考

