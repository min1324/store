package store

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"sync/atomic"
	"unsafe"
)

// A Entry provides an atomic load and store of a consistently typed value.
// The zero value for a Entry returns nil from Load.
// Once Entry has been called, a Entry must not be copied.
//
// A Entry must not be copied after first use.
type Entry struct {
	data any
}

// ifaceWords is interface{} internal representation.
type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

var firstStoreInProgress byte

var (
	nilAny   = new(any)
	nilIface = (*ifaceWords)(unsafe.Pointer(nilAny))
)

// Load returns the value set by the most recent Store.
// It returns nil if there has been no call to Store for this Value.
func (e *Entry) Load() (val any) {
	vp := (*ifaceWords)(unsafe.Pointer(e))
	typ := atomic.LoadPointer(&vp.typ)
	if typ == nil || typ == unsafe.Pointer(&firstStoreInProgress) {
		// First store not yet completed.
		return nil
	}
	data := atomic.LoadPointer(&vp.data)
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	vlp.typ = typ
	vlp.data = data
	if val == nilAny {
		return nil
	}
	return
}

// Store sets the value of the Value to x.
// All calls to Store for a given Value must use values of the same concrete type.
// Store of an inconsistent type panics, as does Store(nil).
func (e *Entry) Store(val any) {
	if val == nil {
		val = nilAny
	}
	vp := (*ifaceWords)(unsafe.Pointer(e))
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			// Attempt to start first store.
			// Disable preemption so that other goroutines can use
			// active spin wait to wait for completion.
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, unsafe.Pointer(&firstStoreInProgress)) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, vlp.data)
			atomic.StorePointer(&vp.typ, vlp.typ)
			runtime_procUnpin()
			return
		}
		if typ == unsafe.Pointer(&firstStoreInProgress) {
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		// First store completed. Check type and overwrite data.
		atomic.StorePointer(&vp.data, vlp.data)
		atomic.StorePointer(&vp.typ, vlp.typ)
		return
	}
}

// Swap stores new into Value and returns the previous value. It returns nil if
// the Value is empty.
//
// All calls to Swap for a given Value must use values of the same concrete
// type. Swap of an inconsistent type panics, as does Swap(nil).
func (e *Entry) Swap(new any) (old any) {
	vp := (*ifaceWords)(unsafe.Pointer(e))
	np := (*ifaceWords)(unsafe.Pointer(&new))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
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
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		// First store completed. Check type and overwrite data.
		// if typ != np.typ {
		// 	panic("sync/atomic: swap of inconsistently typed value into Value")
		// }
		op := (*ifaceWords)(unsafe.Pointer(&old))
		op.typ, op.data = np.typ, atomic.SwapPointer(&vp.data, np.data)
		return old
	}
}

// CompareAndSwap executes the compare-and-swap operation for the Value.
//
// All calls to CompareAndSwap for a given Value must use values of the same
// concrete type. CompareAndSwap of an inconsistent type panics, as does
// CompareAndSwap(old, nil).
func (e *Entry) CompareAndSwap(old, new any) (swapped bool) {
	vp := (*ifaceWords)(unsafe.Pointer(e))
	np := (*ifaceWords)(unsafe.Pointer(&new))
	// op := (*ifaceWords)(unsafe.Pointer(&old))
	// if op.typ != nil && np.typ != op.typ {
	// 	panic("sync/atomic: compare and swap of inconsistently typed values")
	// }
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			if old != nil {
				return false
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
			// First store in progress. Wait.
			// Since we disable preemption around the first store,
			// we can wait with active spinning.
			continue
		}
		// First store completed. Check type and overwrite data.
		// if typ != np.typ {
		// 	panic("sync/atomic: compare and swap of inconsistently typed value into Value")
		// }
		// Compare old and current via runtime equality check.
		// This allows value types to be compared, something
		// not offered by the package functions.
		// CompareAndSwapPointer below only ensures vp.data
		// has not changed since LoadPointer.
		data := atomic.LoadPointer(&vp.data)
		var i any
		(*ifaceWords)(unsafe.Pointer(&i)).typ = typ
		(*ifaceWords)(unsafe.Pointer(&i)).data = data
		if i != old {
			return false
		}
		if atomic.CompareAndSwapPointer(&vp.data, data, np.data) {
			atomic.StorePointer(&vp.typ, np.typ)
			return true
		}
	}
}

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin()

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
