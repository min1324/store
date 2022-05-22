package store_test

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"math/rand"
	"runtime"
	"store"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func TestValue(t *testing.T) {
	var v *store.Value
	v = &store.Value{}
	if v.Load() != nil {
		t.Fatal("initial Value is not nil")
	}
	v.Store(42)
	x := v.Load()
	if xx, ok := x.(int); !ok || xx != 42 {
		t.Fatalf("wrong value: got %+v, want 42", x)
	}
	v.Store(84)
	x = v.Load()
	if xx, ok := x.(int); !ok || xx != 84 {
		t.Fatalf("wrong value: got %+v, want 84", x)
	}
	v.Store(nil)
	x = v.Load()
	if _, ok := x.(int); ok {
		t.Fatalf("wrong value: got %+v, want nil", x)
	}

	v = &store.Value{}
	if xx := v.Swap(nil); xx != nil {
		t.Fatalf("Swap wrong value: got %+v, want nil", xx)
	}
	if xx, ok := v.Swap(1).(int); ok {
		t.Fatalf("Swap wrong value: got %+v, want nil", xx)
	}
	if xx, ok := v.Swap(2).(int); !ok || xx != 1 {
		t.Fatalf("Swap wrong value: got %+v, want 1", xx)
	}
	if xx := v.Swap(nil); xx != 2 {
		t.Fatalf("Swap wrong value: got %+v, want 2", xx)
	}

	v = &store.Value{}
	p := v.Load()
	if !v.CompareAndSwap(nil, 1) {
		t.Fatalf("cas wrong value: %+v, old nil, new 1", p)
	}
	p = v.Load()
	if !v.CompareAndSwap(1, 2) {
		t.Fatalf("cas wrong value:  %+v, old 1, new 2", p)
	}
	p = v.Load()
	if !v.CompareAndSwap(2, nil) {
		t.Fatalf("cas wrong value:  %+v, old 2, new nil", p)
	}
}

func TestValueLarge(t *testing.T) {
	var v store.Value
	v.Store("foo")
	x := v.Load()
	if xx, ok := x.(string); !ok || xx != "foo" {
		t.Fatalf("wrong value: got %+v, want foo", x)
	}
	v.Store("barbaz")
	x = v.Load()
	if xx, ok := x.(string); !ok || xx != "barbaz" {
		t.Fatalf("wrong value: got %+v, want barbaz", x)
	}
}

func TestValuePanic(t *testing.T) {
	const badErr = "store: store of inconsistently typed value into Value"
	var v store.Value
	v.Store(42)
	func() {
		defer func() {
			err := recover()
			if err != badErr {
				t.Fatalf("inconsistent store panic: got '%v', want '%v'", err, badErr)
			}
		}()
		v.Store("foo")
	}()
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
	for _, test := range tests {
		var v store.Value
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
}

func BenchmarkValueRead(b *testing.B) {
	var v store.Value
	v.Store(new(int))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			x := v.Load().(*int)
			if *x != 0 {
				b.Fatalf("wrong value: got %v, want 0", *x)
			}
		}
	})
}

var Value_ValueSwapTests = []struct {
	init any
	new  any
	want any
	err  any
}{
	{init: nil, new: nil, want: nil, err: nil},
	{init: nil, new: true, want: nil, err: nil},
	{init: true, new: "", err: "store: swap of inconsistently typed value into Value"},
	{init: true, new: false, want: true, err: nil},
	{init: true, new: nil, want: true, err: nil},
}

func TestValue_Swap(t *testing.T) {
	for i, tt := range Value_ValueSwapTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var v store.Value
			if tt.init != nil {
				v.Store(tt.init)
			}
			defer func() {
				err := recover()
				switch {
				case tt.err == nil && err != nil:
					t.Errorf("should not panic, got %v", err)
				case tt.err != nil && err == nil:
					t.Errorf("should panic %v, got <nil>", tt.err)
				}
			}()
			if got := v.Swap(tt.new); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if got := v.Load(); got != tt.new {
				t.Errorf("got %v, want %v", got, tt.new)
			}
		})
	}
}

func TestValueSwapConcurrent(t *testing.T) {
	var v store.Value
	var count uint64
	var g sync.WaitGroup
	var m, n uint64 = 10000, 10000
	if testing.Short() {
		m = 1000
		n = 1000
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
}

var Value_CompareAndSwapTests = []struct {
	init any
	new  any
	old  any
	want bool
	err  any
}{
	{init: nil, new: nil, old: nil, want: true, err: nil},
	{init: nil, new: true, old: "", err: "store: compare and swap of inconsistently typed values into Value"},
	{init: nil, new: true, old: true, want: false, err: nil},
	{init: nil, new: true, old: nil, want: true, err: nil},
	{init: true, new: "", err: "store: compare and swap of inconsistently typed value into Value"},
	{init: true, new: true, old: false, want: false, err: nil},
	{init: true, new: true, old: true, want: true, err: nil},
	{init: true, new: nil, old: true, want: true, err: nil},
	{init: true, new: nil, old: nil, want: false, err: nil},
	{init: heapA, new: struct{ uint }{1}, old: heapB, want: true, err: nil},
}

func TestValue_CompareAndSwap(t *testing.T) {
	for i, tt := range Value_CompareAndSwapTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var v store.Value
			if tt.init != nil {
				v.Store(tt.init)
			}
			defer func() {
				err := recover()
				switch {
				case tt.err == nil && err != nil:
					t.Errorf("got %v, wanted no panic", err)
				case tt.err != nil && err == nil:
					t.Errorf("did not panic, want %v", tt.err)
				}
			}()
			if got := v.CompareAndSwap(tt.old, tt.new); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueCompareAndSwapConcurrent(t *testing.T) {
	var v store.Value
	var w sync.WaitGroup
	v.Store(0)
	m, n := 1000, 100
	if testing.Short() {
		m = 100
		n = 100
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
		t.Errorf("did not get to %v, stopped at %v", m*n, stop)
	}
}
