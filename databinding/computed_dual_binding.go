package databinding

import (
	"reflect"
)

// computedDualBinding is a Binding[T] computed from two
// ObservableValues. Whenever either source changes, computerFn is
// reapplied and — if the result differs — pushed into the binding's
// internal target property and republished under the binding's own
// scope.
//
// Unexported because the Kotlin original is `internal class`;
// callers reach it via the (still-to-be-ported) GenericBindings
// helpers.
type computedDualBinding[S0, S1, T any] struct {
	baseBinding[T]
	source0   ObservableValue[S0]
	source1   ObservableValue[S1]
	computer  func(S0, S1) T
}

func newComputedDualBinding[S0, S1, T any](
	source0 ObservableValue[S0],
	source1 ObservableValue[S1],
	computer func(S0, S1) T,
) *computedDualBinding[S0, S1, T] {
	target := newDefaultProperty[T](computer(source0.Value(), source1.Value()))
	b := &computedDualBinding[S0, S1, T]{
		baseBinding: newBaseBinding[T]("ComputedDualBinding", target),
		source0:     source0,
		source1:     source1,
		computer:    computer,
	}
	sub0 := source0.OnChange(func(event ObservableValueChanged[S0]) {
		runWithDisposeOnFailure(b, func() {
			oldValue := computer(event.OldValue, source1.Value())
			newValue := computer(event.NewValue, source1.Value())
			if !reflect.DeepEqual(oldValue, newValue) {
				target.SetValue(newValue)
				b.publishChange(oldValue, newValue, event, event.ChangeType(), b)
			}
		})
	})
	sub1 := source1.OnChange(func(event ObservableValueChanged[S1]) {
		runWithDisposeOnFailure(b, func() {
			oldValue := computer(source0.Value(), event.OldValue)
			newValue := computer(source0.Value(), event.NewValue)
			if !reflect.DeepEqual(oldValue, newValue) {
				target.SetValue(newValue)
				b.publishChange(oldValue, newValue, event, event.ChangeType(), b)
			}
		})
	})
	b.subscriptions = append(b.subscriptions, sub0, sub1)
	return b
}
