package store_test

import (
	"fmt"
	"store"
	"sync/atomic"
	"testing"
)

// NOTE each time use list,iface must init using iface.Store(nil)
var benchList = []struct {
	iface
	name string
}{
	{&atomic.Value{}, "atomic"},
	{&store.Entry{}, "Node"},
}

func BenchmarkValueRead(b *testing.B) {
	for _, v := range benchList {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", v.name), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					x := v.Load().(int)
					if x != 0 {
						b.Fatalf("%s wrong value: got %v, want 0", v.name, x)
					}
				}
			})
		})
	}
}

func BenchmarkValueStore(b *testing.B) {
	for _, v := range benchList {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", v.name), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					v.Store(i)
					i++
				}
			})
		})
	}
}

func BenchmarkValueStoreLoad(b *testing.B) {
	for _, v := range benchList {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", v.name), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				var i = 0
				for pb.Next() {
					v.Store(i)
					x := v.Load().(int)
					if x != i {
						b.Fatalf("%s wrong value: got %v, want %v", v.name, x, i)
					}
				}
			})
		})
	}
}

func BenchmarkValueSwap(b *testing.B) {
	for _, v := range benchList {
		var i int64
		v.Store(i)
		b.Run(fmt.Sprintf("%s", v.name), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					v.Swap(atomic.AddInt64(&i, 1))
				}
			})
		})
	}
}

func BenchmarkValueCAS(b *testing.B) {
	for _, v := range benchList {
		var i = int64(0)
		v.Store(i)
		b.Run(fmt.Sprintf("%s", v.name), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					old := v.Load()
					v.CompareAndSwap(old, old)
					// v.CompareAndSwap(i, atomic.AddInt64(&i, 1))
				}
			})
		})
	}
}
