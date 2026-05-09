package databinding

import (
	"log/slog"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// defaultPropertySetProperty is a SetProperty[P] whose elements are
// themselves observable values. P must be comparable so it can serve
// as the underlying set's element type; in practice callers will use
// concrete pointer property types which Go reports as comparable.
type defaultPropertySetProperty[T any, P interface {
	ObservableValue[T]
	comparable
}] struct {
	*defaultSetProperty[P]
	uniqueProperties map[core.UUID]struct {
		prop P
		sub  events.Subscription
	}
}

// NewPropertySetProperty builds a set-of-properties with automatic
// change forwarding. SetPropertyChange events are published whenever
// any contained property fires a change.
func NewPropertySetProperty[T any, P interface {
	ObservableValue[T]
	comparable
}](initial []P, name string, validator PropertyValidator[[]P]) SetProperty[P] {
	if name == "" {
		name = "DefaultPropertySetProperty"
	}
	seen := make(map[P]struct{}, len(initial))
	deduped := make([]P, 0, len(initial))
	for _, v := range initial {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		deduped = append(deduped, v)
	}
	inner := &defaultSetProperty[P]{
		basePropertyState: newBasePropertyState[[]P](deduped, name, validator),
	}
	p := &defaultPropertySetProperty[T, P]{
		defaultSetProperty: inner,
		uniqueProperties: make(map[core.UUID]struct {
			prop P
			sub  events.Subscription
		}),
	}
	inner.self = p
	for _, elem := range deduped {
		p.subscribeToChanges(elem)
	}
	return p
}

func (p *defaultPropertySetProperty[T, P]) Add(element P) []P {
	p.subscribeToChanges(element)
	return p.defaultSetProperty.Add(element)
}

func (p *defaultPropertySetProperty[T, P]) AddAll(elements []P) []P {
	for _, e := range elements {
		p.subscribeToChanges(e)
	}
	return p.defaultSetProperty.AddAll(elements)
}

func (p *defaultPropertySetProperty[T, P]) Remove(element P) []P {
	p.unsubscribeFromChanges(element)
	return p.defaultSetProperty.Remove(element)
}

func (p *defaultPropertySetProperty[T, P]) RemoveAll(elements []P) []P {
	for _, e := range elements {
		p.unsubscribeFromChanges(e)
	}
	return p.defaultSetProperty.RemoveAll(elements)
}

func (p *defaultPropertySetProperty[T, P]) RemoveAllWhere(predicate func(P) bool) []P {
	for _, v := range p.Value() {
		if predicate(v) {
			p.unsubscribeFromChanges(v)
		}
	}
	return p.defaultSetProperty.RemoveAllWhere(predicate)
}

func (p *defaultPropertySetProperty[T, P]) RetainAll(elements []P) []P {
	keep := make(map[core.UUID]struct{}, len(elements))
	for _, e := range elements {
		keep[e.ID()] = struct{}{}
	}
	for _, v := range p.Value() {
		if _, ok := keep[v.ID()]; !ok {
			p.unsubscribeFromChanges(v)
		}
	}
	return p.defaultSetProperty.RetainAll(elements)
}

func (p *defaultPropertySetProperty[T, P]) Clear() []P {
	for _, v := range p.Value() {
		p.unsubscribeFromChanges(v)
	}
	if len(p.uniqueProperties) > 0 {
		slog.Warn("remaining property subscriptions after clearing set; possible bug")
	}
	return p.defaultSetProperty.Clear()
}

func (p *defaultPropertySetProperty[T, P]) subscribeToChanges(elem P) {
	if _, exists := p.uniqueProperties[elem.ID()]; exists {
		return
	}
	sub := elem.OnChange(func(ovc ObservableValueChanged[T]) {
		p.updateCurrentValue(SetPropertyChange{ChangeEvent: ovc}, func(s []P) []P { return s })
	})
	p.uniqueProperties[elem.ID()] = struct {
		prop P
		sub  events.Subscription
	}{prop: elem, sub: sub}
}

func (p *defaultPropertySetProperty[T, P]) unsubscribeFromChanges(elem P) {
	ps, ok := p.uniqueProperties[elem.ID()]
	if !ok {
		return
	}
	delete(p.uniqueProperties, elem.ID())
	ps.sub.Dispose(core.DisposedManually)
}
