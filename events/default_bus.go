package events

import (
	"fmt"
	"log/slog"

	"github.com/hexworks/cobalt-go/core"
)

type subscriberKey struct {
	scope EventScope
	key   string
}

type eventSubscriptions struct {
	subs []*busSubscription
}

type defaultEventBus struct {
	closed        bool
	subscriptions map[subscriberKey]*eventSubscriptions
}

func newDefaultEventBus() *defaultEventBus {
	return &defaultEventBus{
		subscriptions: map[subscriberKey]*eventSubscriptions{},
	}
}

func (b *defaultEventBus) FetchSubscribers(key string, scope EventScope) []Subscription {
	b.checkOpen()
	entry, ok := b.subscriptions[subscriberKey{scope: scope, key: key}]
	if !ok {
		return nil
	}
	out := make([]Subscription, len(entry.subs))
	for i, s := range entry.subs {
		out[i] = s
	}
	return out
}

func (b *defaultEventBus) Subscribe(key string, scope EventScope, fn func(Event) CallbackResult) Subscription {
	b.checkOpen()
	slog.Debug("subscribing", "key", key, "scope", scope)
	sub := &busSubscription{
		bus:       b,
		scope:     scope,
		key:       key,
		callback:  fn,
		disposeSt: core.NotDisposed,
	}
	subKey := subscriberKey{scope: scope, key: key}
	if entry, ok := b.subscriptions[subKey]; ok {
		entry.subs = append(entry.subs, sub)
	} else {
		b.subscriptions[subKey] = &eventSubscriptions{subs: []*busSubscription{sub}}
	}
	return sub
}

func (b *defaultEventBus) Publish(event Event, scope EventScope) {
	b.checkOpen()
	slog.Debug("publishing", "key", event.Key(), "scope", scope)
	entry, ok := b.subscriptions[subscriberKey{scope: scope, key: event.Key()}]
	if !ok {
		return
	}
	// Snapshot before iterating so callbacks may dispose themselves
	// (or others) without mutating the slice we're walking.
	snapshot := make([]*busSubscription, len(entry.subs))
	copy(snapshot, entry.subs)
	for _, sub := range snapshot {
		b.deliver(sub, event)
	}
}

func (b *defaultEventBus) deliver(sub *busSubscription, event Event) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			slog.Warn("cancelling failed subscription", "err", err)
			func() {
				defer func() {
					if rr := recover(); rr != nil {
						slog.Warn("failed to cancel subscription", "err", rr)
					}
				}()
				sub.Dispose(core.DisposedByException{Err: err})
			}()
		}
	}()
	if _, dispose := sub.callback(event).(disposeSubscription); dispose {
		sub.Dispose(core.DisposedManually)
	}
}

func (b *defaultEventBus) CancelScope(scope EventScope) {
	b.checkOpen()
	slog.Debug("cancelling scope", "scope", scope)
	var targets []*busSubscription
	for k, entry := range b.subscriptions {
		if k.scope == scope {
			targets = append(targets, entry.subs...)
		}
	}
	for _, sub := range targets {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Warn("cancelling subscription failed while cancelling scope", "err", r)
				}
			}()
			sub.Dispose(core.DisposedManually)
		}()
	}
}

func (b *defaultEventBus) Close() {
	b.closed = true
	var all []*busSubscription
	for _, entry := range b.subscriptions {
		all = append(all, entry.subs...)
	}
	for _, sub := range all {
		sub.Dispose(core.DisposedManually)
	}
}

func (b *defaultEventBus) checkOpen() {
	if b.closed {
		panic("This Event Bus is already closed.")
	}
}

type busSubscription struct {
	bus       *defaultEventBus
	scope     EventScope
	key       string
	callback  func(Event) CallbackResult
	disposeSt core.DisposeState
}

func (s *busSubscription) isSubscription() {}

func (s *busSubscription) DisposeState() core.DisposeState { return s.disposeSt }

func (s *busSubscription) Dispose(state core.DisposeState) {
	slog.Debug("cancelling subscription", "scope", s.scope, "key", s.key)
	s.disposeSt = state
	key := subscriberKey{scope: s.scope, key: s.key}
	entry, ok := s.bus.subscriptions[key]
	if !ok {
		return
	}
	for i, candidate := range entry.subs {
		if candidate == s {
			entry.subs = append(entry.subs[:i], entry.subs[i+1:]...)
			break
		}
	}
	if len(entry.subs) == 0 {
		delete(s.bus.subscriptions, key)
	}
}
