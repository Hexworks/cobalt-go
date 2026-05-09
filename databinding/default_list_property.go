package databinding

import "reflect"

// defaultListProperty is the canonical ListProperty implementation.
// Embeds basePropertyState[[]T] for storage / event routing; adds
// the list-specific reads and mutators on top. Mutators go through
// updateCurrentValue with the matching ListChange variant so
// downstream ListBindings can replay element-level changes.
type defaultListProperty[T any] struct {
	basePropertyState[[]T]
}

// NewListProperty constructs a ListProperty seeded with initial.
// initial is defensively copied so later mutation of the caller's
// slice does not affect the property's state. name defaults to
// "DefaultListProperty" when blank; validator defaults to "always
// allow" when nil.
func NewListProperty[T any](initial []T, name string, validator PropertyValidator[[]T]) ListProperty[T] {
	if name == "" {
		name = "DefaultListProperty"
	}
	cloned := append([]T{}, initial...)
	p := &defaultListProperty[T]{
		basePropertyState: newBasePropertyState[[]T](cloned, name, validator),
	}
	p.self = p
	return p
}

// Size implements ObservableList.
func (p *defaultListProperty[T]) Size() int { return len(p.Value()) }

// IsEmpty implements ObservableList.
func (p *defaultListProperty[T]) IsEmpty() bool { return len(p.Value()) == 0 }

// Contains implements ObservableList. Uses reflect.DeepEqual because
// T is unconstrained.
func (p *defaultListProperty[T]) Contains(element T) bool {
	for _, v := range p.Value() {
		if reflect.DeepEqual(v, element) {
			return true
		}
	}
	return false
}

// ContainsAll implements ObservableList. O(n*m).
func (p *defaultListProperty[T]) ContainsAll(elements []T) bool {
	for _, e := range elements {
		if !p.Contains(e) {
			return false
		}
	}
	return true
}

// Get implements ObservableList. Panics on out-of-range index, same
// shape as Go slice indexing.
func (p *defaultListProperty[T]) Get(index int) T { return p.Value()[index] }

// IndexOf implements ObservableList.
func (p *defaultListProperty[T]) IndexOf(element T) int {
	for i, v := range p.Value() {
		if reflect.DeepEqual(v, element) {
			return i
		}
	}
	return -1
}

// LastIndexOf implements ObservableList.
func (p *defaultListProperty[T]) LastIndexOf(element T) int {
	val := p.Value()
	for i := len(val) - 1; i >= 0; i-- {
		if reflect.DeepEqual(val[i], element) {
			return i
		}
	}
	return -1
}

// SubList implements ObservableList. Returns a fresh slice so the
// caller cannot mutate the property's backing array.
func (p *defaultListProperty[T]) SubList(fromIndex, toIndex int) []T {
	src := p.Value()[fromIndex:toIndex]
	return append([]T{}, src...)
}

// --- mutators -------------------------------------------------------------

func (p *defaultListProperty[T]) Add(element T) []T {
	result, _ := p.updateCurrentValue(ListAdd[T]{Element: element}, func(s []T) []T {
		out := make([]T, len(s)+1)
		copy(out, s)
		out[len(s)] = element
		return out
	})
	return result
}

func (p *defaultListProperty[T]) AddAt(index int, element T) []T {
	result, _ := p.updateCurrentValue(ListAddAt[T]{Index: index, Element: element}, func(s []T) []T {
		out := make([]T, 0, len(s)+1)
		out = append(out, s[:index]...)
		out = append(out, element)
		out = append(out, s[index:]...)
		return out
	})
	return result
}

func (p *defaultListProperty[T]) Set(index int, element T) []T {
	result, _ := p.updateCurrentValue(ListSet[T]{Index: index, Element: element}, func(s []T) []T {
		out := append([]T{}, s...)
		out[index] = element
		return out
	})
	return result
}

func (p *defaultListProperty[T]) Remove(element T) []T {
	result, _ := p.updateCurrentValue(ListRemove[T]{Element: element}, func(s []T) []T {
		for i, v := range s {
			if reflect.DeepEqual(v, element) {
				out := make([]T, 0, len(s)-1)
				out = append(out, s[:i]...)
				out = append(out, s[i+1:]...)
				return out
			}
		}
		return append([]T{}, s...)
	})
	return result
}

func (p *defaultListProperty[T]) RemoveAt(index int) []T {
	result, _ := p.updateCurrentValue(ListRemoveAt{Index: index}, func(s []T) []T {
		out := make([]T, 0, len(s)-1)
		out = append(out, s[:index]...)
		out = append(out, s[index+1:]...)
		return out
	})
	return result
}

func (p *defaultListProperty[T]) AddAll(elements []T) []T {
	cloned := append([]T{}, elements...)
	result, _ := p.updateCurrentValue(ListAddAll[T]{Elements: cloned}, func(s []T) []T {
		out := make([]T, 0, len(s)+len(elements))
		out = append(out, s...)
		out = append(out, elements...)
		return out
	})
	return result
}

func (p *defaultListProperty[T]) AddAllAt(index int, c []T) []T {
	cloned := append([]T{}, c...)
	result, _ := p.updateCurrentValue(ListAddAllAt[T]{Index: index, C: cloned}, func(s []T) []T {
		out := make([]T, 0, len(s)+len(c))
		out = append(out, s[:index]...)
		out = append(out, c...)
		out = append(out, s[index:]...)
		return out
	})
	return result
}

func (p *defaultListProperty[T]) RemoveAll(elements []T) []T {
	cloned := append([]T{}, elements...)
	result, _ := p.updateCurrentValue(ListRemoveAll[T]{Elements: cloned}, func(s []T) []T {
		out := make([]T, 0, len(s))
		for _, v := range s {
			if !containsDeep(elements, v) {
				out = append(out, v)
			}
		}
		return out
	})
	return result
}

func (p *defaultListProperty[T]) RemoveAllWhere(predicate func(T) bool) []T {
	result, _ := p.updateCurrentValue(ListRemoveAllWhen[T]{Predicate: predicate}, func(s []T) []T {
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

func (p *defaultListProperty[T]) RetainAll(elements []T) []T {
	cloned := append([]T{}, elements...)
	result, _ := p.updateCurrentValue(ListRetainAll[T]{Elements: cloned}, func(s []T) []T {
		out := make([]T, 0, len(s))
		for _, v := range s {
			if containsDeep(elements, v) {
				out = append(out, v)
			}
		}
		return out
	})
	return result
}

func (p *defaultListProperty[T]) Clear() []T {
	result, _ := p.updateCurrentValue(ListClear, func(s []T) []T { return []T{} })
	return result
}

// UpdateFrom implements WritableValue.
func (p *defaultListProperty[T]) UpdateFrom(observable ObservableValue[[]T], action BindingAction) Binding[[]T] {
	return UpdateFromConverter[[]T, []T](p, observable, action, func(t []T) []T { return t })
}

// Bind implements Property.
func (p *defaultListProperty[T]) Bind(other Property[[]T], action BindingAction) Binding[[]T] {
	return BindWithConverter[[]T, []T](p, other, action, IdentityConverter[[]T]{})
}

// AsDelegate implements Property.
func (p *defaultListProperty[T]) AsDelegate() PropertyDelegate[[]T] {
	return newDefaultPropertyDelegate[[]T](p)
}

// containsDeep is the shared linear-scan helper for list mutators
// that compare element-by-element. Kept package-private; element
// type T any rules out using equality directly.
func containsDeep[T any](xs []T, e T) bool {
	for _, x := range xs {
		if reflect.DeepEqual(x, e) {
			return true
		}
	}
	return false
}
