package databinding

import (
	"errors"
	"strconv"
	"testing"

	"github.com/hexworks/cobalt-go/core"
	"github.com/stretchr/testify/assert"
)

const (
	bcbTestFOO = "foo"
	bcbTestONE = "1"
	bcbTestTWO = "2"
)

type intStringConverter struct{}

func (intStringConverter) Convert(source int) string  { return strconv.Itoa(source) }
func (intStringConverter) ConvertBack(target string) int {
	n, err := strconv.Atoi(target)
	if err != nil {
		panic(err)
	}
	return n
}

func bindTargetToOther(target Property[string], other Property[int]) Binding[string] {
	return BindWithConverter[int, string](target, other, UpdateOnBind, intStringConverter{})
}

func TestBidirectionalConverterBinding_CrossTypeBindUpdatesTarget(t *testing.T) {
	target := ToProperty(bcbTestONE)
	other := ToProperty(2)

	bindTargetToOther(target, other)

	assert.Equal(t, bcbTestTWO, target.Value())
}

func TestBidirectionalConverterBinding_OtherUpdateFlowsToTarget(t *testing.T) {
	target := ToProperty(bcbTestONE)
	other := ToProperty(2)

	bindTargetToOther(target, other)

	other.SetValue(1)

	assert.Equal(t, bcbTestONE, target.Value())
}

func TestBidirectionalConverterBinding_TargetUpdateFlowsToOther(t *testing.T) {
	target := ToProperty(bcbTestONE)
	other := ToProperty(2)

	bindTargetToOther(target, other)

	target.SetValue(bcbTestONE)

	assert.Equal(t, 1, other.Value())
}

func TestBidirectionalConverterBinding_TargetUpdateExceptionDisposesBinding(t *testing.T) {
	target := ToProperty(bcbTestONE)
	other := ToProperty(2)

	binding := bindTargetToOther(target, other)

	target.SetValue(bcbTestFOO)

	state := binding.DisposeState()
	disposedByException, ok := state.(core.DisposedByException)
	assert.True(t, ok, "Binding should have been disposed by exception, got %T", state)
	if ok {
		assert.True(t, errors.Is(disposedByException.Err, disposedByException.Err))
	}
}
