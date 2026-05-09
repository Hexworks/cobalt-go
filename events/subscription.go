package events

import "github.com/hexworks/cobalt-go/core"

// Subscription is the handle returned by EventBus.Subscribe. Disposing
// it removes the underlying callback from the bus.
type Subscription interface {
	core.Disposable
	isSubscription()
}

// CallbackResult is returned by a subscriber callback to tell the bus
// whether to keep or drop the subscription after the call. Sealed:
// only KeepSubscription and DisposeSubscription implement it.
type CallbackResult interface {
	isCallbackResult()
}

type keepSubscription struct{}

func (keepSubscription) isCallbackResult() {}

type disposeSubscription struct{}

func (disposeSubscription) isCallbackResult() {}

// KeepSubscription signals that the subscription should remain active
// after the callback returns.
var KeepSubscription CallbackResult = keepSubscription{}

// DisposeSubscription signals that the subscription should be removed
// from the bus after the callback returns.
var DisposeSubscription CallbackResult = disposeSubscription{}
