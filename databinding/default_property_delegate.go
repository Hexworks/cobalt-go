package databinding

// defaultPropertyDelegate wraps a Property so it satisfies the
// PropertyDelegate marker interface. The wrapper forwards every
// method to the underlying property; the additional `isPropertyDelegate`
// method is what makes the Go interface assertion work.
//
// Kotlin's DefaultPropertyDelegate added getValue/setValue with
// KProperty arguments so it could plug into `var x by ...` syntax.
// Go has no equivalent, so the wrapper carries no extra behaviour.
type defaultPropertyDelegate[T any] struct {
	Property[T]
}

func newDefaultPropertyDelegate[T any](p Property[T]) *defaultPropertyDelegate[T] {
	return &defaultPropertyDelegate[T]{Property: p}
}

func (*defaultPropertyDelegate[T]) isPropertyDelegate() {}
