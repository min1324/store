package atomany

// 本文件包含 atomany 包的文档和最佳实践

/*
atomany - 高性能 Go 原子操作库

概述
----
atomany 是一个高性能的 Go 原子操作库，扩展了标准库 sync/atomic.Value 的功能。
它提供了两种原子值类型：Value 和 Entry，在保留原子操作语义的同时，
增加了对 nil 值和任意类型数据的支持。

核心特性
--------
1. 原子操作保证：提供完整的原子操作语义（Load、Store、Swap、CompareAndSwap）
2. nil 值支持：Value 和 Entry 均支持存储和读取 nil 值
3. 类型灵活性：Entry 支持存储任意类型的数据，可在运行时动态切换类型
4. 高性能：基于 unsafe.Pointer 实现，性能接近标准库
5. 线程安全：所有操作都是线程安全的，无需额外的同步机制
6. 接口统一：Value 和 Entry 实现统一的 Interface 接口

类型选择指南
-----------

使用 Value 的场景：
- 需要类型安全保证
- 存储相同类型的数据
- 性能敏感的场景
- 需要编译时类型检查

使用 Entry 的场景：
- 需要存储不同类型的数据
- 需要在运行时动态切换数据类型
- 需要直接访问底层指针
- 读多写少的场景

使用 TypedValue 的场景：
- 需要类型安全但不想使用类型断言
- 需要泛型支持
- Go 1.18+ 环境

使用 OptimizedEntry 的场景：
- 高并发环境
- 需要减少 CPU 使用
- CompareAndSwap 操作频繁

使用带统计信息的场景：
- 需要监控原子操作性能
- 需要分析操作模式
- 调试和性能优化

最佳实践
--------

1. 初始化
   - Value 首次 Store 必须是非 nil 值
   - 如需从 nil 开始初始化，使用 CompareAndSwap(nil, x)
   - 使用 NewTypedValue 创建类型安全的原子值

2. 并发安全
   - 避免频繁 Store 操作，减少内存分配和 GC 压力
   - 使用 CompareAndSwap 进行条件更新
   - 合理使用 Swap，比 Store + Load 更高效
   - 使用 Value 时确保类型一致性

3. 性能优化
   - 读多写少场景使用 Entry
   - 写多读少场景使用 Value
   - 高并发场景使用 OptimizedEntry
   - 使用 GetOrStore 避免重复初始化

4. 错误处理
   - 使用 TryLoad 检查是否已初始化
   - 使用 TryStore 避免覆盖已有值
   - 捕获并处理 panic（类型不匹配等）
   - 使用预定义的错误类型进行错误检查

5. 监控和调试
   - 使用 EntryWithStats 或 ValueWithStats 收集统计信息
   - 定期检查统计信息以发现性能问题
   - 在生产环境中谨慎使用统计功能

常见使用场景
----------

1. 配置热更新
   使用 Value 或 TypedValue 存储配置，实现无锁的配置更新。

2. 状态机实现
   使用 Entry 存储状态，支持不同类型的状态转换。

3. 缓存实现
   使用 Entry 或 GetOrStore 实现简单的原子缓存。

4. 单例模式
   使用 GetOrStore 实现线程安全的单例。

5. 标志位管理
   使用 Value 存储布尔标志，实现无锁的标志切换。

性能考虑
--------

1. 内存分配
   - Entry 的 Store 操作有 1 次内存分配
   - Value 的 Store 操作有 0 次内存分配
   - 频繁 Store 会增加 GC 压力

2. CPU 使用
   - 高并发下的自旋会增加 CPU 使用
   - 使用 OptimizedEntry 可以减少 CPU 使用
   - CompareAndSwap 在竞争激烈时性能下降

3. 缓存友好性
   - Value 更适合 CPU 缓存
   - Entry 的指针间接访问可能影响缓存性能

4. 并发度
   - 读操作可以完全并发
   - 写操作需要原子操作，有一定开销
   - 高并发写场景考虑使用其他同步原语

注意事项
--------

1. 类型安全
   - Value 要求类型一致性，类型不匹配会 panic
   - Entry 不检查类型，使用时需自行确保类型正确
   - TypedValue 提供编译时类型检查

2. nil 值
   - Value 首次 Store 不能为 nil
   - Entry 可以随时存储 nil
   - Load 可能返回 nil，需要检查

3. 内存可见性
   - 所有操作都保证内存可见性
   - 不需要额外的同步机制
   - 与其他同步原语混用时需谨慎

4. 性能权衡
   - 类型安全 vs 灵活性
   - 内存分配 vs 操作速度
   - CPU 使用 vs 响应延迟

版本兼容性
----------

- Go 1.18+ 支持泛型（TypedValue）
- Go 1.21+ 推荐使用
- 与标准库 sync/atomic.Value 兼容

相关资源
--------

- Go 标准库 sync/atomic 包
- Go 内存模型文档
- 并发编程最佳实践

示例代码
--------

参见 examples.go 文件中的详细示例。
*/
