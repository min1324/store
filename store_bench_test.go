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
		{&store.Value{}, "Entry"},
		{&store.Entry{}, "store"},
	} {
		f(v.name, v.iface)
	}
}

func BenchmarkRead(b *testing.B) {
	benchFunc(func(name string, v iface) {
		v.Store(0)
		b.Run(fmt.Sprintf("%s", name), func(b *testing.B) {
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
