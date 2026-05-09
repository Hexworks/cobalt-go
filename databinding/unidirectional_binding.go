package databinding

// unidirectionalBinding propagates changes from source to target
// only. Whenever source emits a change the binding pushes the
// converted value into target via updateWithEvent — which means the
// target sees the change as if it had been driven by an external
// event, complete with the source's trace for cycle detection.
type unidirectionalBinding[S, T any] struct {
	baseBinding[T]
	source    ObservableValue[S]
	converter Converter[S, T]
}

// newUnidirectionalBinding wires source.OnChange to target, then
// returns the live binding. The subscription is stored on the
// embedded baseBinding so Dispose cancels it.
func newUnidirectionalBinding[S, T any](
	source ObservableValue[S],
	target internalProperty[T],
	converter Converter[S, T],
) *unidirectionalBinding[S, T] {
	b := &unidirectionalBinding[S, T]{
		baseBinding: newBaseBinding[T]("UnidirectionalBinding", target),
		source:      source,
		converter:   converter,
	}
	sub := source.OnChange(func(event ObservableValueChanged[S]) {
		runWithDisposeOnFailure(b, func() {
			oldValue := converter.Convert(event.OldValue)
			newValue := converter.Convert(event.NewValue)
			if target.updateWithEvent(oldValue, newValue, event) {
				b.publishChange(oldValue, newValue, event, event.ChangeType(), b)
			}
		})
	})
	b.subscriptions = append(b.subscriptions, sub)
	return b
}
