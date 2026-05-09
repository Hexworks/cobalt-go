package events_test

import (
	"errors"
	"testing"

	"github.com/hexworks/cobalt-go/core"
	"github.com/hexworks/cobalt-go/events"
	"github.com/stretchr/testify/assert"
)

type testSource struct {
	id core.UUID
}

func (s testSource) ID() core.UUID { return s.id }

func newTestSource() testSource {
	return testSource{id: core.RandomUUID()}
}

type testEvent struct {
	emitter events.EventSource
	trace   []events.Event
}

func (e testEvent) Key() string                 { return testEventDescriptor.Key }
func (e testEvent) Emitter() events.EventSource { return e.emitter }
func (e testEvent) Trace() []events.Event       { return e.trace }

var testEventDescriptor = events.EventDescriptor[testEvent]{Key: "TestEvent"}

type testScopeT struct{}

// testScope mirrors the Kotlin `object TestScope : EventScope`.
var testScope events.EventScope = testScopeT{}

func newTestScope(_ *testing.T) events.EventScope { return testScope }

func TestSubscriptionCancellationDoesNotAffectOtherSubscriptions(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()

	sub0 := events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {})
	sub1 := events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {})

	sub0.Dispose(core.DisposedManually)

	assert.False(t, core.Disposed(sub1))
	_ = src
}

func TestRemainingSubscriberStillNotifiedAfterPeerDisposed(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()

	sub0Notified := false
	sub1Notified := false

	sub0 := events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		sub0Notified = true
	})
	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		sub1Notified = true
	})

	sub0.Dispose(core.DisposedManually)
	target.Publish(testEvent{emitter: src}, events.ApplicationScope)

	assert.False(t, sub0Notified)
	assert.True(t, sub1Notified)
}

func TestSubscriberNotifiedOnMatchingEvent(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()

	notified := false

	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		notified = true
	})

	target.Publish(testEvent{emitter: src}, events.ApplicationScope)

	assert.True(t, notified, "Subscriber was not notified.")
}

func TestSubscriberNotifiedOnMatchingEventAndScope(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()
	scope := newTestScope(t)

	notified := false
	events.SimpleSubscribeTo(target, testEventDescriptor, scope, func(testEvent) {
		notified = true
	})

	target.Publish(testEvent{emitter: src}, scope)

	assert.True(t, notified, "Subscriber was not notified.")
}

func TestSubscriberNotNotifiedWhenScopeMismatches(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()
	scope := newTestScope(t)

	notified := false
	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		notified = true
	})

	target.Publish(testEvent{emitter: src}, scope)

	assert.False(t, notified, "Subscriber should not have been notified.")
}

func TestSubscribeToWithCallbackResult(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()

	notified := false

	events.SubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) events.CallbackResult {
		notified = true
		return events.KeepSubscription
	})

	target.Publish(testEvent{emitter: src}, events.ApplicationScope)

	assert.True(t, notified, "Subscriber was not notified.")
}

func TestMultipleSubscribersAllNotified(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()

	var notifications []int

	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		notifications = append(notifications, 0)
	})
	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		notifications = append(notifications, 1)
	})

	target.Publish(testEvent{emitter: src}, events.ApplicationScope)

	assert.Equal(t, []int{0, 1}, notifications, "All subscribers should have been notified.")
}

func TestOnlyMatchingScopeNotified(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()
	scope := newTestScope(t)

	var notifications []events.EventScope
	events.SimpleSubscribeTo(target, testEventDescriptor, scope, func(testEvent) {
		notifications = append(notifications, scope)
	})
	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		notifications = append(notifications, events.ApplicationScope)
	})

	target.Publish(testEvent{emitter: src}, scope)

	assert.Equal(t, []events.EventScope{scope}, notifications,
		"Only the subscriber with TestScope should have been notified.")
}

func TestSubscriberPresentAfterSubscribe(t *testing.T) {
	target := events.NewEventBus()

	sub := events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {})

	assert.Equal(t,
		[]events.Subscription{sub},
		target.FetchSubscribers(testEventDescriptor.Key, events.ApplicationScope),
		"Subscribers should contain the registered subscription.")
}

func TestSubscriberAbsentAfterDispose(t *testing.T) {
	target := events.NewEventBus()

	events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {}).
		Dispose(core.DisposedManually)

	assert.Empty(t, target.FetchSubscribers(testEventDescriptor.Key, events.ApplicationScope),
		"Subscribers should be empty.")
}

func TestCallbackPanicCancelsSubscriptionWithException(t *testing.T) {
	target := events.NewEventBus()
	src := newTestSource()

	boom := errors.New("boom")
	expectedState := core.DisposedByException{Err: boom}

	sub := events.SimpleSubscribeTo(target, testEventDescriptor, events.ApplicationScope, func(testEvent) {
		panic(boom)
	})

	target.Publish(testEvent{emitter: src}, events.ApplicationScope)

	assert.Equal(t, expectedState, sub.DisposeState(),
		"Subscriber should have been cancelled with exception")
}
