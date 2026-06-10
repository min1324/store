package atomany

import (
	"sync"
	"sync/atomic"
)

// RWEntry 优化读多写少场景的原子值
// 使用读写锁优化读取性能，适用于读操作远多于写操作的场景
//
// 使用场景：
//   - 读多写少的场景（读操作占比 > 80%）
//   - 需要批量读取多个值
//   - 需要保证读取期间值不被修改
//
// 注意：
//   - 写操作性能会降低
//   - 不适用于写多读少的场景
type RWEntry struct {
	Entry
	mu sync.RWMutex
}

// Load 原子地返回最近一次 Store 设置的值
// 使用读锁保证并发读取性能
func (e *RWEntry) Load() any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Entry.Load()
}

// Store 原子地将值存储到 Entry 中
// 使用写锁保证写入的原子性
func (e *RWEntry) Store(val any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Entry.Store(val)
}

// Swap 原子地交换 Entry 中的值并返回旧值
// 使用写锁保证交换的原子性
func (e *RWEntry) Swap(new any) any {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Entry.Swap(new)
}

// CompareAndSwap 执行原子的比较并交换操作
// 使用写锁保证 CAS 的原子性
func (e *RWEntry) CompareAndSwap(old, new any) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Entry.CompareAndSwap(old, new)
}

// TryLoad 尝试加载值，如果未初始化则返回 false
// 使用读锁保证并发读取性能
func (e *RWEntry) TryLoad() (val any, ok bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Entry.TryLoad()
}

// TryStore 尝试存储值，如果已存在则返回 false
// 使用写锁保证存储的原子性
func (e *RWEntry) TryStore(val any) (ok bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Entry.TryStore(val)
}

// GetOrStore 获取当前值，如果未初始化则存储新值并返回
// 使用写锁保证操作的原子性
func (e *RWEntry) GetOrStore(val any) (actual any, loaded bool) {
	// 先尝试读锁读取
	e.mu.RLock()
	p := atomic.LoadPointer(&e.p)
	if p != nil {
		e.mu.RUnlock()
		return ptr2any(p), true
	}
	e.mu.RUnlock()

	// 需要写入，获取写锁
	e.mu.Lock()
	defer e.mu.Unlock()

	// 再次检查，防止在获取写锁期间其他 goroutine 已经写入
	p = atomic.LoadPointer(&e.p)
	if p != nil {
		return ptr2any(p), true
	}

	newPtr := any2ptr(val)
	if atomic.CompareAndSwapPointer(&e.p, nil, newPtr) {
		return val, false
	}
	return ptr2any(atomic.LoadPointer(&e.p)), true
}

// BatchLoad 批量加载多个值
// 适用于需要同时读取多个值的场景
//
// 参数：
//   - entries: 要读取的多个 RWEntry
//
// 返回：
//   - values: 对应的值列表
func BatchLoad(entries ...*RWEntry) []any {
	values := make([]any, len(entries))
	for i, e := range entries {
		e.mu.RLock()
		values[i] = e.Entry.Load()
	}

	// 所有读取完成后统一释放锁
	for _, e := range entries {
		e.mu.RUnlock()
	}

	return values
}

// RWValue 优化读多写少场景的类型安全原子值
// 使用读写锁优化读取性能，适用于读操作远多于写操作的场景
//
// 使用场景：
//   - 读多写少的场景（读操作占比 > 80%）
//   - 需要批量读取多个值
//   - 需要保证读取期间值不被修改
//   - 需要类型安全保证
//
// 注意：
//   - 写操作性能会降低
//   - 不适用于写多读少的场景
type RWValue struct {
	Value
	mu sync.RWMutex
}

// Load 原子地返回最近一次 Store 设置的值
// 使用读锁保证并发读取性能
func (v *RWValue) Load() any {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Value.Load()
}

// Store 原子地将值存储到 Value 中
// 使用写锁保证写入的原子性
func (v *RWValue) Store(val any) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Value.Store(val)
}

// Swap 原子地交换 Value 中的值并返回旧值
// 使用写锁保证交换的原子性
func (v *RWValue) Swap(new any) any {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.Value.Swap(new)
}

// CompareAndSwap 执行原子的比较并交换操作
// 使用写锁保证 CAS 的原子性
func (v *RWValue) CompareAndSwap(old, new any) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.Value.CompareAndSwap(old, new)
}

// BatchLoadValues 批量加载多个类型安全的值
// 适用于需要同时读取多个值的场景
//
// 参数：
//   - values: 要读取的多个 RWValue
//
// 返回：
//   - 对应的值列表
func BatchLoadValues(values ...*RWValue) []any {
	result := make([]any, len(values))
	for i, v := range values {
		v.mu.RLock()
		result[i] = v.Value.Load()
	}

	// 所有读取完成后统一释放锁
	for _, v := range values {
		v.mu.RUnlock()
	}

	return result
}

// CacheEntry 提供缓存优化的原子值
// 使用缓存减少频繁的 Load 操作开销
//
// 使用场景：
//   - 值不经常变化
//   - 频繁读取相同值
//   - 可以容忍短暂的缓存不一致
//
// 注意：
//   - 缓存可能不是最新的
//   - 需要调用 Refresh 刷新缓存
type CacheEntry struct {
	Entry
	cache     any
	cacheOnce sync.Once
}

// Load 返回缓存的值
// 如果缓存未初始化，则加载并缓存当前值
func (c *CacheEntry) Load() any {
	c.cacheOnce.Do(func() {
		c.cache = c.Entry.Load()
	})
	return c.cache
}

// Refresh 刷新缓存，使其与当前值同步
func (c *CacheEntry) Refresh() {
	c.cache = c.Entry.Load()
}

// Store 存储新值并刷新缓存
func (c *CacheEntry) Store(val any) {
	c.Entry.Store(val)
	c.Refresh()
}

// Swap 交换值并刷新缓存
func (c *CacheEntry) Swap(new any) any {
	old := c.Entry.Swap(new)
	c.Refresh()
	return old
}

// CompareAndSwap 比较并交换，成功后刷新缓存
func (c *CacheEntry) CompareAndSwap(old, new any) bool {
	swapped := c.Entry.CompareAndSwap(old, new)
	if swapped {
		c.Refresh()
	}
	return swapped
}
