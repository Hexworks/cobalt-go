package databinding

// BindTransform exposes ObservableValue[S] as a Binding[T] by piping
// every value through transformer. Equivalent to the Kotlin
// `ObservableValue<S>.bindTransform(transformer)` extension; lives
// as a free function because Go has no extension methods.
func BindTransform[S, T any](source ObservableValue[S], transformer func(S) T) Binding[T] {
	return NewComputedBinding[S, T](source, transformer)
}

// BindCompute creates a binding whose value is computer(source0,
// source1) and which refreshes whenever either source changes.
// Free-function counterpart of the internal ComputedDualBinding
// constructor.
func BindCompute[S0, S1, T any](
	source0 ObservableValue[S0],
	source1 ObservableValue[S1],
	computer func(S0, S1) T,
) Binding[T] {
	return newComputedDualBinding[S0, S1, T](source0, source1, computer)
}

// BindWithConverter creates a bidirectional binding between target
// and other, converting in both directions. Cross-type analogue of
// Property.Bind, lifted out because Go forbids generic methods on
// interfaces.
func BindWithConverter[S, T any](
	target Property[T],
	other Property[S],
	action BindingAction,
	converter IsomorphicConverter[S, T],
) Binding[T] {
	if any(target) == any(other) {
		panic("Can't bind a property to itself.")
	}
	intTarget, ok := target.(internalProperty[T])
	if !ok {
		panic("Can only bind Properties which implement internalProperty.")
	}
	intOther, ok := other.(internalProperty[S])
	if !ok {
		panic("Can only bind Properties which implement internalProperty.")
	}
	if action == UpdateOnBind {
		intTarget.SetValue(converter.Convert(other.Value()))
	}
	return newBidirectionalBinding[S, T](intOther, intTarget, converter)
}

// UpdateFromConverter creates a unidirectional binding driving
// target from observable through a plain conversion function.
// Cross-type analogue of WritableValue.UpdateFrom.
func UpdateFromConverter[S, T any](
	target WritableValue[T],
	observable ObservableValue[S],
	action BindingAction,
	converter func(S) T,
) Binding[T] {
	intTarget, ok := target.(internalProperty[T])
	if !ok {
		panic("Can only update properties that implement internalProperty.")
	}
	if any(target) == any(observable) {
		panic("Can't bind a property to itself.")
	}
	if action == UpdateOnBind {
		intTarget.SetValue(converter(observable.Value()))
	}
	return newUnidirectionalBinding[S, T](observable, intTarget, ConverterFromFunc(converter))
}
