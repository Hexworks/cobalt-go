package databinding

// PropertyValidator decides whether a property may transition from
// oldValue to newValue. nil is treated as "always allow" by
// NewProperty / DefaultProperty.
type PropertyValidator[T any] func(oldValue, newValue T) bool

// alwaysValid is the default validator used when callers pass nil.
func alwaysValid[T any](_, _ T) bool { return true }

// Property is a Value that can be read, written and observed.
// Adds bidirectional binding on top of WritableValue + ObservableValue.
//
// The Kotlin original overloaded bind with a converting variant; that
// overload lives as the free function BindWithConverter because Go
// forbids generic methods on interfaces.
type Property[T any] interface {
	WritableValue[T]
	ObservableValue[T]
	// Bind creates a bidirectional binding between the receiver and
	// other. Both sides stay in sync until the returned Binding is
	// disposed.
	Bind(other Property[T], action BindingAction) Binding[T]
	// AsDelegate wraps the receiver in a PropertyDelegate. Kotlin's
	// `var x by property.asDelegate()` syntax has no Go equivalent,
	// so AsDelegate exists primarily to preserve API parity; the
	// returned value behaves like the underlying Property.
	AsDelegate() PropertyDelegate[T]
}

// PropertyDelegate is a Property tagged for use as a Kotlin property
// delegate. Go has no `by` syntax, so this interface is structurally
// identical to Property and acts as a marker rather than adding new
// behaviour.
type PropertyDelegate[T any] interface {
	Property[T]
	isPropertyDelegate()
}

// internalProperty is the privileged interface used by bindings to
// drive a property without going through the validating public
// setters. Unexported because outside packages have no reason to
// reach for the bypass path; every property type defined here
// implements it.
type internalProperty[T any] interface {
	Property[T]
	// updateWithEvent applies newValue using event's trace for cycle
	// detection. Returns true when the value actually changed.
	updateWithEvent(oldValue, newValue T, event AnyObservableValueChanged) bool
	// propertyScope returns the scope this property publishes under.
	propertyScope() PropertyScope
}
