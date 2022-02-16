package main

import (
	"fmt"
	"sync"
)

type Info struct {
	Age int
}

func main() {
	pool := sync.Pool{
		New: func() interface{} {
			return &Info{
				Age: 1,
			}
		},
	}
	infoObject := pool.Get().(*Info)
	fmt.Println(infoObject.Age) // print 1
	pool.Put(infoObject)
}
