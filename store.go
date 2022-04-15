package store

import (
	"sync/atomic"
	"unsafe"
)

// A Entry provides an atomic load and store of value.
// The zero value for a Entry returns nil from Load.
type Entry struct {
	p unsafe.Pointer
}

type any = interface{}

func ptr2any(p unsafe.Pointer) any {
	if p == nil {
		return nil
	}
	return *(*any)(p)
}

// Ptr returns e.p value as a unsafe.Ptr.
func (e *Entry) Ptr() (p unsafe.Pointer) {
	return e.p
}

// Load returns the value set by the most recent Store.
func (e *Entry) Load() (val any) {
	return ptr2any(atomic.LoadPointer(&e.p))
}

// Store sets the value of the Value to x.
func (e *Entry) Store(val any) {
	atomic.StorePointer(&e.p, unsafe.Pointer(&val))
}

// Swap stores new into Value and returns the previous value.
// It returns nil if the Value is empty.
func (e *Entry) Swap(new any) (old any) {
	return ptr2any(atomic.SwapPointer(&e.p, unsafe.Pointer(&new)))
}

// CompareAndSwap executes the compare-and-swap operation for the Value.
func (e *Entry) CompareAndSwap(old, new any) (swapped bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if ptr2any(p) != old {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(&new)) {
			return true
		}
	}
}

// //go:linkname runtime_procPin runtime.procPin
// func runtime_procPin()

// //go:linkname runtime_procUnpin runtime.procUnpin
// func runtime_procUnpin()
