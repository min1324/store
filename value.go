package store

import (
	"sync/atomic"
	"unsafe"
)

// A Value provides an atomic load and store of the same type value.
// it can store nil value.
type Value struct {
	data any
}

// ifaceWords is interface{} internal representation.
type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

// wrap nil value in entry
var empty = unsafe.Pointer(new(any))

// Load returns the Any set by the most recent Store.
// It returns nil if there has been no call to Store for this Any.
func (s *Value) Load() (val any) {
	vp := (*ifaceWords)(unsafe.Pointer(s))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil || typ == unsafe.Pointer(&empty) {
			// First store not yet completed.
			return nil
		}
		data := atomic.LoadPointer(&vp.data)
		if data == empty {
			return nil
		}
		// return *(*any)(unsafe.Pointer(&ifaceWords{typ: typ, data: data}))
		vlp := (*ifaceWords)(unsafe.Pointer(&val))
		vlp.typ = typ
		vlp.data = data
		return
	}
}

// Store sets the Any of the Any to x.
// All calls to Store for a given Any must use Anys of the same concrete type.
// Store of an inconsistent type panics, as does Store(nil).
func (s *Value) Store(val any) {
	vp := (*ifaceWords)(unsafe.Pointer(s))
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion.
			if val == nil {
				// not init store nil, return
				return
			}
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&empty)) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, vlp.data)
			atomic.StorePointer(&vp.typ, vlp.typ)
			runtime_procUnpin()
			return
		}
		if typ == empty {
			continue
		}
		if val == nil {
			// wrap nil value
			vlp.typ = atomic.LoadPointer(&vp.typ)
			vlp.data = empty
		}
		// First store completed. Check type and overwrite data.
		if typ != vlp.typ {
			panic("store: store of inconsistently typed value into Value")
		}
		atomic.StorePointer(&vp.data, vlp.data)
		return
	}
}

// Swap stores new into Any and returns the previous Any. It returns nil if
// the Any is empty.
//
// All calls to Swap for a given Any must use Anys of the same concrete
// type. Swap of an inconsistent type panics, as does Swap(nil).
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
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(^uintptr(0))) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return nil
		}
		if uintptr(typ) == ^uintptr(0) {
			continue
		}
		if new == nil {
			// wrap nil value
			np.typ = atomic.LoadPointer(&vp.typ)
			np.data = empty
		}
		// First store completed. Check type and overwrite data.
		if typ != np.typ {
			panic("store: swap of inconsistently typed value into Value")
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

// CompareAndSwap executes the compare-and-swap operation for the Any.
//
// All calls to CompareAndSwap for a given Any must use Anys of the same
// concrete type. CompareAndSwap of an inconsistent type panics, as does
// CompareAndSwap(old, nil).
func (s *Value) CompareAndSwap(old, new any) (swapped bool) {
	vp := (*ifaceWords)(unsafe.Pointer(s))
	op := (*ifaceWords)(unsafe.Pointer(&old))
	np := (*ifaceWords)(unsafe.Pointer(&new))
	if new == nil {
		// wrap nil value
		np.typ = atomic.LoadPointer(&vp.typ)
		np.data = empty
	}
	if op.typ != nil && np.typ != op.typ {
		panic("store: compare and swap of inconsistently typed values")
	}
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			if old != nil {
				return false
			}
			if new == nil {
				// typ == old == new == nil
				return true
			}
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion; and so that
			// GC does not see the fake type accidentally.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(^uintptr(0))) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, np.data)
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return true
		}
		if uintptr(typ) == ^uintptr(0) {
			continue
		}
		// First store completed. Check type and overwrite data.
		if typ != np.typ {
			panic("store: compare and swap of inconsistently typed value into Value")
		}

		data := atomic.LoadPointer(&vp.data)
		var i = *(*any)(unsafe.Pointer(&ifaceWords{typ: typ, data: data}))
		// var i any
		// (*ifaceWords)(unsafe.Pointer(&i)).typ = typ
		// (*ifaceWords)(unsafe.Pointer(&i)).data = data
		if i != old {
			return false
		}
		return atomic.CompareAndSwapPointer(&vp.data, data, np.data)
	}
}

// Disable/enable preemption, implemented in runtime.
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
