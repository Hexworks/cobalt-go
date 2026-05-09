package databinding

import "fmt"

// ValueValidationFailedError is the Go counterpart of Kotlin's
// ValueValidationFailedException. BaseProperty raises one of these
// when a PropertyValidator rejects a new value, and the surrounding
// TransformValue / UpdateValue boundary converts it into a
// ValueValidationFailed result.
type ValueValidationFailedError struct {
	NewValue any
	Message  string
}

// NewValueValidationFailedError builds an error describing a rejected
// value with the standard message format used by BaseProperty.
func NewValueValidationFailedError(newValue any, message string) *ValueValidationFailedError {
	return &ValueValidationFailedError{NewValue: newValue, Message: message}
}

func (e *ValueValidationFailedError) Error() string {
	return e.Message
}

// String renders the error for fmt's %v / %s verbs so log output is
// consistent with the Kotlin RuntimeException.toString form.
func (e *ValueValidationFailedError) String() string {
	return fmt.Sprintf("ValueValidationFailedException: %s", e.Message)
}

// ValueValidationResult is the sealed outcome of a validating write.
// Only ValueValidationSuccessful and ValueValidationFailed implement
// it.
type ValueValidationResult[T any] interface {
	// IsSuccessful reports whether the write was accepted.
	IsSuccessful() bool
	// GetValue returns the value associated with the result — the
	// stored value for success, the rejected candidate for failure.
	GetValue() T
	isValueValidationResult()
}

// ValueValidationSuccessful carries the value that is now stored on
// the property.
type ValueValidationSuccessful[T any] struct {
	Value T
}

func (ValueValidationSuccessful[T]) IsSuccessful() bool   { return true }
func (s ValueValidationSuccessful[T]) GetValue() T        { return s.Value }
func (ValueValidationSuccessful[T]) isValueValidationResult() {}

// ValueValidationFailed carries the rejected candidate and the cause.
type ValueValidationFailed[T any] struct {
	Value T
	Cause *ValueValidationFailedError
}

func (ValueValidationFailed[T]) IsSuccessful() bool       { return false }
func (f ValueValidationFailed[T]) GetValue() T            { return f.Value }
func (ValueValidationFailed[T]) isValueValidationResult() {}
