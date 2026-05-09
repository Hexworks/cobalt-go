package databinding

import "github.com/hexworks/cobalt-go/core"

// defaultPropertyMapProperty is a MapProperty[K, P] whose values are
// observable properties. Changes from any contained P are
// republished as MapPropertyChange events.
//
// The Kotlin original wraps inner changes in ListPropertyChange (a
// likely typo upstream). We preserve the intent — that the outer
// fires the matching collection-change variant — by using
// MapPropertyChange here, so MapProperty consumers see a coherent
// MapChange feed.
type defaultPropertyMapProperty[K comparable, T any, P ObservableValue[T]] struct {
	*defaultMapProperty[K, P]
	uniqueProperties map[core.UUID]propertySubscription[P]
}

// NewPropertyMapProperty builds a MapProperty[K, P] with automatic
// per-value change forwarding.
func NewPropertyMapProperty[K comparable, T any, P ObservableValue[T]](
	initial map[K]P,
	name string,
	validator PropertyValidator[map[K]P],
) MapProperty[K, P] {
	if name == "" {
		name = "DefaultPropertyMapProperty"
	}
	cloned := cloneMap(initial)
	inner := &defaultMapProperty[K, P]{
		basePropertyState: newBasePropertyState[map[K]P](cloned, name, validator),
	}
	p := &defaultPropertyMapProperty[K, T, P]{
		defaultMapProperty: inner,
		uniqueProperties:   make(map[core.UUID]propertySubscription[P]),
	}
	inner.self = p
	for _, elem := range cloned {
		p.subscribeToChanges(elem)
	}
	return p
}

func (p *defaultPropertyMapProperty[K, T, P]) Put(key K, value P) map[K]P {
	p.subscribeToChanges(value)
	return p.defaultMapProperty.Put(key, value)
}

func (p *defaultPropertyMapProperty[K, T, P]) PutAll(m map[K]P) map[K]P {
	for _, v := range m {
		p.subscribeToChanges(v)
	}
	return p.defaultMapProperty.PutAll(m)
}

func (p *defaultPropertyMapProperty[K, T, P]) Remove(key K) map[K]P {
	if existing, ok := p.Value()[key]; ok {
		p.unsubscribeFromChanges(existing)
	}
	return p.defaultMapProperty.Remove(key)
}

func (p *defaultPropertyMapProperty[K, T, P]) RemoveWithValue(key K, value P) map[K]P {
	if existing, ok := p.Value()[key]; ok {
		p.unsubscribeFromChanges(existing)
	}
	return p.defaultMapProperty.RemoveWithValue(key, value)
}

func (p *defaultPropertyMapProperty[K, T, P]) Clear() map[K]P {
	for _, v := range p.Value() {
		p.unsubscribeFromChanges(v)
	}
	return p.defaultMapProperty.Clear()
}

func (p *defaultPropertyMapProperty[K, T, P]) subscribeToChanges(elem P) {
	if _, exists := p.uniqueProperties[elem.ID()]; exists {
		return
	}
	sub := elem.OnChange(func(ovc ObservableValueChanged[T]) {
		p.updateCurrentValue(MapPropertyChange{ChangeEvent: ovc}, func(m map[K]P) map[K]P { return m })
	})
	p.uniqueProperties[elem.ID()] = propertySubscription[P]{prop: elem, sub: sub}
}

func (p *defaultPropertyMapProperty[K, T, P]) unsubscribeFromChanges(elem P) {
	ps, ok := p.uniqueProperties[elem.ID()]
	if !ok {
		return
	}
	delete(p.uniqueProperties, elem.ID())
	ps.sub.Dispose(core.DisposedManually)
}

