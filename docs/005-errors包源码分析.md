## errors 

```go
// The error built-in interface type is the conventional interface for
// representing an error condition, with the nil value representing no error.
type error interface {
	Error() string
}
``` 

任何实现这个error interface 的类型/结构体都都可以赋值给error 


第三方库: github.com/pkg/errors

```go
// Wrap annotates cause with a message.
func Wrap(cause error, message string) error
// Cause unwraps an annotated error.
func Cause(err error) error
```
通过 Wrap 可以将一个错误，加上一个字符串，“包装”成一个新的错误；通过 Cause 则可以进行相反的操作，将里层的错误还原。

```go
func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open failed")
	}
	defer f.Close()
	
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	return buf, nil
}
```

示例 一个错误可能被处理多次 

```go
func Write(w io.Writer, buf []byte) error { 
	_, err := w.Write(buf)
	if err != nil {
		// annotated error goes to log file
		log.Println("unable to write:", err)
	
		// unannotated error returned to caller return err
		return err
	}
	return nil
}
```
优化方案

```go
func Write(w io.Write, buf []byte) error {
	_, err := w.Write(buf)
	return errors.Wrap(err, "write failed")
}
```

在golang 1.13 开始，为了支持 wrapping，fmt.Errorf 增加了 %w 的格式，并且在 error 包增加了三个函数：errors.Unwrap，errors.Is，errors.As。

fmt.Errorf

使用 fmt.Errorf 加上 %w 格式符来生成一个嵌套的 error，它并没有像 pkg/errors 那样使用一个 Wrap 函数来嵌套 error，非常简洁。

Unwrap

func Unwrap(err error) error

将嵌套的 error 解析出来，多层嵌套需要调用 Unwrap 函数多次，才能获取最里层的 error。

```go
func Unwrap(err error) error {
    // 判断是否实现了 Unwrap 方法
	u, ok := err.(interface {
		Unwrap() error
	})
	// 如果不是，返回 nil
	if !ok {
		return nil
	}
	// 调用 Unwrap 方法返回被嵌套的 error
	return u.Unwrap()
}
```
对 err 进行断言，看它是否实现了 Unwrap 方法，如果是，调用它的 Unwrap 方法。否则，返回 nil。

Is

func Is(err, target error) bool

判断 err 是否和 target 是同一类型，或者 err 嵌套的 error 有没有和 target 是同一类型的，如果是，则返回 true。

```go
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	
	// 无限循环，比较 err 以及嵌套的 error
	for {
		if isComparable && err == target {
			return true
		}
		// 调用 error 的 Is 方法，这里可以自定义实现
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// 返回被嵌套的下一层的 error
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}
```

通过一个无限循环，使用 Unwrap 不断地将 err 里层嵌套的 error 解开，再看被解开的 error 是否实现了 Is 方法，并且调用它的 Is 方法，当两者都返回 true 的时候，整个函数返回 true。

As

func As(err error, target interface{}) bool

从 err 错误链里找到和 target 相等的并且设置 target 所指向的变量。

```go
func As(err error, target interface{}) bool {
    // target 不能为 nil
	if target == nil {
		panic("errors: target cannot be nil")
	}
	
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	
	// target 必须是一个非空指针
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	
	// 保证 target 是一个接口类型或者实现了 Error 接口
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
	    // 使用反射判断是否可被赋值，如果可以就赋值并且返回true
		if reflectlite.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(err))
			return true
		}
		
		// 调用 error 自定义的 As 方法，实现自己的类型断言代码
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		// 不断地 Unwrap，一层层的获取嵌套的 error
		err = Unwrap(err)
	}
	return false
}
```

返回 true 的条件是错误链里的 err 能被赋值到 target 所指向的变量；或者 err 实现的 As(interface{}) bool 方法返回 true。

前者，会将 err 赋给 target 所指向的变量；后者，由 As 函数提供这个功能。

如果 target 不是一个指向“实现了 error 接口的类型或者其它接口类型”的非空的指针的时候，函数会 panic。

