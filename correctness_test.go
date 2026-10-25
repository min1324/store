package store_test

import (
	"store"
	"sync"
	"testing"
)

type Value = store.Value
type Entry = store.Entry

// TestValueBasic 测试 Value 的基本功能
func TestValueBasic(t *testing.T) {
	var v Value

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

// TestValueNilSupport 测试 Value 的 nil 支持
func TestValueNilSupport(t *testing.T) {
	var v Value

	// 初始化
	v.Store(42)

	// 存储 nil
	v.Store(nil)
	if val := v.Load(); val != nil {
		t.Errorf("存储 nil 后 Load 应返回 nil, 得到 %v", val)
	}

	// 从 nil 恢复
	v.Store(100)
	if val := v.Load(); val != 100 {
		t.Errorf("从 nil 恢复后 Load 应返回 100, 得到 %v", val)
	}
}

// TestValueConcurrent 测试 Value 的并发安全性
func TestValueConcurrent_Value(t *testing.T) {
	var v Value
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
}

// TestEntryBasic 测试 Entry 的基本功能
func TestEntryBasic(t *testing.T) {
	var e Entry

	// 测试未初始化的 Load
	if val := e.Load(); val != nil {
		t.Errorf("未初始化的 Load 应返回 nil, 得到 %v", val)
	}

	// 测试 Store
	e.Store(42)
	if val := e.Load(); val != 42 {
		t.Errorf("Store 后 Load 应返回 42, 得到 %v", val)
	}

	// 测试 Swap
	old := e.Swap(100)
	if old != 42 {
		t.Errorf("Swap 应返回旧值 42, 得到 %v", old)
	}
	if val := e.Load(); val != 100 {
		t.Errorf("Swap 后 Load 应返回 100, 得到 %v", val)
	}

	// 测试 CompareAndSwap
	swapped := e.CompareAndSwap(100, 200)
	if !swapped {
		t.Error("CompareAndSwap 应成功")
	}
	if val := e.Load(); val != 200 {
		t.Errorf("CAS 后 Load 应返回 200, 得到 %v", val)
	}

	swapped = e.CompareAndSwap(100, 300)
	if swapped {
		t.Error("CompareAndSwap 应失败")
	}
}

// TestEntryMultiType 测试 Entry 的多类型支持
func TestEntryMultiType(t *testing.T) {
	var e Entry

	// 存储不同类型的值
	e.Store(42)
	if val := e.Load(); val != 42 {
		t.Errorf("应返回 42, 得到 %v", val)
	}

	e.Store("hello")
	if val := e.Load(); val != "hello" {
		t.Errorf("应返回 'hello', 得到 %v", val)
	}

	e.Store([]int{1, 2, 3})
	if val := e.Load(); len(val.([]int)) != 3 {
		t.Errorf("应返回长度为 3 的切片, 得到 %v", val)
	}
}

// TestEntryNilSupport 测试 Entry 的 nil 支持
func TestEntryNilSupport(t *testing.T) {
	var e Entry

	// 直接存储 nil
	e.Store(nil)
	if val := e.Load(); val != nil {
		t.Errorf("存储 nil 后 Load 应返回 nil, 得到 %v", val)
	}

	// 从 nil 恢复
	e.Store(100)
	if val := e.Load(); val != 100 {
		t.Errorf("从 nil 恢复后 Load 应返回 100, 得到 %v", val)
	}
}

// TestEntryConcurrent 测试 Entry 的并发安全性
func TestEntryConcurrent(t *testing.T) {
	var e Entry
	e.Store(0)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				e.CompareAndSwap(id, id+1)
				e.Load()
				e.Swap(id)

			}
		}(i)
	}

	wg.Wait()
}
