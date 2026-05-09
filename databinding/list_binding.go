package databinding

import (
	"reflect"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// ListBinding observes an ObservableList[S], maps each element
// through converter, and exposes the result as a Binding[[]T] driven
// by an internal ListProperty[T] target. Source list changes are
// replayed onto the target through the same converter so the target
// stays a one-to-one mapping of the source.
type ListBinding[S, T any] struct {
	baseBinding[[]T]
	source    ObservableList[S]
	converter func(S) T
}

// NewListBinding builds a ListBinding seeded from source.Value().
func NewListBinding[S, T any](source ObservableList[S], converter func(S) T) *ListBinding[S, T] {
	initial := make([]T, 0, source.Size())
	for _, v := range source.Value() {
		initial = append(initial, converter(v))
	}
	target := &defaultListProperty[T]{
		basePropertyState: newBasePropertyState[[]T](initial, "ListBinding", nil),
	}
	target.self = target
	b := &ListBinding[S, T]{
		baseBinding: newBaseBinding[[]T]("ListBinding", target),
		source:      source,
		converter:   converter,
	}
	sub := source.OnChange(func(event ObservableValueChanged[[]S]) {
		runWithDisposeOnFailure(b, func() {
			if traceContainsID(event.Trace(), b.id) {
				return
			}
			oldValue := target.Value()
			result := target.TransformValue(func(s []T) []T {
				return applyListChange[S, T](s, event.ChangeType(), converter)
			})
			if result.IsSuccessful() && !reflect.DeepEqual(oldValue, result.GetValue()) {
				b.publishChange(oldValue, result.GetValue(), event, event.ChangeType(), b)
			}
		})
	})
	b.subscriptions = append(b.subscriptions, sub)
	return b
}

// applyListChange maps a single source-side ListChange onto a
// target slice through converter. Unknown / non-ListChange types
// leave the slice untouched. Every returned slice is a fresh
// allocation. ListRemoveAllWhen is only honoured when S == T because
// the source-side predicate can't be safely reinterpreted otherwise.
func applyListChange[S, T any](current []T, ct ChangeType, converter func(S) T) []T {
	switch c := ct.(type) {
	case ListAdd[S]:
		out := make([]T, len(current)+1)
		copy(out, current)
		out[len(current)] = converter(c.Element)
		return out
	case ListAddAt[S]:
		out := make([]T, 0, len(current)+1)
		out = append(out, current[:c.Index]...)
		out = append(out, converter(c.Element))
		out = append(out, current[c.Index:]...)
		return out
	case ListRemove[S]:
		t := converter(c.Element)
		out := make([]T, 0, len(current))
		removed := false
		for _, v := range current {
			if !removed && reflect.DeepEqual(v, t) {
				removed = true
				continue
			}
			out = append(out, v)
		}
		return out
	case ListRemoveAt:
		out := make([]T, 0, len(current)-1)
		out = append(out, current[:c.Index]...)
		out = append(out, current[c.Index+1:]...)
		return out
	case ListSet[S]:
		out := append([]T{}, current...)
		out[c.Index] = converter(c.Element)
		return out
	case ListAddAll[S]:
		out := make([]T, 0, len(current)+len(c.Elements))
		out = append(out, current...)
		for _, v := range c.Elements {
			out = append(out, converter(v))
		}
		return out
	case ListAddAllAt[S]:
		out := make([]T, 0, len(current)+len(c.C))
		out = append(out, current[:c.Index]...)
		for _, v := range c.C {
			out = append(out, converter(v))
		}
		out = append(out, current[c.Index:]...)
		return out
	case ListRemoveAll[S]:
		mapped := make([]T, 0, len(c.Elements))
		for _, v := range c.Elements {
			mapped = append(mapped, converter(v))
		}
		out := make([]T, 0, len(current))
		for _, v := range current {
			if !containsDeep(mapped, v) {
				out = append(out, v)
			}
		}
		return out
	case ListRemoveAllWhen[S]:
		if pred, ok := any(c.Predicate).(func(T) bool); ok {
			out := make([]T, 0, len(current))
			for _, v := range current {
				if !pred(v) {
					out = append(out, v)
				}
			}
			return out
		}
		return append([]T{}, current...)
	case ListRetainAll[S]:
		mapped := make([]T, 0, len(c.Elements))
		for _, v := range c.Elements {
			mapped = append(mapped, converter(v))
		}
		out := make([]T, 0, len(current))
		for _, v := range current {
			if containsDeep(mapped, v) {
				out = append(out, v)
			}
		}
		return out
	case listClear:
		return []T{}
	case ListPropertyChange:
		return append([]T{}, current...)
	}
	return append([]T{}, current...)
}

// traceContainsID reports whether trace carries an emitter with the
// given id. Shared by ListBinding / SetBinding for their cycle
// guard.
func traceContainsID(trace []events.Event, id core.UUID) bool {
	for _, e := range trace {
		if e.Emitter() != nil && e.Emitter().ID() == id {
			return true
		}
	}
	return false
}
