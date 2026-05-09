package databinding

import (
	"cmp"
	"fmt"
)

// Numeric covers the integer and floating-point built-ins. Used as
// the constraint for arithmetic expression bindings (Plus, Minus,
// Times, Div). Unsigned types are included for completeness but
// BindNegate restricts itself to Signed.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Signed restricts to numeric types that support unary minus.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// BindNegate creates a binding holding -v.Value(). Kotlin had one
// bindNegate per numeric type; the Go port collapses them via the
// Signed constraint.
func BindNegate[T Signed](v ObservableValue[T]) Binding[T] {
	return BindTransform[T, T](v, func(x T) T { return -x })
}

// BindPlus exposes a + b as a binding. Kotlin allowed a Number on
// the right-hand side and narrowed back to the receiver's type; in
// Go we require both operands share T because that's the only sane
// thing to do without runtime numeric conversions.
func BindPlus[T Numeric](a, b ObservableValue[T]) Binding[T] {
	return BindCompute[T, T, T](a, b, func(x, y T) T { return x + y })
}

// BindMinus exposes a - b as a binding.
func BindMinus[T Numeric](a, b ObservableValue[T]) Binding[T] {
	return BindCompute[T, T, T](a, b, func(x, y T) T { return x - y })
}

// BindTimes exposes a * b as a binding.
func BindTimes[T Numeric](a, b ObservableValue[T]) Binding[T] {
	return BindCompute[T, T, T](a, b, func(x, y T) T { return x * y })
}

// BindDiv exposes a / b as a binding. For integer T the result
// follows Go's integer-division semantics (toward zero).
func BindDiv[T Numeric](a, b ObservableValue[T]) Binding[T] {
	return BindCompute[T, T, T](a, b, func(x, y T) T { return x / y })
}

// BindGreaterThan exposes a > b as a Binding[bool]. Works for every
// cmp.Ordered (numerics, strings); the Kotlin original limited it
// per numeric type because Kotlin's Comparable hierarchy doesn't
// translate cleanly into generic bounds.
func BindGreaterThan[T cmp.Ordered](a, b ObservableValue[T]) Binding[bool] {
	return BindCompute[T, T, bool](a, b, func(x, y T) bool { return x > y })
}

// BindLessThan exposes a < b as a Binding[bool].
func BindLessThan[T cmp.Ordered](a, b ObservableValue[T]) Binding[bool] {
	return BindCompute[T, T, bool](a, b, func(x, y T) bool { return x < y })
}

// BindGreaterThanOrEqual exposes a >= b as a Binding[bool].
func BindGreaterThanOrEqual[T cmp.Ordered](a, b ObservableValue[T]) Binding[bool] {
	return BindCompute[T, T, bool](a, b, func(x, y T) bool { return x >= y })
}

// BindLessThanOrEqual exposes a <= b as a Binding[bool].
func BindLessThanOrEqual[T cmp.Ordered](a, b ObservableValue[T]) Binding[bool] {
	return BindCompute[T, T, bool](a, b, func(x, y T) bool { return x <= y })
}

// BindEquals exposes a == b as a Binding[bool]. Replaces the per-
// type bindEqualsWith helpers; the constraint comparable covers
// every value type Kotlin exposed bindings for (numerics, strings,
// bools, structs that happen to be comparable).
func BindEquals[T comparable](a, b ObservableValue[T]) Binding[bool] {
	return BindCompute[T, T, bool](a, b, func(x, y T) bool { return x == y })
}

// BindToString routes v.Value() through fmt.Sprint, so any type
// formatted by fmt becomes a Binding[string]. Replaces the per-type
// bindToString helpers in the Kotlin original.
func BindToString[T any](v ObservableValue[T]) Binding[string] {
	return BindTransform[T, string](v, func(x T) string { return fmt.Sprint(x) })
}
