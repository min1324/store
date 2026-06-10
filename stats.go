package atomany

import (
	"sync/atomic"
)

// Stats 提供原子值的统计信息
type Stats struct {
	LoadCount    uint64 // Load 操作次数
	StoreCount   uint64 // Store 操作次数
	SwapCount    uint64 // Swap 操作次数
	CASCount     uint64 // CompareAndSwap 操作次数
	CASSuccess   uint64 // CompareAndSwap 成功次数
	CASFailed    uint64 // CompareAndSwap 失败次数
}

// Reset 重置所有统计计数器
func (s *Stats) Reset() {
	atomic.StoreUint64(&s.LoadCount, 0)
	atomic.StoreUint64(&s.StoreCount, 0)
	atomic.StoreUint64(&s.SwapCount, 0)
	atomic.StoreUint64(&s.CASCount, 0)
	atomic.StoreUint64(&s.CASSuccess, 0)
	atomic.StoreUint64(&s.CASFailed, 0)
}

// EntryWithStats 提供带统计信息的原子值
type EntryWithStats struct {
	Entry
	stats Stats
}

// Load 原子地返回最近一次 Store 设置的值，并更新统计信息
func (e *EntryWithStats) Load() any {
	atomic.AddUint64(&e.stats.LoadCount, 1)
	return e.Entry.Load()
}

// Store 原子地将值存储到 Entry 中，并更新统计信息
func (e *EntryWithStats) Store(val any) {
	atomic.AddUint64(&e.stats.StoreCount, 1)
	e.Entry.Store(val)
}

// Swap 原子地交换 Entry 中的值并返回旧值，并更新统计信息
func (e *EntryWithStats) Swap(new any) any {
	atomic.AddUint64(&e.stats.SwapCount, 1)
	return e.Entry.Swap(new)
}

// CompareAndSwap 执行原子的比较并交换操作，并更新统计信息
func (e *EntryWithStats) CompareAndSwap(old, new any) bool {
	atomic.AddUint64(&e.stats.CASCount, 1)
	swapped := e.Entry.CompareAndSwap(old, new)
	if swapped {
		atomic.AddUint64(&e.stats.CASSuccess, 1)
	} else {
		atomic.AddUint64(&e.stats.CASFailed, 1)
	}
	return swapped
}

// GetStats 获取当前统计信息
func (e *EntryWithStats) GetStats() Stats {
	return Stats{
		LoadCount:  atomic.LoadUint64(&e.stats.LoadCount),
		StoreCount: atomic.LoadUint64(&e.stats.StoreCount),
		SwapCount:  atomic.LoadUint64(&e.stats.SwapCount),
		CASCount:   atomic.LoadUint64(&e.stats.CASCount),
		CASSuccess: atomic.LoadUint64(&e.stats.CASSuccess),
		CASFailed:  atomic.LoadUint64(&e.stats.CASFailed),
	}
}

// ResetStats 重置统计信息
func (e *EntryWithStats) ResetStats() {
	e.stats.Reset()
}

// ValueWithStats 提供带统计信息的类型安全原子值
type ValueWithStats struct {
	Value
	stats Stats
}

// Load 原子地返回最近一次 Store 设置的值，并更新统计信息
func (v *ValueWithStats) Load() any {
	atomic.AddUint64(&v.stats.LoadCount, 1)
	return v.Value.Load()
}

// Store 原子地将值存储到 Value 中，并更新统计信息
func (v *ValueWithStats) Store(val any) {
	atomic.AddUint64(&v.stats.StoreCount, 1)
	v.Value.Store(val)
}

// Swap 原子地交换 Value 中的值并返回旧值，并更新统计信息
func (v *ValueWithStats) Swap(new any) any {
	atomic.AddUint64(&v.stats.SwapCount, 1)
	return v.Value.Swap(new)
}

// CompareAndSwap 执行原子的比较并交换操作，并更新统计信息
func (v *ValueWithStats) CompareAndSwap(old, new any) bool {
	atomic.AddUint64(&v.stats.CASCount, 1)
	swapped := v.Value.CompareAndSwap(old, new)
	if swapped {
		atomic.AddUint64(&v.stats.CASSuccess, 1)
	} else {
		atomic.AddUint64(&v.stats.CASFailed, 1)
	}
	return swapped
}

// GetStats 获取当前统计信息
func (v *ValueWithStats) GetStats() Stats {
	return Stats{
		LoadCount:  atomic.LoadUint64(&v.stats.LoadCount),
		StoreCount: atomic.LoadUint64(&v.stats.StoreCount),
		SwapCount:  atomic.LoadUint64(&v.stats.SwapCount),
		CASCount:   atomic.LoadUint64(&v.stats.CASCount),
		CASSuccess: atomic.LoadUint64(&v.stats.CASSuccess),
		CASFailed:  atomic.LoadUint64(&v.stats.CASFailed),
	}
}

// ResetStats 重置统计信息
func (v *ValueWithStats) ResetStats() {
	v.stats.Reset()
}
