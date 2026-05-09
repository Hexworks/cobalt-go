package databinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLongExpressions_NegatedBindingValueIsNegative(t *testing.T) {
	prop := ToProperty[int64](2)

	binding := BindNegate[int64](prop)

	assert.Equal(t, int64(-2), binding.Value())
}

func TestLongExpressions_AdditionBindingValueIsSum(t *testing.T) {
	prop := ToProperty[int64](2)
	other := ToProperty[int64](4)

	sum := BindPlus[int64](prop, other)

	assert.Equal(t, int64(6), sum.Value())
}

func TestLongExpressions_SumBindingTracksPropertyChange(t *testing.T) {
	prop := ToProperty[int64](2)
	other := ToProperty[int64](4)

	sum := BindPlus[int64](prop, other)

	prop.SetValue(5)

	assert.Equal(t, int64(9), sum.Value())
}

func TestLongExpressions_ComplexBindingTracksPropertyChange(t *testing.T) {
	sumPart0 := ToProperty[int64](2)
	sumPart1 := ToProperty[int64](4)

	diffPart0 := ToProperty[int64](8)
	diffPart1 := ToProperty[int64](2)

	sum := BindPlus[int64](sumPart0, sumPart1)
	diff := BindMinus[int64](diffPart0, diffPart1)

	complex := BindTimes[int64](sum, diff)

	assert.Equal(t, int64(36), complex.Value())

	diffPart1.SetValue(7)

	assert.Equal(t, int64(6), complex.Value())
}
