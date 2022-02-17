
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



