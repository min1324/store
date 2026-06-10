package atomany

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// Value 提供了对相同类型值的原子加载和存储操作。
// Value 是 sync/atomic.Value 的增强版本，支持存储 nil 值。
//
// 特性：
//   - 原子操作：所有操作都是原子的，保证并发安全
//   - 类型安全：要求所有存储的值类型一致
//   - nil 支持：初始化后允许存储和读取 nil 值
//
// 使用场景：
//   - 需要类型安全保证的原子操作
//   - 需要存储 nil 值
//   - 性能敏感的场景
//
// 注意：首次 Store 不能为 nil，如需从 nil 开始初始化，请使用 CompareAndSwap(nil, x)。
type Value struct {
	data any
}

// ifaceWords 表示 interface{} 的内部结构
// 包含类型信息和数据指针
type ifaceWords struct {
	typ  unsafe.Pointer // 类型指针
	data unsafe.Pointer // 数据指针
}

// maxSpins 定义最大自旋次数
const maxSpins = 10

// empty 用于包装 nil 值
// 这是一个特殊的指针，用于表示 nil 值
var empty = unsafe.Pointer(new(any))

// Load 原子地返回最近一次 Store 设置的值
// 如果 Value 未被初始化，则返回 nil。
//
// 返回：
//   - 最近一次存储的值，如果未初始化则返回 nil
//
// 注意：
//   - 该操作是原子的，保证并发安全
//   - 读取过程中会验证类型一致性，防止撕裂读取
func (v *Value) Load() (val any) {
	vp := (*ifaceWords)(unsafe.Pointer(v))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil || typ == unsafe.Pointer(&empty) {
			// 首次存储尚未完成
			return nil
		}
		data := atomic.LoadPointer(&vp.data)
		if data == empty {
			return nil
		}
		// 验证类型在读取数据期间未发生变化
		// 这防止了 typ 和 data 来自不同存储的撕裂读取
		if atomic.LoadPointer(&vp.typ) != typ {
			continue
		}
		vlp := (*ifaceWords)(unsafe.Pointer(&val))
		vlp.typ = typ
		vlp.data = data
		return
	}
}

// Store 原子地将值存储到 Value 中
//
// 参数：
//   - val: 要存储的值，首次存储不能为 nil
//
// 注意：
//   - 首次 Store 必须是非 nil 值，以建立类型
//   - 所有后续 Store 必须使用相同的具体类型
//   - 类型不一致会 panic
//   - 初始化后可以存储 nil 值
func (s *Value) Store(val any) {
	vp := (*ifaceWords)(unsafe.Pointer(s))
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			// 尝试开始首次存储
			// 禁用抢占，以便其他 goroutine 可以使用主动自旋等待完成
			if val == nil {
				// 首次存储不能为 nil，请使用 CompareAndSwap(nil, x) 代替
				// 必须先用非 nil 值初始化以建立类型
				panic("atomany: store of nil value into uninitialized Value")
			}
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&empty)) {
				runtime_procUnpin()
				continue
			}
			// 完成首次存储
			atomic.StorePointer(&vp.data, vlp.data)
			atomic.StorePointer(&vp.typ, vlp.typ)
			runtime_procUnpin()
			return
		}
		if typ == unsafe.Pointer(&empty) {
			continue
		}
		if val == nil {
			// 包装 nil 值
			vlp.typ = atomic.LoadPointer(&vp.typ)
			vlp.data = empty
		}
		// 首次存储已完成。检查类型并覆盖数据
		if typ != vlp.typ {
			panic("atomany: store of inconsistently typed value into Value")
		}
		atomic.StorePointer(&vp.data, vlp.data)
		return
	}
}

// Swap 原子地交换 Value 中的值并返回旧值
// 如果 Value 未被初始化，则返回 nil。
//
// 参数：
//   - new: 要存储的新值
//
// 返回：
//   - 之前的旧值，如果未初始化则返回 nil
//
// 注意：
//   - 所有 Swap 调用必须使用相同的具体类型
//   - 类型不一致会 panic
//   - 未初始化时不能 Swap(nil)
func (s *Value) Swap(new any) (old any) {
	vp := (*ifaceWords)(unsafe.Pointer(s))
	np := (*ifaceWords)(unsafe.Pointer(&new))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			if new == nil {
				return nil
			}
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion; and so that
			// GC does not see the fake type accidentally.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&empty)) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return nil
		}
		if typ == unsafe.Pointer(&empty) {
			continue
		}
		if new == nil {
			// wrap nil value
			np.typ = atomic.LoadPointer(&vp.typ)
			np.data = empty
		}
		// First store completed. Check type and overwrite data.
		if typ != np.typ {
			panic("atomany: swap of inconsistently typed value into Value")
		}
		// Complete first store.
		data := atomic.SwapPointer(&vp.data, np.data)
		if data == empty {
			return nil
		}
		op := (*ifaceWords)(unsafe.Pointer(&old))
		op.typ, op.data = typ, data
		return old
	}
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
//   - 所有 CompareAndSwap 调用必须使用相同的具体类型
//   - 类型不一致会 panic
//   - 未初始化时可以使用 CompareAndSwap(nil, x) 进行首次赋值
func (v *Value) CompareAndSwap(old, new any) (swapped bool) {
	vp := (*ifaceWords)(unsafe.Pointer(&v))
	op := (*ifaceWords)(unsafe.Pointer(&old))
	np := (*ifaceWords)(unsafe.Pointer(&new))

	if new == nil {
		np.typ = atomic.LoadPointer(&vp.typ)
		np.data = empty
	}

	if op.typ != nil && np.typ != op.typ {
		panic("atomany: compare and swap of inconsistently typed values")
	}

	spins := 0

	for {
		typ := atomic.LoadPointer(&vp.typ)

		// 首次存储。检查类型并覆盖数据
		if typ == nil {
			if old != nil {
				return false
			}
			if new == nil {
				return true
			}

			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&empty)) {
				runtime_procUnpin()
				spins++
				if spins > maxSpins {
					runtime.Gosched()
					spins = 0
				}
				continue
			}

			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return true
		}

		// 首次存储尚未完成
		if typ == unsafe.Pointer(&empty) {
			spins++
			if spins > maxSpins {
				runtime.Gosched()
				spins = 0
			}
			continue
		}

		data := atomic.LoadPointer(&vp.data)
		if data == empty {
			if old != nil {
				return false
			}
		}

		// 检查类型是否匹配
		if atomic.LoadPointer(&vp.typ) != typ {
			spins++
			if spins > maxSpins {
				runtime.Gosched()
				spins = 0
			}
			continue
		}

		// 检查值是否匹配
		if old != nil {
			if data == empty {
				return false
			}
			if *(*any)(unsafe.Pointer(&ifaceWords{typ, data})) != old {
				return false
			}
		}

		if typ != np.typ {
			panic("atomany: compare and swap of inconsistently typed value into Value")
		}

		// 比较并交换数据
		if atomic.CompareAndSwapPointer(&vp.data, data, np.data) {
			return true
		}

		spins++
		if spins > maxSpins {
			runtime.Gosched()
			spins = 0
		}
	}
}

// Disable/enable preemption, implemented in runtime.
//
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
