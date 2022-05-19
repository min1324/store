package store_test

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import (
	"fmt"
	"math/rand"
	"runtime"
	"store"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

type any = interface{}

type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

type iface interface {
	Load() (val any)
	Store(val any)
	Swap(new any) (old any)
	CompareAndSwap(old, new any) (swapped bool)
}

func fmtfn(name string, got, want any) string {
	return fmt.Sprintf("%+s wrong value: got %+v, want %+v\n", name, got, want)
}

func newFactor(f func(name string, e iface)) {
	for _, v := range []struct {
		iface
		name string
	}{
		{iface: &store.Value{}, name: "store"},
		// {iface: &store.Any{}, name: "Any"},
	} {
		f(v.name, v.iface)
	}
}

func TestValue(t *testing.T) {
	newFactor(func(name string, v iface) {
		if v.Load() != nil {
			t.Fatal(name + "initial Value is not nil")
		}

		v.Store(nil)
		if xx := v.Load(); xx != nil {
			t.Fatalf("%s wrong value nil: got %v, want nil", name, xx)
		}
		var valueNil = unsafe.Pointer(new(interface{}))
		v.Store(valueNil)
		if xx := v.Load(); xx != valueNil {
			t.Fatalf("wrong value vnil: got %+v, want %v", xx, valueNil)
		}
		v.Store(42)
		x := v.Load()
		if xx, ok := x.(int); !ok || xx != 42 {
			t.Fatalf("wrong value: got %+v, want 42", x)
		}
		v.Store(int64(84))
		x = v.Load()
		if xx, ok := x.(int64); !ok || xx != 84 {
			t.Fatalf("wrong value: got %+v, want 84", x)
		}
	})
}

func TestValueConcurrent(t *testing.T) {
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
	newFactor(func(name string, v iface) {
		for _, test := range tests {
			done := make(chan bool, p)
			for i := 0; i < p; i++ {
				go func() {
					r := rand.New(rand.NewSource(rand.Int63()))
					expected := true
				loop:
					for j := 0; j < N; j++ {
						x := test[r.Intn(len(test))]
						v.Store(x)
						x = v.Load()
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
	})
}

var Value_SwapTests = []struct {
	init any
	new  any
	want any
	err  any
}{
	{init: nil, new: "asd", want: nil},
	{init: nil, new: true, want: nil},
	{init: nil, new: nil, want: nil},
	{init: true, new: nil, want: true},
	{init: true, new: false, want: true},
	{init: true, new: true, want: true},
	{init: false, new: true, want: false},
}

func TestValue_Swap(t *testing.T) {
	newFactor(func(name string, v iface) {
		for i, tt := range Value_SwapTests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				v.Store(tt.init)
				defer func() {
					err := recover()
					switch {
					case err != nil:
						t.Errorf(" should not panic, got %v", err)
					}
				}()
				if got := v.Swap(tt.new); got != tt.want {
					t.Error(fmtfn(name+" Swap ", got, tt.want))
				}
				if got := v.Load(); got != tt.new {
					t.Error(fmtfn(name+" load ", got, tt.want))
				}
			})
		}
	})
}

func TestValueSwapConcurrent(t *testing.T) {
	newFactor(func(name string, v iface) {
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
					if old := v.Swap(new); old != nil {
						c += old.(uint64)
					}
				}
				atomic.AddUint64(&count, c)
				g.Done()
			}()
		}
		g.Wait()
		if want, got := (m*n-1)*(m*n)/2, count+v.Load().(uint64); got != want {
			t.Errorf("sum from 0 to %d was %d, want %v", m*n-1, got, want)
		}
	})
}

var heapA, heapB = struct{ uint }{0}, struct{ uint }{0}

var Value_CompareAndSwapTests = []struct {
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

func TestValue_CompareAndSwap(t *testing.T) {
	newFactor(func(name string, v iface) {
		for _, tt := range Value_CompareAndSwapTests {
			if tt.init != nil {
				v.Store(tt.init)
			}
			defer func() {
				err := recover()
				switch {
				case err != nil:
					t.Errorf(" got %v, wanted no panic", err)
				}
			}()
			if got := v.CompareAndSwap(tt.old, tt.new); got != tt.want {
				t.Errorf(fmtfn("", got, tt.want))
			}
		}
	})
}

func TestValueCompareAndSwapConcurrent(t *testing.T) {
	newFactor(func(name string, v iface) {
		var w sync.WaitGroup
		v.Store(0)
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
					if v.CompareAndSwap(j, j+1) {
						j += m
					}
				}
				w.Done()
			}()
		}
		w.Wait()
		if stop := v.Load().(int); stop != m*n {
			t.Errorf(" did not get to %v, stopped at %v", m*n, stop)
		}
	})
}
