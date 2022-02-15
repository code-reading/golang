// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bufio implements buffered I/O. It wraps an io.Reader or io.Writer
// object, creating another object (Reader or Writer) that also implements
// the interface but provides buffering and some help for textual I/O.
// bufio 包实现了带缓存的 I/O 操作
// 它封装一个 io.Reader 或 io.Writer 对象
// 使其具有缓存和一些文本读写功能
package bufio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	defaultBufSize = 4096 // 默认4kb
)

var (
	ErrInvalidUnreadByte = errors.New("bufio: invalid use of UnreadByte")
	ErrInvalidUnreadRune = errors.New("bufio: invalid use of UnreadRune")
	ErrBufferFull        = errors.New("bufio: buffer full")
	ErrNegativeCount     = errors.New("bufio: negative count")
)

// Buffered input.

// Reader implements buffering for an io.Reader object.
type Reader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int // last byte read for UnreadByte; -1 means invalid
	lastRuneSize int // size of last rune read for UnreadRune; -1 means invalid
}

const minReadBufferSize = 16
const maxConsecutiveEmptyReads = 100

// NewReaderSize returns a new Reader whose buffer has at least the specified
// size. If the argument io.Reader is already a Reader with large enough
// size, it returns the underlying Reader. (基础reader)
// 初始化Reader size
// 如果rd 是reader 且其buf 大于要设置的size  则直接返回
// 否则 检测 最小读buf size
// 否则 new 一个新的Reader 并为其设置size

// 可以根据实际情况 初始化一个合适size的Reader
func NewReaderSize(rd io.Reader, size int) *Reader {
	// Is it already a Reader?
	b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}
	if size < minReadBufferSize {
		size = minReadBufferSize
	}
	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

// NewReader returns a new Reader whose buffer has the default size.
// new 一个Reader  默认bufsize 是 4kb
// 不知道要多大的reader时 可以直接用这个，默认4kb大小， 一般时一个页的大小
// 弊端是 在某些情况下可能太大浪费内存或者太小不够用，
// 最好还是使用NewReaderSize根据情况自定义缓存大小。
func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

// Size returns the size of the underlying buffer in bytes.
func (b *Reader) Size() int { return len(b.buf) }

// Reset discards any buffered data, resets all state, and switches
// the buffered reader to read from r.
// 丢失buffered 数据, 重置reader所有状态， 并将缓存切换到r
// 这里丢弃buffered 数据不是清空buf 而是设置lastByte 和 lastRuneSize 为非法位置
/*
package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	s := strings.NewReader("ABCEFG")
	str := strings.NewReader("123455")
	br := bufio.NewReader(s)
	b, _ := br.ReadString('\n')
	fmt.Println(b)     //ABCEFG
	br.Reset(str)
	b, _ = br.ReadString('\n')
	fmt.Println(b)     //123455
}
*/
func (b *Reader) Reset(r io.Reader) {
	b.reset(b.buf, r)
}

func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{
		buf:          buf,
		rd:           r,
		lastByte:     -1,
		lastRuneSize: -1,
	}
}

var errNegativeRead = errors.New("bufio: reader returned negative count from Read")

// fill reads a new chunk into the buffer.
// 读取一块数据到buffer中
// fill()把剩余未读长度的数据复制到缓存头部并且r重置为0，相当于把未读数据移动到头部。
// 同时尽量从io中读取数据写入缓存，有可能不能写满。
func (b *Reader) fill() {
	// Slide existing data to beginning.
	if b.r > 0 { // 把buf剩余可读的数据复制到最前面
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}

	if b.w >= len(b.buf) { // 缓存已经溢出
		panic("bufio: tried to fill full buffer")
	}
	// maxConsecutiveEmptyReads = 100
	// 最大连续空读次数100次
	// Read new data: try a limited number of times.
	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := b.rd.Read(b.buf[b.w:]) // 从io中读取数据并写入缓存buffer中
		if n < 0 {
			panic(errNegativeRead)
		}
		b.w += n // 更新写入缓存的长度
		if err != nil {
			b.err = err
			return
		}
		if n > 0 {
			return
		} // n == 0时 会循环尝试从io中读取, 最多100次
	}
	b.err = io.ErrNoProgress // 读取了100次， n都为0
}

func (b *Reader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

// Peek returns the next n bytes without advancing the reader. The bytes stop
// being valid at the next read call. If Peek returns fewer than n bytes, it
// also returns an error explaining why the read is short. The error is
// ErrBufferFull if n is larger than b's buffer size.
//
// Calling Peek prevents a UnreadByte or UnreadRune call from succeeding
// until the next read operation.
// Peek 直接返回n字节内容并不会写入到reader中
// 在下一次read 调用前 引用的bytes是有效的， 因为read后会修改游标;
// 如果Peek返回的字节数小于n, 会同时返回一条error说明原因
// 如果 n 比缓存的buffer size 还要大 则会返回一个ErrBufferFull 错误
// Peek 可以使得后继UnreadByte和UnReadRune调用， 直到下一次read操作
/*
	Peek
	1.判断当前缓存数据小于要读取的n字节, 并且缓存没有满 且没有报错 则先去从io中读取数据到缓存
	2.如果读取的n自己大于缓存， 则返回缓存数据并报ErrBufferFull错误
	3.如果缓存数据不够, 则返回可能的缓存数据并报ErrBufferFull
	4.如果要读取的n比缓存小, 且有足够的缓存数据, 则返回对应的缓存数据，不报错;


	不管怎样都会返回缓存可读取的切片但是没有移动读游标，修改返回的切片会影响缓存中的数据。

	Peek返回错误不为空时，一种情况是你Peek的长度比缓存都大，那么数据永远不够，所以传入参数时要注意别比缓存大。
别一种情况是，从io里尝试读数据了但还是准备不够你需要的长度，比如网络tcp的数据一开始没有到达，等下一轮你再调用Peek时可能缓存就足够你读了。
其实就算有错误也会返回你可读取的切片。
如果错误为空，恭喜你，数据都准备好啦。
Peek不会移动读游标，如果直接使用Peek返回的切片可以配合Discard来跳过指定字节的数据不再读取也就是移动读游标。

Peek 返回缓存中前n个字节的切片， 并不会修改游标, 所以在读取缓存 这些数据还是可以读取的，如果修改了Peek返回的字节切片
那么缓存中对应的数据也会被修改
通过Peek 返回的值 可以修改缓存中的数据， 但是不能修改底层io.Reader中的数据

func main() {
s := strings.NewReader("ABCDEFG")
br := bufio.NewReader(s)

b, _ := br.Peek(5)
fmt.Printf("%s\n", b)
// ABCDE

b[0] = 'a'
b, _ = br.Peek(5)
fmt.Printf("%s\n", b)
// aBCDE
}

*/
func (b *Reader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrNegativeCount
	}
	// 初始化 最后一次读取字节的位置和runesize 为invalid
	b.lastByte = -1
	b.lastRuneSize = -1

	// b.w-b.r < n 缓存小于读取的n字节
	// b.w-b.r < len(b.buf) 缓存小于reader的buf大小
	// 且 reader 没有报错
	// 则 直接调用fill() 先填充buffer
	// 剩余可读小于n而且小于缓存时从io里fill数据到缓存
	for b.w-b.r < n && b.w-b.r < len(b.buf) && b.err == nil {
		b.fill() // b.w-b.r < len(b.buf) => buffer is not full
	}
	// 如果 n 要大于 reader 的buffer大小
	if n > len(b.buf) {
		return b.buf[b.r:b.w], ErrBufferFull // 直接返回全部缓存数据, 并且返回ErrBufferFull
	}

	// 0 <= n <= len(b.buf)  n 小于reader 缓存大小
	var err error
	if avail := b.w - b.r; avail < n { // 当前缓存数据小于读取的n字节
		// not enough data in buffer
		n = avail
		err = b.readErr()
		if err == nil {
			err = ErrBufferFull // 如果reader 没有错 则报ErrBufferFull
		}
	}
	return b.buf[b.r : b.r+n], err // 返回可能的缓存数据， 并说明原因
}

// Discard skips the next n bytes, returning the number of bytes discarded.
//
// If Discard skips fewer than n bytes, it also returns an error.
// If 0 <= n <= b.Buffered(), Discard is guaranteed to succeed without
// reading from the underlying io.Reader.
// 跳过n个字节不读取
func (b *Reader) Discard(n int) (discarded int, err error) {
	if n < 0 {
		return 0, ErrNegativeCount
	}
	if n == 0 {
		return
	}
	remain := n
	for {
		skip := b.Buffered() // 返回当前缓存中可读缓存数据的长度
		if skip == 0 {
			b.fill()            // 从io中读取数据到缓存
			skip = b.Buffered() // 再次读取这个长度
		}
		if skip > remain { // 如果长度大于要读取的remain 则设置为remain
			skip = remain
		}
		b.r += skip    // 修改游标，这里为什么不能直接修改为 b.r +=remain ?
		remain -= skip // 修改剩余要丢弃的字节数目
		if remain == 0 {
			return n, nil
		}
		// 如果要丢弃的字节数大于所有的缓存数据, 则会报错
		// 注意此时b.r的游标也修改为最后的数据位置了
		// 返回 n - remain 已经丢弃的数据位置
		if b.err != nil {
			return n - remain, b.readErr()
		}
	}
}

// Read reads data into p.
// It returns the number of bytes read into p.
// The bytes are taken from at most one Read on the underlying Reader,
// hence n may be less than len(p).
// To read exactly len(p) bytes, use io.ReadFull(b, p).
// At EOF, the count will be zero and err will be io.EOF.
// 将缓存数据读取到字节数组中
// 返回读取的字节长度, 没有数据会返回 EOF
// 注意 len(p) 可能会大于 返回的n 大小
// 如果需要精确读取len(p)  可以使用io.ReadFull(b,p)
/*
	Read 具体逻辑
	1.检测待装数据的字节数组长度， 如果为0 则直接返回
	2.检测缓存中是否有数据可以读取, 如果没有,
	2.1 如果p的空间大于缓存的空间， 则直接从io中读取数据到p中并返回;
	2.2 否则先读取数据到缓存中;
	3.将缓存中的数据读取到p中并返回;
	3.如果缓存中有数据则不会去读取io中的数据;

	p的空间有可能被填满，也有可能不满，返回的n说明读取了多少个字节。
	Read尽量先从缓存中读取数据。当前缓存无数据可读时先从io中读取填充到缓存里，
	然后从缓存中复制。返回读取到的数据长度不一定，小于或者等于Read要求的长度。

// Read 从 b 中读出数据到 p 中，返回读出的字节数
// 如果 p 的大小 >= 缓存的总大小，而且缓存不为空
// 则只能读出缓存中的数据，不会从底层 io.Reader 中提取数据
// 如果 p 的大小 >= 缓存的总大小，而且缓存为空
// 则直接从底层 io.Reader 向 p 中读出数据，不经过缓存
// 只有当 b 中无可读数据时，才返回 (0, io.EOF)

func main() {
s := strings.NewReader("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
br := bufio.NewReader(s)
b := make([]byte, 20)

n, err := br.Read(b)
fmt.Printf("%-20s %-2v %v\n", b[:n], n, err)
// ABCDEFGHIJKLMNOPQRST 20 <nil>

n, err = br.Read(b)
fmt.Printf("%-20s %-2v %v\n", b[:n], n, err)
// UVWXYZ1234567890 16 <nil>

n, err = br.Read(b)
fmt.Printf("%-20s %-2v %v\n", b[:n], n, err)
// 0 EOF
}
*/
func (b *Reader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		if b.Buffered() > 0 {
			return 0, nil
		}
		return 0, b.readErr()
	}
	if b.r == b.w { // 缓存中无数据可以读取
		if b.err != nil { // 查看是否报错
			return 0, b.readErr()
		}
		if len(p) >= len(b.buf) { // p的空间大于等于缓存
			// Large read, empty buffer.
			// Read directly into p to avoid copy.
			n, b.err = b.rd.Read(p) // 直接从io中把数据读取到p中
			if n < 0 {
				panic(errNegativeRead)
			}
			if n > 0 {
				b.lastByte = int(p[n-1])
				b.lastRuneSize = -1
			}
			return n, b.readErr()
		}
		// One read.
		// Do not use b.fill, which will loop.
		// 无数据可以读取, 表示buf中数据无用了则重置r和w的游标
		b.r = 0
		b.w = 0
		n, b.err = b.rd.Read(b.buf) // 从io中读取到缓存
		if n < 0 {
			panic(errNegativeRead)
		}
		if n == 0 {
			return 0, b.readErr()
		}
		b.w += n // 更新缓存写入了多少数据
	}

	// copy as much as we can
	n = copy(p, b.buf[b.r:b.w]) //赋值缓存中的数据到p中
	b.r += n                    // 读了多少
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = -1
	return n, nil
}

// ReadByte reads and returns a single byte.
// If no byte is available, returns an error.
// ReadByte 从 b 中读出一个字节并返回
// 如果 b 中无可读数据，则返回一个错误
func (b *Reader) ReadByte() (byte, error) {
	b.lastRuneSize = -1
	for b.r == b.w {
		if b.err != nil { // 如果err 不为nil 读取err ,原err设置为nil
			return 0, b.readErr()
		}
		b.fill() // buffer is empty
	}
	c := b.buf[b.r] // 读取一个字节
	b.r++           // read加一
	b.lastByte = int(c)
	return c, nil
}

// UnreadByte unreads the last byte. Only the most recently read byte can be unread.
//
// UnreadByte returns an error if the most recent method called on the
// Reader was not a read operation. Notably, Peek is not considered a
// read operation.
// UnreadByte 撤消最后一次读出的字节
// 只有最后读出的字节可以被撤消
// 无论任何操作，只要有内容被读出，就可以用 UnreadByte 撤消一个字节
/*
func main() {
s := strings.NewReader("ABCDEFG")
br := bufio.NewReader(s)

c, _ := br.ReadByte()
fmt.Printf("%c\n", c)
// A

c, _ = br.ReadByte()
fmt.Printf("%c\n", c)
// B

br.UnreadByte()
c, _ = br.ReadByte()
fmt.Printf("%c\n", c)
// B
}
*/
func (b *Reader) UnreadByte() error {
	if b.lastByte < 0 || b.r == 0 && b.w > 0 { // 如果读已经是0 并且已经再写了 则没法撤销了
		return ErrInvalidUnreadByte
	}
	// b.r > 0 || b.w == 0
	if b.r > 0 { // 读取大于0 直接撤销
		b.r--
	} else { // 如果读写都为0 则写加一
		// b.r == 0 && b.w == 0
		b.w = 1
	}
	b.buf[b.r] = byte(b.lastByte) // 写入最后一次字节
	b.lastByte = -1
	b.lastRuneSize = -1
	return nil
}

// ReadRune reads a single UTF-8 encoded Unicode character and returns the
// rune and its size in bytes. If the encoded rune is invalid, it consumes one byte
// and returns unicode.ReplacementChar (U+FFFD) with a size of 1.
// ReadRune 从 b 中读出一个 UTF8 编码的字符并返回
// 同时返回该字符的 UTF8 编码长度
// 如果 UTF8 序列无法解码出一个正确的 Unicode 字符
// 则只读出 b 中的一个字节，并返回 U+FFFD 字符，size 返回 1
/*
func main() {
s := strings.NewReader("你好，世界！")
br := bufio.NewReader(s)

c, size, _ := br.ReadRune()
fmt.Printf("%c %v\n", c, size)
// 你 3

c, size, _ = br.ReadRune()
fmt.Printf("%c %v\n", c, size)
// 好 3

br.UnreadRune()
c, size, _ = br.ReadRune()
fmt.Printf("%c %v\n", c, size)
// 好 3
}
*/
func (b *Reader) ReadRune() (r rune, size int, err error) {
	// 注意这里是一个循环， 每次都装数据到缓存buf中
	for b.r+utf8.UTFMax > b.w && !utf8.FullRune(b.buf[b.r:b.w]) && b.err == nil && b.w-b.r < len(b.buf) {
		b.fill() // b.w-b.r < len(buf) => buffer is not full
	}
	b.lastRuneSize = -1
	if b.r == b.w {
		return 0, 0, b.readErr()
	}
	r, size = rune(b.buf[b.r]), 1
	if r >= utf8.RuneSelf {
		r, size = utf8.DecodeRune(b.buf[b.r:b.w])
	}
	b.r += size
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = size // 标志读取的rune size, 可用于 UnreadRune() 撤销
	return r, size, nil
}

// UnreadRune unreads the last rune. If the most recent method called on
// the Reader was not a ReadRune, UnreadRune returns an error. (In this
// regard it is stricter than UnreadByte, which will unread the last byte
// from any read operation.)
// UnreadRune 撤消最后一次读出的 Unicode 字符
// 如果最后一次执行的不是 ReadRune 操作，则返回一个错误
// 因此，UnreadRune 比 UnreadByte 更严格
func (b *Reader) UnreadRune() error {
	if b.lastRuneSize < 0 || b.r < b.lastRuneSize {
		return ErrInvalidUnreadRune
	}
	b.r -= b.lastRuneSize
	b.lastByte = -1
	b.lastRuneSize = -1
	return nil
}

// Buffered returns the number of bytes that can be read from the current buffer.
// 返回当前缓存中缓存数据的长度
/*
func main() {
s := strings.NewReader("你好，世界！")
br := bufio.NewReader(s)

fmt.Println(br.Buffered())
// 0

br.Peek(1) // 通过触发fill() 填充缓存
fmt.Println(br.Buffered())
// 18
}
*/
func (b *Reader) Buffered() int { return b.w - b.r }

// ReadSlice reads until the first occurrence of delim in the input,
// returning a slice pointing at the bytes in the buffer.
// The bytes stop being valid at the next read.
// If ReadSlice encounters an error before finding a delimiter,
// it returns all the data in the buffer and the error itself (often io.EOF).
// ReadSlice fails with error ErrBufferFull if the buffer fills without a delim.
// Because the data returned from ReadSlice will be overwritten
// by the next I/O operation, most clients should use
// ReadBytes or ReadString instead.
// ReadSlice returns err != nil if and only if line does not end in delim.

// ReadSlice 在 b 中查找 delim 并返回 delim 及其之前的所有数据的切片
// 该操作会读出数据，返回的切片是已读出数据的引用
// 切片中的数据在下一次读取操作之前是有效的
//
// 如果 ReadSlice 在找到 delim 之前遇到错误
// 则读出缓存中的所有数据并返回，同时返回遇到的错误（通常是 io.EOF）
// 如果在整个缓存中都找不到 delim，则 err 返回 ErrBufferFull
// 如果 ReadSlice 能找到 delim，则 err 始终返回 nil
//
// 因为返回的切片中的数据有可能被下一次读写操作修改
// 因此大多数操作应该使用 ReadBytes 或 ReadString，它们返回的不是数据引用

// ReadSlice 会把数据读取出来， 之后缓存中没有这个数据了
// ReadSlice 可供用户指定分隔符读取, 每次都是搜索到第一次遇到delim结束
/*
func main() {
s := strings.NewReader("ABC DEF GHI JKL")
br := bufio.NewReader(s)

w, _ := br.ReadSlice(' ')
fmt.Printf("%q\n", w)
// "ABC "

w, _ = br.ReadSlice(' ')
fmt.Printf("%q\n", w)
// "DEF "

w, _ = br.ReadSlice(' ')
fmt.Printf("%q\n", w)
// "GHI "
}
*/
func (b *Reader) ReadSlice(delim byte) (line []byte, err error) {
	s := 0 // search start index
	for {
		// Search buffer.
		if i := bytes.IndexByte(b.buf[b.r+s:b.w], delim); i >= 0 { // i > 0 表示找到了
			i += s
			line = b.buf[b.r : b.r+i+1] // 读取delim 之前的数据到line中， 注意line 是[]byte
			b.r += i + 1                // 修改了读的游标
			break
		}

		// Pending error?
		if b.err != nil {
			line = b.buf[b.r:b.w] // 有错误 返回所有缓存数据
			b.r = b.w             // 设置游标到写的位置，表示缓存中没有数据了
			err = b.readErr()     // 读取错误， 并重置b的错误为nil
			break
		}

		// Buffer full?
		if b.Buffered() >= len(b.buf) { // 如果缓存数据大于 b的缓存大小
			b.r = b.w
			line = b.buf // 直接返回 line = b.buf
			err = ErrBufferFull
			break
		}

		s = b.w - b.r // do not rescan area we scanned before

		b.fill() // buffer is not full
	}

	// Handle last byte, if any.
	if i := len(line) - 1; i >= 0 {
		b.lastByte = int(line[i]) // 记录最后一个byte, 可用于unReadByte
		b.lastRuneSize = -1
	}

	return
}

// ReadLine is a low-level line-reading primitive. Most callers should use
// ReadBytes('\n') or ReadString('\n') instead or use a Scanner.
//
// ReadLine tries to return a single line, not including the end-of-line bytes.
// If the line was too long for the buffer then isPrefix is set and the
// beginning of the line is returned. The rest of the line will be returned
// from future calls. isPrefix will be false when returning the last fragment
// of the line. The returned buffer is only valid until the next call to
// ReadLine. ReadLine either returns a non-nil line or it returns an error,
// never both.
//
// The text returned from ReadLine does not include the line end ("\r\n" or "\n").
// No indication or error is given if the input ends without a final line end.
// Calling UnreadByte after ReadLine will always unread the last byte read
// (possibly a character belonging to the line end) even if that byte is not
// part of the line returned by ReadLine.
// 与ReadSlice 不同的是, ReadLine 是以换行符来分割
// ReadLine 是一个低级的原始的行读取操作
// 大多数情况下，应该使用 ReadBytes('\n') 或 ReadString('\n')
// 或者使用一个 Scanner
//
// ReadLine 通过调用 ReadSlice 方法实现，返回的也是缓存的切片
// ReadLine 尝试返回一个单行数据，不包括行尾标记（\n 或 \r\n）
// 如果在缓存中找不到行尾标记，则设置 isPrefix 为 true，表示查找未完成
// 同时读出缓存中的数据并作为切片返回
// 只有在当前缓存中找到行尾标记，才将 isPrefix 设置为 false，表示查找完成
// 可以多次调用 ReadLine 来读出一行
// 返回的数据在下一次读取操作之前是有效的
// 如果 ReadLine 无法获取任何数据，则返回一个错误信息（通常是 io.EOF）
/*
func main() {
s := strings.NewReader("ABC\nDEF\r\nGHI\r\nJKL")
br := bufio.NewReader(s)

w, isPrefix, _ := br.ReadLine()
fmt.Printf("%q %v\n", w, isPrefix)
// "ABC" false

w, isPrefix, _ = br.ReadLine()
fmt.Printf("%q %v\n", w, isPrefix)
// "DEF" false

w, isPrefix, _ = br.ReadLine()
fmt.Printf("%q %v\n", w, isPrefix)
// "GHI" false
}
*/
func (b *Reader) ReadLine() (line []byte, isPrefix bool, err error) {
	line, err = b.ReadSlice('\n')
	if err == ErrBufferFull { // 在找到\n之前 已经缓存的数据大于b本身的buffer size
		// Handle the case where "\r\n" straddles the buffer.
		if len(line) > 0 && line[len(line)-1] == '\r' {
			// Put the '\r' back on buf and drop it from line.
			// Let the next call to ReadLine check for "\r\n".
			if b.r == 0 {
				// should be unreachable
				panic("bufio: tried to rewind past start of buffer")
			}
			b.r--
			line = line[:len(line)-1] // 去掉\r
		}
		return line, true, nil // 直接返回, isPrefix= true 表示没有找到行尾标记\n 就已经遇到了ErrBufferFull
	}

	if len(line) == 0 { // 如果没有找到数据
		if err != nil {
			line = nil
		}
		return
	}
	err = nil

	if line[len(line)-1] == '\n' {
		drop := 1
		if len(line) > 1 && line[len(line)-2] == '\r' {
			drop = 2
		}
		line = line[:len(line)-drop] // 去掉\r\n
	}
	return
}

// collectFragments reads until the first occurrence of delim in the input. It
// returns (slice of full buffers, remaining bytes before delim, total number
// of bytes in the combined first two elements, error).
// The complete result is equal to
// `bytes.Join(append(fullBuffers, finalFragment), nil)`, which has a
// length of `totalLen`. The result is structured in this way to allow callers
// to minimize allocations and copies.
// 收集碎片, 直到遇到第一个输入的delim 为止.
// 这种返回结构方式可以是的调用者以最小的代价分配和复制这些结果
func (b *Reader) collectFragments(delim byte) (fullBuffers [][]byte, finalFragment []byte, totalLen int, err error) {
	var frag []byte
	// Use ReadSlice to look for delim, accumulating full buffers.
	for { // 一直到读到一个delim 或者引起 ErrBufferFull 为止
		var e error
		frag, e = b.ReadSlice(delim) // ReadSlice 能找到 delim，则 err 始终返回 nil
		if e == nil {                // got final fragment
			break
		}
		if e != ErrBufferFull { // unexpected error // 如果没有找到delim
			err = e
			break
		}

		// Make a copy of the buffer.
		buf := make([]byte, len(frag))
		copy(buf, frag)
		fullBuffers = append(fullBuffers, buf)
		totalLen += len(buf)
	}

	totalLen += len(frag)
	return fullBuffers, frag, totalLen, err
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadBytes encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often io.EOF).
// ReadBytes returns err != nil if and only if the returned data does not end in
// delim.
// For simple uses, a Scanner may be more convenient.

// ReadBytes 在 b 中查找 delim 并读出 delim 及其之前的所有数据
// 如果 ReadBytes 在找到 delim 之前遇到错误
// 则返回遇到错误之前的所有数据，同时返回遇到的错误（通常是 io.EOF）
// 只有当 ReadBytes 找不到 delim 时，err 才不为 nil
// 对于简单的用途，使用 Scanner 可能更方便
// 注意 ReadBytes 返回是带了delim的， 除非没有找到delim
/*
func main() {
s := strings.NewReader("ABC DEF GHI JKL")
br := bufio.NewReader(s)

w, _ := br.ReadBytes(' ')
fmt.Printf("%q\n", w)
// "ABC "

w, _ = br.ReadBytes(' ')
fmt.Printf("%q\n", w)
// "DEF "

w, _ = br.ReadBytes(' ')
fmt.Printf("%q\n", w)
// "GHI "
}
*/
func (b *Reader) ReadBytes(delim byte) ([]byte, error) {
	full, frag, n, err := b.collectFragments(delim)
	// Allocate new buffer to hold the full pieces and the fragment.
	buf := make([]byte, n) // 申请n buf大小
	n = 0
	// Copy full pieces and fragment in.
	for i := range full {
		n += copy(buf[n:], full[i]) // 将full copy 到buf中
	}
	copy(buf[n:], frag) // 最后将frag copy 到buf中
	return buf, err
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
// If ReadString encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often io.EOF).
// ReadString returns err != nil if and only if the returned data does not end in
// delim.
// For simple uses, a Scanner may be more convenient.
// ReadString 功能同 ReadBytes，只不过返回的是一个字符串
/*
func main() {
s := strings.NewReader("ABC DEF GHI JKL")
br := bufio.NewReader(s)

w, _ := br.ReadString(' ')
fmt.Printf("%q\n", w)
// "ABC "

w, _ = br.ReadString(' ')
fmt.Printf("%q\n", w)
// "DEF "

w, _ = br.ReadString(' ')
fmt.Printf("%q\n", w)
// "GHI "
}
*/
func (b *Reader) ReadString(delim byte) (string, error) {
	full, frag, n, err := b.collectFragments(delim)
	// Allocate new buffer to hold the full pieces and the fragment.
	var buf strings.Builder
	buf.Grow(n)
	// Copy full pieces and fragment in.
	for _, fb := range full {
		buf.Write(fb)
	}
	buf.Write(frag)
	return buf.String(), err
}

// WriteTo implements io.WriterTo.
// This may make multiple calls to the Read method of the underlying Reader.
// If the underlying reader supports the WriteTo method,
// this calls the underlying WriteTo without buffering.
// 将缓存数据写入到w中
/*
func main() {
s := strings.NewReader("ABCEFG")
br := bufio.NewReader(s)
b := bytes.NewBuffer(make([]byte, 0))

br.WriteTo(b)
fmt.Printf("%s\n", b)
// ABCEFG
}
*/
func (b *Reader) WriteTo(w io.Writer) (n int64, err error) {
	n, err = b.writeBuf(w) // 将缓存数据写入到w中
	if err != nil {
		return
	}

	if r, ok := b.rd.(io.WriterTo); ok {
		m, err := r.WriteTo(w) // 将io中的数据写入到w中
		n += m
		return n, err
	}

	if w, ok := w.(io.ReaderFrom); ok {
		m, err := w.ReadFrom(b.rd)
		n += m
		return n, err
	}

	if b.w-b.r < len(b.buf) { // 缓存数据不够
		b.fill() // buffer not full
	}

	for b.r < b.w {
		// b.r < b.w => buffer is not empty
		m, err := b.writeBuf(w)
		n += m
		if err != nil {
			return n, err
		}
		b.fill() // buffer is empty
	}

	if b.err == io.EOF {
		b.err = nil
	}

	return n, b.readErr()
}

var errNegativeWrite = errors.New("bufio: writer returned negative count from Write")

// writeBuf writes the Reader's buffer to the writer.
// 将reader buffer 写入到w
func (b *Reader) writeBuf(w io.Writer) (int64, error) {
	n, err := w.Write(b.buf[b.r:b.w])
	if n < 0 {
		panic(errNegativeWrite)
	}
	b.r += n
	return int64(n), err
}

// buffered output

// Writer implements buffering for an io.Writer object.
// If an error occurs writing to a Writer, no more data will be
// accepted and all subsequent writes, and Flush, will return the error.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.Writer.
// Writer 实现了带缓存的 io.Writer 对象
// 如果在向 Writer 中写入数据的过程中遇到错误
// 则 Writer 不会再接受任何数据
// 而且后续的写入操作都将返回错误信息
type Writer struct {
	err error
	buf []byte
	n   int
	wr  io.Writer
}

// NewWriterSize returns a new Writer whose buffer has at least the specified
// size. If the argument io.Writer is already a Writer with large enough
// size, it returns the underlying Writer.
// NewWriterSize 将 wr 封装成一个拥有 size 大小缓存的 bufio.Writer 对象
// 如果 wr 的基类型就是 bufio.Writer 类型，而且拥有足够的缓存
// 则直接将 wr 转换为基类型并返回
func NewWriterSize(w io.Writer, size int) *Writer {
	// Is it already a Writer?
	b, ok := w.(*Writer)
	if ok && len(b.buf) >= size {
		return b
	}
	if size <= 0 {
		size = defaultBufSize
	}
	return &Writer{
		buf: make([]byte, size),
		wr:  w,
	}
}

// NewWriter returns a new Writer whose buffer has the default size.
// NewWriter 相当于 NewWriterSize(wr, 4096)
func NewWriter(w io.Writer) *Writer {
	return NewWriterSize(w, defaultBufSize)
}

// Size returns the size of the underlying buffer in bytes.
func (b *Writer) Size() int { return len(b.buf) }

// Reset discards any unflushed buffered data, clears any error, and
// resets b to write its output to w.
// 重置所有错误信息（如果有错误的话), 并且丢弃所有未刷盘的缓存数据
// Reset丢弃任何没有写入的缓存数据，清除任何错误并且重新将b指定它的输出结果指向w
/*
package main

import (
	"bufio"
	"bytes"
	"fmt"
)

func main() {
	b := bytes.NewBuffer(make([]byte, 0))
	bw := bufio.NewWriter(b)
	bw.WriteString("123")
	c := bytes.NewBuffer(make([]byte, 0))
	bw.Reset(c)
	bw.WriteString("456")
	bw.Flush()
	fmt.Println(b)       //输出为空
	fmt.Println(c)　　//输出456
}
*/
func (b *Writer) Reset(w io.Writer) {
	b.err = nil
	b.n = 0
	b.wr = w
}

// Flush writes any buffered data to the underlying io.Writer.
// Flush 将缓存中的数据提交到底层的 io.Writer 中
func (b *Writer) Flush() error {
	if b.err != nil {
		return b.err
	}
	if b.n == 0 {
		return nil
	}
	n, err := b.wr.Write(b.buf[0:b.n]) // 将缓存数据全部写入 wr中
	if n < b.n && err == nil {
		err = io.ErrShortWrite // 没有写满
	}
	if err != nil { // 写错误
		if n > 0 && n < b.n {
			copy(b.buf[0:b.n-n], b.buf[n:b.n]) // copy
		}
		b.n -= n
		b.err = err
		return err
	}
	b.n = 0 // n =0 表示清空buffer
	return nil
}

// Available returns how many bytes are unused in the buffer.
// Available 返回缓存中的可以空间
func (b *Writer) Available() int { return len(b.buf) - b.n }

// Buffered returns the number of bytes that have been written into the current buffer.
// Buffered 返回缓存中未提交的数据长度
func (b *Writer) Buffered() int { return b.n }

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
// Write 将 p 中的数据写入 b 中，返回写入的字节数
// 如果写入的字节数小于 p 的长度，则返回一个错误信息
func (b *Writer) Write(p []byte) (nn int, err error) {
	for len(p) > b.Available() && b.err == nil { // 如果写入数据长度大于能写的缓存大小
		var n int
		if b.Buffered() == 0 { // 当前没有要提交的写入数据
			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			n, b.err = b.wr.Write(p) // 则直接将p 写入到wr中
		} else { // 之前存在写入未提交的数据, 则copy 到buf中
			n = copy(b.buf[b.n:], p)
			b.n += n
			b.Flush() // 将buf刷入wr中
		}
		nn += n   // 记录写入的字节大小
		p = p[n:] // 丢弃已写入的数据
	}
	if b.err != nil {
		return nn, b.err
	}
	n := copy(b.buf[b.n:], p) // 如果当前缓存可以直接写入， 则直接将p copy到缓存中
	b.n += n
	nn += n
	return nn, nil
}

// WriteByte writes a single byte.
// 写入单个字节
func (b *Writer) WriteByte(c byte) error {
	if b.err != nil {
		return b.err
	}
	if b.Available() <= 0 && b.Flush() != nil {
		return b.err
	}
	b.buf[b.n] = c
	b.n++
	return nil
}

// WriteRune writes a single Unicode code point, returning
// the number of bytes written and any error.
// 写入单个rune
/*
func main() {
    b := bytes.NewBuffer(make([]byte, 0))
    bw := bufio.NewWriter(b)
    bw.WriteByte('H')
    bw.WriteByte('e')
    bw.WriteByte('l')
    bw.WriteByte('l')
    bw.WriteByte('o') // 写入单个字节
    bw.WriteByte(' ')
    bw.WriteRune('世') // 写入单个字符
    bw.WriteRune('界')
    bw.WriteRune('！')
    bw.Flush()
    fmt.Println(b) // Hello 世界！

}

*/
func (b *Writer) WriteRune(r rune) (size int, err error) {
	// Compare as uint32 to correctly handle negative runes.
	if uint32(r) < utf8.RuneSelf {
		err = b.WriteByte(byte(r))
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	if b.err != nil {
		return 0, b.err
	}
	n := b.Available()
	if n < utf8.UTFMax {
		if b.Flush(); b.err != nil {
			return 0, b.err
		}
		n = b.Available()
		if n < utf8.UTFMax {
			// Can only happen if buffer is silly small.
			return b.WriteString(string(r))
		}
	}
	size = utf8.EncodeRune(b.buf[b.n:], r)
	b.n += size
	return size, nil
}

// WriteString writes a string.
// It returns the number of bytes written.
// If the count is less than len(s), it also returns an error explaining
// why the write is short.
// WriteString 同 Write，只不过写入的是字符串
/*
func main() {
    b := bytes.NewBuffer(make([]byte, 0))
    bw := bufio.NewWriter(b)
    fmt.Println(bw.Available()) // 4096
    fmt.Println(bw.Buffered())  // 0

    bw.WriteString("ABCDEFGH")
    fmt.Println(bw.Available()) // 4088
    fmt.Println(bw.Buffered())  // 8
    fmt.Printf("%q\n", b)       // ""

    bw.Flush()
    fmt.Println(bw.Available()) // 4096
    fmt.Println(bw.Buffered())  // 0
    fmt.Printf("%q\n", b)       // "ABCEFG"
}
*/
func (b *Writer) WriteString(s string) (int, error) {
	nn := 0
	for len(s) > b.Available() && b.err == nil {
		n := copy(b.buf[b.n:], s)
		b.n += n
		nn += n
		s = s[n:]
		b.Flush()
	}
	if b.err != nil {
		return nn, b.err
	}
	n := copy(b.buf[b.n:], s)
	b.n += n
	nn += n
	return nn, nil
}

// ReadFrom implements io.ReaderFrom. If the underlying writer
// supports the ReadFrom method, and b has no buffered data yet,
// this calls the underlying ReadFrom without buffering.
// ReadFrom 实现了 io.ReaderFrom 接口
// 从reader中读取数据到指定的buffer中
/*
func main() {
b := bytes.NewBuffer(make([]byte, 0))
s := strings.NewReader("Hello 世界！")
bw := bufio.NewWriter(b)
bw.ReadFrom(s)
//bw.Flush()            //ReadFrom无需使用Flush，其自己已经写入．
fmt.Println(b) // Hello 世界！
}
*/
func (b *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.Buffered() == 0 { // 缓存中没有数据
		if w, ok := b.wr.(io.ReaderFrom); ok {
			n, err = w.ReadFrom(r)
			b.err = err
			return n, err
		}
	}
	var m int
	for {
		if b.Available() == 0 { // 没有可写的缓存空间了
			if err1 := b.Flush(); err1 != nil {
				return n, err1
			}
		}
		nr := 0
		for nr < maxConsecutiveEmptyReads { // 最大连续空读次数 100
			m, err = r.Read(b.buf[b.n:])
			if m != 0 || err != nil {
				break
			}
			nr++
		}
		if nr == maxConsecutiveEmptyReads {
			return n, io.ErrNoProgress
		}
		b.n += m
		n += int64(m)
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		// If we filled the buffer exactly, flush preemptively.
		if b.Available() == 0 {
			err = b.Flush()
		} else {
			err = nil
		}
	}
	return n, err
}

// buffered input and output

// ReadWriter stores pointers to a Reader and a Writer.
// It implements io.ReadWriter.
// ReadWriter 集成了 bufio.Reader 和 bufio.Writer
// 它实现了 io.ReadWriter 接口
type ReadWriter struct {
	*Reader
	*Writer
}

// NewReadWriter allocates a new ReadWriter that dispatches to r and w.
/*
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func main() {
	b := bytes.NewBuffer(make([]byte, 0))
	bw := bufio.NewWriter(b)
	s := strings.NewReader("123")
	br := bufio.NewReader(s)
	rw := bufio.NewReadWriter(br, bw)
	p, _ := rw.ReadString('\n')
	fmt.Println(string(p))              //123
	rw.WriteString("asdf")
	rw.Flush()
	fmt.Println(b)                          //asdf
}
*/
func NewReadWriter(r *Reader, w *Writer) *ReadWriter {
	return &ReadWriter{r, w}
}
