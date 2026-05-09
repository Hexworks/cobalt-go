package databinding

// ComputedBinding wraps an ObservableValue[S] with a transformer
// function and exposes the result as a Binding[T]. Whenever the
// source changes the binding recomputes the converted value and
// pushes it into its internal target property.
//
// Exported because the Kotlin original is a public class — callers
// build computed bindings directly when they don't want to use the
// BindTransform helper.
type ComputedBinding[S, T any] struct {
	baseBinding[T]
	source    ObservableValue[S]
	transform func(S) T
}

// NewComputedBinding constructs a ComputedBinding seeded with the
// transformed current value of source. The returned binding starts
// observing source immediately.
func NewComputedBinding[S, T any](source ObservableValue[S], transform func(S) T) *ComputedBinding[S, T] {
	target := newDefaultProperty[T](transform(source.Value()))
	b := &ComputedBinding[S, T]{
		baseBinding: newBaseBinding[T]("ComputedBinding", target),
		source:      source,
		transform:   transform,
	}
	sub := source.OnChange(func(event ObservableValueChanged[S]) {
		runWithDisposeOnFailure(b, func() {
			oldValue := transform(event.OldValue)
			newValue := transform(event.NewValue)
			// Kotlin's ComputedBinding always publishes ScalarChange
			// regardless of the upstream change kind; pass it
			// explicitly through publishChange so the binding's
			// downstream observers don't see misleading collection
			// change variants.
			if target.updateWithEvent(oldValue, newValue, event) {
				b.publishChange(oldValue, newValue, event, ScalarChange, b)
			}
		})
	})
	b.subscriptions = append(b.subscriptions, sub)
	return b
}
