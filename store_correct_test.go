package store_test

import (
	"math/rand"
	"runtime"
	"store"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

type Interface interface {
	store.Interface

	TryLoad() (val any, loaded bool)
	TryStore(val any) (ok bool)
	LoadOrStore(val any) (actual any, loaded bool)
}

type Value = store.Value
type Entry = store.Entry

// TestValueBasic 测试 Value 的基本功能
func TestBasic_Value(t *testing.T) {
	testBasic(t, &Value{})
}

func TestBasic_Entry(t *testing.T) {
	testBasic(t, &Entry{})
}

func testBasic(t *testing.T, v Interface) {
	// 测试未初始化的 Load
	if val := v.Load(); val != nil {
		t.Errorf("未初始化的 Load 应返回 nil, 得到 %v", val)
	}

	// 测试 Store
	v.Store(42)
	if val := v.Load(); val != 42 {
		t.Errorf("Store 后 Load 应返回 42, 得到 %v", val)
	}

	// 测试 Swap
	old := v.Swap(100)
	if old != 42 {
		t.Errorf("Swap 应返回旧值 42, 得到 %v", old)
	}
	if val := v.Load(); val != 100 {
		t.Errorf("Swap 后 Load 应返回 100, 得到 %v", val)
	}

	// 测试 CompareAndSwap
	swapped := v.CompareAndSwap(100, 200)
	if !swapped {
		t.Error("CompareAndSwap 应成功")
	}
	if val := v.Load(); val != 200 {
		t.Errorf("CAS 后 Load 应返回 200, 得到 %v", val)
	}

	swapped = v.CompareAndSwap(100, 300)
	if swapped {
		t.Error("CompareAndSwap 应失败")
	}
}

// TestEntryNilSupport 测试 Entry 的 nil 支持
func TestNilSupport_Value(t *testing.T) {
	testNilSupport(t, &Value{})
}
func TestNilSupport_Entry(t *testing.T) {
	testNilSupport(t, &Entry{})
}

func testNilSupport(t *testing.T, s Interface) {
	s.Store(100)
	if val := s.Load(); val != 100 {
		t.Errorf("存储 100 后 Load 应返回 100, 得到 %v", val)
	}

	s.Store(nil)
	if val := s.Load(); val != nil {
		t.Errorf("存储 nil 后 Load 应返回 nil, 得到 %v", val)
	}

	s.Store(200)
	if val := s.Load(); val != 200 {
		t.Errorf("从 nil 恢复后 Load 应返回 200, 得到 %v", val)
	}
}

// TestEntryMultiType 测试 Entry 的多类型支持
func TestMultiType_Entry(t *testing.T) {
	testMultiType(t, &Entry{})
}

func testMultiType(t *testing.T, v Interface) {

	// 存储不同类型的值
	v.Store(42)
	if val := v.Load(); val != 42 {
		t.Errorf("应返回 42, 得到 %v", val)
	}

	if !v.CompareAndSwap(42, 100) {
		t.Error("CAS 应成功")
	}
	if val := v.Load(); val != 100 {
		t.Errorf("应返回 100, 得到 %v", val)
	}

	v.Store("hello")
	if val := v.Load(); val != "hello" {
		t.Errorf("应返回 'hello', 得到 %v", val)
	}

	v.Swap([]int{1, 2, 3})
	if val := v.Load(); len(val.([]int)) != 3 {
		t.Errorf("应返回长度为 3 的切片, 得到 %v", val)
	}
}

// TestValueConcurrent 测试 Value 的并发安全性
func TestValueConcurrent_Value(t *testing.T) {
	testConcurrent(t, &Value{})
}

func TestValueConcurrent_Entry(t *testing.T) {
	testConcurrent(t, &Entry{})
}

func testConcurrent(t *testing.T, v Interface) {
	v.Store(0)

	var wg sync.WaitGroup
	// 启动多个 goroutine 进行并发操作
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				v.Load()
				v.Swap(id)
				v.CompareAndSwap(id, id+1)
			}
		}(i)
	}

	wg.Wait()
	// 只要没有 panic 就算通过
	t.Log("pass")
}

// ==================================================================

func TestUnion_Value(t *testing.T) {
	testUnion(t, &Value{})
}
func TestUnion_Entry(t *testing.T) {
	testUnion(t, &Entry{})
}

func testUnion(t *testing.T, v Interface) {
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

func TestOperationsCombined_Value(t *testing.T) {
	operations_Combined(t, &Value{})
}

func TestOperationsCombined_Entry(t *testing.T) {
	operations_Combined(t, &Entry{})
}

func operations_Combined(t *testing.T, e Interface) {
	testValue := "test"
	newValue := "new"
	updatedValue := "updated"

	// 1. 测试初始状态
	val, ok := e.TryLoad()
	if val != nil || ok {
		t.Error("初始状态应该为未初始化")
	}

	// 2. 测试 LoadOrStore - 第一次调用应该存储并返回 false
	actual, loaded := e.LoadOrStore(testValue)
	if actual != testValue || loaded {
		t.Errorf("LoadOrStore 第一次调用: 期望 (%v, false)，得到 (%v, %v)", testValue, actual, loaded)
	}

	// 3. 测试 LoadOrStore - 第二次调用应该加载并返回 true
	actual, loaded = e.LoadOrStore(newValue)
	if actual != testValue || !loaded {
		t.Errorf("LoadOrStore 第二次调用: 期望 (%v, true)，得到 (%v, %v)", testValue, actual, loaded)
	}

	// 4. 测试 Load - 应该返回最近存储的值
	val = e.Load()
	if val != testValue {
		t.Errorf("Load: 期望 %v，得到 %v", testValue, val)
	}

	// 5. 测试 Store - 存储新值
	e.Store(newValue)
	val = e.Load()
	if val != newValue {
		t.Errorf("Store: 期望 %v，得到 %v", newValue, val)
	}

	// 6. 测试 Swap - 交换值并返回旧值
	old := e.Swap(updatedValue)
	if old != newValue {
		t.Errorf("Swap: 期望 %v，得到 %v", newValue, old)
	}
	val = e.Load()
	if val != updatedValue {
		t.Errorf("Swap后Load: 期望 %v，得到 %v", updatedValue, val)
	}

	// 7. 测试 CompareAndSwap - 成功的情况
	swapped := e.CompareAndSwap(updatedValue, testValue)
	if !swapped {
		t.Error("CompareAndSwap 应该成功")
	}
	val = e.Load()
	if val != testValue {
		t.Errorf("CompareAndSwap后Load: 期望 %v，得到 %v", testValue, val)
	}

	// 8. 测试 CompareAndSwap - 失败的情况
	swapped = e.CompareAndSwap(newValue, updatedValue)
	if swapped {
		t.Error("CompareAndSwap 应该失败")
	}
	val = e.Load()
	if val != testValue {
		t.Errorf("CompareAndSwap失败后Load: 期望 %v，得到 %v", testValue, val)
	}

	// 9. 测试 TryStore - 在已初始化的 Entry 上应该失败
	if e.TryStore(newValue) {
		t.Error("TryStore 应该失败")
	}
	val = e.Load()
	if val != testValue {
		t.Errorf("TryStore失败后Load: 期望 %v，得到 %v", testValue, val)
	}

	// 10. 测试 TryStore - 在新的 Entry 上应该成功
	var newEntry Entry
	if !newEntry.TryStore(testValue) {
		t.Error("TryStore 应该成功")
	}
	val = newEntry.Load()
	if val != testValue {
		t.Errorf("TryStore成功后Load: 期望 %v，得到 %v", testValue, val)
	}

	// 11. 测试操作序列：Store -> Load -> Swap -> Load -> CompareAndSwap -> Load
	e.Store("sequence")
	val = e.Load()
	if val != "sequence" {
		t.Errorf("Store->Load: 期望 'sequence'，得到 %v", val)
	}

	old = e.Swap("swap")
	if old != "sequence" {
		t.Errorf("Swap: 期望 'sequence'，得到 %v", old)
	}
	val = e.Load()
	if val != "swap" {
		t.Errorf("Swap后Load: 期望 'swap'，得到 %v", val)
	}

	swapped = e.CompareAndSwap("swap", "cas")
	if !swapped {
		t.Error("CompareAndSwap 应该成功")
	}
	val = e.Load()
	if val != "cas" {
		t.Errorf("CompareAndSwap后Load: 期望 'cas'，得到 %v", val)
	}

	// 12. 测试 nil 值的处理
	e.Store(nil)
	val = e.Load()
	if val != nil {
		t.Errorf("Store nil 后 Load: 期望 nil，得到 %v", val)
	}

	old = e.Swap("not nil")
	if old != nil {
		t.Errorf("Swap nil: 期望 nil，得到 %v", old)
	}
	val = e.Load()
	if val != "not nil" {
		t.Errorf("Swap nil 后 Load: 期望 'not nil'，得到 %v", val)
	}

	swapped = e.CompareAndSwap("not nil", nil)
	if !swapped {
		t.Error("CompareAndSwap(not nil, nil) 应该成功")
	}
	val = e.Load()
	if val != nil {
		t.Errorf("CompareAndSwap(not nil, nil) 后 Load: 期望 nil，得到 %v", val)
	}
}
