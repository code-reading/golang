package main

import (
	"container/heap"
	"fmt"
)

// Item 优先级队列项Item
type Item struct {
	value    string // 数据
	priority int    // 优先级
	index    int    // 堆数组中的索引
}

// PriorityQueue  优先级队列
type PriorityQueue []*Item

// 接着heap 接口实现优先级队列， 需要实现 heap 的Push 和Pop
// 及其继承sort.Interface 的 Len, Less, Swap 这五个接口

// Len 不涉及修改pq 操作的 可以不需要使用指针
func (p PriorityQueue) Len() int {
	return len(p)
}

// Less 业务定义: 优先级大的 表示优先级高 这里用 > 大于号
func (p PriorityQueue) Less(i, j int) bool {
	return p[i].priority > p[j].priority
}

// Swap 交换i和j的位置
func (p PriorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j
}

// 涉及到修改堆数组的时候需要用指针
// Push 加入新节点, 注意 堆接口Push 的入口参数是interface{}
func (p *PriorityQueue) Push(x interface{}) {
	// 获得加入新节点的下标位置
	n := len(*p)
	item := x.(*Item)
	item.index = n          // 保存数组索引
	(*p) = append(*p, item) // 添加item
}

// Pop 出口参数也是interface
func (p *PriorityQueue) Pop() interface{} {
	old := *p        // 用临时变量old 指向优先级队列p
	n := len(old)    // 获取队列长度
	item := old[n-1] // 获取最后一个参数
	// 注意事项
	// 防止内存泄露
	old[n-1] = nil
	// 防止索引被引用
	item.index = -1
	*p = old[:n-1] // 修改队列长度
	return item    // 返回堆顶元素
}

// update  更新索引和值
func (p *PriorityQueue) update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(p, item.index) // 索引在更新堆化这里用到, 因为重新调整堆时 需要知道是哪个节点要调整
}

func main() {
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}
	pq := make(PriorityQueue, len(items)) // 初始化优先级队列, 注意这里填写了len 后面就不要用append 了否则前面len个数据是Nil
	i := 0
	for item, priority := range items {
		pq[i] = &Item{
			value:    item,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq) // 构建最小堆
	// 插入一个新项，并修改其优先级
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&pq, item)
	pq.update(item, item.value, 5) // 修改优先级

	// 读取优先级队列
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("%.2d:%s ", item.priority, item.value)
	}
	// 05:orange 04:pear 03:banana 02:apple
}
