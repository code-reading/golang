## sync.Pool 源码分析

sync.Pool 是 Golang 内置的对象池技术，可用于缓存临时对象，避免因频繁建立临时对象所带来的消耗以及对 GC 造成的压力。

需要注意的是，sync.Pool 缓存的对象随时可能被无通知的清除，因此不能将 sync.Pool 用于存储持久对象的场景。

sync.Pool 作为 goroutine 内置的官方库，其设计非常精妙。sync.Pool 不仅是并发安全的，而且实现了 lock free，里面有许多值得学习的知识点。

## sync.Pool 主要特点
- 利用 GMP 的特性，为每个 P 创建了一个本地对象池 poolLocal，尽量减少并发冲突。

- 每个 poolLocal 都有一个 private 对象，优先存取 private 对象，可以避免进入复杂逻辑。

- 在 Get 和 Put 期间，利用 pin 锁定当前 P，防止 goroutine 被抢占，造成程序混乱。

- 在获取对象期间，利用对象窃取的机制，从其他 P 的本地对象池以及 victim 中获取对象。

- 充分利用 CPU Cache 特性，提升程序性能。