package databinding

import "strings"

// BindStringIsEmpty exposes len(v.Value()) == 0 as a Binding[bool].
func BindStringIsEmpty(v ObservableValue[string]) Binding[bool] {
	return BindTransform[string, bool](v, func(s string) bool { return len(s) == 0 })
}

// BindStringIsBlank exposes "is the string empty or whitespace-only"
// as a Binding[bool]. Mirrors Kotlin's String.isBlank by stripping
// Unicode whitespace before checking emptiness.
func BindStringIsBlank(v ObservableValue[string]) Binding[bool] {
	return BindTransform[string, bool](v, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
}

// BindStringConcat exposes a + b (concatenation) as a Binding[string].
func BindStringConcat(a, b ObservableValue[string]) Binding[string] {
	return BindCompute[string, string, string](a, b, func(x, y string) string { return x + y })
}

// BindStringEqualsIgnoreCase exposes case-insensitive equality of a
// and b as a Binding[bool]. Uses strings.EqualFold so the
// comparison handles Unicode correctly the same way Kotlin's
// equals(ignoreCase = true) does.
func BindStringEqualsIgnoreCase(a, b ObservableValue[string]) Binding[bool] {
	return BindCompute[string, string, bool](a, b, strings.EqualFold)
}

// BindStringLength exposes len(v.Value()) as a Binding[int]. Note
// that this counts bytes, not runes — matching Go conventions
// rather than Kotlin's UTF-16 code-unit counting.
func BindStringLength(v ObservableValue[string]) Binding[int] {
	return BindTransform[string, int](v, func(s string) int { return len(s) })
}
