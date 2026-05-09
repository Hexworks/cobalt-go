package events

// EventBus broadcasts Events to subscribers keyed by EventScope and
// the Event's key. The port keeps generic-free methods on the
// interface (Go forbids generic interface methods); use the package
// level SubscribeTo / SimpleSubscribeTo helpers for type-safe wiring.
type EventBus interface {
	// FetchSubscribers returns every subscription registered for the
	// given key/scope. The returned slice is a defensive copy.
	FetchSubscribers(key string, scope EventScope) []Subscription

	// Subscribe registers fn for events with the given key and scope
	// and returns the resulting Subscription.
	Subscribe(key string, scope EventScope, fn func(Event) CallbackResult) Subscription

	// Publish dispatches event to every subscriber matching scope and
	// event.Key().
	Publish(event Event, scope EventScope)

	// CancelScope disposes every subscription registered under scope.
	CancelScope(scope EventScope)

	// Close disposes every subscription and prevents further use of
	// the bus.
	Close()
}

// NewEventBus constructs a fresh EventBus.
func NewEventBus() EventBus {
	return newDefaultEventBus()
}

type applicationScope struct{}

// ApplicationScope is the default EventScope, mirroring the Kotlin
// ApplicationScope singleton. The port keeps it explicit at every
// call site (Go has no default arguments), so callers must pass it
// rather than relying on a default.
var ApplicationScope EventScope = applicationScope{}

// SubscribeTo is the generic counterpart of EventBus.Subscribe. The
// callback receives the event already asserted to its descriptor type.
func SubscribeTo[E Event](
	bus EventBus,
	descriptor EventDescriptor[E],
	scope EventScope,
	fn func(E) CallbackResult,
) Subscription {
	return bus.Subscribe(descriptor.Key, scope, func(e Event) CallbackResult {
		return fn(e.(E))
	})
}

// SimpleSubscribeTo registers a fire-and-forget callback that always
// keeps its subscription. Equivalent to the Kotlin simpleSubscribeTo
// helper.
func SimpleSubscribeTo[E Event](
	bus EventBus,
	descriptor EventDescriptor[E],
	scope EventScope,
	fn func(E),
) Subscription {
	return SubscribeTo(bus, descriptor, scope, func(e E) CallbackResult {
		fn(e)
		return KeepSubscription
	})
}
