package atomany_test

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"atomany"
)

// ---------------------------------------------------------------------------
// Entry basic tests
// ---------------------------------------------------------------------------

func TestEntryLoadNil(t *testing.T) {
	var e atomany.Entry
	if e.Load() != nil {
		t.Fatal("initial Entry is not nil")
	}
}

func TestEntryStoreAndLoad(t *testing.T) {
	var e atomany.Entry
	e.Store(42)
	x := e.Load()
	if xx, ok := x.(int); !ok || xx != 42 {
		t.Fatalf("wrong value: got %+v, want 42", x)
	}
	e.Store(84)
	x = e.Load()
	if xx, ok := x.(int); !ok || xx != 84 {
		t.Fatalf("wrong value: got %+v, want 84", x)
	}
}

func TestEntryStoreNil(t *testing.T) {
	var e atomany.Entry
	e.Store(42)
	e.Store(nil)
	x := e.Load()
	if x != nil {
		t.Fatalf("wrong value: got %+v, want nil", x)
	}
}

func TestEntryStoreNilFirst(t *testing.T) {
	var e atomany.Entry
	e.Store(nil)
	if e.Load() != nil {
		t.Fatal("Entry should load nil after Store(nil)")
	}
	e.Store(1)
	if e.Load() != 1 {
		t.Fatalf("Entry should load 1, got %v", e.Load())
	}
}

func TestEntryStoreDifferentTypes(t *testing.T) {
	// Entry allows storing different types (unlike Value)
	var e atomany.Entry
	e.Store(42)
	e.Store("hello")
	x := e.Load()
	if xx, ok := x.(string); !ok || xx != "hello" {
		t.Fatalf("wrong value: got %+v, want hello", x)
	}
}

func TestEntryPtr(t *testing.T) {
	var e atomany.Entry
	e.Store(1)
	p := e.Ptr()
	if p == nil {
		t.Fatal("Ptr() should not return nil after Store")
	}
	q := *(*any)(p)
	if xx := q.(int); xx != 1 {
		t.Fatalf("wrong value ptr: got %+v, want 1", xx)
	}
}

func TestEntryPtrNil(t *testing.T) {
	var e atomany.Entry
	p := e.Ptr()
	if p != nil {
		t.Fatal("Ptr() should return nil for uninitialized Entry")
	}
}

func TestEntryUnsafePointerValue(t *testing.T) {
	var e atomany.Entry
	valueNil := unsafe.Pointer(new(interface{}))
	e.Store(valueNil)
	if xx := e.Load(); xx != valueNil {
		t.Fatalf("wrong value: got %+v, want %+v", xx, valueNil)
	}
}

func TestEntryInitStoreNil(t *testing.T) {
	var e atomany.Entry
	e.Store(nil)
	if xx := e.Load(); xx != nil {
		t.Fatalf("store nil: got %+v, want nil", xx)
	}
	e.Store(42)
	x := e.Load()
	if xx, ok := x.(int); !ok || xx != 42 {
		t.Fatalf("wrong value: got %+v, want 42", x)
	}
	e.Store(int64(84))
	x = e.Load()
	if xx, ok := x.(int64); !ok || xx != 84 {
		t.Fatalf("wrong value: got %+v, want 84", x)
	}
}

// ---------------------------------------------------------------------------
// Entry Swap tests
// ---------------------------------------------------------------------------

var entrySwapTests = []struct {
	init any
	new  any
	want any
}{
	{init: nil, new: "asd", want: nil},
	{init: nil, new: true, want: nil},
	{init: nil, new: nil, want: nil},
	{init: true, new: nil, want: true},
	{init: true, new: false, want: true},
	{init: true, new: true, want: true},
	{init: false, new: true, want: false},
}

func TestEntrySwap(t *testing.T) {
	for i, tt := range entrySwapTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var e atomany.Entry
			if tt.init != nil {
				e.Store(tt.init)
			}
			if got := e.Swap(tt.new); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if got := e.Load(); got != tt.new {
				t.Errorf("got %v, want %v", got, tt.new)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Entry CompareAndSwap tests
// ---------------------------------------------------------------------------

var entryCASTests = []struct {
	init any
	new  any
	old  any
	want bool
}{
	{init: nil, old: nil, new: nil, want: true},
	{init: nil, old: "", new: true, want: false},
	{init: nil, old: true, new: true, want: false},
	{init: nil, old: nil, new: true, want: true},
	{init: nil, old: 0, new: 0, want: false},
	{init: 0, old: 0, new: 0, want: true},
	{init: true, old: nil, new: "", want: false},
	{init: true, old: false, new: true, want: false},
	{init: true, old: true, new: true, want: true},
	{init: true, old: true, new: nil, want: true},
	{init: 2, old: 2, new: int64(2), want: true},
	{init: 2, old: int64(2), new: 2, want: false},
	{init: heapA, old: heapB, new: struct{ uint }{1}, want: true},
}

func TestEntryCompareAndSwap(t *testing.T) {
	for i, tt := range entryCASTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var e atomany.Entry
			if tt.init != nil {
				e.Store(tt.init)
			}
			defer func() {
				err := recover()
				if err != nil {
					t.Errorf("got %v, wanted no panic", err)
				}
			}()
			if got := e.CompareAndSwap(tt.old, tt.new); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Entry concurrent tests
// ---------------------------------------------------------------------------

func TestEntryConcurrent(t *testing.T) {
	tests := [][]any{
		{uint16(0), ^uint16(0), uint16(1 + 2<<8), uint16(3 + 4<<8)},
		{uint32(0), ^uint32(0), uint32(1 + 2<<16), uint32(3 + 4<<16)},
		{uint64(0), ^uint64(0), uint64(1 + 2<<32), uint64(3 + 4<<32)},
		{complex(0, 0), complex(1, 2), complex(3, 4), complex(5, 6)},
	}
	p := 4 * runtime.GOMAXPROCS(0)
	N := int(1e5)
	if testing.Short() {
		p /= 2
		N = 1e3
	}
	for _, test := range tests {
		var e atomany.Entry
		done := make(chan bool, p)
		for i := 0; i < p; i++ {
			go func() {
				r := rand.New(rand.NewSource(rand.Int63()))
				expected := true
			loop:
				for j := 0; j < N; j++ {
					x := test[r.Intn(len(test))]
					e.Store(x)
					x = e.Load()
					for _, x1 := range test {
						if x == x1 {
							continue loop
						}
					}
					t.Logf("loaded unexpected value %+v, want %+v", x, test)
					expected = false
					break
				}
				done <- expected
			}()
		}
		for i := 0; i < p; i++ {
			if !<-done {
				t.FailNow()
			}
		}
	}
}

func TestEntrySwapConcurrent(t *testing.T) {
	var e atomany.Entry
	var count uint64
	var g sync.WaitGroup
	var m, n uint64 = 100, 100
	if testing.Short() {
		m = 10
		n = 10
	}
	for i := uint64(0); i < m*n; i += n {
		i := i
		g.Add(1)
		go func() {
			var c uint64
			for new := i; new < i+n; new++ {
				if old := e.Swap(new); old != nil {
					c += old.(uint64)
				}
			}
			atomic.AddUint64(&count, c)
			g.Done()
		}()
	}
	g.Wait()
	if want, got := (m*n-1)*(m*n)/2, count+e.Load().(uint64); got != want {
		t.Errorf("sum from 0 to %d was %d, want %v", m*n-1, got, want)
	}
}

func TestEntryCASConcurrent(t *testing.T) {
	var e atomany.Entry
	var w sync.WaitGroup
	e.Store(0)
	m, n := 100, 100
	if testing.Short() {
		m = 10
		n = 10
	}
	for i := 0; i < m; i++ {
		i := i
		w.Add(1)
		go func() {
			for j := i; j < m*n; runtime.Gosched() {
				if e.CompareAndSwap(j, j+1) {
					j += m
				}
			}
			w.Done()
		}()
	}
	w.Wait()
	if stop := e.Load().(int); stop != m*n {
		t.Errorf("did not get to %v, stopped at %v", m*n, stop)
	}
}

// ---------------------------------------------------------------------------
// Entry benchmarks
// ---------------------------------------------------------------------------

func BenchmarkEntryRead(b *testing.B) {
	var e atomany.Entry
	e.Store(new(int))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			x := e.Load().(*int)
			if *x != 0 {
				b.Fatalf("wrong value: got %v, want 0", *x)
			}
		}
	})
}

func BenchmarkEntryStore(b *testing.B) {
	var e atomany.Entry
	e.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			e.Store(i)
			i++
		}
	})
}

func BenchmarkEntryStoreLoad(b *testing.B) {
	var e atomany.Entry
	e.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		var i = 0
		for pb.Next() {
			e.Store(i)
			x := e.Load().(int)
			if x != i {
				b.Fatalf("wrong value: got %v, want %v", x, i)
			}
		}
	})
}

func BenchmarkEntrySwap(b *testing.B) {
	var e atomany.Entry
	var i int64
	e.Store(int64(0))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.Swap(atomic.AddInt64(&i, 1))
		}
	})
}

func BenchmarkEntryCAS(b *testing.B) {
	var e atomany.Entry
	e.Store(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			j := e.Load().(int)
			e.CompareAndSwap(j, j+1)
		}
	})
}
