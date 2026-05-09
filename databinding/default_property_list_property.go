package databinding

import (
	"log/slog"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// propertySubscription pairs a child property with the subscription
// the outer collection holds against it. Stored in
// uniqueProperties so the outer can dispose the child subscriptions
// when an element is removed.
type propertySubscription[P any] struct {
	prop P
	sub  events.Subscription
}

// defaultPropertyListProperty is a ListProperty[P] whose elements
// are themselves observable values. Whenever any contained element
// fires a change event, the outer list republishes a
// ListPropertyChange wrapping the inner event so listeners on the
// outer list see element-level mutations too.
//
// Embeds a *defaultListProperty[P] for storage + the standard list
// mutator surface; overrides every mutator to keep the child
// subscriptions in lockstep with the underlying slice.
type defaultPropertyListProperty[T any, P ObservableValue[T]] struct {
	*defaultListProperty[P]
	uniqueProperties map[core.UUID]propertySubscription[P]
}

// NewPropertyListProperty builds a list-of-properties where every
// contained P is auto-subscribed for change forwarding. The returned
// value satisfies ListProperty[P]. Mutators handle subscribe /
// unsubscribe bookkeeping; raw SetValue / UpdateValue do NOT —
// matches the Kotlin behaviour where the value setter bypasses the
// per-element tracking.
func NewPropertyListProperty[T any, P ObservableValue[T]](initial []P, name string, validator PropertyValidator[[]P]) ListProperty[P] {
	if name == "" {
		name = "DefaultPropertyListProperty"
	}
	cloned := append([]P{}, initial...)
	inner := &defaultListProperty[P]{
		basePropertyState: newBasePropertyState[[]P](cloned, name, validator),
	}
	p := &defaultPropertyListProperty[T, P]{
		defaultListProperty: inner,
		uniqueProperties:    make(map[core.UUID]propertySubscription[P]),
	}
	inner.self = p
	for _, elem := range cloned {
		p.subscribeToChanges(elem)
	}
	return p
}

func (p *defaultPropertyListProperty[T, P]) Add(element P) []P {
	p.subscribeToChanges(element)
	return p.defaultListProperty.Add(element)
}

func (p *defaultPropertyListProperty[T, P]) AddAt(index int, element P) []P {
	p.subscribeToChanges(element)
	return p.defaultListProperty.AddAt(index, element)
}

func (p *defaultPropertyListProperty[T, P]) Set(index int, element P) []P {
	current := p.Value()
	if index < len(current) {
		p.unsubscribeFromChanges(current[index])
		p.subscribeToChanges(element)
	}
	return p.defaultListProperty.Set(index, element)
}

func (p *defaultPropertyListProperty[T, P]) Remove(element P) []P {
	p.unsubscribeFromChanges(element)
	return p.defaultListProperty.Remove(element)
}

func (p *defaultPropertyListProperty[T, P]) RemoveAt(index int) []P {
	current := p.Value()
	if index < len(current) {
		p.unsubscribeFromChanges(current[index])
	}
	return p.defaultListProperty.RemoveAt(index)
}

func (p *defaultPropertyListProperty[T, P]) AddAll(elements []P) []P {
	for _, e := range elements {
		p.subscribeToChanges(e)
	}
	return p.defaultListProperty.AddAll(elements)
}

func (p *defaultPropertyListProperty[T, P]) AddAllAt(index int, c []P) []P {
	for _, e := range c {
		p.subscribeToChanges(e)
	}
	return p.defaultListProperty.AddAllAt(index, c)
}

func (p *defaultPropertyListProperty[T, P]) RemoveAll(elements []P) []P {
	for _, e := range elements {
		p.unsubscribeFromChanges(e)
	}
	return p.defaultListProperty.RemoveAll(elements)
}

func (p *defaultPropertyListProperty[T, P]) RemoveAllWhere(predicate func(P) bool) []P {
	for _, v := range p.Value() {
		if predicate(v) {
			p.unsubscribeFromChanges(v)
		}
	}
	return p.defaultListProperty.RemoveAllWhere(predicate)
}

func (p *defaultPropertyListProperty[T, P]) RetainAll(elements []P) []P {
	keep := make(map[core.UUID]struct{}, len(elements))
	for _, e := range elements {
		keep[e.ID()] = struct{}{}
	}
	for _, v := range p.Value() {
		if _, ok := keep[v.ID()]; !ok {
			p.unsubscribeFromChanges(v)
		}
	}
	return p.defaultListProperty.RetainAll(elements)
}

func (p *defaultPropertyListProperty[T, P]) Clear() []P {
	for _, v := range p.Value() {
		p.unsubscribeFromChanges(v)
	}
	if len(p.uniqueProperties) > 0 {
		slog.Warn("remaining property subscriptions after clearing list; possible bug")
	}
	return p.defaultListProperty.Clear()
}

// subscribeToChanges wires elem's OnChange into the outer's
// ListPropertyChange republisher. Idempotent per element id.
func (p *defaultPropertyListProperty[T, P]) subscribeToChanges(elem P) {
	if _, exists := p.uniqueProperties[elem.ID()]; exists {
		return
	}
	sub := elem.OnChange(func(ovc ObservableValueChanged[T]) {
		p.updateCurrentValue(ListPropertyChange{ChangeEvent: ovc}, func(s []P) []P { return s })
	})
	p.uniqueProperties[elem.ID()] = propertySubscription[P]{prop: elem, sub: sub}
}

func (p *defaultPropertyListProperty[T, P]) unsubscribeFromChanges(elem P) {
	ps, ok := p.uniqueProperties[elem.ID()]
	if !ok {
		return
	}
	delete(p.uniqueProperties, elem.ID())
	ps.sub.Dispose(core.DisposedManually)
}
