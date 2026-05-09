package databinding

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dspNumbers1To3 = []int{1, 2, 3}
	dspNumbers1To4 = []int{1, 2, 3, 4}
)

func sortedInts(xs []int) []int {
	out := append([]int{}, xs...)
	sort.Ints(out)
	return out
}

func TestDefaultSetProperty_AddChangesValue(t *testing.T) {
	target := ToSetProperty(dspNumbers1To3)

	target.Add(4)

	assert.Equal(t, dspNumbers1To4, sortedInts(target.Value()))
}

func TestDefaultSetProperty_ChangeEmittedAsEvent(t *testing.T) {
	target := ToSetProperty(dspNumbers1To3)
	var newValue []int
	target.OnChange(func(e ObservableValueChanged[[]int]) {
		newValue = e.NewValue
	})

	target.Add(4)

	assert.Equal(t, dspNumbers1To4, sortedInts(newValue))
}

func TestDefaultSetProperty_BoundTakesOtherValue(t *testing.T) {
	target := ToSetProperty(dspNumbers1To3)
	other := ToSetProperty(dspNumbers1To4)

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, dspNumbers1To4, sortedInts(target.Value()))
}
