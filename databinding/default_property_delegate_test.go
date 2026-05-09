package databinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	delegateTestXUL = "XUL"
	delegateTestQUX = "QUX"
)

func TestDefaultPropertyDelegate_DelegateHasSamePropertyValue(t *testing.T) {
	property := ToProperty(delegateTestXUL)

	target := property.AsDelegate()

	assert.Equal(t, delegateTestXUL, target.Value())
}

func TestDefaultPropertyDelegate_DelegateUpdatedWhenPropertyUpdates(t *testing.T) {
	property := ToProperty(delegateTestXUL)

	target := property.AsDelegate()
	property.SetValue(delegateTestQUX)

	assert.Equal(t, delegateTestQUX, target.Value())
}
