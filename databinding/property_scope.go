package databinding

import "github.com/hexworks/cobalt-go/core"

// PropertyScope is the EventScope under which a property publishes
// its change events. Wrapping a UUID gives each property its own
// scope on the singleton Cobalt bus, so subscribers only receive
// changes from the property they're watching.
//
// Implements events.EventScope (= any) by being a plain comparable
// struct.
type PropertyScope struct {
	ID core.UUID
}

// NewPropertyScope returns a PropertyScope for id.
func NewPropertyScope(id core.UUID) PropertyScope {
	return PropertyScope{ID: id}
}
