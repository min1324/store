package store_test

import (
	"fmt"
	"store"
	"unsafe"
)

type any = interface{}

type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

type iface interface {
	store.Interface
}

func fmtfn(name string, got, want any) string {
	return fmt.Sprintf("%+s wrong value: got %+v, want %+v\n", name, got, want)
}

func newFactor(f func(name string, e iface)) {
	for _, v := range []struct {
		iface
		name string
	}{
		{iface: &store.Entry{}, name: "store"},
	} {
		f(v.name, v.iface)
	}
}
