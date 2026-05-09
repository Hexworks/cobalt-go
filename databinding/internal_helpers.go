package databinding

import (
	"fmt"
	"log/slog"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
)

// disposeSubscriptions disposes every Subscription in subs. Failures
// are logged but never propagated so one bad listener can't stop the
// rest from being cleaned up. Mirrors the Kotlin
// disposeSubscriptions() extension.
func disposeSubscriptions(subs []events.Subscription) {
	for _, s := range subs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Warn("cancelling subscription failed", "err", r)
				}
			}()
			s.Dispose(core.DisposedManually)
		}()
	}
}

// runWithDisposeOnFailure executes fn and, if it panics, disposes b
// with DisposedByException carrying the recovered error. Every
// binding callback runs through this so a converter panic can't take
// down the event bus's publish loop.
func runWithDisposeOnFailure(b core.Disposable, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			b.Dispose(core.DisposedByException{Err: err})
		}
	}()
	fn()
}
