package atomany

// TypedValue 提供类型安全的原子值
// 使用泛型确保类型安全，避免运行时类型断言
type TypedValue[T any] struct {
	v Value
}

// NewTypedValue 创建一个新的类型安全原子值
//
// 参数：
//   - val: 初始值
//
// 返回：
//   - 类型安全的原子值
func NewTypedValue[T any](val T) *TypedValue[T] {
	tv := &TypedValue[T]{}
	tv.v.Store(val)
	return tv
}

// Load 原子地返回当前值
//
// 返回：
//   - 当前存储的值
func (tv *TypedValue[T]) Load() T {
	return tv.v.Load().(T)
}

// Store 原子地存储新值
//
// 参数：
//   - val: 要存储的新值
func (tv *TypedValue[T]) Store(val T) {
	tv.v.Store(val)
}

// Swap 原子地交换值并返回旧值
//
// 参数：
//   - val: 要存储的新值
//
// 返回：
//   - 之前的旧值
func (tv *TypedValue[T]) Swap(val T) T {
	return tv.v.Swap(val).(T)
}

// CompareAndSwap 执行原子的比较并交换操作
//
// 参数：
//   - old: 期望的当前值
//   - new: 要设置的新值
//
// 返回：
//   - true 如果交换成功，false 否则
func (tv *TypedValue[T]) CompareAndSwap(old, new T) bool {
	return tv.v.CompareAndSwap(old, new)
}

// TryLoad 尝试加载值，如果未初始化则返回 false
//
// 返回：
//   - val: 当前存储的值
//   - ok: 如果已初始化则为 true，否则为 false
func (tv *TypedValue[T]) TryLoad() (val T, ok bool) {
	v, loaded := tv.v.Load(), true
	if v == nil {
		var zero T
		return zero, false
	}
	return v.(T), loaded
}

// TryStore 尝试存储值，如果已存在则返回 false
//
// 参数：
//   - val: 要存储的值
//
// 返回：
//   - true 如果存储成功，false 如果已存在
func (tv *TypedValue[T]) TryStore(val T) (ok bool) {
	return tv.v.CompareAndSwap(nil, val)
}

// GetOrStore 获取当前值，如果未初始化则存储新值并返回
//
// 参数：
//   - val: 如果未初始化则存储的值
//
// 返回：
//   - actual: 实际存储的值
//   - loaded: 如果值已存在则为 true，否则为 false
func (tv *TypedValue[T]) GetOrStore(val T) (actual T, loaded bool) {
	if v := tv.v.Load(); v != nil {
		return v.(T), true
	}
	tv.v.Store(val)
	return val, false
}
