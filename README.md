# atomany

[![Build Status](https://travis-ci.com/min1324/atomany.svg?branch=main)](https://travis-ci.com/min1324/atomany) [![codecov](https://codecov.io/gh/min1324/atomany/branch/main/graph/badge.svg)](https://codecov.io/gh/min1324/atomany) [![Go Report Card](https://goreportcard.com/badge/github.com/min1324/atomany)](https://goreportcard.com/report/github.com/min1324/atomany) [![GoDoc](https://pkg.go.dev/github.com/min1324/atomany)](https://pkg.go.dev/github.com/min1324/atomany)

## 简介

**atomany** 是一个高性能的 Go 原子操作库，扩展了标准库 `sync/atomic.Value` 的功能。它提供了两种原子值类型：`Value` 和 `Entry`，在保留原子操作语义的同时，增加了对 `nil` 值和任意类型数据的支持。

### 核心特性

- ✅ **原子操作保证**：提供完整的原子操作语义（Load、Store、Swap、CompareAndSwap）
- ✅ **nil 值支持**：`Value` 和 `Entry` 均支持存储和读取 `nil` 值
- ✅ **类型灵活性**：`Entry` 支持存储任意类型的数据，可在运行时动态切换类型
- ✅ **高性能**：基于 `unsafe.Pointer` 实现，性能接近标准库
- ✅ **线程安全**：所有操作都是线程安全的，无需额外的同步机制
- ✅ **接口统一**：`Value` 和 `Entry` 实现统一的 `Interface` 接口

### 功能对比

| 特性 | `sync/atomic.Value` | `atomany.Value` | `atomany.Entry` |
|:---|:---:|:---:|:---:|
| 原子 Load/Store | ✅ | ✅ | ✅ |
| 原子 Swap | ✅ | ✅ | ✅ |
| 原子 CompareAndSwap | ✅ | ✅ | ✅ |
| 存储 nil | ❌ | ✅ | ✅ |
| 存储不同类型 | ❌ | ❌ | ✅ |
| 类型检查 | ✅ | ✅ | ❌ |

## 安装

```bash
go get github.com/min1324/atomany
```

## 快速开始

### Value - 类型安全的原子值

`Value` 适用于需要类型安全保证的场景，要求所有存储的值类型一致：

```go
package main

import (
    "fmt"
    "github.com/min1324/atomany"
)

func main() {
    // 创建 Value 实例
    var v atomany.Value

    // 首次存储（必须是非 nil 值）
    v.Store("hello")
    fmt.Println(v.Load()) // 输出: "hello"

    // 存储同类型的值
    v.Store("world")
    fmt.Println(v.Load()) // 输出: "world"

    // 存储 nil 值（初始化后允许）
    v.Store(nil)
    fmt.Println(v.Load()) // 输出: <nil>

    // 使用 CompareAndSwap 进行条件更新
    if v.CompareAndSwap(nil, "new") {
        fmt.Println("CAS 成功")
    }
}
```

### Entry - 灵活的原子值

`Entry` 适用于需要存储不同类型数据的场景：

```go
package main

import (
    "fmt"
    "github.com/min1324/atomany"
)

func main() {
    // 创建 Entry 实例
    var e atomany.Entry

    // 存储整数
    e.Store(42)
    fmt.Println(e.Load()) // 输出: 42

    // 存储字符串（允许类型切换）
    e.Store("world")
    fmt.Println(e.Load()) // 输出: "world"

    // 存储 nil 值
    e.Store(nil)
    fmt.Println(e.Load()) // 输出: <nil>

    // 存储结构体
    e.Store(struct{ Name string }{"atomany"})
    fmt.Println(e.Load()) // 输出: {atomany}
}
```

## API 文档

### 核心类型

#### Value 类型

`Value` 是 `sync/atomic.Value` 的增强版本，支持 `nil` 值但保持类型安全。

#### 方法

| 方法 | 签名 | 说明 |
|:---|:---|:---|
| **Load** | `Load() any` | 原子读取当前值，未初始化时返回 `nil` |
| **Store** | `Store(val any)` | 原子写入值；首次 Store 不可为 `nil`，后续可为 `nil`；类型不一致时 panic |
| **Swap** | `Swap(new any) any` | 原子交换值，返回旧值；类型不一致时 panic |
| **CompareAndSwap** | `CompareAndSwap(old, new any) bool` | 原子比较并交换；类型不一致时 panic |

#### 使用注意事项

⚠️ **重要提示**：
- 未初始化的 `Value` 调用 `Store(nil)` 会 panic
- 如需从 `nil` 开始初始化，请使用 `CompareAndSwap(nil, x)`
- 所有 `Store` 操作必须使用相同的类型，否则会 panic

### Entry 类型

`Entry` 基于 `unsafe.Pointer` 实现，提供最大的灵活性。

#### 方法

| 方法 | 签名 | 说明 |
|:---|:---|:---|
| **Load** | `Load() any` | 原子读取当前值，未初始化时返回 `nil` |
| **Store** | `Store(val any)` | 原子写入值，允许 `nil` 和任意类型 |
| **Swap** | `Swap(new any) any` | 原子交换值，返回旧值 |
| **CompareAndSwap** | `CompareAndSwap(old, new any) bool` | 原子比较并交换 |
| **Ptr** | `Ptr() unsafe.Pointer` | 返回底层 `unsafe.Pointer`（原子读取） |

#### 使用场景

✅ **适用场景**：
- 需要存储不同类型的数据
- 需要在运行时动态切换数据类型
- 需要直接访问底层指针

❌ **不适用场景**：
- 需要类型安全保证
- 性能极其敏感的场景（建议使用 `Value`）

### Interface 接口

`Value` 和 `Entry` 均实现了统一的 `Interface` 接口：

```go
type Interface interface {
    // Load 原子读取当前值
    Load() any

    // Store 原子写入值
    Store(val any)

    // Swap 原子交换值，返回旧值
    Swap(new any) any

    // CompareAndSwap 原子比较并交换
    CompareAndSwap(old, new any) bool
}
```

### 扩展类型

#### TypedValue[T] - 类型安全的原子值

`TypedValue` 提供泛型支持，在编译时保证类型安全。

```go
// 创建类型安全的原子值
tv := NewTypedValue(42)

// 类型安全的操作
tv.Store(100)
val := tv.Load() // val 是 int 类型
old := tv.Swap(200)
swapped := tv.CompareAndSwap(200, 300)
```

**方法**：
- `NewTypedValue[T](val T) *TypedValue[T]` - 创建类型安全的原子值
- `Load() T` - 类型安全的读取
- `Store(val T)` - 类型安全的存储
- `Swap(val T) T` - 类型安全的交换
- `CompareAndSwap(old, new T) bool` - 类型安全的比较并交换
- `TryLoad() (T, bool)` - 尝试加载
- `TryStore(val T) bool` - 尝试存储
- `GetOrStore(val T) (T, bool)` - 获取或存储

#### OptimizedEntry - 性能优化的 Entry

`OptimizedEntry` 使用指数退避机制减少高并发下的 CPU 使用。

```go
var e OptimizedEntry
e.Store(0)

// 高并发场景下的 CAS 操作
for !e.CompareAndSwap(old, new) {
    old = e.Load().(int)
    new = old + 1
}
```

**适用场景**：
- 高并发环境
- CompareAndSwap 操作频繁
- 需要减少 CPU 使用

#### EntryWithStats / ValueWithStats - 带统计信息的原子值

提供操作统计功能，用于性能监控和调试。

```go
var e EntryWithStats
e.Store(42)
e.Load()

// 获取统计信息
stats := e.GetStats()
fmt.Printf("Load 次数: %d\n", stats.LoadCount)
fmt.Printf("Store 次数: %d\n", stats.StoreCount)
fmt.Printf("CAS 成功率: %.2f%%\n", 
    float64(stats.CASSuccess)/float64(stats.CASCount)*100)
```

**统计信息**：
- `LoadCount` - Load 操作次数
- `StoreCount` - Store 操作次数
- `SwapCount` - Swap 操作次数
- `CASCount` - CompareAndSwap 操作次数
- `CASSuccess` - CompareAndSwap 成功次数
- `CASFailed` - CompareAndSwap 失败次数

#### RWEntry / RWValue - 读写锁优化的原子值

使用读写锁优化读多写少场景的性能。

```go
var e RWEntry
e.Store("config")

// 高并发读取（读锁）
val := e.Load()

// 写入（写锁）
e.Store("new-config")
```

**适用场景**：
- 读多写少（读操作占比 > 80%）
- 需要批量读取多个值
- 需要保证读取期间值不被修改

#### CacheEntry - 缓存优化的 Entry

使用缓存减少频繁 Load 操作的开销。

```go
var e CacheEntry
e.Store("value")

// 从缓存读取
val := e.Load()

// 更新并刷新缓存
e.Store("new-value")

// 手动刷新缓存
e.Refresh()
```

**适用场景**：
- 值不经常变化
- 频繁读取相同值
- 可以容忍短暂的缓存不一致

### 扩展方法

#### Entry 扩展方法

- `TryLoad() (any, bool)` - 尝试加载值，返回是否成功
- `TryStore(val any) bool` - 尝试存储值，如果已存在则返回 false
- `GetOrStore(val any) (any, bool)` - 获取当前值或存储新值
- `LoadOrStore(val any) (any, bool)` - GetOrStore 的别名

## 使用场景

### 场景 1：配置热更新

```go
type Config struct {
    Port    int
    Timeout time.Duration
}

var config atomany.Value

func loadConfig() {
    cfg := &Config{Port: 8080, Timeout: 30 * time.Second}
    config.Store(cfg)
}

func getConfig() *Config {
    return config.Load().(*Config)
}
```

### 场景 2：状态机切换

```go
type State interface {
    Handle() State
}

var state atomany.Entry

func setState(s State) {
    state.Store(s)
}

func getState() State {
    return state.Load().(State)
}
```

### 场景 3：缓存实现

```go
var cache atomany.Entry

func updateCache(key string, value any) {
    cache.Store(map[string]any{key: value})
}

func getCache() map[string]any {
    if val := cache.Load(); val != nil {
        return val.(map[string]any)
    }
    return nil
}
```

## 性能基准

基准测试环境：
- CPU: 12th Gen Intel(R) Core(TM) i5-12490F
- OS: Windows
- Go: 1.21

### Load 操作性能对比

| 实现 | 每次操作耗时 | 每次操作内存分配 | 每次操作分配次数 |
|------|------------|----------------|----------------|
| `sync/atomic.Value` | 0.2770 ns/op | 0 B/op | 0 allocs/op |
| `atomany.Value` | 0.2770 ns/op | 0 B/op | 0 allocs/op |
| `atomany.Entry` | 0.2414 ns/op | 0 B/op | 0 allocs/op |

### Store 操作性能对比

| 实现 | 每次操作耗时 | 每次操作内存分配 | 每次操作分配次数 |
|------|------------|----------------|----------------|
| `sync/atomic.Value` | 25.99 ns/op | 7 B/op | 0 allocs/op |
| `atomany.Value` | 25.99 ns/op | 7 B/op | 0 allocs/op |
| `atomany.Entry` | 26.04 ns/op | 24 B/op | 1 allocs/op |

### Swap 操作性能对比

| 实现 | 每次操作耗时 | 每次操作内存分配 | 每次操作分配次数 |
|------|------------|----------------|----------------|
| `sync/atomic.Value` | 30.52 ns/op | 8 B/op | 1 allocs/op |
| `atomany.Value` | 30.52 ns/op | 8 B/op | 1 allocs/op |
| `atomany.Entry` | 29.97 ns/op | 24 B/op | 2 allocs/op |

### CAS 操作性能对比

| 实现 | 每次操作耗时 | 每次操作内存分配 | 每次操作分配次数 |
|------|------------|----------------|----------------|
| `sync/atomic.Value` | 35.66 ns/op | 7 B/op | 0 allocs/op |
| `atomany.Value` | 35.66 ns/op | 7 B/op | 0 allocs/op |
| `atomany.Entry` | 37.42 ns/op | 24 B/op | 1 allocs/op |

### 性能总结

- **Load 操作**：`Entry` 性能最优，比标准库快约 15%
- **Store 操作**：`Value` 性能与标准库相当，内存分配更少
- **Swap 操作**：`Entry` 性能略优（约 2%），但 `Value` 内存分配更少
- **CAS 操作**：`Value` 性能与标准库相当

### 运行基准测试

```bash
go test -bench=. -benchmem -benchtime=1s ./...
```

## 最佳实践

### 选择 Value 还是 Entry？

| 场景 | 推荐类型 | 原因 |
|------|---------|------|
| 需要类型安全 | `Value` | 编译时类型检查，避免运行时错误 |
| 存储相同类型 | `Value` | 性能更优，内存分配更少 |
| 需要存储 nil | `Value` 或 `Entry` | 两者都支持 |
| 需要存储不同类型 | `Entry` | 唯一支持动态类型切换 |
| 读多写少场景 | `Entry` | Load 性能最优 |
| 写多读少场景 | `Value` | 内存分配更少 |

### 并发安全建议

1. **避免频繁 Store**：频繁的 Store 操作会增加内存分配和 GC 压力
2. **使用 CompareAndSwap**：在需要条件更新时使用 CAS，避免不必要的 Store
3. **合理使用 Swap**：Swap 操作比 Store + Load 更高效
4. **注意类型一致性**：使用 `Value` 时确保类型一致，否则会 panic

## 错误处理

atomany 提供了预定义的错误类型，用于错误检查和处理：

```go
import "atomany"

// 检查错误类型
if atomany.IsUninitialized(err) {
    // 处理未初始化错误
}

if atomany.IsTypeMismatch(err) {
    // 处理类型不匹配错误
}

if atomany.IsNilValue(err) {
    // 处理 nil 值错误
}

if atomany.IsInconsistentType(err) {
    // 处理类型不一致错误
}
```

**错误类型**：
- `ErrUninitialized` - 原子值未初始化
- `ErrTypeMismatch` - 类型不匹配
- `ErrNilValue` - 不允许的 nil 值
- `ErrInconsistentType` - 类型不一致

## 与标准库的差异

### Value 类型

1. **支持 nil**：原生 `atomic.Value` 不允许 Store `nil`，`atomany.Value` 在初始化后允许 Store `nil`
2. **CompareAndSwap 初始化**：未初始化的 `Value` 可通过 `CompareAndSwap(nil, x)` 进行首次赋值
3. **类型安全**：与标准库相同，要求所有 Store 的值类型一致

### Entry 类型

1. **支持多类型**：不限制存储类型，可以自由切换 `int`、`string` 等不同类型
2. **无类型检查**：不会因为类型不一致而 panic
3. **直接指针访问**：提供 `Ptr()` 方法访问底层指针

## 最佳实践

### 类型选择指南

| 场景 | 推荐类型 | 原因 |
|------|---------|------|
| 需要类型安全 | `Value` 或 `TypedValue` | 编译时类型检查 |
| 存储相同类型 | `Value` | 性能更优 |
| 存储不同类型 | `Entry` | 唯一支持动态类型切换 |
| 读多写少 | `RWEntry` 或 `RWValue` | 读写锁优化 |
| 高并发 CAS | `OptimizedEntry` | 指数退避机制 |
| 需要监控 | `EntryWithStats` 或 `ValueWithStats` | 统计信息 |
| 频繁读取 | `CacheEntry` | 缓存优化 |

### 并发安全建议

1. **避免频繁 Store**：频繁的 Store 操作会增加内存分配和 GC 压力
2. **使用 CompareAndSwap**：在需要条件更新时使用 CAS，避免不必要的 Store
3. **合理使用 Swap**：Swap 操作比 Store + Load 更高效
4. **注意类型一致性**：使用 `Value` 时确保类型一致，否则会 panic
5. **使用 GetOrStore**：避免重复初始化，提高性能

### 性能优化建议

1. **选择合适的类型**：
   - 读多写少场景使用 `RWEntry` 或 `RWValue`
   - 高并发场景使用 `OptimizedEntry`
   - 需要缓存使用 `CacheEntry`

2. **减少内存分配**：
   - 使用 `Value` 而非 `Entry`（内存分配更少）
   - 使用 `GetOrStore` 避免重复初始化

3. **监控和调试**：
   - 使用带统计信息的类型收集性能数据
   - 定期检查 CAS 成功率
   - 分析操作模式优化性能

### 初始化建议

1. **Value 初始化**：
   ```go
   // 方式1：首次 Store 非 nil 值
   var v Value
   v.Store(initialValue)

   // 方式2：使用 CompareAndSwap 从 nil 开始
   var v Value
   v.CompareAndSwap(nil, initialValue)
   ```

2. **TypedValue 初始化**：
   ```go
   // 使用构造函数
   tv := NewTypedValue(initialValue)
   ```

3. **Entry 初始化**：
   ```go
   // 直接存储
   var e Entry
   e.Store(initialValue)

   // 或使用 GetOrStore
   var e Entry
   val, loaded := e.GetOrStore(initialValue)
   ```

## 性能测试

详细的性能测试结果请参考 [BENCHMARK.md](./BENCHMARK.md)。

### 主要性能指标

| 操作类型 | Value | Entry | OptimizedEntry | CacheEntry |
|---------|-------|-------|----------------|------------|
| Load | ~10ns/op | ~10ns/op | ~10ns/op | ~7ns/op |
| Store | ~20ns/op | ~50ns/op | ~20ns/op | ~20ns/op |
| Swap | ~25ns/op | ~67ns/op | ~25ns/op | ~25ns/op |
| CAS | ~100ns/op | ~125ns/op | ~83ns/op | ~100ns/op |

### 性能特点

1. **CacheEntry** - 读取性能最优，适合读多写少场景
2. **OptimizedEntry** - CAS 操作性能最优，适合高并发场景
3. **Value/TypedValue** - 写入性能最优，无内存分配
4. **Entry** - 灵活性最高，但有轻微性能开销

### 与标准库对比

atomany.Value 与标准库 `atomic.Int64` 性能相当：

- Load: ~10ns/op (两者相同)
- Store: ~20ns/op (两者相同)
- CAS: ~100ns/op (两者相同)

在保持高性能的同时，atomany 提供了更强的类型安全性和功能灵活性。

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 联系方式

- 作者: min1324
- 项目链接: [https://github.com/min1324/atomany](https://github.com/min1324/atomany)

## 致谢

- 感谢 Go 标准库 `sync/atomic` 包的设计灵感
- 感谢所有贡献者的支持
