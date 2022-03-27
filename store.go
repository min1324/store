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

// Load returns the value set by the most recent Store.
func (e *Entry) Load() (val any) {
	p := atomic.LoadPointer(&e.p)
	if p == nil {
		return nil
	}
	return *(*any)(p)
}

// Store sets the value of the Value to x.
func (e *Entry) Store(val any) {
	atomic.StorePointer(&e.p, unsafe.Pointer(&val))
}

// Swap stores new into Value and returns the previous value.
// It returns nil if the Value is empty.
func (e *Entry) Swap(new any) (old any) {
	p := atomic.SwapPointer(&e.p, unsafe.Pointer(&new))
	if p == nil {
		return nil
	}
	return *(*any)(p)
}

// CompareAndSwap executes the compare-and-swap operation for the Value.
func (e *Entry) CompareAndSwap(old, new any) (swapped bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil {
			if old != nil {
				return false
			}
			// old == p == nil
			if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(&new)) {
				return true
			}
		} else {
			// p != nil
			// runtime_procPin()
			if *(*any)(p) != old {
				// runtime_procUnpin()
				return false
			}
			// p == old
			if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(&new)) {
				// runtime_procUnpin()
				return true
			}
			// runtime_procUnpin()
		}
	}
}

// //go:linkname runtime_procPin runtime.procPin
// func runtime_procPin()

// //go:linkname runtime_procUnpin runtime.procUnpin
// func runtime_procUnpin()
