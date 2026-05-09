// Package atom provides a minimal mutable single-value container used
// internally by the databinding package.
//
// The Kotlin original used Atom as a thread-safe holder built around
// kotlinx atomics. This port is single-threaded by design, so Atom is
// a plain wrapper that exists to mirror the upstream layering and to
// keep the call sites in BaseProperty readable.
package atom

// Atom holds a single mutable value of type T.
type Atom[T any] struct {
	value T
}

// New returns an Atom initialised with v.
func New[T any](v T) *Atom[T] {
	return &Atom[T]{value: v}
}

// Get returns the current value.
func (a *Atom[T]) Get() T {
	return a.value
}

// Transform applies fn to the current value, stores the result and
// returns it.
func (a *Atom[T]) Transform(fn func(T) T) T {
	a.value = fn(a.value)
	return a.value
}
