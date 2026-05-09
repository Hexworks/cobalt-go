package databinding

// Converter turns a value of type S into a value of type T. The
// Kotlin original declared the interface so users could plug in
// non-lambda implementations on Kotlin/JS; Go has no equivalent
// limitation, but the interface is preserved so the binding APIs
// match the upstream signatures.
type Converter[S, T any] interface {
	Convert(source S) T
}

// IsomorphicConverter is a Converter that can also convert back, so
// it can drive bidirectional bindings between properties of
// different types.
type IsomorphicConverter[S, T any] interface {
	Converter[S, T]
	ConvertBack(target T) S
}

// IdentityConverter is the converter used when both ends of a
// binding share the same type. Useful as a generic zero value for
// IsomorphicConverter[T, T] in Property.Bind.
type IdentityConverter[T any] struct{}

// Convert returns source unchanged.
func (IdentityConverter[T]) Convert(source T) T { return source }

// ConvertBack returns target unchanged.
func (IdentityConverter[T]) ConvertBack(target T) T { return target }

// converterFunc adapts a plain function to the Converter interface.
type converterFunc[S, T any] func(S) T

func (f converterFunc[S, T]) Convert(s S) T { return f(s) }

// ConverterFromFunc lifts a plain function into a Converter. This is
// the Go equivalent of the Kotlin `((S) -> T).toConverter()`
// extension.
func ConverterFromFunc[S, T any](fn func(S) T) Converter[S, T] {
	return converterFunc[S, T](fn)
}

// ReverseConverter swaps the type parameters of an
// IsomorphicConverter. Equivalent to IsomorphicConverter.reverseConverter
// in Kotlin. Returns a Converter (not an IsomorphicConverter): the
// reversed view doesn't need to convert back through the original
// direction.
func ReverseConverter[S, T any](c IsomorphicConverter[S, T]) Converter[T, S] {
	return converterFunc[T, S](c.ConvertBack)
}
