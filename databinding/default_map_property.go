package databinding

import "reflect"

// defaultMapProperty is the canonical MapProperty implementation.
// Storage is a plain `map[K]V`; mutators clone-on-write via cloneMap
// so the change event's old value and new value are independent
// snapshots.
type defaultMapProperty[K comparable, V any] struct {
	basePropertyState[map[K]V]
}

// NewMapProperty constructs a MapProperty seeded with a defensive
// copy of initial. name defaults to "DefaultMapProperty" when blank.
func NewMapProperty[K comparable, V any](initial map[K]V, name string, validator PropertyValidator[map[K]V]) MapProperty[K, V] {
	if name == "" {
		name = "DefaultMapProperty"
	}
	cloned := cloneMap(initial)
	p := &defaultMapProperty[K, V]{
		basePropertyState: newBasePropertyState[map[K]V](cloned, name, validator),
	}
	p.self = p
	return p
}

func (p *defaultMapProperty[K, V]) Size() int     { return len(p.Value()) }
func (p *defaultMapProperty[K, V]) IsEmpty() bool { return len(p.Value()) == 0 }

func (p *defaultMapProperty[K, V]) ContainsKey(key K) bool {
	_, ok := p.Value()[key]
	return ok
}

func (p *defaultMapProperty[K, V]) ContainsValue(value V) bool {
	for _, v := range p.Value() {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

func (p *defaultMapProperty[K, V]) Get(key K) (V, bool) {
	v, ok := p.Value()[key]
	return v, ok
}

func (p *defaultMapProperty[K, V]) Keys() []K {
	m := p.Value()
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func (p *defaultMapProperty[K, V]) Values() []V {
	m := p.Value()
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func (p *defaultMapProperty[K, V]) Put(key K, value V) map[K]V {
	result, _ := p.updateCurrentValue(MapPut[K, V]{Key: key, Value: value}, func(m map[K]V) map[K]V {
		out := cloneMap(m)
		out[key] = value
		return out
	})
	return result
}

func (p *defaultMapProperty[K, V]) PutAll(other map[K]V) map[K]V {
	cloned := cloneMap(other)
	result, _ := p.updateCurrentValue(MapPutAll[K, V]{M: cloned}, func(m map[K]V) map[K]V {
		out := cloneMap(m)
		for k, v := range other {
			out[k] = v
		}
		return out
	})
	return result
}

func (p *defaultMapProperty[K, V]) Remove(key K) map[K]V {
	result, _ := p.updateCurrentValue(MapRemove[K]{Key: key}, func(m map[K]V) map[K]V {
		out := cloneMap(m)
		delete(out, key)
		return out
	})
	return result
}

func (p *defaultMapProperty[K, V]) RemoveWithValue(key K, value V) map[K]V {
	result, _ := p.updateCurrentValue(MapRemoveWithValue[K, V]{Key: key, Value: value}, func(m map[K]V) map[K]V {
		out := cloneMap(m)
		if cur, ok := out[key]; ok && reflect.DeepEqual(cur, value) {
			delete(out, key)
		}
		return out
	})
	return result
}

func (p *defaultMapProperty[K, V]) Clear() map[K]V {
	result, _ := p.updateCurrentValue(MapClear, func(m map[K]V) map[K]V { return map[K]V{} })
	return result
}

func (p *defaultMapProperty[K, V]) UpdateFrom(observable ObservableValue[map[K]V], action BindingAction) Binding[map[K]V] {
	return UpdateFromConverter[map[K]V, map[K]V](p, observable, action, func(t map[K]V) map[K]V { return t })
}

func (p *defaultMapProperty[K, V]) Bind(other Property[map[K]V], action BindingAction) Binding[map[K]V] {
	return BindWithConverter[map[K]V, map[K]V](p, other, action, IdentityConverter[map[K]V]{})
}

func (p *defaultMapProperty[K, V]) AsDelegate() PropertyDelegate[map[K]V] {
	return newDefaultPropertyDelegate[map[K]V](p)
}

// cloneMap returns a shallow copy of m. nil maps clone to non-nil
// empty maps so the property's backing storage is never nil — saves
// a nil check in every mutator.
func cloneMap[K comparable, V any](m map[K]V) map[K]V {
	out := make(map[K]V, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
