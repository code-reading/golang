// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package unsafe contains operations that step around the type safety of Go programs.

	Packages that import unsafe may be non-portable and are not protected by the
	Go 1 compatibility guidelines.
	unsafe包包含一些可以绕过Go语言中的类型安全的操作，如果引用了unsafe包可能会导致不稳定，而且也不会受Go 1的兼容性所保护
*/
package unsafe

// ArbitraryType is here for the purposes of documentation only and is not actually
// part of the unsafe package. It represents the type of an arbitrary Go expression.
// ArbitraryType 可以表示任意一个go 表达式， 没其它用意
type ArbitraryType int

// IntegerType is here for the purposes of documentation only and is not actually
// part of the unsafe package. It represents any arbitrary integer type.
// IntegerType 表示任意一个整数类型， 没有其它用意
type IntegerType int

// Pointer represents a pointer to an arbitrary type. There are four special operations
// available for type Pointer that are not available for other types:
// Pointer 可以表示执行一个任意类型， 如下几种操作是其它类型不具有的
// 大概意思就是 unsafe.Pointer 可以用于不同类型指针的转换
/*
示例
v1 := uint(1)
v2 := int(2)

p := &v1		// p的类型是(uint *)
p = &v2			// 会报错，不能把(int *) 赋给(uint *)

// 可以通过unsafe.Pointer实现
p = (*uint)(unsafe.Pointer(&v2))
*/
//	- A pointer value of any type can be converted to a Pointer.
//	- A Pointer can be converted to a pointer value of any type.
//	- A uintptr can be converted to a Pointer.
//	- A Pointer can be converted to a uintptr.
// Pointer therefore allows a program to defeat the type system and read and write
// arbitrary memory. It should be used with extreme care.
/*
unsafe 包主要是可以使得用户绕过go的类型规范检查，能够对指针以及其指向的区域进行读写操作。但是使用时要格外小心。
*/
//
// The following patterns involving Pointer are valid.
// Code not using these patterns is likely to be invalid today
// or to become invalid in the future.
// Even the valid patterns below come with important caveats.
// 下面将会介绍哪些格式是允许的，其他的操作方法都是禁止的，
// 就算现在不符合下面规范的一些写法可以编译通过，但是未来很可能就无法通过。
//
// Running "go vet" can help find uses of Pointer that do not conform to these patterns,
// but silence from "go vet" is not a guarantee that the code is valid.
//
// (1) Conversion of a *T1 to Pointer to *T2.
//
// Provided that T2 is no larger than T1 and that the two share an equivalent
// memory layout, this conversion allows reinterpreting data of one type as
// data of another type. An example is the implementation of
// math.Float64bits:
// 这里的第一个例子是，把一种数值类型转换成另一种的方法，用的是float64转成uint64，
// 基本过程就是先把存储f值的地址的这个数值转换成unsafe.Pointer，将这个Pointer地址转换成 *uint64类型的地址，
// 之后按照使用uint64类型指针那样，调用这个位置的值就可以了。这样就可以绕过Go语言的类型检查，将二进制强制转换数值类型。

//	func Float64bits(f float64) uint64 {
//		return *(*uint64)(unsafe.Pointer(&f))
//	}
//
// (2) Conversion of a Pointer to a uintptr (but not back to Pointer).
// Converting a Pointer to a uintptr produces the memory address of the value
// pointed at, as an integer. The usual use for such a uintptr is to print it.
//把一个Pointer转换成uintptr类型，通常这个用法是打印该地址
// 通常来说，你得到了这个uintptr类型的地址数值，Go是不允许你再把它转换回Pointer类型的。

// Conversion of a uintptr back to Pointer is not valid in general.
// 通常情况， 不能将一个uintptr 转换为一个Pointer
// A uintptr is an integer, not a reference.
// uintptr 是一个整数，不是一个引用
// Converting a Pointer to a uintptr creates an integer value
// with no pointer semantics.
// 将一个Pointer 转换为一个uintptr 相当于创建一个没有指针指向的整数值
// Even if a uintptr holds the address of some object,
// the garbage collector will not update that uintptr's value
// if the object moves, nor will that uintptr keep the object
// from being reclaimed.
// 就算一个uintptr 保存了执行某个对象的地址
// gc回收了这个对象， 也不会更新uintptr的内容
// 这样 uintptr 的内容就和object 不会保持一致了
// The remaining patterns enumerate the only valid conversions
// from uintptr to Pointer.
// 唯一有效的转换方式
// (3) Conversion of a Pointer to a uintptr and back, with arithmetic.
//
// If p points into an allocated object, it can be advanced through the object
// by conversion to uintptr, addition of an offset, and conversion back to Pointer.
// 一个point 执行一个分配了的对象，可以通过这个对象转换成一个uintptr, 如果给这个uintptr的对象 再加上转换之前的一个offset
// 那么 uintptr 可以再次转换为一个point 类型
//	p = unsafe.Pointer(uintptr(p) + offset)
//
// The most common use of this pattern is to access fields in a struct
// or elements of an array:
// 这种模式 常用于读取数组中的字段
//	// equivalent to f := unsafe.Pointer(&s.f)
//	f := unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + unsafe.Offsetof(s.f))
//
//	// equivalent to e := unsafe.Pointer(&x[i])
//	e := unsafe.Pointer(uintptr(unsafe.Pointer(&x[0])) + i*unsafe.Sizeof(x[0]))
//
// It is valid both to add and to subtract offsets from a pointer in this way.
// It is also valid to use &^ to round pointers, usually for alignment.
// In all cases, the result must continue to point into the original allocated object.
//
// Unlike in C, it is not valid to advance a pointer just beyond the end of
// its original allocation:
//
//	// INVALID: end points outside allocated space.
//	var s thing
//	end = unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + unsafe.Sizeof(s))
//
//	// INVALID: end points outside allocated space.
//	b := make([]byte, n)
//	end = unsafe.Pointer(uintptr(unsafe.Pointer(&b[0])) + uintptr(n))
//
// Note that both conversions must appear in the same expression, with only
// the intervening arithmetic between them:
//
//	// INVALID: uintptr cannot be stored in variable
//	// before conversion back to Pointer.
//	u := uintptr(p)
//	p = unsafe.Pointer(u + offset)
//
// Note that the pointer must point into an allocated object, so it may not be nil.
//
//	// INVALID: conversion of nil pointer
//	u := unsafe.Pointer(nil)
//	p := unsafe.Pointer(uintptr(u) + offset)
//
// (4) Conversion of a Pointer to a uintptr when calling syscall.Syscall.
//
// The Syscall functions in package syscall pass their uintptr arguments directly
// to the operating system, which then may, depending on the details of the call,
// reinterpret some of them as pointers.
// That is, the system call implementation is implicitly converting certain arguments
// back from uintptr to pointer.
//
// If a pointer argument must be converted to uintptr for use as an argument,
// that conversion must appear in the call expression itself:
//
//	syscall.Syscall(SYS_READ, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(n))
//
// The compiler handles a Pointer converted to a uintptr in the argument list of
// a call to a function implemented in assembly by arranging that the referenced
// allocated object, if any, is retained and not moved until the call completes,
// even though from the types alone it would appear that the object is no longer
// needed during the call.
//
// For the compiler to recognize this pattern,
// the conversion must appear in the argument list:
//
//	// INVALID: uintptr cannot be stored in variable
//	// before implicit conversion back to Pointer during system call.
//	u := uintptr(unsafe.Pointer(p))
//	syscall.Syscall(SYS_READ, uintptr(fd), u, uintptr(n))
//
// (5) Conversion of the result of reflect.Value.Pointer or reflect.Value.UnsafeAddr
// from uintptr to Pointer.
//
// Package reflect's Value methods named Pointer and UnsafeAddr return type uintptr
// instead of unsafe.Pointer to keep callers from changing the result to an arbitrary
// type without first importing "unsafe". However, this means that the result is
// fragile and must be converted to Pointer immediately after making the call,
// in the same expression:
//
//	p := (*int)(unsafe.Pointer(reflect.ValueOf(new(int)).Pointer()))
//
// As in the cases above, it is invalid to store the result before the conversion:
//
//	// INVALID: uintptr cannot be stored in variable
//	// before conversion back to Pointer.
//	u := reflect.ValueOf(new(int)).Pointer()
//	p := (*int)(unsafe.Pointer(u))
//
// (6) Conversion of a reflect.SliceHeader or reflect.StringHeader Data field to or from Pointer.
//
// As in the previous case, the reflect data structures SliceHeader and StringHeader
// declare the field Data as a uintptr to keep callers from changing the result to
// an arbitrary type without first importing "unsafe". However, this means that
// SliceHeader and StringHeader are only valid when interpreting the content
// of an actual slice or string value.
//
//	var s string
//	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
//	hdr.Data = uintptr(unsafe.Pointer(p))              // case 6 (this case)
//	hdr.Len = n
//
// In this usage hdr.Data is really an alternate way to refer to the underlying
// pointer in the string header, not a uintptr variable itself.
//
// In general, reflect.SliceHeader and reflect.StringHeader should be used
// only as *reflect.SliceHeader and *reflect.StringHeader pointing at actual
// slices or strings, never as plain structs.
// A program should not declare or allocate variables of these struct types.
//
//	// INVALID: a directly-declared header will not hold Data as a reference.
//	var hdr reflect.StringHeader
//	hdr.Data = uintptr(unsafe.Pointer(p))
//	hdr.Len = n
//	s := *(*string)(unsafe.Pointer(&hdr)) // p possibly already lost
//
type Pointer *ArbitraryType

// Sizeof takes an expression x of any type and returns the size in bytes
// of a hypothetical variable v as if v was declared via var v = x.
// The size does not include any memory possibly referenced by x.
// For instance, if x is a slice, Sizeof returns the size of the slice
// descriptor, not the size of the memory referenced by the slice.
// The return value of Sizeof is a Go constant.
//Sizeof接受任何类型的表达式x，并以字节为单位返回一个假设等于x的变量的大小。
//该大小不包括x可能引用的任何内存。
//例如，如果x是一个切片，Sizeof返回切片描述符的大小，而不是切片引用的内存大小。
//返回值是一个常量

func Sizeof(x ArbitraryType) uintptr

// Offsetof returns the offset within the struct of the field represented by x,
// which must be of the form structValue.field. In other words, it returns the
// number of bytes between the start of the struct and the start of the field.
// The return value of Offsetof is a Go constant.
// 返回x 结构体的偏移，x 必须是一个结构体，返回的是从结构体开始到结束字段之间的大小
func Offsetof(x ArbitraryType) uintptr

// Alignof takes an expression x of any type and returns the required alignment
// of a hypothetical variable v as if v was declared via var v = x.
// It is the largest value m such that the address of v is always zero mod m.
// It is the same as the value returned by reflect.TypeOf(x).Align().
// As a special case, if a variable s is of struct type and f is a field
// within that struct, then Alignof(s.f) will return the required alignment
// of a field of that type within a struct. This case is the same as the
// value returned by reflect.TypeOf(s.f).FieldAlign().
// The return value of Alignof is a Go constant.
//Alignof接受任何类型的表达式x并返回所需的对齐方式
//它是最大值m，因此v的地址总是0 mod m。
//它与reflect返回的值相同。reflect.TypeOf(x).Align()。
//作为特例，如果变量s是结构类型，而f是字段
//在该结构中，Alignof（s.f）将返回所需的对齐方式
//结构中该类型的字段。与reflect.TypeOf(s.f).FieldAlig()效果一样。

func Alignof(x ArbitraryType) uintptr

// The function Add adds len to ptr and returns the updated pointer
// Pointer(uintptr(ptr) + uintptr(len)).
// The len argument must be of integer type or an untyped constant.
// A constant len argument must be representable by a value of type int;
// if it is an untyped constant it is given type int.
// The rules for valid uses of Pointer still apply.
//函数Add将len添加到ptr并返回更新后的指针
//Pointer(uintptr(ptr)+uintptr(len))
//len参数必须是整型或非整型常量。
//常量len参数必须由int类型的值表示；
//如果它是一个非类型化常量，则被赋予int类型。
//指针的有效使用规则仍然适用。
func Add(ptr Pointer, len IntegerType) Pointer

// The function Slice returns a slice whose underlying array starts at ptr
// and whose length and capacity are len.
// Slice(ptr, len) is equivalent to
//
//	(*[len]ArbitraryType)(unsafe.Pointer(ptr))[:]
//
// except that, as a special case, if ptr is nil and len is zero,
// Slice returns nil.
//
// The len argument must be of integer type or an untyped constant.
// A constant len argument must be non-negative and representable by a value of type int;
// if it is an untyped constant it is given type int.
// At run time, if len is negative, or if ptr is nil and len is not zero,
// a run-time panic occurs.
/*
//函数根据切片的指针ptr, 和长度len 返回其底层数组
//Slice(ptr, len)相当于
//(*[len]ArbiryType)(unsafe.Pointer(ptr))[：]
//
//如果ptr为nil，len为零，
//Slice返回nil。
//
//len参数必须是整型或非整型常量。
//常量len参数必须是非负的，并且可以用int类型的值表示；
//如果它是一个非类型化常量，则被赋予int类型。
// 如果len 是一个负数，或者ptr == nil && len != 0 则会触发 run-time panic
*/
func Slice(ptr *ArbitraryType, len IntegerType) []ArbitraryType
