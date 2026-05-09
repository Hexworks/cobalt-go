package databinding

// BindBoolNot exposes !v.Value() as a binding. Kotlin used
// `ObservableValue<Boolean>.bindNot()`; Go has no extension methods
// so the receiver becomes a regular parameter.
func BindBoolNot(v ObservableValue[bool]) Binding[bool] {
	return BindTransform[bool, bool](v, func(b bool) bool { return !b })
}

// BindBoolAnd exposes a && b.
func BindBoolAnd(a, b ObservableValue[bool]) Binding[bool] {
	return BindCompute[bool, bool, bool](a, b, func(x, y bool) bool { return x && y })
}

// BindBoolOr exposes a || b.
func BindBoolOr(a, b ObservableValue[bool]) Binding[bool] {
	return BindCompute[bool, bool, bool](a, b, func(x, y bool) bool { return x || y })
}

// BindBoolXor exposes a XOR b.
func BindBoolXor(a, b ObservableValue[bool]) Binding[bool] {
	return BindCompute[bool, bool, bool](a, b, func(x, y bool) bool { return x != y })
}
