package databinding

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dlpNumbers1To3 = []int{1, 2, 3}
	dlpNumbers1To4 = []int{1, 2, 3, 4}
)

func TestDefaultListProperty_AddChangesValue(t *testing.T) {
	target := ToListProperty(dlpNumbers1To3)

	target.Add(4)

	assert.Equal(t, dlpNumbers1To4, target.Value())
}

func TestDefaultListProperty_ChangeEmittedAsEvent(t *testing.T) {
	target := ToListProperty(dlpNumbers1To3)
	var newValue []int

	target.OnChange(func(e ObservableValueChanged[[]int]) {
		newValue = e.NewValue
	})

	target.Add(4)

	assert.Equal(t, dlpNumbers1To4, newValue)
}

func TestDefaultListProperty_HierarchicalListBinding(t *testing.T) {
	list0 := ToListProperty([]int{1})
	list1 := ToListProperty([]int{2})
	list2 := ToListProperty([]int{3})

	step := BindListPlus[int](list0, list1)
	binding := BindListPlus[int](step, list2)

	assert.Equal(t, []int{1, 2, 3}, binding.Value())
}

func TestDefaultListProperty_HierarchicalListBindingStaysCorrectOnModify(t *testing.T) {
	list0 := ToListProperty([]int{1})
	list1 := ToListProperty([]int{3})
	list2 := ToListProperty([]int{5})

	step := BindListPlus[int](list0, list1)
	binding := BindListPlus[int](step, list2)

	list0.Add(2)
	list1.Add(4)
	list2.Add(6)

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, binding.Value())
}

func TestDefaultListProperty_BoundTakesOtherValue(t *testing.T) {
	target := ToListProperty(dlpNumbers1To3)
	other := ToListProperty(dlpNumbers1To4)

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, dlpNumbers1To4, target.Value())
}

func TestDefaultListProperty_TransformBindingValueCorrect(t *testing.T) {
	prop := ToListProperty([]int{1, 2})

	binding := BindMapList[int, string](prop, strconv.Itoa)

	assert.Equal(t, []string{"1", "2"}, binding.Value())
}

func TestDefaultListProperty_TransformBindingTracksSourceAdds(t *testing.T) {
	prop := ToListProperty([]int{1, 2})

	binding := BindMapList[int, string](prop, strconv.Itoa)

	prop.Add(3)

	assert.Equal(t, []string{"1", "2", "3"}, binding.Value())
}

func TestDefaultListProperty_TransformBindingEmitsCorrectEvent(t *testing.T) {
	prop := ToListProperty([]int{1, 2})

	binding := BindMapList[int, string](prop, strconv.Itoa)

	var changes []ObservableValueChanged[[]string]
	binding.OnChange(func(e ObservableValueChanged[[]string]) {
		changes = append(changes, e)
	})

	prop.Add(3)

	require.Equal(t, 1, len(changes))
	change := changes[0]
	expected := NewObservableValueChanged[[]string](
		[]string{"1", "2"},
		[]string{"1", "2", "3"},
		change.ObservableValue,
		ListAdd[int]{Element: 3},
		change.Emitter(),
		change.Trace(),
	)
	assert.Equal(t, expected, change)
}

func TestDefaultListProperty_HierarchicalTransformBindingTracksChanges(t *testing.T) {
	prop := ToListProperty([]int{1, 2})

	binding := BindMapList[int, string](prop, strconv.Itoa)

	derived := BindListPlus[string](binding, ToListProperty([]string{"foo"}))

	prop.Add(3)

	assert.Equal(t, []string{"1", "2", "3", "foo"}, derived.Value())
}

func TestDefaultListProperty_ArrayOfListsReduced(t *testing.T) {
	prop0 := ToListProperty([]int{0})
	prop1 := ToListProperty([]int{1})
	prop2 := ToListProperty([]int{2})
	prop3 := ToListProperty([]int{3})
	prop4 := ToListProperty([]int{4})

	var binding ObservableValue[[]int] = prop0
	for _, p := range []ObservableValue[[]int]{prop1, prop2, prop3, prop4} {
		binding = BindListPlus[int](binding, p)
	}

	assert.Equal(t, []int{0, 1, 2, 3, 4}, binding.Value())
}

func TestDefaultListProperty_ArrayOfListsReducedAfterMutation(t *testing.T) {
	prop0 := ToListProperty([]int{0})
	prop1 := ToListProperty([]int{1})
	prop2 := ToListProperty([]int{2})
	prop3 := ToListProperty([]int{3})
	prop4 := ToListProperty([]int{5})

	var binding ObservableValue[[]int] = prop0
	for _, p := range []ObservableValue[[]int]{prop1, prop2, prop3, prop4} {
		binding = BindListPlus[int](binding, p)
	}

	prop3.Add(4)

	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, binding.Value())
}

func TestDefaultListProperty_MapTransformationOnSetExecutesOnce(t *testing.T) {
	prop := ToListProperty([]int{0, 1, 2})

	transformationCount := 0
	BindMapList[int, int](prop, func(x int) int {
		transformationCount++
		return x + 1
	})

	transformationCount = 0
	prop.Set(0, 5)

	assert.Equal(t, 1, transformationCount)
}
