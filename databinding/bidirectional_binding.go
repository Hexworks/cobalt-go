package databinding

// bidirectionalBinding keeps source and target in sync. A change on
// either side flows through the converter (or its reverse) to the
// other; cycle detection on each property's updateWithEvent prevents
// the resulting ping-pong from running away.
type bidirectionalBinding[S, T any] struct {
	baseBinding[T]
	source           internalProperty[S]
	converter        IsomorphicConverter[S, T]
	reverseConverter Converter[T, S]
}

func newBidirectionalBinding[S, T any](
	source internalProperty[S],
	target internalProperty[T],
	converter IsomorphicConverter[S, T],
) *bidirectionalBinding[S, T] {
	b := &bidirectionalBinding[S, T]{
		baseBinding:      newBaseBinding[T]("BidirectionalBinding", target),
		source:           source,
		converter:        converter,
		reverseConverter: ReverseConverter[S, T](converter),
	}
	srcSub := source.OnChange(func(event ObservableValueChanged[S]) {
		runWithDisposeOnFailure(b, func() {
			oldValue := converter.Convert(event.OldValue)
			newValue := converter.Convert(event.NewValue)
			if target.updateWithEvent(oldValue, newValue, event) {
				b.publishChange(oldValue, newValue, event, event.ChangeType(), b)
			}
		})
	})
	tgtSub := target.OnChange(func(event ObservableValueChanged[T]) {
		runWithDisposeOnFailure(b, func() {
			oldValue := b.reverseConverter.Convert(event.OldValue)
			newValue := b.reverseConverter.Convert(event.NewValue)
			if source.updateWithEvent(oldValue, newValue, event) {
				// Re-publish through the binding so observers see the
				// converted T-typed change, not the raw S-typed one.
				b.publishChange(event.OldValue, event.NewValue, event, event.ChangeType(), b)
			}
		})
	})
	b.subscriptions = append(b.subscriptions, srcSub, tgtSub)
	return b
}
