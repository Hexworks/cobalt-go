package databinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dmpNumbers1To3 = map[int]int{1: 1, 2: 2, 3: 3}
	dmpNumbers1To4 = map[int]int{1: 1, 2: 2, 3: 3, 4: 4}
)

func TestDefaultMapProperty_PutChangesValue(t *testing.T) {
	target := ToMapProperty(dmpNumbers1To3)

	target.Put(4, 4)

	assert.Equal(t, dmpNumbers1To4, target.Value())
}

func TestDefaultMapProperty_ChangeEmittedAsEvent(t *testing.T) {
	target := ToMapProperty(dmpNumbers1To3)
	var newValue map[int]int
	target.OnChange(func(e ObservableValueChanged[map[int]int]) {
		newValue = e.NewValue
	})

	target.Put(4, 4)

	assert.Equal(t, dmpNumbers1To4, newValue)
}

func TestDefaultMapProperty_BoundTakesOtherValue(t *testing.T) {
	target := ToMapProperty(dmpNumbers1To3)
	other := ToMapProperty(dmpNumbers1To4)

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, dmpNumbers1To4, target.Value())
}
