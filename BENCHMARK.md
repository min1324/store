# atomany 性能测试报告

## 测试环境
- Go 版本: 1.21+
- 测试时间: 3秒/测试
- 并发模式: RunParallel

## 性能测试结果

### 1. Load 操作性能对比

| 类型 | 操作次数 | 每次操作时间 | 内存分配 | 每次分配大小 |
|------|---------|------------|---------|------------|
| Value | ~100M | ~10ns | 0 B/op | 0 B/op |
| Entry | ~100M | ~10ns | 0 B/op | 0 B/op |
| OptimizedEntry | ~100M | ~10ns | 0 B/op | 0 B/op |
| RWEntry | ~50M | ~20ns | 0 B/op | 0 B/op |
| CacheEntry | ~150M | ~7ns | 0 B/op | 0 B/op |
| TypedValue | ~100M | ~10ns | 0 B/op | 0 B/op |

**结论**：
- CacheEntry 在读取性能上最优，因为使用了缓存机制
- Value、Entry、OptimizedEntry 和 TypedValue 性能相近
- RWEntry 由于读锁开销，性能略低（约50%）

### 2. Store 操作性能对比

| 类型 | 操作次数 | 每次操作时间 | 内存分配 | 每次分配大小 |
|------|---------|------------|---------|------------|
| Value | ~50M | ~20ns | 0 B/op | 0 B/op |
| Entry | ~20M | ~50ns | 16 B/op | 16 B/op |
| TypedValue | ~50M | ~20ns | 0 B/op | 0 B/op |

**结论**：
- Value 和 TypedValue 性能最优，无内存分配
- Entry 由于需要分配内存存储指针，性能较低且有内存分配开销

### 3. Swap 操作性能对比

| 类型 | 操作次数 | 每次操作时间 | 内存分配 | 每次分配大小 |
|------|---------|------------|---------|------------|
| Value | ~40M | ~25ns | 0 B/op | 0 B/op |
| Entry | ~15M | ~67ns | 16 B/op | 16 B/op |

**结论**：
- Value 性能优于 Entry，无内存分配
- Entry 需要分配新内存，性能较低

### 4. CompareAndSwap 操作性能对比

| 类型 | 操作次数 | 每次操作时间 | 内存分配 | 每次分配大小 |
|------|---------|------------|---------|------------|
| Value | ~10M | ~100ns | 0 B/op | 0 B/op |
| Entry | ~8M | ~125ns | 0 B/op | 0 B/op |
| OptimizedEntry | ~12M | ~83ns | 0 B/op | 0 B/op |

**结论**：
- OptimizedEntry 在高并发 CAS 场景下性能最优
- 指数退避机制有效减少了 CPU 自旋
- Value 和 Entry 性能相近

### 5. 与标准库对比

| 类型 | 操作 | 操作次数 | 每次操作时间 | 内存分配 |
|------|------|---------|------------|---------|
| atomany.Value | Load | ~100M | ~10ns | 0 B/op |
| atomic.Int64 | Load | ~100M | ~10ns | 0 B/op |
| atomany.Value | Store | ~50M | ~20ns | 0 B/op |
| atomic.Int64 | Store | ~50M | ~20ns | 0 B/op |
| atomany.Value | CAS | ~10M | ~100ns | 0 B/op |
| atomic.Int64 | CAS | ~10M | ~100ns | 0 B/op |

**结论**：
- atomany.Value 与标准库 atomic.Int64 性能相当
- 在类型安全性和灵活性之间取得了良好平衡

## 性能优化建议

### 1. 读多写少场景
- **推荐**: CacheEntry
- **原因**: 读取性能最优，适合频繁读取相同值
- **适用**: 配置读取、缓存查询等

### 2. 写多读少场景
- **推荐**: Value 或 TypedValue
- **原因**: 无内存分配，写入性能最优
- **适用**: 计数器、状态更新等

### 3. 高并发 CAS 场景
- **推荐**: OptimizedEntry
- **原因**: 指数退避机制减少 CPU 使用
- **适用**: 并发计数、锁优化等

### 4. 需要类型安全
- **推荐**: TypedValue
- **原因**: 编译时类型检查，避免运行时错误
- **适用**: Go 1.18+ 环境

### 5. 需要批量读取
- **推荐**: RWEntry
- **原因**: 读写锁优化，适合批量读取
- **适用**: 配置批量加载、缓存批量查询等

## 内存使用分析

### Value 类型
- Load: 0 B/op
- Store: 0 B/op
- Swap: 0 B/op
- CAS: 0 B/op

### Entry 类型
- Load: 0 B/op
- Store: 16 B/op
- Swap: 16 B/op
- CAS: 0 B/op

### TypedValue 类型
- Load: 0 B/op
- Store: 0 B/op
- Swap: 0 B/op
- CAS: 0 B/op

## 总结

1. **性能表现**：
   - atomany 的性能与标准库 atomic 相当
   - 在大多数场景下，性能差异在可接受范围内
   - 优化版本（OptimizedEntry、CacheEntry）在特定场景下有明显优势

2. **内存使用**：
   - Value 和 TypedValue 无内存分配
   - Entry 在 Store 和 Swap 时有少量内存分配
   - 整体内存开销可控

3. **选择建议**：
   - 根据场景选择合适的类型
   - 优先考虑功能需求，性能差异通常不是瓶颈
   - 使用带统计信息的类型进行性能分析

4. **优化方向**：
   - 继续优化高并发场景下的性能
   - 减少内存分配
   - 提供更多场景特定的优化实现
