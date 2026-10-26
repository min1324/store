package store_test

// import (
// 	"fmt"
// 	"math/rand"
// 	"runtime"
// 	"sync"
// 	"testing"
// 	"time"
// )

// // TestEntry_BasicOperations 测试 Entry 的基本操作
// func TestEntry_BasicOperations(t *testing.T) {
// 	var e Entry

// 	// 测试初始状态
// 	val, ok := e.TryLoad()
// 	if val != nil || ok {
// 		t.Error("初始状态应该为未初始化")
// 	}

// 	// 测试 Store 和 Load
// 	e.Store("test")
// 	if val := e.Load(); val != "test" {
// 		t.Errorf("期望 'test'，得到 %v", val)
// 	}

// 	// 测试 Swap
// 	old := e.Swap("new")
// 	if old != "test" {
// 		t.Errorf("期望 'test'，得到 %v", old)
// 	}
// 	if val := e.Load(); val != "new" {
// 		t.Errorf("期望 'new'，得到 %v", val)
// 	}

// 	// 测试 CompareAndSwap
// 	if !e.CompareAndSwap("new", "updated") {
// 		t.Error("CompareAndSwap 失败")
// 	}
// 	if e.CompareAndSwap("wrong", "should not happen") {
// 		t.Error("CompareAndSwap 不应该成功")
// 	}
// 	if val := e.Load(); val != "updated" {
// 		t.Errorf("期望 'updated'，得到 %v", val)
// 	}

// 	// 测试 TryStore
// 	if e.TryStore("should fail") {
// 		t.Error("TryStore 不应该成功")
// 	}
// 	if val := e.Load(); val != "updated" {
// 		t.Errorf("期望 'updated'，得到 %v", val)
// 	}

// 	// 测试 TryStore 在空 Entry 上
// 	var empty Entry
// 	if !empty.TryStore("first") {
// 		t.Error("TryStore 应该成功")
// 	}
// 	if empty.TryStore("second") {
// 		t.Error("TryStore 不应该成功")
// 	}
// }

// // TestEntry_NilValues 测试处理 nil 值
// func TestEntry_NilValues(t *testing.T) {
// 	var e Entry

// 	// 存储 nil
// 	e.Store(nil)
// 	if val := e.Load(); val != nil {
// 		t.Errorf("期望 nil，得到 %v", val)
// 	}

// 	// Swap nil
// 	old := e.Swap("not nil")
// 	if old != nil {
// 		t.Errorf("期望 nil，得到 %v", old)
// 	}
// 	e.Store(nil)

// 	// CompareAndSwap nil
// 	if !e.CompareAndSwap(nil, "not nil") {
// 		t.Error("CompareAndSwap(nil, not nil) 应该成功")
// 	}
// 	if e.CompareAndSwap(nil, "should not happen") {
// 		t.Error("CompareAndSwap(nil, should not happen) 不应该成功")
// 	}

// 	// TryStore nil
// 	var empty Entry
// 	if !empty.TryStore(nil) {
// 		t.Error("TryStore(nil) 应该成功")
// 	}
// 	if empty.TryStore(nil) {
// 		t.Error("TryStore(nil) 不应该成功")
// 	}
// }

// // TestEntry_TypeSwitching 测试类型切换
// func TestEntry_TypeSwitching(t *testing.T) {
// 	var e Entry

// 	// 存储不同类型的值
// 	e.Store("string")
// 	if val := e.Load(); val != "string" {
// 		t.Errorf("期望 'string'，得到 %v", val)
// 	}

// 	e.Store(42)
// 	if val := e.Load(); val != 42 {
// 		t.Errorf("期望 42，得到 %v", val)
// 	}

// 	e.Store(true)
// 	if val := e.Load(); val != true {
// 		t.Errorf("期望 true，得到 %v", val)
// 	}

// 	e.Store(nil)
// 	if val := e.Load(); val != nil {
// 		t.Errorf("期望 nil，得到 %v", val)
// 	}
// }

// // TestEntry_LoadOrStore 测试 LoadOrStore
// func TestEntry_LoadOrStore(t *testing.T) {
// 	var e Entry

// 	// 第一次调用应该存储并返回 false
// 	actual, loaded := e.LoadOrStore("first")
// 	if actual != "first" || loaded {
// 		t.Errorf("期望 ('first', false)，得到 (%v, %v)", actual, loaded)
// 	}

// 	// 第二次调用应该加载并返回 true
// 	actual, loaded = e.LoadOrStore("second")
// 	if actual != "first" || !loaded {
// 		t.Errorf("期望 ('first', true)，得到 (%v, %v)", actual, loaded)
// 	}
// }

// func TestEntry_Operations_Combined(t *testing.T) {
// 	var e Entry
// 	testValue := "test"
// 	newValue := "new"
// 	updatedValue := "updated"

// 	// 1. 测试初始状态
// 	val, ok := e.TryLoad()
// 	if val != nil || ok {
// 		t.Error("初始状态应该为未初始化")
// 	}

// 	// 2. 测试 LoadOrStore - 第一次调用应该存储并返回 false
// 	actual, loaded := e.LoadOrStore(testValue)
// 	if actual != testValue || loaded {
// 		t.Errorf("LoadOrStore 第一次调用: 期望 (%v, false)，得到 (%v, %v)", testValue, actual, loaded)
// 	}

// 	// 3. 测试 LoadOrStore - 第二次调用应该加载并返回 true
// 	actual, loaded = e.LoadOrStore(newValue)
// 	if actual != testValue || !loaded {
// 		t.Errorf("LoadOrStore 第二次调用: 期望 (%v, true)，得到 (%v, %v)", testValue, actual, loaded)
// 	}

// 	// 4. 测试 Load - 应该返回最近存储的值
// 	val = e.Load()
// 	if val != testValue {
// 		t.Errorf("Load: 期望 %v，得到 %v", testValue, val)
// 	}

// 	// 5. 测试 Store - 存储新值
// 	e.Store(newValue)
// 	val = e.Load()
// 	if val != newValue {
// 		t.Errorf("Store: 期望 %v，得到 %v", newValue, val)
// 	}

// 	// 6. 测试 Swap - 交换值并返回旧值
// 	old := e.Swap(updatedValue)
// 	if old != newValue {
// 		t.Errorf("Swap: 期望 %v，得到 %v", newValue, old)
// 	}
// 	val = e.Load()
// 	if val != updatedValue {
// 		t.Errorf("Swap后Load: 期望 %v，得到 %v", updatedValue, val)
// 	}

// 	// 7. 测试 CompareAndSwap - 成功的情况
// 	swapped := e.CompareAndSwap(updatedValue, testValue)
// 	if !swapped {
// 		t.Error("CompareAndSwap 应该成功")
// 	}
// 	val = e.Load()
// 	if val != testValue {
// 		t.Errorf("CompareAndSwap后Load: 期望 %v，得到 %v", testValue, val)
// 	}

// 	// 8. 测试 CompareAndSwap - 失败的情况
// 	swapped = e.CompareAndSwap(newValue, updatedValue)
// 	if swapped {
// 		t.Error("CompareAndSwap 应该失败")
// 	}
// 	val = e.Load()
// 	if val != testValue {
// 		t.Errorf("CompareAndSwap失败后Load: 期望 %v，得到 %v", testValue, val)
// 	}

// 	// 9. 测试 TryStore - 在已初始化的 Entry 上应该失败
// 	if e.TryStore(newValue) {
// 		t.Error("TryStore 应该失败")
// 	}
// 	val = e.Load()
// 	if val != testValue {
// 		t.Errorf("TryStore失败后Load: 期望 %v，得到 %v", testValue, val)
// 	}

// 	// 10. 测试 TryStore - 在新的 Entry 上应该成功
// 	var newEntry Entry
// 	if !newEntry.TryStore(testValue) {
// 		t.Error("TryStore 应该成功")
// 	}
// 	val = newEntry.Load()
// 	if val != testValue {
// 		t.Errorf("TryStore成功后Load: 期望 %v，得到 %v", testValue, val)
// 	}

// 	// 11. 测试操作序列：Store -> Load -> Swap -> Load -> CompareAndSwap -> Load
// 	e.Store("sequence")
// 	val = e.Load()
// 	if val != "sequence" {
// 		t.Errorf("Store->Load: 期望 'sequence'，得到 %v", val)
// 	}

// 	old = e.Swap("swap")
// 	if old != "sequence" {
// 		t.Errorf("Swap: 期望 'sequence'，得到 %v", old)
// 	}
// 	val = e.Load()
// 	if val != "swap" {
// 		t.Errorf("Swap后Load: 期望 'swap'，得到 %v", val)
// 	}

// 	swapped = e.CompareAndSwap("swap", "cas")
// 	if !swapped {
// 		t.Error("CompareAndSwap 应该成功")
// 	}
// 	val = e.Load()
// 	if val != "cas" {
// 		t.Errorf("CompareAndSwap后Load: 期望 'cas'，得到 %v", val)
// 	}

// 	// 12. 测试 nil 值的处理
// 	e.Store(nil)
// 	val = e.Load()
// 	if val != nil {
// 		t.Errorf("Store nil 后 Load: 期望 nil，得到 %v", val)
// 	}

// 	old = e.Swap("not nil")
// 	if old != nil {
// 		t.Errorf("Swap nil: 期望 nil，得到 %v", old)
// 	}
// 	val = e.Load()
// 	if val != "not nil" {
// 		t.Errorf("Swap nil 后 Load: 期望 'not nil'，得到 %v", val)
// 	}

// 	swapped = e.CompareAndSwap("not nil", nil)
// 	if !swapped {
// 		t.Error("CompareAndSwap(not nil, nil) 应该成功")
// 	}
// 	val = e.Load()
// 	if val != nil {
// 		t.Errorf("CompareAndSwap(not nil, nil) 后 Load: 期望 nil，得到 %v", val)
// 	}
// }

// func TestEntry_Operations_Concurrent_WithTimeout(t *testing.T) {
// 	const (
// 		goroutines = 10
// 		iterations = 1000
// 		timeout    = 10 * time.Second
// 	)

// 	var e Entry
// 	var wg sync.WaitGroup
// 	errors := make(chan error, 100)
// 	done := make(chan struct{})

// 	// 设置超时
// 	timeoutChan := make(chan bool, 1)
// 	go func() {
// 		time.Sleep(timeout)
// 		timeoutChan <- true
// 	}()

// 	// 启动多个 goroutine 并发执行各种操作
// 	for i := 0; i < goroutines; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			for j := 0; j < iterations; j++ {
// 				select {
// 				case <-done:
// 					return
// 				default:
// 					// 随机选择一个操作执行
// 					switch rand.Intn(5) {
// 					case 0: // LoadOrStore
// 						val, loaded := e.LoadOrStore(id)
// 						if !loaded && val != id {
// 							errors <- fmt.Errorf("LoadOrStore: 期望 %d，得到 %v", id, val)
// 						}
// 					case 1: // Load
// 						val := e.Load()
// 						_ = val
// 					case 2: // Store
// 						e.Store(id)
// 					case 3: // Swap
// 						old := e.Swap(id + 1000)
// 						_ = old
// 					case 4: // CompareAndSwap
// 						swapped := e.CompareAndSwap(id, id+2000)
// 						if swapped {
// 							val := e.Load()
// 							if val != id+2000 {
// 								errors <- fmt.Errorf("CompareAndSwap后Load: 期望 %d，得到 %v", id+2000, val)
// 							}
// 						}
// 					}
// 					runtime.Gosched()
// 				}
// 			}
// 		}(i)
// 	}

// 	// 等待完成或超时
// 	select {
// 	case <-timeoutChan:
// 		close(done)
// 		t.Errorf("测试超时，超过 %v", timeout)
// 	case <-time.After(2 * time.Second):
// 		// 等待一小段时间让 goroutine 完成当前操作
// 	}

// 	wg.Wait()
// 	close(errors)

// 	// 报告所有错误
// 	for err := range errors {
// 		t.Error(err)
// 	}
// }
