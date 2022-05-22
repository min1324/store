package store

// any is an alias for interface{} and is equivalent to interface{} in all ways.
type any = interface{}

// Interface store interface
type Interface interface {
	// Load returns the Any set by the most recent Store.
	// It returns nil if there has been no call to Store for this Any.
	Load() (val any)

	// Store sets the Any of the Any to x.
	// All calls to Store for a given Any must use Anys of the same concrete type.
	// Store of an inconsistent type panics, as does Store(nil).
	Store(val any)

	// Swap stores new into Any and returns the previous Any. It returns nil if
	// the Any is empty.
	//
	// All calls to Swap for a given Any must use Anys of the same concrete
	// type. Swap of an inconsistent type panics, as does Swap(nil).
	Swap(new any) (old any)

	// CompareAndSwap executes the compare-and-swap operation for the Any.
	//
	// All calls to CompareAndSwap for a given Any must use Anys of the same
	// concrete type. CompareAndSwap of an inconsistent type panics, as does
	// CompareAndSwap(old, nil).
	CompareAndSwap(old, new any) (swapped bool)
}
