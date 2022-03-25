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

var storeInProgress = unsafe.Pointer(new(any))

// Load returns the value set by the most recent Store.
// It returns nil if there has been no call to Store for this Value.
func (e *Entry) Load() (val any) {
	vp := (*ifaceWords)(unsafe.Pointer(e))
	typ := atomic.LoadPointer(&vp.typ)
	if typ == nil || typ == storeInProgress {
		// First store not yet completed.
		return nil
	}
	data := atomic.LoadPointer(&vp.data)
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	vlp.typ = typ
	vlp.data = data

	return
}

// Store sets the value of the Value to x.
// All calls to Store for a given Value must use values of the same concrete type.
// Store of an inconsistent type panics, as does Store(nil).
func (e *Entry) Store(val any) {
	vp := (*ifaceWords)(unsafe.Pointer(e))
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			runtime_procPin()
			if !atomic.CompareAndSwapPointer(&vp.typ, nil, storeInProgress) {
				runtime_procUnpin()
				continue
			}
			// Complete first store.
			atomic.StorePointer(&vp.data, vlp.data)
			atomic.StorePointer(&vp.typ, vlp.typ)
			runtime_procUnpin()
			return
		}
		if typ == storeInProgress {
			continue
		}
		runtime_procPin()
		atomic.StorePointer(&vp.data, vlp.data)
		atomic.StorePointer(&vp.typ, vlp.typ)
		runtime_procUnpin()
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
		if typ == storeInProgress {
			continue
		}
		runtime_procPin()
		if !atomic.CompareAndSwapPointer(&vp.typ, typ, storeInProgress) {
			runtime_procUnpin()
			continue
		}
		op := (*ifaceWords)(unsafe.Pointer(&old))
		atomic.StorePointer(&op.data, atomic.SwapPointer(&vp.data, np.data))
		atomic.StorePointer(&op.typ, typ)
		atomic.StorePointer(&vp.typ, np.typ)
		runtime_procUnpin()
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
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == storeInProgress {
			continue
		}
		runtime_procPin()
		if !atomic.CompareAndSwapPointer(&vp.typ, typ, storeInProgress) {
			runtime_procUnpin()
			continue
		}
		data := atomic.LoadPointer(&vp.data)
		var i any
		(*ifaceWords)(unsafe.Pointer(&i)).typ = typ
		(*ifaceWords)(unsafe.Pointer(&i)).data = data
		if i != old {
			// compare(init,old) not equal, restore typ
			// atomic.CompareAndSwapPointer(&vp.typ, storeInProgress, typ)
			atomic.StorePointer(&vp.typ, typ)
			runtime_procUnpin()
			return false
		}
		if atomic.CompareAndSwapPointer(&vp.data, data, np.data) {
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return true
		}
		runtime_procUnpin()
	}
}

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin()

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
