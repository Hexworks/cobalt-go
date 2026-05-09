package databinding

import "reflect"

// SetBinding observes an ObservableSet[S], maps each element
// through converter, and exposes the result as a Binding[[]T] driven
// by an internal SetProperty[T] target. Element-level SetChanges on
// the source are replayed through the converter so the target stays
// a one-to-one mapping.
type SetBinding[S, T comparable] struct {
	baseBinding[[]T]
	source    ObservableSet[S]
	converter func(S) T
}

// NewSetBinding builds a SetBinding seeded from source.Value().
func NewSetBinding[S, T comparable](source ObservableSet[S], converter func(S) T) *SetBinding[S, T] {
	seen := make(map[T]struct{}, source.Size())
	initial := make([]T, 0, source.Size())
	for _, v := range source.Value() {
		mapped := converter(v)
		if _, ok := seen[mapped]; ok {
			continue
		}
		seen[mapped] = struct{}{}
		initial = append(initial, mapped)
	}
	target := &defaultSetProperty[T]{
		basePropertyState: newBasePropertyState[[]T](initial, "SetBinding", nil),
	}
	target.self = target
	b := &SetBinding[S, T]{
		baseBinding: newBaseBinding[[]T]("SetBinding", target),
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
				return applySetChange[S, T](s, event.ChangeType(), converter)
			})
			if result.IsSuccessful() && !reflect.DeepEqual(oldValue, result.GetValue()) {
				b.publishChange(oldValue, result.GetValue(), event, event.ChangeType(), b)
			}
		})
	})
	b.subscriptions = append(b.subscriptions, sub)
	return b
}

func applySetChange[S, T comparable](current []T, ct ChangeType, converter func(S) T) []T {
	switch c := ct.(type) {
	case SetAdd[S]:
		t := converter(c.Element)
		for _, v := range current {
			if v == t {
				return append([]T{}, current...)
			}
		}
		out := make([]T, len(current)+1)
		copy(out, current)
		out[len(current)] = t
		return out
	case SetRemove[S]:
		t := converter(c.Element)
		out := make([]T, 0, len(current))
		for _, v := range current {
			if v != t {
				out = append(out, v)
			}
		}
		return out
	case SetAddAll[S]:
		index := make(map[T]struct{}, len(current))
		for _, v := range current {
			index[v] = struct{}{}
		}
		out := append([]T{}, current...)
		for _, v := range c.Elements {
			t := converter(v)
			if _, ok := index[t]; ok {
				continue
			}
			index[t] = struct{}{}
			out = append(out, t)
		}
		return out
	case SetRemoveAll[S]:
		remove := make(map[T]struct{}, len(c.Elements))
		for _, v := range c.Elements {
			remove[converter(v)] = struct{}{}
		}
		out := make([]T, 0, len(current))
		for _, v := range current {
			if _, ok := remove[v]; !ok {
				out = append(out, v)
			}
		}
		return out
	case SetRemoveAllWhen[S]:
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
	case SetRetainAll[S]:
		keep := make(map[T]struct{}, len(c.Elements))
		for _, v := range c.Elements {
			keep[converter(v)] = struct{}{}
		}
		out := make([]T, 0, len(current))
		for _, v := range current {
			if _, ok := keep[v]; ok {
				out = append(out, v)
			}
		}
		return out
	case setClear:
		return []T{}
	case SetPropertyChange:
		return append([]T{}, current...)
	}
	return append([]T{}, current...)
}
