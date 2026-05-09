package databinding

import (
	"strconv"
	"testing"

	"github.com/hexworks/cobalt-go/core"
	"github.com/stretchr/testify/assert"
)

const (
	cdbTestONE = "1"
	cdbTestTWO = "2"
)

func newCdbTarget(stringValue Property[string], intValue Property[int]) Binding[bool] {
	return BindCompute[string, int, bool](stringValue, intValue, func(s string, i int) bool {
		n, err := strconv.Atoi(s)
		if err != nil {
			return false
		}
		return n == i
	})
}

func TestComputedDualBinding_InitialValueComputed(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)

	target := newCdbTarget(stringValue, intValue)

	assert.Equal(t, true, target.Value())
}

func TestComputedDualBinding_StringMismatchesIntBindingFalse(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	stringValue.SetValue(cdbTestTWO)

	assert.Equal(t, false, target.Value())
}

func TestComputedDualBinding_IntMismatchesStringBindingFalse(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	intValue.SetValue(2)

	assert.Equal(t, false, target.Value())
}

func TestComputedDualBinding_IntBackToEqualBindingTrue(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	intValue.SetValue(2)
	intValue.SetValue(1)

	assert.Equal(t, true, target.Value())
}

func TestComputedDualBinding_DisposedAccessPanics(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	target.Dispose(core.DisposedManually)

	assert.Panics(t, func() {
		_ = target.Value()
	})
}

func TestComputedDualBinding_IntChangeNotifiesSubscribers(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	notified := false
	target.OnChange(func(_ ObservableValueChanged[bool]) {
		notified = true
	})

	intValue.SetValue(4)

	assert.True(t, notified, "No notification happened.")
}

func TestComputedDualBinding_StringChangeNotifiesSubscribers(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	notified := false
	target.OnChange(func(_ ObservableValueChanged[bool]) {
		notified = true
	})

	stringValue.SetValue(cdbTestTWO)

	assert.True(t, notified, "No notification happened.")
}

func TestComputedDualBinding_NoChangeNoNotification(t *testing.T) {
	stringValue := ToProperty(cdbTestONE)
	intValue := ToProperty(1)
	target := newCdbTarget(stringValue, intValue)

	notified := false
	target.OnChange(func(_ ObservableValueChanged[bool]) {
		notified = true
	})

	stringValue.SetValue(cdbTestONE)

	assert.False(t, notified, "No notification should have happened.")
}
