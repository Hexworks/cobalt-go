// Package cobalt holds the singleton event bus used by the
// databinding package. The path-internal location (under
// databinding/internal/) restricts visibility to databinding code, so
// downstream consumers can't accidentally publish on or close the
// shared bus.
package cobalt

import "github.com/hexworks/cobalt-go/events"

// bus is the process-wide bus that backs every property's change
// notifications. The Kotlin original is a single object on the JVM
// classpath; we mirror that with a package-level var initialised
// once at first use.
var bus events.EventBus = events.NewEventBus()

// EventBus returns the singleton bus.
func EventBus() events.EventBus {
	return bus
}

// ResetForTest replaces the singleton with a fresh bus. Tests that
// exercise teardown semantics (Close, CancelScope) need a clean
// instance per case; production code must never call this.
func ResetForTest() {
	bus = events.NewEventBus()
}
