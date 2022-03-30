package store

import (
	"sync/atomic"
	"unsafe"
)

type Any struct {
	data any
}

// ifaceWords is interface{} internal representation.
type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

var changeInProgress = unsafe.Pointer(new(any))

// Load returns the Any set by the most recent Store.
// It returns nil if there has been no call to Store for this Any.
func (v *Any) Load() (val any) {
	for {
		vp := (*ifaceWords)(unsafe.Pointer(v))
		typ := atomic.LoadPointer(&vp.typ)
		if typ == nil {
			return nil
		}
		runtime_procPin()
		if typ == unsafe.Pointer(&changeInProgress) {
			runtime_procUnpin()
			continue
		}
		vlp := (*ifaceWords)(unsafe.Pointer(&val))
		vlp.typ = typ
		vlp.data = atomic.LoadPointer(&vp.data)
		runtime_procUnpin()
		return
	}
}

// Store sets the Any of the Any to x.
// All calls to Store for a given Any must use Anys of the same concrete type.
// Store of an inconsistent type panics, as does Store(nil).
func (v *Any) Store(val any) {
	vp := (*ifaceWords)(unsafe.Pointer(v))
	vlp := (*ifaceWords)(unsafe.Pointer(&val))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == changeInProgress {
			continue
		}
		runtime_procPin()
		if !atomic.CompareAndSwapPointer(&vp.typ, typ, unsafe.Pointer(&changeInProgress)) {
			runtime_procUnpin()
			continue
		}
		atomic.StorePointer(&vp.data, vlp.data)
		atomic.StorePointer(&vp.typ, vlp.typ)
		runtime_procUnpin()
		return
	}
}

// Swap stores new into Any and returns the previous Any. It returns nil if
// the Any is empty.
//
// All calls to Swap for a given Any must use Anys of the same concrete
// type. Swap of an inconsistent type panics, as does Swap(nil).
func (v *Any) Swap(new any) (old any) {
	vp := (*ifaceWords)(unsafe.Pointer(v))
	np := (*ifaceWords)(unsafe.Pointer(&new))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == changeInProgress {
			continue
		}
		// Attempt to start first store.
		// Disable preemption so that other goroutines can use
		// active spin wait to wait for completion; and so that
		// GC does not see the fake type accidentally.
		runtime_procPin()
		if !atomic.CompareAndSwapPointer(&vp.typ, typ, changeInProgress) {
			runtime_procUnpin()
			continue
		}
		// Complete first store.
		op := (*ifaceWords)(unsafe.Pointer(&old))
		op.typ, op.data = typ, atomic.SwapPointer(&vp.data, np.data)
		atomic.StorePointer(&vp.typ, np.typ)
		runtime_procUnpin()
		return old
	}
}

// CompareAndSwap executes the compare-and-swap operation for the Any.
//
// All calls to CompareAndSwap for a given Any must use Anys of the same
// concrete type. CompareAndSwap of an inconsistent type panics, as does
// CompareAndSwap(old, nil).
func (v *Any) CompareAndSwap(old, new any) (swapped bool) {
	vp := (*ifaceWords)(unsafe.Pointer(v))
	np := (*ifaceWords)(unsafe.Pointer(&new))
	for {
		typ := atomic.LoadPointer(&vp.typ)
		if typ == changeInProgress {
			continue
		}
		data := atomic.LoadPointer(&vp.data)
		var i any
		(*ifaceWords)(unsafe.Pointer(&i)).typ = typ
		(*ifaceWords)(unsafe.Pointer(&i)).data = data
		if i != old {
			return false
		}
		runtime_procPin()
		if atomic.CompareAndSwapPointer(&vp.data, data, np.data) {
			atomic.StorePointer(&vp.typ, np.typ)
			runtime_procUnpin()
			return true
		}
		runtime_procUnpin()
	}
}

// Disable/enable preemption, implemented in runtime.
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin()

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
