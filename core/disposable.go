// Package core ports the foundational types of the Kotlin cobalt.core
// module: UUID, Disposable / DisposeState, Predicate, and the Identity
// function. These primitives have no dependency on the events or
// databinding packages and are safe to import from anywhere.
package core

// Disposable is an object holding resources that can be released by
// calling Dispose.
//
// Note: the Kotlin original exposed `disposeWhen`/`keepWhile` infix
// helpers tied to ObservableValue. Those live in the databinding
// package in this port to keep core free of databinding dependencies.
type Disposable interface {
	// DisposeState returns the current disposal state.
	DisposeState() DisposeState
	// Dispose releases the resources held by this Disposable and
	// records the provided state. Callers that do not care about the
	// reason should pass DisposedManually.
	Dispose(state DisposeState)
}

// Disposed reports whether d has been disposed. Equivalent to
// d.DisposeState().IsDisposed().
func Disposed(d Disposable) bool {
	return d.DisposeState().IsDisposed()
}
