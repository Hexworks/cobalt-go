// Package databinding ports the Kotlin cobalt.databinding package. It
// flattens the upstream api/internal split into a single Go package
// because the original sub-packages reference each other heavily and
// Go forbids the resulting import cycles. Path-internal helpers (the
// singleton event bus, for example) live under databinding/internal/.
package databinding

import "github.com/hexworks/cobalt-go/events"

// Value wraps an arbitrary value of type T. Read-only counterpart of
// WritableValue. Kotlin used `out T`; Go has no variance so callers
// that need a covariant view must wire it themselves.
type Value[T any] interface {
	Value() T
}

// WritableValue is a Value whose contents can be replaced or
// transformed and that can be bound to an upstream ObservableValue.
//
// Kotlin overloaded updateFrom with a converting variant taking an
// (S) -> T function. Go forbids generic methods on interfaces, so the
// converting overload lives as a free function (UpdateFromConverter,
// added in stage 2 alongside the binding implementations).
type WritableValue[T any] interface {
	Value[T]
	// SetValue replaces the current value. Equivalent to writing to
	// the Kotlin `value` setter; emits a change event when the new
	// value differs from the old.
	SetValue(v T)
	// UpdateValue is SetValue but returns the validation outcome
	// instead of throwing.
	UpdateValue(newValue T) ValueValidationResult[T]
	// TransformValue applies fn to the current value, stores the
	// result and returns the validation outcome.
	TransformValue(transformer func(oldValue T) T) ValueValidationResult[T]
	// UpdateFrom starts updating this WritableValue from observable.
	// If action is UpdateOnBind the value is refreshed immediately;
	// otherwise it only updates on observable's next change.
	UpdateFrom(observable ObservableValue[T], action BindingAction) Binding[T]
}

// ObservableValue is a Value that can be observed for changes through
// the Cobalt event bus. Implementors are also EventSources, so each
// carries a stable UUID used to scope event delivery.
type ObservableValue[T any] interface {
	Value[T]
	events.EventSource
	// Name returns the human-readable label used in debug output.
	Name() string
	// OnChange registers fn to receive every change of this value.
	// The returned subscription disposes the listener when disposed.
	// Note: setting the value to its current value does NOT fire fn.
	OnChange(fn func(ObservableValueChanged[T])) events.Subscription
}
