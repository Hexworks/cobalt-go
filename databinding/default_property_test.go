package databinding

import (
	"testing"

	"github.com/hexworks/cobalt-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPropertyTestXUL = "XUL"
	defaultPropertyTestQUX = "QUX"
	defaultPropertyTestBAZ = "BAZ"
	defaultPropertyTestONE = "1"
	defaultPropertyTestTWO = "2"
)

func TestDefaultProperty_TargetChangeNotifiesListenerWithProperEvent(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)

	var change *ObservableValueChanged[string]
	target.OnChange(func(e ObservableValueChanged[string]) {
		c := e
		change = &c
	})

	target.SetValue(defaultPropertyTestQUX)

	require.NotNil(t, change, "No change happened.")
	expected := NewObservableValueChanged[string](
		defaultPropertyTestXUL, defaultPropertyTestQUX,
		target, ScalarChange, target, nil,
	)
	assert.Equal(t, expected, *change)
}

func TestDefaultProperty_BoundPropertyTakesOtherValue(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, defaultPropertyTestQUX, target.Value())
}

func TestDefaultProperty_BoundPropertyUpdatesWhenOtherChanges(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	target.Bind(other, UpdateOnBind)

	other.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestBAZ, target.Value())
}

func TestDefaultProperty_AllBoundPropertyValuesUpdate(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)
	bound := ToProperty(defaultPropertyTestQUX)

	target.Bind(other, UpdateOnBind)
	bound.Bind(other, UpdateOnBind)

	other.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestBAZ, target.Value())
	assert.Equal(t, defaultPropertyTestBAZ, bound.Value())
}

func TestDefaultProperty_CircularBindingNoStackOverflow(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other0 := ToProperty(defaultPropertyTestQUX)
	other1 := ToProperty(defaultPropertyTestBAZ)

	target.Bind(other0, UpdateOnBind)
	other0.Bind(other1, UpdateOnBind)
	other1.Bind(target, UpdateOnBind)

	assert.Equal(t, defaultPropertyTestBAZ, target.Value())
	assert.Equal(t, defaultPropertyTestBAZ, other0.Value())
	assert.Equal(t, defaultPropertyTestBAZ, other1.Value())
}

func TestDefaultProperty_SettingValueInCircularBindingNoDeadlock(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other0 := ToProperty(defaultPropertyTestQUX)
	other1 := ToProperty(defaultPropertyTestBAZ)

	target.Bind(other0, UpdateOnBind)
	other0.Bind(other1, UpdateOnBind)
	other1.Bind(target, UpdateOnBind)
	target.SetValue(defaultPropertyTestXUL)

	assert.Equal(t, defaultPropertyTestXUL, target.Value())
	assert.Equal(t, defaultPropertyTestXUL, other0.Value())
	assert.Equal(t, defaultPropertyTestXUL, other1.Value())
}

func TestDefaultProperty_BidirectionalBindTargetUpdated(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, defaultPropertyTestQUX, target.Value())
}

func TestDefaultProperty_BidirectionalBindTargetChangePropagatesToOther(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	target.Bind(other, UpdateOnBind)
	target.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestBAZ, other.Value())
}

func TestDefaultProperty_BidirectionalBindOtherChangePropagatesToTarget(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	target.Bind(other, UpdateOnBind)
	other.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestBAZ, target.Value())
}

func TestDefaultProperty_BidirectionalBindBindingHasSameValueAsTarget(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	binding := target.Bind(other, UpdateOnBind)
	other.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestBAZ, binding.Value())
}

func TestDefaultProperty_DisposedBindingTargetDoesntUpdateOnOtherChange(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	binding := target.Bind(other, UpdateOnBind)
	binding.Dispose(core.DisposedManually)

	other.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestQUX, target.Value())
}

func TestDefaultProperty_DisposedBindingOtherDoesntUpdateOnTargetChange(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(defaultPropertyTestQUX)

	binding := target.Bind(other, UpdateOnBind)
	binding.Dispose(core.DisposedManually)

	target.SetValue(defaultPropertyTestBAZ)

	assert.Equal(t, defaultPropertyTestQUX, other.Value())
}

func TestDefaultProperty_BoundWithConverterTargetUpdated(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(1)

	UpdateFromConverter[int, string](target, other, UpdateOnBind, func(i int) string {
		return defaultPropertyTestONE
	})

	assert.Equal(t, defaultPropertyTestONE, target.Value())
}

func TestDefaultProperty_BoundWithConverterTargetUpdatesWhenOtherChanges(t *testing.T) {
	target := ToProperty(defaultPropertyTestXUL)
	other := ToProperty(1)

	UpdateFromConverter[int, string](target, other, UpdateOnBind, func(i int) string {
		switch i {
		case 1:
			return defaultPropertyTestONE
		case 2:
			return defaultPropertyTestTWO
		}
		return ""
	})

	other.SetValue(2)

	assert.Equal(t, defaultPropertyTestTWO, target.Value())
}
