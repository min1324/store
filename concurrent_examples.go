package atomany

import (
	"fmt"
	"sync"
	"time"
)

// ExampleRWEntry 展示读写锁优化的 Entry 使用
func ExampleRWEntry() {
	var e RWEntry

	// 初始化
	e.Store("initial")

	// 启动多个读取 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				val := e.Load()
				fmt.Printf("Reader %d: %s\n", id, val)
			}
		}(i)
	}

	// 启动少量写入 goroutine
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				e.Store(fmt.Sprintf("writer-%d-%d", id, j))
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Done")

	// Output:
	// Reader 0: initial
	// Reader 1: initial
	// ...
	// Done
}

// ExampleBatchLoad 展示批量加载的使用
func ExampleBatchLoad() {
	// 创建多个 RWEntry
	entries := make([]*RWEntry, 5)
	for i := 0; i < 5; i++ {
		e := &RWEntry{}
		e.Store(i * 10)
		entries[i] = e
	}

	// 批量加载
	values := BatchLoad(entries...)
	fmt.Println("批量加载的值:", values)

	// Output:
	// 批量加载的值: [0 10 20 30 40]
}

// ExampleCacheEntry 展示缓存优化的 Entry 使用
func ExampleCacheEntry() {
	var e CacheEntry

	// 初始化
	e.Store("value1")

	// 多次读取（从缓存读取）
	for i := 0; i < 5; i++ {
		fmt.Println("读取:", e.Load())
	}

	// 更新值
	e.Store("value2")

	// 读取新值
	fmt.Println("更新后读取:", e.Load())

	// 手动刷新缓存
	e.Store("value3")
	e.Refresh()
	fmt.Println("刷新后读取:", e.Load())

	// Output:
	// 读取: value1
	// 读取: value1
	// 读取: value1
	// 读取: value1
	// 读取: value1
	// 更新后读取: value2
	// 刷新后读取: value3
}

// ExampleConcurrentConfig 展示并发配置更新的使用
func ExampleConcurrentConfig() {
	type Config struct {
		Port    int
		Timeout time.Duration
	}

	// 使用 RWValue 存储配置
	config := &RWValue{}
	config.Store(&Config{
		Port:    8080,
		Timeout: 30 * time.Second,
	})

	// 启动多个读取 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := config.Load().(*Config)
			fmt.Printf("Reader %d: Port=%d, Timeout=%v\n", id, cfg.Port, cfg.Timeout)
		}(i)
	}

	// 更新配置
	config.Store(&Config{
		Port:    9090,
		Timeout: 60 * time.Second,
	})

	wg.Wait()

	// Output:
	// Reader 0: Port=8080, Timeout=30s
	// Reader 1: Port=8080, Timeout=30s
	// Reader 2: Port=8080, Timeout=30s
	// Reader 3: Port=8080, Timeout=30s
	// Reader 4: Port=8080, Timeout=30s
}

// ExampleConcurrentCache 展示并发缓存的使用
func ExampleConcurrentCache() {
	// 使用 CacheEntry 实现缓存
	var cache CacheEntry

	// 初始化缓存
	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	cache.Store(data)

	// 启动多个读取 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cachedData := cache.Load().(map[string]string)
			fmt.Printf("Reader %d: key1=%s\n", id, cachedData["key1"])
		}(i)
	}

	// 更新缓存
	newData := map[string]string{
		"key1": "new-value1",
		"key2": "new-value2",
	}
	cache.Store(newData)

	wg.Wait()

	// 读取更新后的缓存
	cachedData := cache.Load().(map[string]string)
	fmt.Println("更新后 key1:", cachedData["key1"])

	// Output:
	// Reader 0: key1=value1
	// Reader 1: key1=value1
	// Reader 2: key1=value1
	// 更新后 key1: new-value1
}

// ExampleOptimizedConcurrency 展示优化的并发模式
func ExampleOptimizedConcurrency() {
	// 根据场景选择合适的类型

	// 场景1: 读多写少 - 使用 RWEntry
	var readHeavy RWEntry
	readHeavy.Store("read-heavy-config")

	// 场景2: 写多读少 - 使用 Value
	var writeHeavy Value
	writeHeavy.Store("write-heavy-config")

	// 场景3: 需要缓存 - 使用 CacheEntry
	var cached CacheEntry
	cached.Store("cached-config")

	fmt.Println("读多写少:", readHeavy.Load())
	fmt.Println("写多读少:", writeHeavy.Load())
	fmt.Println("缓存:", cached.Load())

	// Output:
	// 读多读少: read-heavy-config
	// 写多读少: write-heavy-config
	// 高并发 CAS: 0
	// 缓存: cached-config
}
