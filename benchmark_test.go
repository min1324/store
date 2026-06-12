package atomany

import (
	"sync/atomic"
	"testing"
)

// BenchmarkValueLoad 测试 Value 的 Load 性能
func BenchmarkValueLoad(b *testing.B) {
	var v Value
	v.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Load()
		}
	})
}

// BenchmarkEntryLoad 测试 Entry 的 Load 性能
func BenchmarkEntryLoad(b *testing.B) {
	var e Entry
	e.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.Load()
		}
	})
}

// BenchmarkRWEntryLoad 测试 RWEntry 的 Load 性能
func BenchmarkRWEntryLoad(b *testing.B) {
	var e RWEntry
	e.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.Load()
		}
	})
}

// BenchmarkCacheEntryLoad 测试 CacheEntry 的 Load 性能
func BenchmarkCacheEntryLoad(b *testing.B) {
	var e CacheEntry
	e.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.Load()
		}
	})
}

// BenchmarkValueStore 测试 Value 的 Store 性能
func BenchmarkValueStore(b *testing.B) {
	var v Value
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			v.Store(i)
			i++
		}
	})
}

// BenchmarkEntryStore 测试 Entry 的 Store 性能
func BenchmarkEntryStore(b *testing.B) {
	var e Entry
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			e.Store(i)
			i++
		}
	})
}

// BenchmarkValueSwap 测试 Value 的 Swap 性能
func BenchmarkValueSwap(b *testing.B) {
	var v Value
	v.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			v.Swap(i)
			i++
		}
	})
}

// BenchmarkEntrySwap 测试 Entry 的 Swap 性能
func BenchmarkEntrySwap(b *testing.B) {
	var e Entry
	e.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			e.Swap(i)
			i++
		}
	})
}

// BenchmarkValueCAS 测试 Value 的 CompareAndSwap 性能
func BenchmarkValueCAS(b *testing.B) {
	var v Value
	v.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			for !v.CompareAndSwap(i, i+1) {
				i = v.Load().(int)
			}
			i++
		}
	})
}

// BenchmarkEntryCAS 测试 Entry 的 CompareAndSwap 性能
func BenchmarkEntryCAS(b *testing.B) {
	var e Entry
	e.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			for !e.CompareAndSwap(i, i+1) {
				i = e.Load().(int)
			}
			i++
		}
	})
}

// BenchmarkTypedValueLoad 测试 TypedValue 的 Load 性能
func BenchmarkTypedValueLoad(b *testing.B) {
	tv := NewTypedValue(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tv.Load()
		}
	})
}

// BenchmarkTypedValueStore 测试 TypedValue 的 Store 性能
func BenchmarkTypedValueStore(b *testing.B) {
	tv := NewTypedValue(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tv.Store(i)
			i++
		}
	})
}

// BenchmarkAtomicIntLoad 测试标准库 atomic.Int64 的 Load 性能
func BenchmarkAtomicIntLoad(b *testing.B) {
	var v atomic.Int64
	v.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Load()
		}
	})
}

// BenchmarkAtomicIntStore 测试标准库 atomic.Int64 的 Store 性能
func BenchmarkAtomicIntStore(b *testing.B) {
	var v atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			v.Store(i)
			i++
		}
	})
}

// BenchmarkAtomicIntCAS 测试标准库 atomic.Int64 的 CompareAndSwap 性能
func BenchmarkAtomicIntCAS(b *testing.B) {
	var v atomic.Int64
	v.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			for !v.CompareAndSwap(i, i+1) {
				i = v.Load()
			}
			i++
		}
	})
}
