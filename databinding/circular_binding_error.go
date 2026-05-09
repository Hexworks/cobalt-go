package databinding

// circularBindingError is the Kotlin CircularBindingException port.
// Raised by BaseProperty.updateWithEvent (stage 2) when an event's
// trace already contains the receiver, then caught in the same
// function so the cycle is logged and dropped rather than crashing
// the bus. Unexported because it never crosses a package boundary.
type circularBindingError struct {
	Message string
}

func (e *circularBindingError) Error() string {
	return e.Message
}
