package databinding

import (
	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// DisposeWhen disposes d as soon as condition becomes true. If
// condition is already true the disposal happens immediately. The
// subscription installed on condition disposes itself once it fires
// so the helper never leaks beyond a single trigger.
//
// Ported from the Kotlin Disposable.disposeWhen infix extension. It
// lives in the databinding package (not core) to keep core free of
// any ObservableValue dependency — the original sits in core because
// Kotlin allows the cycle; Go forbids it.
func DisposeWhen(d core.Disposable, condition ObservableValue[bool]) {
	if condition.Value() {
		d.Dispose(core.DisposedManually)
		return
	}
	var sub events.Subscription
	sub = condition.OnChange(func(ovc ObservableValueChanged[bool]) {
		if ovc.NewValue {
			if sub != nil {
				sub.Dispose(core.DisposedManually)
			}
			d.Dispose(core.DisposedManually)
		}
	})
}

// KeepWhile keeps d alive as long as condition stays true. If
// condition is already false the disposal happens immediately;
// otherwise a subscription waits for the first false transition,
// disposes d and tears itself down.
//
// Ported from the Kotlin Disposable.keepWhile infix extension. Same
// cycle reasoning as DisposeWhen.
func KeepWhile(d core.Disposable, condition ObservableValue[bool]) {
	if !condition.Value() {
		d.Dispose(core.DisposedManually)
		return
	}
	var sub events.Subscription
	sub = condition.OnChange(func(ovc ObservableValueChanged[bool]) {
		if !ovc.NewValue {
			if sub != nil {
				sub.Dispose(core.DisposedManually)
			}
			d.Dispose(core.DisposedManually)
		}
	})
}
