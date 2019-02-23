// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors implements functions to manipulate errors.
// errors 包通过实现error接口, 即实现Error()string 方法来支持自定义操作错误
package errors

// New returns an error that formats as the given text.
//返回一个指定文本的error
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
//errorString 是对error接口的一种实现
type errorString struct {
	s string
}

//通过errorString指针e对象实现Error()string方法, 从而实现error接口
func (e *errorString) Error() string {
	return e.s
}
