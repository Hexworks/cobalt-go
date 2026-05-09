package databinding

import (
	"reflect"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// listBindingDecorator wraps a Binding[[]T] so it satisfies the full
// ObservableListBinding[T] interface — i.e. the list-reading
// convenience methods plus the underlying Binding contract. Used by
// `BindMapList` / collection helpers that need to expose a list-y
// surface on top of a generic ComputedBinding output.
type listBindingDecorator[T any] struct {
	binding Binding[[]T]
	name    string
}

// NewObservableListBinding decorates an existing Binding[[]T] with
// the ObservableList reading methods. name defaults to
// "ListBindingDecorator" when blank.
func NewObservableListBinding[T any](binding Binding[[]T], name string) ObservableListBinding[T] {
	if name == "" {
		name = "ListBindingDecorator"
	}
	return &listBindingDecorator[T]{binding: binding, name: name}
}

func (d *listBindingDecorator[T]) ID() core.UUID    { return d.binding.ID() }
func (d *listBindingDecorator[T]) Name() string     { return d.name }
func (d *listBindingDecorator[T]) Value() []T       { return d.binding.Value() }
func (d *listBindingDecorator[T]) Size() int        { return len(d.Value()) }
func (d *listBindingDecorator[T]) IsEmpty() bool    { return len(d.Value()) == 0 }
func (d *listBindingDecorator[T]) Get(index int) T  { return d.Value()[index] }
func (d *listBindingDecorator[T]) Contains(e T) bool {
	for _, v := range d.Value() {
		if reflect.DeepEqual(v, e) {
			return true
		}
	}
	return false
}

func (d *listBindingDecorator[T]) ContainsAll(es []T) bool {
	for _, e := range es {
		if !d.Contains(e) {
			return false
		}
	}
	return true
}

func (d *listBindingDecorator[T]) IndexOf(e T) int {
	for i, v := range d.Value() {
		if reflect.DeepEqual(v, e) {
			return i
		}
	}
	return -1
}

func (d *listBindingDecorator[T]) LastIndexOf(e T) int {
	cur := d.Value()
	for i := len(cur) - 1; i >= 0; i-- {
		if reflect.DeepEqual(cur[i], e) {
			return i
		}
	}
	return -1
}

func (d *listBindingDecorator[T]) SubList(from, to int) []T {
	src := d.Value()[from:to]
	return append([]T{}, src...)
}

func (d *listBindingDecorator[T]) OnChange(fn func(ObservableValueChanged[[]T])) events.Subscription {
	return d.binding.OnChange(fn)
}

func (d *listBindingDecorator[T]) DisposeState() core.DisposeState { return d.binding.DisposeState() }
func (d *listBindingDecorator[T]) Dispose(s core.DisposeState)     { d.binding.Dispose(s) }
