package databinding

import "github.com/hexworks/cobalt-go/events"

// ObservableValueChangedKey is the routing key used by every
// ObservableValueChanged event regardless of T. The bus only switches
// on this string; the typed payload is recovered via the generic
// SubscribeTo helper in the events package or via the descriptor
// returned by ObservableValueChangedDescriptor.
const ObservableValueChangedKey = "ObservableValueChanged"

// AnyObservableValueChanged is the type-erased view of every
// ObservableValueChanged[T]. It exists because the change-type
// hierarchy (ListPropertyChange / SetPropertyChange /
// MapPropertyChange) needs to carry an inner change event whose
// value type isn't known to the outer container, and because the
// updateWithEvent path on bound properties consumes events without
// caring about the source's T.
type AnyObservableValueChanged interface {
	events.Event
	// ChangeType returns the mutation kind that produced the event.
	ChangeType() ChangeType
	isObservableValueChanged()
}

// ObservableValueChanged is fired whenever an ObservableValue[T]
// transitions to a new value. The struct is passed by value
// throughout the port — the Kotlin original is a data class with
// structural equality and no mutability.
type ObservableValueChanged[T any] struct {
	OldValue        T
	NewValue        T
	ObservableValue ObservableValue[T]
	// changeType is unexported because we expose it through the
	// ChangeType() method so the type-erased AnyObservableValueChanged
	// interface can also reach it. Field name would otherwise clash
	// with the method.
	changeType ChangeType
	emitter    events.EventSource
	trace      []events.Event
}

// NewObservableValueChanged constructs a change event. trace may be
// nil for the root change in a chain; later listeners prepend
// themselves to it as the change propagates.
func NewObservableValueChanged[T any](
	oldValue, newValue T,
	observable ObservableValue[T],
	changeType ChangeType,
	emitter events.EventSource,
	trace []events.Event,
) ObservableValueChanged[T] {
	return ObservableValueChanged[T]{
		OldValue:        oldValue,
		NewValue:        newValue,
		ObservableValue: observable,
		changeType:      changeType,
		emitter:         emitter,
		trace:           trace,
	}
}

// Key implements events.Event.
func (ObservableValueChanged[T]) Key() string { return ObservableValueChangedKey }

// Emitter implements events.Event.
func (e ObservableValueChanged[T]) Emitter() events.EventSource { return e.emitter }

// Trace implements events.Event.
func (e ObservableValueChanged[T]) Trace() []events.Event { return e.trace }

// ChangeType implements AnyObservableValueChanged.
func (e ObservableValueChanged[T]) ChangeType() ChangeType { return e.changeType }

func (ObservableValueChanged[T]) isObservableValueChanged() {}

// ObservableValueChangedDescriptor returns the typed descriptor used
// with events.SubscribeTo / events.SimpleSubscribeTo. Every T shares
// the same routing key — the descriptor exists so the generic helper
// can assert the inbound Event back to ObservableValueChanged[T].
func ObservableValueChangedDescriptor[T any]() events.EventDescriptor[ObservableValueChanged[T]] {
	return events.EventDescriptor[ObservableValueChanged[T]]{Key: ObservableValueChangedKey}
}
