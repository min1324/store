package atomany

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// Entry 提供了对任意类型值的原子加载和存储操作。
// Entry 的零值在调用 Load 时返回 nil。
// Entry 允许存储任意类型的值，包括 nil，并且可以在运行时动态切换类型。
//
// 使用场景：
//   - 需要存储不同类型的数据
//   - 需要在运行时动态切换数据类型
//   - 需要直接访问底层指针
//
// 注意：Entry 不提供类型安全保证，使用时需要自行确保类型正确性。
type Entry struct {
	p unsafe.Pointer
}

// ptr2any 将指针转换为 any 值
// 该函数通过解引用指针来获取接口值，避免了额外的内存分配。
//
// 参数：
//   - p: 指向接口值的指针
//
// 返回：
//   - 接口值，如果 p 为 nil 则返回 nil
func ptr2any(p unsafe.Pointer) any {
	if p == nil {
		return nil
	}
	// 直接解引用指针获取接口值
	return *(*any)(p)
}

// any2ptr 将 any 转换为指针
// 该函数返回指向接口值的指针，而不是接口值内部的数据指针。
// 这样可以确保 GC 正确跟踪接口值的生命周期。
//
// 参数：
//   - val: 要转换的接口值
//
// 返回：
//   - 指向接口值的指针
func any2ptr(val any) unsafe.Pointer {
	return unsafe.Pointer(&val)
}

// Ptr 原子地返回 Entry 的底层指针
//
// 返回：
//   - Entry 的底层指针，如果 Entry 未初始化则返回 nil
func (e *Entry) Ptr() (p unsafe.Pointer) {
	return atomic.LoadPointer(&e.p)
}

// Load 原子地返回最近一次 Store 设置的值
// 如果 Entry 未被初始化，则返回 nil。
//
// 返回：
//   - 最近一次存储的值，如果未初始化则返回 nil
func (e *Entry) Load() (val any) {
	return ptr2any(atomic.LoadPointer(&e.p))
}

// Store 原子地将值存储到 Entry 中
// 该方法允许存储任意类型的值，包括 nil。
//
// 参数：
//   - val: 要存储的值，可以是任意类型或 nil
//
// 注意：
//   - 该操作是原子的，保证并发安全
//   - 旧值会被保留直到新值存储完成，防止 GC 提前回收
func (e *Entry) Store(val any) {
	old := atomic.SwapPointer(&e.p, any2ptr(val))
	if old != nil {
		// 保持旧值存活直到新指针存储完成，
		// 防止 GC 提前回收旧值。
		_ = *(*any)(old)
	}
}

// Swap 原子地交换 Entry 中的值并返回旧值
// 如果 Entry 未被初始化，则返回 nil。
//
// 参数：
//   - new: 要存储的新值
//
// 返回：
//   - 之前的旧值，如果未初始化则返回 nil
func (e *Entry) Swap(new any) (old any) {
	p := atomic.SwapPointer(&e.p, any2ptr(new))
	return ptr2any(p)
}

// CompareAndSwap 执行原子的比较并交换操作
// 该方法仅在当前值等于 old 时，才将值设置为 new。
//
// 参数：
//   - old: 期望的当前值
//   - new: 要设置的新值
//
// 返回：
//   - true 如果交换成功，false 否则
//
// 注意：
//   - 该操作是原子的，保证并发安全
//   - 如果当前值与 old 不相等，则不会修改 Entry 的值
func (e *Entry) CompareAndSwap(old, new any) (swapped bool) {
	newPtr := any2ptr(new)
	spins := 0
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil {
			if old == nil {
				// 尝试 CAS 操作
				if atomic.CompareAndSwapPointer(&e.p, nil, newPtr) {
					return true
				}
				spins++
				if spins > maxSpins {
					runtime.Gosched()
					spins = 0
				}
				continue
			}
			return false
		}
		// 直接比较指针值，避免创建接口值
		if *(*any)(p) != old {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, newPtr) {
			return true
		}
		spins++
		if spins > maxSpins {
			runtime.Gosched()
			spins = 0
		}
	}
}

// TryLoad 尝试加载值，如果未初始化则返回 false
//
// 返回：
//   - val: 当前存储的值，如果未初始化则为 nil
//   - ok: 如果已初始化则为 true，否则为 false
func (e *Entry) TryLoad() (val any, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil {
		return nil, false
	}
	return ptr2any(p), true
}

// TryStore 尝试存储值，如果已存在则返回 false
//
// 参数：
//   - val: 要存储的值
//
// 返回：
//   - true 如果存储成功，false 如果已存在
func (e *Entry) TryStore(val any) (ok bool) {
	newPtr := any2ptr(val)
	return atomic.CompareAndSwapPointer(&e.p, nil, newPtr)
}

// GetOrStore 获取当前值，如果未初始化则存储新值并返回
//
// 参数：
//   - val: 如果未初始化则存储的值
//
// 返回：
//   - actual: 实际存储的值
//   - loaded: 如果值已存在则为 true，否则为 false
func (e *Entry) GetOrStore(val any) (actual any, loaded bool) {
	p := atomic.LoadPointer(&e.p)
	if p != nil {
		return ptr2any(p), true
	}
	newPtr := any2ptr(val)
	if atomic.CompareAndSwapPointer(&e.p, nil, newPtr) {
		return val, false
	}
	return ptr2any(atomic.LoadPointer(&e.p)), true
}

// LoadOrStore 获取当前值，如果未初始化则存储新值并返回
// 与 GetOrStore 相同，提供更符合 Go 命名习惯的接口
//
// 参数：
//   - val: 如果未初始化则存储的值
//
// 返回：
//   - actual: 实际存储的值
//   - loaded: 如果值已存在则为 true，否则为 false
func (e *Entry) LoadOrStore(val any) (actual any, loaded bool) {
	return e.GetOrStore(val)
}
