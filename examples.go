package atomany

import (
	"fmt"
	"time"
)

// ExampleTypedValue 展示类型安全原子值的使用
func ExampleTypedValue() {
	// 创建类型安全的原子值
	tv := NewTypedValue(42)

	// 读取值
	fmt.Println("初始值:", tv.Load())

	// 存储新值
	tv.Store(100)
	fmt.Println("新值:", tv.Load())

	// 交换值
	old := tv.Swap(200)
	fmt.Println("交换前的值:", old)
	fmt.Println("交换后的值:", tv.Load())

	// 比较并交换
	swapped := tv.CompareAndSwap(200, 300)
	fmt.Println("CAS 成功:", swapped)
	fmt.Println("当前值:", tv.Load())

	// 尝试加载
	val, ok := tv.TryLoad()
	fmt.Printf("TryLoad: val=%v, ok=%v\n", val, ok)

	// 尝试存储
	ok = tv.TryStore(400)
	fmt.Println("TryStore (已存在):", ok)

	// 获取或存储
	actual, loaded := tv.GetOrStore(500)
	fmt.Printf("GetOrStore: actual=%v, loaded=%v\n", actual, loaded)

	// Output:
	// 初始值: 42
	// 新值: 100
	// 交换前的值: 100
	// 交换后的值: 200
	// CAS 成功: true
	// 当前值: 300
	// TryLoad: val=300, ok=true
	// TryStore (已存在): false
	// GetOrStore: actual=300, loaded=true
}

// ExampleExtendedEntry 展示扩展功能的使用
func ExampleExtendedEntry() {
	var e Entry

	// 尝试加载（未初始化）
	val, ok := e.TryLoad()
	fmt.Printf("TryLoad (未初始化): val=%v, ok=%v\n", val, ok)

	// 尝试存储（首次）
	ok = e.TryStore(42)
	fmt.Println("TryStore (首次):", ok)

	// 尝试存储（已存在）
	ok = e.TryStore(100)
	fmt.Println("TryStore (已存在):", ok)

	// 获取或存储
	actual, loaded := e.GetOrStore(200)
	fmt.Printf("GetOrStore: actual=%v, loaded=%v\n", actual, loaded)

	// LoadOrStore (同 GetOrStore)
	actual, loaded = e.LoadOrStore(300)
	fmt.Printf("LoadOrStore: actual=%v, loaded=%v\n", actual, loaded)

	// Output:
	// TryLoad (未初始化): val=<nil>, ok=false
	// TryStore (首次): true
	// TryStore (已存在): false
	// GetOrStore: actual=42, loaded=true
	// LoadOrStore: actual=42, loaded=true
}

// ExampleStats 展示统计功能的使用
func ExampleStats() {
	var e EntryWithStats

	// 执行一些操作
	for i := 0; i < 10; i++ {
		e.Store(i)
		e.Load()
	}

	// 获取统计信息
	stats := e.GetStats()
	fmt.Printf("Load 次数: %d\n", stats.LoadCount)
	fmt.Printf("Store 次数: %d\n", stats.StoreCount)

	// 重置统计
	e.ResetStats()
	stats = e.GetStats()
	fmt.Printf("重置后 Load 次数: %d\n", stats.LoadCount)

	// Output:
	// Load 次数: 10
	// Store 次数: 10
	// 重置后 Load 次数: 0
}

// ExampleConfigHotReload 展示配置热更新的使用场景
func ExampleConfigHotReload() {
	type Config struct {
		Port    int
		Timeout time.Duration
	}

	// 使用类型安全的原子值存储配置
	config := NewTypedValue(&Config{
		Port:    8080,
		Timeout: 30 * time.Second,
	})

	// 读取配置
	cfg := config.Load()
	fmt.Printf("初始配置: Port=%d, Timeout=%v\n", cfg.Port, cfg.Timeout)

	// 热更新配置
	newConfig := &Config{
		Port:    9090,
		Timeout: 60 * time.Second,
	}
	config.Store(newConfig)

	// 读取更新后的配置
	cfg = config.Load()
	fmt.Printf("更新后配置: Port=%d, Timeout=%v\n", cfg.Port, cfg.Timeout)

	// Output:
	// 初始配置: Port=8080, Timeout=30s
	// 更新后配置: Port=9090, Timeout=1m0s
}

// State 接口定义
type State interface {
	Handle() State
}

// InitState 初始状态
type InitState struct{}

func (s *InitState) Handle() State {
	fmt.Println("初始化中...")
	return &RunningState{}
}

// RunningState 运行状态
type RunningState struct{}

func (s *RunningState) Handle() State {
	fmt.Println("运行中...")
	return &StoppedState{}
}

// StoppedState 停止状态
type StoppedState struct{}

func (s *StoppedState) Handle() State {
	fmt.Println("已停止")
	return nil
}

// ExampleStateMachine 展示状态机的使用场景
func ExampleStateMachine() {
	// 使用 Entry 存储状态
	var state Entry
	state.Store(&InitState{})

	// 状态转换
	for {
		current := state.Load().(State)
		if current == nil {
			break
		}
		next := current.Handle()
		if next == nil {
			break
		}
		state.Store(next)
	}

	// Output:
	// 初始化中...
	// 运行中...
	// 已停止
}

// ExampleCache 展示缓存实现的使用场景
func ExampleCache() {
	var cache Entry

	// 更新缓存
	updateCache := func(key string, value any) {
		data := map[string]any{key: value}
		cache.Store(data)
	}

	// 读取缓存
	getCache := func(key string) any {
		if val := cache.Load(); val != nil {
			data := val.(map[string]any)
			return data[key]
		}
		return nil
	}

	// 使用缓存
	updateCache("user:1", "Alice")
	updateCache("user:2", "Bob")

	fmt.Println("user:1 =", getCache("user:1"))
	fmt.Println("user:2 =", getCache("user:2"))
	fmt.Println("user:3 =", getCache("user:3"))

	// Output:
	// user:1 = Alice
	// user:2 = Bob
	// user:3 = <nil>
}
