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

// any is an alias for interface{} and is equivalent to interface{} in all ways.
type any = interface{}

func ptr2any(p unsafe.Pointer) any {
	if p == nil {
		return nil
	}
	return *(*any)(p)
}

// Ptr returns LoadPointer(&e.p) as a unsafe.Ptr,
// so that can use atomic.CompareAndSwapPointer.
func (e *Entry) Ptr() (p unsafe.Pointer) {
	return atomic.LoadPointer(&e.p)
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
	p := atomic.LoadPointer(&e.p)
	if ptr2any(p) != old {
		return false
	}
	return atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(&new))
}
