package databinding

// defaultProperty is the canonical scalar Property implementation. It
// embeds basePropertyState[T] for storage and event routing; only the
// Bind / UpdateFrom / AsDelegate methods live on the wrapper itself
// because they reference free functions that take a `Property[T]`.
type defaultProperty[T any] struct {
	basePropertyState[T]
}

// newDefaultProperty is the unexported variant used internally by
// bindings (ComputedBinding, ComputedDualBinding) that need to spin
// up their own target property without forcing a name or validator
// on the caller.
func newDefaultProperty[T any](initial T) *defaultProperty[T] {
	p := &defaultProperty[T]{
		basePropertyState: newBasePropertyState[T](initial, "DefaultProperty", nil),
	}
	p.self = p
	return p
}

// NewProperty builds a Property holding initial. name defaults to
// "DefaultProperty" when blank; validator defaults to "always allow"
// when nil. The Kotlin overloads with default-arg validator and
// optional name collapse into this single constructor.
func NewProperty[T any](initial T, name string, validator PropertyValidator[T]) Property[T] {
	if name == "" {
		name = "DefaultProperty"
	}
	p := &defaultProperty[T]{
		basePropertyState: newBasePropertyState[T](initial, name, validator),
	}
	p.self = p
	return p
}

// UpdateFrom implements WritableValue.
func (p *defaultProperty[T]) UpdateFrom(observable ObservableValue[T], action BindingAction) Binding[T] {
	return UpdateFromConverter[T, T](p, observable, action, func(t T) T { return t })
}

// Bind implements Property.
func (p *defaultProperty[T]) Bind(other Property[T], action BindingAction) Binding[T] {
	return BindWithConverter[T, T](p, other, action, IdentityConverter[T]{})
}

// AsDelegate implements Property.
func (p *defaultProperty[T]) AsDelegate() PropertyDelegate[T] {
	return newDefaultPropertyDelegate[T](p)
}
