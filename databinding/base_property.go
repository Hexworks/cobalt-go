package databinding

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/databinding/internal/cobalt"
	"github.com/hexworks/cobalt-go/events"
	"github.com/hexworks/cobalt-go/internal/atom"
)

// basePropertyState is the shared storage + event-routing core of
// every Property implementation in the package. defaultProperty and
// the collection variants (defaultListProperty etc.) embed it; the
// outer constructor wires `self` to itself so events published from
// inside this struct carry the outer wrapper as the emitter rather
// than the embedded base. The Kotlin original kept all of this on the
// abstract BaseProperty class — Go composes via embedding instead.
type basePropertyState[T any] struct {
	id        core.UUID
	name      string
	validator PropertyValidator[T]
	backend   *atom.Atom[T]
	scope     PropertyScope
	// self is the ObservableValue exposed in change events. Always
	// equal to the wrapping struct; set immediately after construction.
	self ObservableValue[T]
}

func newBasePropertyState[T any](initial T, name string, validator PropertyValidator[T]) basePropertyState[T] {
	if validator == nil {
		validator = alwaysValid[T]
	}
	id := core.RandomUUID()
	return basePropertyState[T]{
		id:        id,
		name:      name,
		validator: validator,
		backend:   atom.New(initial),
		scope:     PropertyScope{ID: id},
	}
}

func (p *basePropertyState[T]) ID() core.UUID                { return p.id }
func (p *basePropertyState[T]) Name() string                 { return p.name }
func (p *basePropertyState[T]) Value() T                     { return p.backend.Get() }
func (p *basePropertyState[T]) propertyScope() PropertyScope { return p.scope }

func (p *basePropertyState[T]) SetValue(v T) {
	if _, err := p.updateCurrentValue(ScalarChange, func(T) T { return v }); err != nil {
		panic(err)
	}
}

func (p *basePropertyState[T]) UpdateValue(newValue T) ValueValidationResult[T] {
	return p.TransformValue(func(T) T { return newValue })
}

func (p *basePropertyState[T]) TransformValue(transformer func(oldValue T) T) ValueValidationResult[T] {
	result, err := p.updateCurrentValue(ScalarChange, transformer)
	if err != nil {
		return ValueValidationFailed[T]{Value: err.NewValue.(T), Cause: err}
	}
	return ValueValidationSuccessful[T]{Value: result}
}

func (p *basePropertyState[T]) OnChange(fn func(ObservableValueChanged[T])) events.Subscription {
	slog.Debug("subscribing to property changes", "property", p.name, "id", core.Abbreviate(p.id))
	return events.SimpleSubscribeTo(
		cobalt.EventBus(),
		ObservableValueChangedDescriptor[T](),
		p.scope,
		fn,
	)
}

// updateCurrentValue is the write path used by SetValue,
// UpdateValue, TransformValue and by collection mutators that pass a
// non-scalar ChangeType.
func (p *basePropertyState[T]) updateCurrentValue(changeType ChangeType, fn func(T) T) (T, *ValueValidationFailedError) {
	oldValue := p.backend.Get()
	newValue := fn(oldValue)
	if !p.validator(oldValue, newValue) {
		return newValue, NewValueValidationFailedError(newValue, fmt.Sprintf("The given value %v is invalid.", newValue))
	}
	_, scalar := changeType.(scalarChange)
	shouldFire := !scalar || !reflect.DeepEqual(oldValue, newValue)
	p.backend.Transform(func(T) T { return newValue })
	if shouldFire {
		emitter := p.self
		evt := NewObservableValueChanged[T](oldValue, newValue, emitter, changeType, emitter, nil)
		slog.Debug("publishing change event",
			"property", p.name, "oldValue", oldValue, "newValue", newValue)
		cobalt.EventBus().Publish(evt, p.scope)
	}
	return p.backend.Get(), nil
}

// updateWithEvent implements internalProperty. Used by bindings to
// drive this property as their target. Always references self as the
// emitter so downstream listeners see the wrapping property.
func (p *basePropertyState[T]) updateWithEvent(oldValue, newValue T, event AnyObservableValueChanged) bool {
	slog.Debug("updating property via event",
		"property", p.name, "id", core.Abbreviate(p.id), "newValue", newValue)
	for _, e := range event.Trace() {
		if e.Emitter() != nil && e.Emitter().ID() == p.id {
			slog.Warn("circular binding detected; dropping update",
				"property", p.name, "id", core.Abbreviate(p.id))
			return false
		}
	}
	if !p.validator(oldValue, newValue) {
		panic(NewValueValidationFailedError(newValue, fmt.Sprintf("The given value '%v' is invalid.", newValue)))
	}
	if reflect.DeepEqual(oldValue, newValue) {
		return false
	}
	p.backend.Transform(func(T) T { return newValue })
	trace := append([]events.Event{event}, event.Trace()...)
	emitter := p.self
	changed := NewObservableValueChanged[T](
		oldValue, newValue, emitter, event.ChangeType(), emitter, trace,
	)
	cobalt.EventBus().Publish(changed, p.scope)
	return true
}

func (p *basePropertyState[T]) String() string {
	return fmt.Sprintf("%s(id=%s, value=%v)", p.name, core.Abbreviate(p.id), p.backend.Get())
}
