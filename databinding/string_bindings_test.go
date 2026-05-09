package databinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringExpressions_EmptyPropertyBindingTrue(t *testing.T) {
	prop := ToProperty("")

	binding := BindStringIsEmpty(prop)

	assert.Equal(t, true, binding.Value())
}

func TestStringExpressions_NotEmptyPropertyBindingFalse(t *testing.T) {
	prop := ToProperty("")

	binding := BindStringIsEmpty(prop)

	prop.SetValue("foo")

	assert.Equal(t, false, binding.Value())
}

func TestStringExpressions_NotEmptyNegationBindingFalse(t *testing.T) {
	prop := ToProperty("")

	binding := BindBoolNot(BindStringIsEmpty(prop))

	assert.Equal(t, false, binding.Value())
}

func TestStringExpressions_EmptyNegationBindingTrueAfterUpdate(t *testing.T) {
	prop := ToProperty("")

	binding := BindBoolNot(BindStringIsEmpty(prop))

	prop.SetValue("foo")

	assert.Equal(t, true, binding.Value())
}
