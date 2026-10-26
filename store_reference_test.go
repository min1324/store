package store_test

import (
	"fmt"
	"store"
)

type iface interface {
	store.Interface
}

func fmtfn(name string, got, want any) string {
	return fmt.Sprintf("%+s wrong value: got %+v, want %+v\n", name, got, want)
}

// newFactor 是一个工厂函数，用于创建和注册各种接口的实现
// 参数 f 是一个回调函数，接受名称(name)和接口实现(iface)作为参数
func newFactor(f func(name string, e iface)) {
	// 使用匿名结构体数组进行迭代，每个元素包含接口实现和对应的名称
	for _, v := range []struct {
		iface
		name string
	}{
		// 注册 store.Entry 实现，名称为 "store"
		{iface: &store.Entry{}, name: "store"},
	} {
		// 调用回调函数，传递名称和接口实现
		f(v.name, v.iface)
	}
}
