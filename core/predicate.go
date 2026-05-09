package core

// Predicate is a function that tests a value of type T.
type Predicate[T any] func(T) bool

// And returns a predicate that is true when both a and b are true.
func And[T any](a, b Predicate[T]) Predicate[T] {
	return func(v T) bool { return a(v) && b(v) }
}

// Or returns a predicate that is true when either a or b is true.
func Or[T any](a, b Predicate[T]) Predicate[T] {
	return func(v T) bool { return a(v) || b(v) }
}

// Not returns a predicate that is true when p is false.
func Not[T any](p Predicate[T]) Predicate[T] {
	return func(v T) bool { return !p(v) }
}
