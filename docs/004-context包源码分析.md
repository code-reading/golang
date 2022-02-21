
## Context 

[分析示例:coding/context](../coding/context)

[源码位置:/src/context](../go/src/contex)

[代码设计:/design/context](../design/context.pu)

context 用来解决 goroutine 之间退出通知、元数据(token, trace_id等等)传递的功能
### Context 接口 
Context 其实是定义了一个接口, 那么用户可以通过实现这个接口而自定义Context 

```go
type Context interface {
 Deadline() (deadline time.Time, ok bool)
 Done() <-chan struct{}
 Err() error
 Value(key interface{}) interface{}
}
```
Deadlne方法：当Context自动取消或者到了取消时间被取消后返回
Done方法：当Context被取消或者到了deadline返回一个被关闭的channel
Err方法：当Context被取消或者关闭后，返回context取消的原因
Value方法：获取设置的key对应的值

这个接口主要被三个类继承实现，分别是emptyCtx、ValueCtx、cancelCtx，采用匿名接口的写法，这样可以对任意实现了该接口的类型进行重写。

基于emptyCtx, ValueCtx , cancelCtx 这几个类 有派生出如下接个功能函数

```go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
func WithValue(parent Context, key, val interface{}) Context
```

### cancler 接口 

```go
type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
}
```
实现了上面定义的两个方法的 Context，就表明该 Context 是可取消的。源码中有两个类型实现了 canceler 接口：*cancelCtx 和 *timerCtx。注意是加了 * 号的，是这两个结构体的指针实现了 canceler 接口。


### 创建context
context包主要提供了两种方式创建context:
```go
context.Backgroud()
context.TODO()
```
这两个函数其实只是互为别名，没有差别，官方给的定义是：

- context.Background 是上下文的默认值，所有其他的上下文都应该从它衍生（Derived）出来。
- context.TODO 应该只在不确定应该使用哪种上下文时使用；


### context场景
context的作用就是在不同的goroutine之间同步请求特定的数据、取消信号以及处理请求的截止日期。

context 可以在函数之间传递一些公共参数，比如全局日志追踪的签名，trace_id等， 但是不要传递业务参数， 一是因为context传递的参数都是interface{} 类型，需要来回转换; 二是隐性传参，导致代码可读性差，维护很难;

#### 传值

```go
const (
 KEY = "trace_id"
)

func NewRequestID() string {
 return strings.Replace(uuid.New().String(), "-", "", -1)
}

func NewContextWithTraceID() context.Context {
 ctx := context.WithValue(context.Background(), KEY,NewRequestID())
 return ctx
}

func PrintLog(ctx context.Context, message string)  {
 fmt.Printf("%s|info|trace_id=%s|%s",time.Now().Format("2006-01-02 15:04:05") , GetContextValue(ctx, KEY), message)
}

func GetContextValue(ctx context.Context,k string)  string{
 v, ok := ctx.Value(k).(string)
 if !ok{
  return ""
 }
 return v
}

func ProcessEnter(ctx context.Context) {
 PrintLog(ctx, "Golang梦工厂")
}


func main()  {
 ProcessEnter(NewContextWithTraceID())
}
```

output 
```shell
2021-10-31 15:13:25|info|trace_id=7572e295351e478e91b1ba0fc37886c0|Golang梦工厂
Process finished with the exit code 0
```

#### 超时控制
通常健壮的程序都是要设置超时时间的，避免因为服务端长时间响应消耗资源，withTimeout或者withDeadline来做超时控制，当一次请求到达我们设置的超时时间，就会及时取消，不在往下执行。withTimeout和withDeadline作用是一样的，就是传递的时间参数不同而已。

#### 取消控制

WithCancel 
