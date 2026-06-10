package atomany

import (
	"sync"
	"testing"
	"time"
)

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
func TestValueConcurrent(t *testing.T) {
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
				e.Load()
				e.Swap(id)
				e.CompareAndSwap(id, id+1)
			}
		}(i)
	}

	wg.Wait()
}

// TestEntryExtendedMethods 测试 Entry 的扩展方法
func TestEntryExtendedMethods(t *testing.T) {
	var e Entry

	// 测试 TryLoad
	val, ok := e.TryLoad()
	if ok || val != nil {
		t.Errorf("未初始化的 TryLoad 应返回 (nil, false), 得到 (%v, %v)", val, ok)
	}

	// 测试 TryStore
	ok = e.TryStore(42)
	if !ok {
		t.Error("首次 TryStore 应成功")
	}

	ok = e.TryStore(100)
	if ok {
		t.Error("已存在时 TryStore 应失败")
	}

	// 测试 GetOrStore
	val, loaded := e.GetOrStore(200)
	if !loaded || val != 42 {
		t.Errorf("GetOrStore 应返回已存在的值 (42, true), 得到 (%v, %v)", val, loaded)
	}

	// 测试 LoadOrStore
	val, loaded = e.LoadOrStore(300)
	if !loaded || val != 42 {
		t.Errorf("LoadOrStore 应返回已存在的值 (42, true), 得到 (%v, %v)", val, loaded)
	}
}

// TestTypedValueBasic 测试 TypedValue 的基本功能
func TestTypedValueBasic(t *testing.T) {
	tv := NewTypedValue(42)

	// 测试 Load
	if val := tv.Load(); val != 42 {
		t.Errorf("Load 应返回 42, 得到 %v", val)
	}

	// 测试 Store
	tv.Store(100)
	if val := tv.Load(); val != 100 {
		t.Errorf("Store 后 Load 应返回 100, 得到 %v", val)
	}

	// 测试 Swap
	old := tv.Swap(200)
	if old != 100 {
		t.Errorf("Swap 应返回旧值 100, 得到 %v", old)
	}
	if val := tv.Load(); val != 200 {
		t.Errorf("Swap 后 Load 应返回 200, 得到 %v", val)
	}

	// 测试 CompareAndSwap
	swapped := tv.CompareAndSwap(200, 300)
	if !swapped {
		t.Error("CompareAndSwap 应成功")
	}
	if val := tv.Load(); val != 300 {
		t.Errorf("CAS 后 Load 应返回 300, 得到 %v", val)
	}
}

// TestTypedValueExtendedMethods 测试 TypedValue 的扩展方法
func TestTypedValueExtendedMethods(t *testing.T) {
	tv := NewTypedValue(42)

	// 测试 TryLoad
	val, ok := tv.TryLoad()
	if !ok || val != 42 {
		t.Errorf("TryLoad 应返回 (42, true), 得到 (%v, %v)", val, ok)
	}

	// 测试 TryStore - 已存在时应失败
	ok = tv.TryStore(100)
	if ok {
		t.Error("TryStore 应失败（已存在）")
	}

	// 验证值未被更新
	if val := tv.Load(); val != 42 {
		t.Errorf("TryStore 失败后 Load 应仍返回 42, 得到 %v", val)
	}

	// 测试 GetOrStore
	actual, loaded := tv.GetOrStore(200)
	if !loaded || actual != 42 {
		t.Errorf("GetOrStore 应返回已存在的值 (42, true), 得到 (%v, %v)", actual, loaded)
	}

	// 测试未初始化的 TypedValue
	var tv2 TypedValue[int]

	// TryStore 应成功
	ok = tv2.TryStore(100)
	if !ok {
		t.Error("未初始化时 TryStore 应成功")
	}

	// 验证值已被设置
	if val := tv2.Load(); val != 100 {
		t.Errorf("TryStore 后 Load 应返回 100, 得到 %v", val)
	}

	// 再次 TryStore 应失败
	ok = tv2.TryStore(200)
	if ok {
		t.Error("已存在时 TryStore 应失败")
	}
}

// TestEntryWithStatsBasic 测试 EntryWithStats 的基本功能
func TestEntryWithStatsBasic(t *testing.T) {
	var e EntryWithStats

	// 执行一些操作
	e.Store(42)
	e.Load()
	e.Swap(100)
	e.CompareAndSwap(100, 200)

	// 获取统计信息
	stats := e.GetStats()
	if stats.LoadCount != 1 {
		t.Errorf("LoadCount 应为 1, 得到 %d", stats.LoadCount)
	}
	if stats.StoreCount != 1 {
		t.Errorf("StoreCount 应为 1, 得到 %d", stats.StoreCount)
	}
	if stats.SwapCount != 1 {
		t.Errorf("SwapCount 应为 1, 得到 %d", stats.SwapCount)
	}
	if stats.CASCount != 1 {
		t.Errorf("CASCount 应为 1, 得到 %d", stats.CASCount)
	}
	if stats.CASSuccess != 1 {
		t.Errorf("CASSuccess 应为 1, 得到 %d", stats.CASSuccess)
	}

	// 重置统计
	e.ResetStats()
	stats = e.GetStats()
	if stats.LoadCount != 0 || stats.StoreCount != 0 {
		t.Error("重置后统计信息应全为 0")
	}
}

// TestRWEntryBasic 测试 RWEntry 的基本功能
func TestRWEntryBasic(t *testing.T) {
	var e RWEntry

	// 测试 Store 和 Load
	e.Store(42)
	if val := e.Load(); val != 42 {
		t.Errorf("Load 应返回 42, 得到 %v", val)
	}

	// 测试 Swap
	old := e.Swap(100)
	if old != 42 {
		t.Errorf("Swap 应返回旧值 42, 得到 %v", old)
	}

	// 测试 CompareAndSwap
	swapped := e.CompareAndSwap(100, 200)
	if !swapped {
		t.Error("CompareAndSwap 应成功")
	}
	if val := e.Load(); val != 200 {
		t.Errorf("CAS 后 Load 应返回 200, 得到 %v", val)
	}
}

// TestRWEntryConcurrent 测试 RWEntry 的并发安全性
func TestRWEntryConcurrent(t *testing.T) {
	var e RWEntry
	e.Store(0)

	var wg sync.WaitGroup
	// 启动多个读取 goroutine
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				e.Load()
			}
		}()
	}

	// 启动少量写入 goroutine
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				e.Store(id)
				e.Swap(id + 1)
				e.CompareAndSwap(id+1, id+2)
			}
		}(i)
	}

	wg.Wait()
}

// TestCacheEntryBasic 测试 CacheEntry 的基本功能
func TestCacheEntryBasic(t *testing.T) {
	var e CacheEntry

	// 测试 Store 和 Load
	e.Store(42)
	if val := e.Load(); val != 42 {
		t.Errorf("Load 应返回 42, 得到 %v", val)
	}

	// 测试 Swap
	old := e.Swap(100)
	if old != 42 {
		t.Errorf("Swap 应返回旧值 42, 得到 %v", old)
	}

	// 测试 CompareAndSwap
	swapped := e.CompareAndSwap(100, 200)
	if !swapped {
		t.Error("CompareAndSwap 应成功")
	}
	if val := e.Load(); val != 200 {
		t.Errorf("CAS 后 Load 应返回 200, 得到 %v", val)
	}

	// 测试 Refresh
	e.Store(300)
	e.Refresh()
	if val := e.Load(); val != 300 {
		t.Errorf("Refresh 后 Load 应返回 300, 得到 %v", val)
	}
}

// TestCacheEntryCaching 测试 CacheEntry 的缓存功能
func TestCacheEntryCaching(t *testing.T) {
	var e CacheEntry
	e.Store(42)

	// 多次 Load 应从缓存读取
	val1 := e.Load()
	val2 := e.Load()
	if val1 != val2 {
		t.Error("缓存值应一致")
	}

	// 更新值
	e.Store(100)

	// 缓存可能未刷新，需要手动刷新
	e.Refresh()
	if val := e.Load(); val != 100 {
		t.Errorf("Refresh 后 Load 应返回 100, 得到 %v", val)
	}
}

// TestInterfaceCompliance 测试所有类型是否实现 Interface 接口
func TestInterfaceCompliance(t *testing.T) {
	var _ Interface = (*Value)(nil)
	var _ Interface = (*Entry)(nil)
	var _ Interface = (*RWEntry)(nil)
	var _ Interface = (*CacheEntry)(nil)
	var _ Interface = (*EntryWithStats)(nil)
	var _ Interface = (*ValueWithStats)(nil)
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 测试错误检查函数
	if !IsUninitialized(ErrUninitialized) {
		t.Error("IsUninitialized 应返回 true")
	}
	if !IsTypeMismatch(ErrTypeMismatch) {
		t.Error("IsTypeMismatch 应返回 true")
	}
	if !IsNilValue(ErrNilValue) {
		t.Error("IsNilValue 应返回 true")
	}
	if !IsInconsistentType(ErrInconsistentType) {
		t.Error("IsInconsistentType 应返回 true")
	}
}

// TestStateExample 测试状态机示例
func TestStateExample(t *testing.T) {
	var state Entry
	state.Store(&InitState{})

	// 状态转换：Init -> Running -> Stopped -> nil
	for {
		current := state.Load().(State)
		if current == nil {
			break
		}
		next := current.Handle()
		if next == nil {
			// 最后一个状态返回 nil，存储 nil
			state.Store(nil)
			break
		}
		state.Store(next)
	}

	// 最终状态应为 nil
	if state.Load() != nil {
		t.Error("状态转换应完成")
	}
}

// TestConcurrentConfigUpdate 测试并发配置更新
func TestConcurrentConfigUpdate(t *testing.T) {
	type Config struct {
		Port    int
		Timeout time.Duration
	}

	config := NewTypedValue(&Config{
		Port:    8080,
		Timeout: 30 * time.Second,
	})

	var wg sync.WaitGroup
	// 启动多个读取 goroutine
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := config.Load()
			if cfg == nil {
				t.Error("配置不应为 nil")
			}
		}()
	}

	// 启动更新 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		config.Store(&Config{
			Port:    9090,
			Timeout: 60 * time.Second,
		})
	}()

	wg.Wait()
}
