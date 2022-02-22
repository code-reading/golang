// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"internal/reflectlite"
)

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
// 如果 err 实现了Unwrap，则返回Unwrap方法
// 否则返回Nil
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See syscall.Errno.Is for
// an example in the standard library.
// 判断 err 是不是与target 相等
// 通过反射比较
// 错误可能提供了自己的Is 方法， 用于比较 这个错误是不是的逻辑， 这个在业务处理中很方便
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		// 反射判断如果可比较，且相等, 则直接返回true
		if isComparable && err == target {
			return true
		}
		// 否则， 查看是不是有 Is(error） bool 方法
		// 查看某个对象是否有个某个方法， 通过这种断言可以的 ;
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporting target.Is(err). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		// 否则再次查看 err 有没有自己实现Unwrap ,  如果没法 返回false
		// 注意这里是循环， 如果有Unwrap方法， 则err !=nil
		// 那么这里就有一个for循环， 继续重复比较
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
// As 匹配错误链中第一个匹配的错误，匹配则将错误设置到target中返回true
// 如果target 为Nil 或者这个error 的类型， 则会报panics
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	// 这里用的 internal/reflectlite 包
	val := reflectlite.ValueOf(target)
	typ := val.Type() // 如果target 如果不是指针 或者为nil 则panic
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	targetType := typ.Elem() // 不是interface 并且没有实现 error 则panic
	if targetType.Kind() != reflectlite.Interface && !targetType.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	for err != nil {
		// 反射赋值方式
		// 需要注意的是 需要提前判断targetType 符合err的赋值条件
		if reflectlite.TypeOf(err).AssignableTo(targetType) {
			// target.Elem().Set(reflectlite.Value)
			val.Elem().Set(reflectlite.ValueOf(err))
			return true
		}
		// 如果err 实现了As 方法， 则递归调用err.As()方法
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = Unwrap(err) // 递归调用Unwrap
	} // 如果err == nil 直接返回false
	return false
}

var errorType = reflectlite.TypeOf((*error)(nil)).Elem()
