### 什么是 CSP

CSP 全称是 “Communicating Sequential Processes”，这也是 Tony Hoare 在 1978 年发表在 ACM 的一篇论文。论文里指出一门编程语言应该重视 input 和 output 的原语，尤其是并发编程的代码。

大多数的编程语言的并发编程模型是基于线程和内存同步访问控制，Go 的并发编程的模型则用 goroutine 和 channel 来替代。Goroutine 和线程类似，channel 和 mutex (用于内存同步访问控制)类似。

Go 的并发原则非常优秀，目标就是简单：尽量使用 channel；把 goroutine 当作免费的资源，随便用。

Go 并发不要通过共享内存来通信, 而要通过通信来实现内存共享, 它依赖 CSP 模型，基于 channel 实现。

### 什么是channel 

Goroutine 和 channel 是 Go 语言并发编程的 两大基石。Goroutine 用于执行并发任务，channel 用于 goroutine 之间的同步、通信。

Channel 在 gouroutine 间架起了一条管道，在管道里传输数据，实现 gouroutine 间的通信；由于它是线程安全的，所以用起来非常方便；channel 还提供“先进先出”的特性；它还能影响 goroutine 的阻塞和唤醒。

```go
chan T 
// 声明一个双向通道

chan<- T 
// 声明一个只能用于发送的通道

<-chan T 
// 声明一个只能用于接收的通道
```

Go 通过 channel 实现 CSP 通信模型，主要用于 goroutine 之间的消息传递和事件通知。

源码中关于channel 主要的生命周期

```go 
// 创建一个channel 
func makechan(t *chantype, size int) *hchan
// 发送数据
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool 
// 接收数据
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool)
// 关闭 channel 
func closechan(c *hchan)
```

### makechan 
makechan 创建一个channel, 返回一个channel 的指针， *hchan

