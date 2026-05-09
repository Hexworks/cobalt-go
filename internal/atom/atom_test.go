package atom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomTransform(t *testing.T) {
	ref := New(1)

	ref.Transform(func(int) int { return 2 })

	assert.Equal(t, 2, ref.Get())
}
