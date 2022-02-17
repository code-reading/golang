package main

import (
	"container/list"
	"fmt"
	"sync"
)

type Queue struct {
	list.List
	sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{}
}

func (p *Queue) Push(v interface{}) {
	p.Lock()
	defer p.Unlock()
	p.PushFront(v)
}

func (p *Queue) Pop() (v interface{}) {
	p.Lock()
	defer p.Unlock()
	iter := p.Back()
	if iter != nil {
		v = iter.Value
	}
	p.Remove(iter)
	return v
}

func (p *Queue) Dump() {
	for e := p.Front(); e != nil; e = e.Next() {
		fmt.Printf("%v ", e.Value)
	}
}

func main() {
	q := NewQueue()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			q.Push(i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i += 3 {
			q.Push(i)
		}
	}()
	q.Push(10)
	wg.Wait()
	fmt.Println("pop :", q.Pop())
	q.Dump()

}

// pop : 10
// 2 1 0 9 6 3 0
