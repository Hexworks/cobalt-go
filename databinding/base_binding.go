package databinding

import (
	"fmt"
	"log/slog"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/databinding/internal/cobalt"
	"github.com/hexworks/cobalt-go/events"
)

// baseBinding holds the state shared by every Binding implementation
// in the package. Concrete unidirectional / bidirectional / computed
// bindings embed it and only write the subscription wiring.
//
// The Kotlin original parameterised BaseBinding over the source type
// S; we don't, because the binding's observable surface only exposes
// the target's type T. S only lives in the subscription closures of
// the concrete types.
type baseBinding[T any] struct {
	id            core.UUID
	name          string
	target        internalProperty[T]
	subscriptions []events.Subscription
	scope         PropertyScope
	disposeState  core.DisposeState
}

// newBaseBinding initialises the embedded state.
func newBaseBinding[T any](name string, target internalProperty[T]) baseBinding[T] {
	id := core.RandomUUID()
	return baseBinding[T]{
		id:           id,
		name:         name,
		target:       target,
		scope:        PropertyScope{ID: id},
		disposeState: core.NotDisposed,
	}
}

func (b *baseBinding[T]) ID() core.UUID                  { return b.id }
func (b *baseBinding[T]) Name() string                   { return b.name }
func (b *baseBinding[T]) DisposeState() core.DisposeState { return b.disposeState }

// Value implements Value[T]. Reading after Dispose panics — matches
// the Kotlin `require(disposed.not())` check.
func (b *baseBinding[T]) Value() T {
	if b.disposeState.IsDisposed() {
		panic("Can't calculate the value of a Binding which is disposed.")
	}
	return b.target.Value()
}

// Dispose implements core.Disposable. Cancels the binding's scope
// (removing any external listeners attached via OnChange) and
// disposes the internal source subscriptions so we stop reacting to
// upstream changes.
func (b *baseBinding[T]) Dispose(state core.DisposeState) {
	b.disposeState = state
	cobalt.EventBus().CancelScope(b.scope)
	disposeSubscriptions(b.subscriptions)
	b.subscriptions = nil
}

// OnChange implements ObservableValue. External listeners subscribe
// under the binding's own scope, isolated from the source and
// target scopes.
func (b *baseBinding[T]) OnChange(fn func(ObservableValueChanged[T])) events.Subscription {
	return events.SimpleSubscribeTo(
		cobalt.EventBus(),
		ObservableValueChangedDescriptor[T](),
		b.scope,
		fn,
	)
}

// String mirrors Kotlin BaseBinding.toString.
func (b *baseBinding[T]) String() string {
	return fmt.Sprintf("%s(id=%s, value=%v)", b.name, core.Abbreviate(b.id), b.target.Value())
}

// publishChange composes the change event, prepending sourceEvent
// onto the trace so cycle detection downstream can spot the loop.
// asSource is the concrete binding (typically the outer embedding
// struct) so the published event's ObservableValue / Emitter fields
// refer to it rather than to the embedded baseBinding. changeType
// is parameterised because ComputedBinding overrides it with
// ScalarChange regardless of the upstream change kind.
func (b *baseBinding[T]) publishChange(
	oldValue, newValue T,
	sourceEvent AnyObservableValueChanged,
	changeType ChangeType,
	asSource ObservableValue[T],
) {
	trace := append([]events.Event{sourceEvent}, sourceEvent.Trace()...)
	evt := NewObservableValueChanged[T](oldValue, newValue, asSource, changeType, asSource, trace)
	slog.Debug("binding publishing change", "binding", b.name, "newValue", newValue)
	cobalt.EventBus().Publish(evt, b.scope)
}
