// Package events ports the Kotlin cobalt event bus. The bus dispatches
// Event values to subscribers keyed by (EventScope, Event.Key()). The
// port is single-threaded: no goroutines, no locks.
package events

import "github.com/hexworks/cobalt-go/core"

// Event is the common interface for everything that can be sent through
// an EventBus. Each Event has a Key used for routing, an Emitter that
// produced it, and an optional Trace chain showing the events that led
// to it (most recent first).
type Event interface {
	Key() string
	Emitter() EventSource
	Trace() []Event
}

// EventSource is anything that can emit Events and be identified by a
// stable UUID.
type EventSource interface {
	ID() core.UUID
}

// EventScope partitions the event bus into independent namespaces.
// Events published with one scope are invisible to subscribers in
// another. The bus uses scopes as map keys, so any value used here
// must be comparable; empty struct singletons are the idiomatic
// choice.
type EventScope any

// EventDescriptor pairs an Event subtype with the routing key the bus
// uses to look up its subscribers. The type parameter exists so the
// generic SubscribeTo / SimpleSubscribeTo helpers can deliver a typed
// callback without callers writing a type assertion.
type EventDescriptor[E Event] struct {
	Key string
}
