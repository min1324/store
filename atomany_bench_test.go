package atomany_test

import (
	"atomany"
	"sync"
	"sync/atomic"
	"testing"
)

// ============================================================
// 基准测试框架
// ============================================================

// bench 定义基准测试的结构
type bench struct {
	setup func(b *testing.B, m atomany.Interface) // 可选的测试前设置函数
	perG  func(b *testing.B, pb *testing.PB, id int, m atomany.Interface)
}

// benchMap 对多种队列实现运行相同的基准测试
func benchMap(b *testing.B, bench bench) {
	implementations := []struct {
		name  string
		value atomany.Interface
	}{
		{"Value", &atomany.Value{}},
		{"Entry", &atomany.Entry{}},
	}

	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			m := impl.value

			if bench.setup != nil {
				bench.setup(b, m)
			}

			b.ResetTimer()
			var idCounter int64
			b.RunParallel(func(pb *testing.PB) {
				id := int(atomic.AddInt64(&idCounter, 1) - 1)
				bench.perG(b, pb, id, m)
			})
		})
	}
}

// ============================================================
// 混合操作基准测试
// ============================================================

// BenchmarkBalance 均衡的 Load/Store 操作性能
func BenchmarkBalance(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				m.Load()
				m.Store(i)
				i++
			}
		},
	})
}

// ============================================================
// 生产者-消费者模式基准测试
// ============================================================

// BenchmarkProducerConsumer_LockFree 经典生产者-消费者模式 (LockFree)
func BenchmarkProducerConsumer_LockFree(b *testing.B) {
	var val atomany.Value
	var produced, consumed int64

	var wg sync.WaitGroup

	b.ResetTimer()

	// 生产者
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			val.Store(i)
			atomic.AddInt64(&produced, 1)
		}
	}()

	// 消费者
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			for atomic.LoadInt64(&consumed) < atomic.LoadInt64(&produced) {
				val.Load()
				atomic.AddInt64(&consumed, 1)
			}
		}
	}()

	wg.Wait()
	b.ReportMetric(float64(produced)/float64(b.Elapsed().Seconds()), "ops/sec")
}

// ============================================================
// 读写分离基准测试
// ============================================================

// BenchmarkReadHeavy 读多写少场景 (90% 读, 10% 写)
func BenchmarkReadHeavy(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					m.Store(i)
				} else {
					m.Load()
				}
				i++
			}
		},
	})
}

// BenchmarkWriteHeavy 写多读少场景 (10% 读, 90% 写)
func BenchmarkWriteHeavy(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					m.Load()
				} else {
					m.Store(i)
				}
				i++
			}
		},
	})
}

// ============================================================
// 高并发写入基准测试
// ============================================================

// BenchmarkHighContention 高并发写入竞争
func BenchmarkHighContention(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			for pb.Next() {
				m.Store(id)
			}
		},
	})
}

// ============================================================
// 内存分配基准测试
// ============================================================

// BenchmarkValueStoreWithAlloc 带内存分配的 Store 操作
func BenchmarkValueStoreWithAlloc(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			val := m.(*atomany.Value)
			for pb.Next() {
				val.Store(make([]byte, i%1024))
				i++
			}
		},
	})
}

// BenchmarkValueLoadWithAlloc 带内存分配的 Load 操作
func BenchmarkValueLoadWithAlloc(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			val := m.(*atomany.Value)
			for pb.Next() {
				_ = val.Load().([]byte)
			}
		},
	})
}

// ============================================================
// Swap 操作基准测试
// ============================================================

// BenchmarkValueSwapWithAlloc 带内存分配的 Swap 操作
func BenchmarkValueSwapWithAlloc(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			val := m.(*atomany.Value)
			for pb.Next() {
				val.Swap(make([]byte, i%1024))
				i++
			}
		},
	})
}

// ============================================================
// CAS 操作基准测试
// ============================================================

// BenchmarkValueCASWithAlloc 带内存分配的 CAS 操作
func BenchmarkValueCASWithAlloc(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			val := m.(*atomany.Value)
			for pb.Next() {
				old := val.Load()
				val.CompareAndSwap(old, i)
				i++
			}
		},
	})
}

// ============================================================
// Entry 性能测试
// ============================================================

// BenchmarkEntryReadHeavy 读多写少场景
func BenchmarkEntryReadHeavy(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					m.Store(i)
				} else {
					m.Load()
				}
				i++
			}
		},
	})
}

// BenchmarkEntryWriteHeavy 写多读少场景
func BenchmarkEntryWriteHeavy(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					m.Load()
				} else {
					m.Store(i)
				}
				i++
			}
		},
	})
}

// BenchmarkEntryHighContention 高并发写入竞争
func BenchmarkEntryHighContention(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			for pb.Next() {
				m.Store(id)
			}
		},
	})
}

// BenchmarkEntrySwapWithAlloc 带内存分配的 Swap 操作
func BenchmarkEntrySwapWithAlloc(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(make([]byte, 1024))
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				m.Swap(make([]byte, i%1024))
				i++
			}
		},
	})
}

// BenchmarkEntryCASWithAlloc 带内存分配的 CAS 操作
func BenchmarkEntryCASWithAlloc(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				old := m.Load()
				m.CompareAndSwap(old, i)
				i++
			}
		},
	})
}

// ============================================================
// Value vs Entry 对比测试
// ============================================================

// BenchmarkValueVsEntryLoad 对比 Value 和 Entry 的 Load 性能
func BenchmarkValueVsEntryLoad(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			for pb.Next() {
				m.Load()
			}
		},
	})
}

// BenchmarkValueVsEntryStore 对比 Value 和 Entry 的 Store 性能
func BenchmarkValueVsEntryStore(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			i := 0
			for pb.Next() {
				m.Store(i)
				i++
			}
		},
	})
}

// BenchmarkValueVsEntrySwap 对比 Value 和 Entry 的 Swap 性能
func BenchmarkValueVsEntrySwap(b *testing.B) {
	var counter int64
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(int64(0))
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			for pb.Next() {
				m.Swap(atomic.AddInt64(&counter, 1))
			}
		},
	})
}

// BenchmarkValueVsEntryCAS 对比 Value 和 Entry 的 CAS 性能
func BenchmarkValueVsEntryCAS(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m atomany.Interface) {
			m.Store(0)
		},
		perG: func(b *testing.B, pb *testing.PB, id int, m atomany.Interface) {
			for pb.Next() {
				old := m.Load()
				m.CompareAndSwap(old, old.(int)+1)
			}
		},
	})
}
