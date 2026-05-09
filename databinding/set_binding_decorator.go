package databinding

import (
	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// setBindingDecorator wraps a Binding[[]T] so it satisfies the full
// ObservableSetBinding[T] interface.
type setBindingDecorator[T comparable] struct {
	binding Binding[[]T]
	name    string
}

// NewObservableSetBinding decorates an existing Binding[[]T] with
// the ObservableSet reading methods.
func NewObservableSetBinding[T comparable](binding Binding[[]T], name string) ObservableSetBinding[T] {
	if name == "" {
		name = "SetBindingDecorator"
	}
	return &setBindingDecorator[T]{binding: binding, name: name}
}

func (d *setBindingDecorator[T]) ID() core.UUID  { return d.binding.ID() }
func (d *setBindingDecorator[T]) Name() string   { return d.name }
func (d *setBindingDecorator[T]) Value() []T     { return d.binding.Value() }
func (d *setBindingDecorator[T]) Size() int      { return len(d.Value()) }
func (d *setBindingDecorator[T]) IsEmpty() bool  { return len(d.Value()) == 0 }

func (d *setBindingDecorator[T]) Contains(e T) bool {
	for _, v := range d.Value() {
		if v == e {
			return true
		}
	}
	return false
}

func (d *setBindingDecorator[T]) ContainsAll(es []T) bool {
	current := d.Value()
	index := make(map[T]struct{}, len(current))
	for _, v := range current {
		index[v] = struct{}{}
	}
	for _, e := range es {
		if _, ok := index[e]; !ok {
			return false
		}
	}
	return true
}

func (d *setBindingDecorator[T]) OnChange(fn func(ObservableValueChanged[[]T])) events.Subscription {
	return d.binding.OnChange(fn)
}

func (d *setBindingDecorator[T]) DisposeState() core.DisposeState { return d.binding.DisposeState() }
func (d *setBindingDecorator[T]) Dispose(s core.DisposeState)     { d.binding.Dispose(s) }
