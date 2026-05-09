package databinding

// defaultSetProperty is the canonical SetProperty implementation.
// Storage is the same `[]T` as ListProperty; uniqueness is enforced
// by every mutator before delegating to updateCurrentValue.
//
// Element order is not guaranteed. Insertion order is preserved
// where the mutator can do it cheaply (Add appends, Remove keeps the
// surviving order), but RemoveAllWhere / RetainAll iterate the
// current slice so they too keep order. We document "no guarantee"
// to stay free to swap in map-keyed storage later without breaking
// callers.
type defaultSetProperty[T comparable] struct {
	basePropertyState[[]T]
}

// NewSetProperty constructs a SetProperty seeded with the unique
// elements of initial (duplicates dropped, first occurrence wins).
// name defaults to "DefaultSetProperty" when blank.
func NewSetProperty[T comparable](initial []T, name string, validator PropertyValidator[[]T]) SetProperty[T] {
	if name == "" {
		name = "DefaultSetProperty"
	}
	seen := make(map[T]struct{}, len(initial))
	deduped := make([]T, 0, len(initial))
	for _, v := range initial {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		deduped = append(deduped, v)
	}
	p := &defaultSetProperty[T]{
		basePropertyState: newBasePropertyState[[]T](deduped, name, validator),
	}
	p.self = p
	return p
}

func (p *defaultSetProperty[T]) Size() int     { return len(p.Value()) }
func (p *defaultSetProperty[T]) IsEmpty() bool { return len(p.Value()) == 0 }

func (p *defaultSetProperty[T]) Contains(element T) bool {
	for _, v := range p.Value() {
		if v == element {
			return true
		}
	}
	return false
}

func (p *defaultSetProperty[T]) ContainsAll(elements []T) bool {
	current := p.Value()
	index := make(map[T]struct{}, len(current))
	for _, v := range current {
		index[v] = struct{}{}
	}
	for _, e := range elements {
		if _, ok := index[e]; !ok {
			return false
		}
	}
	return true
}

func (p *defaultSetProperty[T]) Add(element T) []T {
	result, _ := p.updateCurrentValue(SetAdd[T]{Element: element}, func(s []T) []T {
		for _, v := range s {
			if v == element {
				return append([]T{}, s...)
			}
		}
		out := make([]T, len(s)+1)
		copy(out, s)
		out[len(s)] = element
		return out
	})
	return result
}

func (p *defaultSetProperty[T]) Remove(element T) []T {
	result, _ := p.updateCurrentValue(SetRemove[T]{Element: element}, func(s []T) []T {
		out := make([]T, 0, len(s))
		for _, v := range s {
			if v != element {
				out = append(out, v)
			}
		}
		return out
	})
	return result
}

func (p *defaultSetProperty[T]) AddAll(elements []T) []T {
	cloned := append([]T{}, elements...)
	result, _ := p.updateCurrentValue(SetAddAll[T]{Elements: cloned}, func(s []T) []T {
		index := make(map[T]struct{}, len(s))
		for _, v := range s {
			index[v] = struct{}{}
		}
		out := append([]T{}, s...)
		for _, e := range elements {
			if _, ok := index[e]; ok {
				continue
			}
			index[e] = struct{}{}
			out = append(out, e)
		}
		return out
	})
	return result
}

func (p *defaultSetProperty[T]) RemoveAll(elements []T) []T {
	cloned := append([]T{}, elements...)
	result, _ := p.updateCurrentValue(SetRemoveAll[T]{Elements: cloned}, func(s []T) []T {
		remove := make(map[T]struct{}, len(elements))
		for _, e := range elements {
			remove[e] = struct{}{}
		}
		out := make([]T, 0, len(s))
		for _, v := range s {
			if _, ok := remove[v]; ok {
				continue
			}
			out = append(out, v)
		}
		return out
	})
	return result
}

func (p *defaultSetProperty[T]) RemoveAllWhere(predicate func(T) bool) []T {
	result, _ := p.updateCurrentValue(SetRemoveAllWhen[T]{Predicate: predicate}, func(s []T) []T {
		out := make([]T, 0, len(s))
		for _, v := range s {
			if !predicate(v) {
				out = append(out, v)
			}
		}
		return out
	})
	return result
}

func (p *defaultSetProperty[T]) RetainAll(elements []T) []T {
	cloned := append([]T{}, elements...)
	result, _ := p.updateCurrentValue(SetRetainAll[T]{Elements: cloned}, func(s []T) []T {
		keep := make(map[T]struct{}, len(elements))
		for _, e := range elements {
			keep[e] = struct{}{}
		}
		out := make([]T, 0, len(s))
		for _, v := range s {
			if _, ok := keep[v]; ok {
				out = append(out, v)
			}
		}
		return out
	})
	return result
}

func (p *defaultSetProperty[T]) Clear() []T {
	result, _ := p.updateCurrentValue(SetClear, func(s []T) []T { return []T{} })
	return result
}

func (p *defaultSetProperty[T]) UpdateFrom(observable ObservableValue[[]T], action BindingAction) Binding[[]T] {
	return UpdateFromConverter[[]T, []T](p, observable, action, func(t []T) []T { return t })
}

func (p *defaultSetProperty[T]) Bind(other Property[[]T], action BindingAction) Binding[[]T] {
	return BindWithConverter[[]T, []T](p, other, action, IdentityConverter[[]T]{})
}

func (p *defaultSetProperty[T]) AsDelegate() PropertyDelegate[[]T] {
	return newDefaultPropertyDelegate[[]T](p)
}
