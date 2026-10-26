package store_test

import (
	"fmt"
	"store"
	"sync/atomic"
	"testing"
)

func benchFunc(f func(name string, e iface)) {
	for _, v := range []struct {
		iface
		name string
	}{
		{&atomic.Value{}, "atomic"},
		{&store.Value{}, "Value"},
		{&store.Entry{}, "Entry"},
	} {
		f(v.name, v.iface)
	}
}

func BenchmarkRead(b *testing.B) {
	benchFunc(func(name string, v iface) {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", name), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					x := v.Load().(int)
					if x != 0 {
						b.Fatalf("%s wrong value: got %v, want 0", name, x)
					}
				}
			})
		})
	})

}

func BenchmarkValueStore(b *testing.B) {
	benchFunc(func(name string, v iface) {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", name), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					v.Store(i)
					i++
				}
			})
		})
	})
}

func BenchmarkValueStoreLoad(b *testing.B) {
	benchFunc(func(name string, v iface) {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", name), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				var i = 0
				for pb.Next() {
					v.Store(i)
					x := v.Load().(int)
					if x != i {
						b.Fatalf("%s wrong value: got %v, want %v", name, x, i)
					}
				}
			})
		})
	})
}

func BenchmarkValueSwap(b *testing.B) {
	benchFunc(func(name string, v iface) {
		var i int64

		b.Run(fmt.Sprintf("%s", name), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					v.Swap(atomic.AddInt64(&i, 1))
				}
			})
		})
	})
}

func BenchmarkValueCAS(b *testing.B) {
	benchFunc(func(name string, v iface) {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", name), func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					j := v.Load().(int)
					v.CompareAndSwap(j, j+1)
					// v.CompareAndSwap(i, atomic.AddInt64(&i, 1))
				}
			})
		})
	})
}

func Benchmark_nil(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
		}
	})
}

// BenchmarkValueLoad 测试 store.Value 的 Load 性能
func BenchmarkLoad_atomic(b *testing.B) {
	var v atomic.Value
	v.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Load()
		}
	})
}

func BenchmarkLoad_Value(b *testing.B) {
	var v store.Value
	v.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Load()
		}
	})
}

// BenchmarkEntryLoad 测试 store.Entry 的 Load 性能
func BenchmarkLoad_Entry(b *testing.B) {
	var e store.Entry
	e.Store(42)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.Load()
		}
	})
}

// BenchmarkValueStore 测试 store.Value 的 Store 性能
func BenchmarkStore_atomic(b *testing.B) {
	var v atomic.Value
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			v.Store(i)
			i++
		}
	})
}

// BenchmarkValueStore 测试 store.Value 的 Store 性能
func BenchmarkStore_Value(b *testing.B) {
	var v store.Value
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			v.Store(i)
			i++
		}
	})
}

// BenchmarkEntryStore 测试 store.Entry 的 Store 性能
func BenchmarkStore_Entry(b *testing.B) {
	var e store.Entry
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			e.Store(i)
			i++
		}
	})
}

func BenchmarkSwap_atomic(b *testing.B) {
	var v atomic.Value
	v.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			v.Swap(i)
			i++
		}
	})
}

// BenchmarkValueSwap 测试 store.Value 的 Swap 性能
func BenchmarkSwap_Value(b *testing.B) {
	var v store.Value
	v.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			v.Swap(i)
			i++
		}
	})
}

// BenchmarkEntrySwap 测试 store.Entry 的 Swap 性能
func BenchmarkSwap_Entry(b *testing.B) {
	var e store.Entry
	e.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			e.Swap(i)
			i++
		}
	})
}

func BenchmarkCAS_atomic(b *testing.B) {
	var v atomic.Value
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

// BenchmarkValueCAS 测试 store.Value 的 CompareAndSwap 性能
func BenchmarkCAS_Value(b *testing.B) {
	var v store.Value
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

// BenchmarkEntryCAS 测试 store.Entry 的 CompareAndSwap 性能
func BenchmarkCAS_Entry(b *testing.B) {
	var e store.Entry
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
