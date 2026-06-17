package events

import "github.com/hexworks/cobalt-go/core"

// Subscription is the handle returned by EventBus.Subscribe. Disposing
// it removes the underlying callback from the bus.
// NOTE: Subscription is a sealed interface, it can't be ipmlemented
// outside of the `events` package
type Subscription interface {
	core.Disposable
	sealed()
}

// CallbackResult is returned by a subscriber callback to tell the bus
// whether to keep or drop the subscription after the call. Sealed:
// only KeepSubscription and DisposeSubscription implement it.
// NOTE: CallbackResult is a sealed interface, it can't be ipmlemented
// outside of the `events` package
type CallbackResult interface {
	sealed()
}

type keepSubscription struct{}

func (keepSubscription) sealed() {}

type disposeSubscription struct{}

func (disposeSubscription) sealed() {}

// KeepSubscription signals that the subscription should remain active
// after the callback returns.
var KeepSubscription CallbackResult = keepSubscription{}

// DisposeSubscription signals that the subscription should be removed
// from the bus after the callback returns.
var DisposeSubscription CallbackResult = disposeSubscription{}
