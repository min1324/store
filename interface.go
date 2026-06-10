package atomany

// any 是 interface{} 的类型别名，在所有方面都等同于 interface{}
type any = interface{}

// Interface 定义了原子值操作的通用接口
// Value 和 Entry 类型都实现了此接口
type Interface interface {
	// Load 原子地返回最近一次 Store 设置的值
	// 如果未初始化，则返回 nil
	Load() (val any)

	// Store 原子地将值存储到原子值中
	// Value 类型要求所有存储的值类型一致
	// Entry 类型允许存储任意类型的值
	Store(val any)

	// Swap 原子地交换值并返回旧值
	// 如果未初始化，则返回 nil
	// Value 类型要求类型一致
	Swap(new any) (old any)

	// CompareAndSwap 执行原子的比较并交换操作
	// 仅在当前值等于 old 时，才将值设置为 new
	// Value 类型要求类型一致
	CompareAndSwap(old, new any) (swapped bool)
}
