// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package context defines the Context type, which carries deadlines,
// cancellation signals, and other request-scoped values across API boundaries
// and between processes.
// 定义了Context类型 通过context类型可以在API直接传递 deadlines, cancellation等信号 已经其它
// 请求范围的值
//
// Incoming requests to a server should create a Context, and outgoing
// calls to servers should accept a Context. The chain of function
// calls between them must propagate the Context, optionally replacing
// it with a derived Context created using WithCancel, WithDeadline,
// WithTimeout, or WithValue. When a Context is canceled, all
// Contexts derived from it are also canceled.
/*
	在服务器直接请求应该携带context, 方法链直接的调用应该传递context, context 可以通过
	WithCancel, WithDeadline, WithTimeout, WithValue 创建， 当context取消时,
	该调用该context 的所有函数或请求也都能收到取消的信号
*/
//
// The WithCancel, WithDeadline, and WithTimeout functions take a
// Context (the parent) and return a derived Context (the child) and a
// CancelFunc. Calling the CancelFunc cancels the child and its
// children, removes the parent's reference to the child, and stops
// any associated timers. Failing to call the CancelFunc leaks the
// child and its children until the parent is canceled or the timer
// fires. The go vet tool checks that CancelFuncs are used on all
// control-flow paths.
/*
 Context 作为 WithCancel, WithDeadline, WithTimeout的父context, 通过Withxxx
 创建其子context, 和一个cancel函数 , 调用cancel函数可以取消其所有的子context 并且移除
 父context 对当前取消的context的引用;
 并且停止与之关联的timers
 Cancel函数调用失败或者忘记调用，会造成其子context 泄露， 这种泄露需要等到其父context 取消成功或者计时结束才会终止
 go vet 工具可以帮助检测所有控制流程中调用的CancelFuncs
*/
// Programs that use Contexts should follow these rules to keep interfaces
// consistent across packages and enable static analysis tools to check context
// propagation:
/*
	程序使用Contexts 需要遵守如下三条规则
	1. 不要将ctx 存储在结构体中, 而是应该显示的在(需要的)函数之间传递和调用
	   context 应该作为函数的第一个参数， 通常可以命名位ctx
	2. 不要传递一个nil Context, 如果不确定要使用哪个context, 应该传递 context.TODO
	3. ctx 值 仅用于传输过程中需要用到的作用域数据和API, 不应该给函数传递可选的参数变量

	context 是协程并发安全的;
*/
// Do not store Contexts inside a struct type; instead, pass a Context
// explicitly to each function that needs it. The Context should be the first
// parameter, typically named ctx:
//
// 	func DoSomething(ctx context.Context, arg Arg) error {
// 		// ... use ctx ...
// 	}
//
// Do not pass a nil Context, even if a function permits it. Pass context.TODO
// if you are unsure about which Context to use.
//
// Use context Values only for request-scoped data that transits processes and
// APIs, not for passing optional parameters to functions.
//
// The same Context may be passed to functions running in different goroutines;
// Contexts are safe for simultaneous use by multiple goroutines.
//
// See https://blog.golang.org/context for example code for a server that uses
// Contexts.
package context

import (
	"errors"
	"internal/reflectlite"
	"sync"
	"sync/atomic"
	"time"
)

// A Context carries a deadline, a cancellation signal, and other values across
// API boundaries.
//
// Context's methods may be called by multiple goroutines simultaneously.
// Context 可以在API之间传递 deadline, cancellation 等信号集其它值
// Context方法 在多协程并发时是安全的;
type Context interface {
	// Deadline returns the time when work done on behalf of this context
	// should be canceled. Deadline returns ok==false when no deadline is
	// set. Successive calls to Deadline return the same results.
	// 如果没有设置Deadline 则会返回false, 多次调用会返回相同的值
	//Deadline()返回该context是否有截止时间(timerCtx),如果有什么时候(time.Time)
	Deadline() (deadline time.Time, ok bool)

	// Done returns a channel that's closed when work done on behalf of this
	// context should be canceled. Done may return nil if this context can
	// never be canceled. Successive calls to Done return the same value.
	// The close of the Done channel may happen asynchronously,
	// after the cancel function returns.
	// 多次调用Done 会返回相同的值  Done chan的关闭 可以是在cancel function
	// 返回之后的异步操作
	// WithCancel arranges for Done to be closed when cancel is called;
	// WithDeadline arranges for Done to be closed when the deadline
	// expires; WithTimeout arranges for Done to be closed when the timeout
	// elapses.
	// Done 在 cancel 调用之后， deadline 过期， 或者timeout 超时之后会被close

	// Done is provided for use in select statements:
	//
	//  // Stream generates values with DoSomething and sends them to out
	//  // until DoSomething returns an error or ctx.Done is closed.
	//  func Stream(ctx context.Context, out chan<- Value) error {
	//  	for {
	//  		v, err := DoSomething(ctx)
	//  		if err != nil {
	//  			return err
	//  		}
	//  		select {
	//  		case <-ctx.Done(): // ctx.Done() 关闭 响应当前case
	//  			return ctx.Err()
	//  		case out <- v:
	//  		}
	//  	}
	//  }
	//
	// See https://blog.golang.org/pipelines for more examples of how to use
	// a Done channel for cancellation.
	//Done() 返回一个只读的channel,使用者通过从此channel读到一个值得知context已结束
	Done() <-chan struct{}

	// If Done is not yet closed, Err returns nil.
	// If Done is closed, Err returns a non-nil error explaining why:
	// Canceled if the context was canceled
	// or DeadlineExceeded if the context's deadline passed.
	// After Err returns a non-nil error, successive calls to Err return the same error.
	/*
		1.如果Done 还没有关闭， Err 返回nil
		2.Done 已经关闭, Err 返回一个non-nil 解释错误原因
		3.多次调用Err() 返回相同的信息, Canceled 或DeadlineExceede 会返回错误
	*/
	//Err()返回context由什么原因结束.手动结束?超过截止时间?
	Err() error

	// Value returns the value associated with this context for key, or nil
	// if no value is associated with key. Successive calls to Value with
	// the same key returns the same result.
	//
	// Use context values only for request-scoped data that transits
	// processes and API boundaries, not for passing optional parameters to
	// functions.
	//
	// A key identifies a specific value in a Context. Functions that wish
	// to store values in Context typically allocate a key in a global
	// variable then use that key as the argument to context.WithValue and
	// Context.Value. A key can be any type that supports equality;
	// packages should define keys as an unexported type to avoid
	// collisions.
	//
	// Packages that define a Context key should provide type-safe accessors
	// for the values stored using that key:
	//
	// 	// Package user defines a User type that's stored in Contexts.
	// 	package user
	//
	// 	import "context"
	//
	// 	// User is the type of value stored in the Contexts.
	// 	type User struct {...}
	//
	// 	// key is an unexported type for keys defined in this package.
	// 	// This prevents collisions with keys defined in other packages.
	// 	type key int
	//
	// 	// userKey is the key for user.User values in Contexts. It is
	// 	// unexported; clients use user.NewContext and user.FromContext
	// 	// instead of using this key directly.
	// 	var userKey key
	//
	// 	// NewContext returns a new Context that carries value u.
	// 	func NewContext(ctx context.Context, u *User) context.Context {
	// 		return context.WithValue(ctx, userKey, u)
	// 	}
	//
	// 	// FromContext returns the User value stored in ctx, if any.
	// 	func FromContext(ctx context.Context) (*User, bool) {
	// 		u, ok := ctx.Value(userKey).(*User)
	// 		return u, ok
	// 	}
	//Value(...)根据提供的key在context中遍历是否有这个key,如果有则返回其value否则返回nil
	Value(key interface{}) interface{}
}

// Canceled is the error returned by Context.Err when the context is canceled.
// Canceled 会返回这个错误
var Canceled = errors.New("context canceled")

// DeadlineExceeded is the error returned by Context.Err when the context's
// deadline passes.
// 超时错误
var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string   { return "context deadline exceeded" }
func (deadlineExceededError) Timeout() bool   { return true }
func (deadlineExceededError) Temporary() bool { return true }

// An emptyCtx is never canceled, has no values, and has no deadline. It is not
// struct{}, since vars of this type must have distinct addresses.
// 空Ctx 没有值，没有deadline,也不是空struct, 因为emptyCtx 必须是不同的地址， 而struct{} 是具有相同地址的
// emtpyCtx 没有cancel, emptyCtx 实现了Context 接口
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil // 返回一个nil 的<-chan  那么读取这个chan 是永远阻塞的 ,直到有值才不会阻塞
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}

func (e *emptyCtx) String() string {
	switch e {
	case background:
		return "context.Background"
	case todo:
		return "context.TODO"
	}
	return "unknown empty Context"
}

// emptyCtx 两种实例类型
// background  和 todo
var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

// Background returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline. It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming
// requests.
// Background  返回一个 non-nil, empty 的Context.
// emtpyCtx 没有cancel, 没有value 和deadline
// 在main 初始化， 测试场景， 或者顶层的context中使用
func Background() Context {
	return background
}

// TODO returns a non-nil, empty Context. Code should use context.TODO when
// it's unclear which Context to use or it is not yet available (because the
// surrounding function has not yet been extended to accept a Context
// parameter).
// 如果不确定使用什么Context 就使用TODO()
func TODO() Context {
	return todo
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc does not wait for the work to stop.
// A CancelFunc may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to a CancelFunc do nothing.
// CancelFunc 执行告诉调用者丢弃未完成的工作, CancelFunc 并不会等work 结束 , 可以在多协程中并发调用
// 第一次调用CancelFunc 之后， 后面在调用CancelFunc 啥事也不做
type CancelFunc func()

// WithCancel returns a copy of parent with a new Done channel. The returned
// context's Done channel is closed when the returned cancel function is called
// or when the parent context's Done channel is closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
// 返回一个cancelCtx 和 cancelFunc, 调用cancelFunc 会关闭cancelCtx中的done channel
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	c := newCancelCtx(parent)
	propagateCancel(parent, &c) //尝试把c与父节点绑定(写入父节点的hashmap)
	// 外部主动调用cancel 那么removeFromParent 就为true 表示从parent 摘除这个ctx
	return &c, func() { c.cancel(true, Canceled) }
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent Context) cancelCtx {
	return cancelCtx{Context: parent}
}

// goroutines counts the number of goroutines ever created; for testing.
var goroutines int32

// propagateCancel arranges for child to be canceled when parent is.
// 父ctx取消的时候，传递cancel 给它的子ctx
func propagateCancel(parent Context, child canceler) {
	// 如果返回nil，说明当前父`context`从来不会被取消，是一个空节点，直接返回即可
	done := parent.Done()
	if done == nil {
		return // parent is never canceled
	}
	// 提前判断一个父context是否被取消，如果取消了也不需要构建关联了，
	// 把当前子节点取消掉并返回
	select {
	case <-done:
		// parent is already canceled
		child.cancel(false, parent.Err())
		return
	default:
	}
	// 这里目的就是找到可以“挂”、“取消”的context
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		// 找到了可以“挂”、“取消”的context，但是已经被取消了，那么这个子节点也不需要
		// 继续挂靠了，取消即可
		if p.err != nil {
			// parent has already been canceled
			child.cancel(false, p.err)
		} else {
			if p.children == nil {
				// 将当前节点挂到父节点的childrn map中，外面调用cancel时可以层层取消
				p.children = make(map[canceler]struct{})
			}
			p.children[child] = struct{}{}
		}
		p.mu.Unlock()
	} else {
		// 没有找到 就开一个goroutine
		atomic.AddInt32(&goroutines, +1)
		go func() {
			select { //当父context为不识别的或emptyctx时创建独立goroutine维护生命周期
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}

// &cancelCtxKey is the key that a cancelCtx returns itself for.
var cancelCtxKey int

// parentCancelCtx returns the underlying *cancelCtx for parent.
// It does this by looking up parent.Value(&cancelCtxKey) to find
// the innermost enclosing *cancelCtx and then checking whether
// parent.Done() matches that *cancelCtx. (If not, the *cancelCtx
// has been wrapped in a custom implementation providing a
// different done channel, in which case we should not bypass it.)
func parentCancelCtx(parent Context) (*cancelCtx, bool) {
	done := parent.Done()
	if done == closedchan || done == nil {
		return nil, false
	}
	p, ok := parent.Value(&cancelCtxKey).(*cancelCtx)
	if !ok {
		return nil, false
	}
	pdone, _ := p.done.Load().(chan struct{})
	if pdone != done {
		return nil, false
	}
	return p, true
}

// removeChild removes a context from its parent.
func removeChild(parent Context, child canceler) {
	// 查找parent ctx
	p, ok := parentCancelCtx(parent)
	if !ok {
		return
	}
	p.mu.Lock()
	// 摘除子ctx
	if p.children != nil {
		delete(p.children, child)
	}
	p.mu.Unlock()
}

// A canceler is a context type that can be canceled directly. The
// implementations are *cancelCtx and *timerCtx.
type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
}

// closedchan is a reusable closed channel.
// 可复用的关闭的channel
var closedchan = make(chan struct{})

func init() {
	close(closedchan) // 初始化就关闭
}

// A cancelCtx can be canceled. When canceled, it also cancels any children
// that implement canceler.
// cancelCtx 也实现了Context 接口， 并且cancel触发之后，其所有子ctx 都会被cancel掉
type cancelCtx struct {
	Context // 这种用法 表示继承了Context接口的方法

	mu sync.Mutex // protects following fields 保护fields读写操作的互斥锁
	// 如果channel 为nil done 就保存一个closedchan 结构
	done atomic.Value // of chan struct{}, created lazily, closed by first cancel call 表示channel 是否关闭
	// children是一个hashmap,key为实现canceler接口的实体,value为空类型.
	// 这个filed的作用是:当本身或父级context传来结束生命周期信号时(调用了自身的cancel方法),
	// 通过这个map寻找所有指向本节点的子节点,并调用他们的cancel方法.
	children map[canceler]struct{} // set to nil by the first cancel call
	err      error                 // set to non-nil by the first cancel call 描述context 生命周期结束的原因
}

func (c *cancelCtx) Value(key interface{}) interface{} {
	if key == &cancelCtxKey { // ???
		return c
	}
	return c.Context.Value(key)
}

func (c *cancelCtx) Done() <-chan struct{} {
	d := c.done.Load()
	if d != nil {
		// 非nil 则没有结束
		return d.(chan struct{}) // chan struct{} 编码优化技巧, struct{}不占大小空间
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d = c.done.Load() // 两次Load
	if d == nil {
		d = make(chan struct{})
		c.done.Store(d) // 存储d 表示已经结束了
	}
	return d.(chan struct{})
}

func (c *cancelCtx) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

type stringer interface {
	String() string
}

func contextName(c Context) string {
	if s, ok := c.(stringer); ok {
		return s.String()
	}
	return reflectlite.TypeOf(c).String()
}

func (c *cancelCtx) String() string {
	return contextName(c.Context) + ".WithCancel"
}

// cancel closes c.done, cancels each of c's children, and, if
// removeFromParent is true, removes c from its parent's children.
// cancel方法 关闭当前channel ， 并且依次关闭它的子ctx
// 这个cancel 是幂等的
// cancel() 方法的功能就是关闭 channel：c.done；递归地取消它的所有子节点；
// 从父节点从删除自己。达到的效果是通过关闭 channel，将取消信号传递给了它的所有子节点。
// goroutine 接收到取消信号的方式就是 select 语句中的读 c.done 被选中。
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	// 取消时传入的error信息不能为nil, context定义了默认error:var Canceled = errors.New("context canceled")
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	// 已经有错误信息了，说明当前节点已经被取消过了
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err

	d, _ := c.done.Load().(chan struct{})
	if d == nil {
		c.done.Store(closedchan)
	} else {
		// 关闭channel ,通知这个channel阻塞的协程
		close(d)
	}
	// 取消子ctx
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err) // 这里是内部调用 这里的removeFromParent = false
	}
	c.children = nil // 手动赋值为nil
	c.mu.Unlock()
	// 把当前节点从父节点中移除，只有在外部父节点调用时才会传true
	// 其他都是传false，内部调用都会因为c.children = nil被剔除出去
	if removeFromParent {
		removeChild(c.Context, c)
	}
}

// WithDeadline returns a copy of the parent context with the deadline adjusted
// to be no later than d. If the parent's deadline is already earlier than d,
// WithDeadline(parent, d) is semantically equivalent to parent. The returned
// context's Done channel is closed when the deadline expires, when the returned
// cancel function is called, or when the parent context's Done channel is
// closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	// 当父context的结束时间早于要设置的时间，则不需要再去单独处理子节点的定时器了
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		// The current deadline is already sooner than the new one.
		return WithCancel(parent)
	}
	c := &timerCtx{
		cancelCtx: newCancelCtx(parent),
		deadline:  d,
	}
	// 将当前节点挂到父节点上
	propagateCancel(parent, c)
	dur := time.Until(d)
	if dur <= 0 { // 过期了也主动摘除这个子ctx
		c.cancel(true, DeadlineExceeded) // deadline has already passed
		return c, func() { c.cancel(false, Canceled) }
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// 如果没被取消，则直接添加一个定时器，定时去取消
	if c.err == nil {
		c.timer = time.AfterFunc(dur, func() {
			c.cancel(true, DeadlineExceeded)
		})
	}
	return c, func() { c.cancel(true, Canceled) }
}

// A timerCtx carries a timer and a deadline. It embeds a cancelCtx to
// implement Done and Err. It implements cancel by stopping its timer then
// delegating to cancelCtx.cancel.
type timerCtx struct {
	cancelCtx
	timer *time.Timer // Under cancelCtx.mu.

	deadline time.Time
}

func (c *timerCtx) Deadline() (deadline time.Time, ok bool) {
	return c.deadline, true
}

func (c *timerCtx) String() string {
	return contextName(c.cancelCtx.Context) + ".WithDeadline(" +
		c.deadline.String() + " [" +
		time.Until(c.deadline).String() + "])"
}

func (c *timerCtx) cancel(removeFromParent bool, err error) {
	// 调用cancelCtx的cancel方法取消掉子节点context
	c.cancelCtx.cancel(false, err)
	if removeFromParent {
		// Remove this timerCtx from its parent cancelCtx's children.
		removeChild(c.cancelCtx.Context, c)
	}
	c.mu.Lock()
	// 停掉定时器，释放资源
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil // 注意 上面stop之后， 这里还是手动赋值nil
	}
	c.mu.Unlock()
}

// WithTimeout returns WithDeadline(parent, time.Now().Add(timeout)).
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete:
//
// 	func slowOperationWithTimeout(ctx context.Context) (Result, error) {
// 		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
// 		defer cancel()  // releases resources if slowOperation completes before timeout elapses
// 		return slowOperation(ctx)
// 	}
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

// WithValue returns a copy of parent in which the value associated with key is
// val.
//
// Use context Values only for request-scoped data that transits processes and
// APIs, not for passing optional parameters to functions.
//
// The provided key must be comparable and should not be of type
// string or any other built-in type to avoid collisions between
// packages using context. Users of WithValue should define their own
// types for keys. To avoid allocating when assigning to an
// interface{}, context keys often have concrete type
// struct{}. Alternatively, exported context key variables' static
// type should be a pointer or interface.
func WithValue(parent Context, key, val interface{}) Context {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	if key == nil {
		panic("nil key")
	}
	if !reflectlite.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return &valueCtx{parent, key, val}
}

// A valueCtx carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Context.
type valueCtx struct {
	Context
	key, val interface{}
}

// stringify tries a bit to stringify v, without using fmt, since we don't
// want context depending on the unicode tables. This is only used by
// *valueCtx.String().
func stringify(v interface{}) string {
	switch s := v.(type) {
	case stringer:
		return s.String()
	case string:
		return s
	}
	return "<not Stringer>"
}

func (c *valueCtx) String() string {
	return contextName(c.Context) + ".WithValue(type " +
		reflectlite.TypeOf(c.key).String() +
		", val " + stringify(c.val) + ")"
}

func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}
