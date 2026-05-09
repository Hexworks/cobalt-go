package databinding

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dplpInitial() []Property[int] {
	return []Property[int]{ToProperty(1), ToProperty(2), ToProperty(3)}
}

var dplpNumbers1To4 = []int{1, 2, 3, 4}

func plpPropValues(ps []Property[int]) []int {
	out := make([]int, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Value())
	}
	return out
}

func TestDefaultPropertyListProperty_AddChangesValue(t *testing.T) {
	target := ToPropertyListProperty[int, Property[int]](dplpInitial())

	target.Add(ToProperty(4))

	assert.Equal(t, dplpNumbers1To4, plpPropValues(target.Value()))
}

func TestDefaultPropertyListProperty_ChangeEmittedAsEvent(t *testing.T) {
	target := ToPropertyListProperty[int, Property[int]](dplpInitial())
	var newValue []Property[int]
	target.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		newValue = e.NewValue
	})

	target.Add(ToProperty(4))

	assert.Equal(t, dplpNumbers1To4, plpPropValues(newValue))
}

func TestDefaultPropertyListProperty_HierarchicalListBinding(t *testing.T) {
	list0 := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(1)})
	list1 := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(2)})
	list2 := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(3)})

	step := BindListPlus[Property[int]](list0, list1)
	binding := BindListPlus[Property[int]](step, list2)

	assert.Equal(t, []int{1, 2, 3}, plpPropValues(binding.Value()))
}

func TestDefaultPropertyListProperty_HierarchicalListBindingStaysCorrectOnModify(t *testing.T) {
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

func TestDefaultPropertyListProperty_BoundTakesOtherValue(t *testing.T) {
	target := ToPropertyListProperty[int, Property[int]](dplpInitial())
	other := ToPropertyListProperty[int, Property[int]]([]Property[int]{
		ToProperty(1), ToProperty(2), ToProperty(3), ToProperty(4),
	})

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, other.Value(), target.Value())
}

func TestDefaultPropertyListProperty_TransformBindingValueCorrect(t *testing.T) {
	prop := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapList[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	assert.Equal(t, []string{"1", "2"}, binding.Value())
}

func TestDefaultPropertyListProperty_TransformBindingTracksSourceAdds(t *testing.T) {
	prop := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapList[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	prop.Add(ToProperty(3))

	assert.Equal(t, []string{"1", "2", "3"}, binding.Value())
}

func TestDefaultPropertyListProperty_TransformBindingEmitsCorrectEvent(t *testing.T) {
	prop := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapList[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	var listChanges []ObservableValueChanged[[]Property[int]]
	var changes []ObservableValueChanged[[]string]

	prop.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		listChanges = append(listChanges, e)
	})
	binding.OnChange(func(e ObservableValueChanged[[]string]) {
		changes = append(changes, e)
	})

	prop.Add(ToProperty(3))

	require.Equal(t, 1, len(changes))
	change := changes[0]
	expected := NewObservableValueChanged[[]string](
		[]string{"1", "2"},
		[]string{"1", "2", "3"},
		change.ObservableValue,
		listChanges[0].ChangeType(),
		change.Emitter(),
		change.Trace(),
	)
	assert.Equal(t, expected, change)
}

func TestDefaultPropertyListProperty_HierarchicalTransformBindingTracksChanges(t *testing.T) {
	prop := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapList[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	derived := BindListPlus[string](binding, ToListProperty([]string{"foo"}))

	prop.Add(ToProperty(3))

	assert.Equal(t, []string{"1", "2", "3", "foo"}, derived.Value())
}

func TestDefaultPropertyListProperty_MapTransformationOnSetExecutesOnce(t *testing.T) {
	prop := ToPropertyListProperty[int, Property[int]]([]Property[int]{
		ToProperty(0), ToProperty(1), ToProperty(2),
	})

	transformationCount := 0
	BindMapList[Property[int], int](prop, func(p Property[int]) int {
		transformationCount++
		return p.Value() + 1
	})

	transformationCount = 0
	prop.Set(0, ToProperty(5))

	assert.Equal(t, 1, transformationCount)
}

func TestDefaultPropertyListProperty_InnerPropertyChangeFiresProperEvent(t *testing.T) {
	target := ToPropertyListProperty[int, Property[int]](dplpInitial())
	prop := target.Get(0)
	var change *ObservableValueChanged[[]Property[int]]
	var innerChange *ObservableValueChanged[int]
	prop.OnChange(func(e ObservableValueChanged[int]) {
		c := e
		innerChange = &c
	})
	target.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		c := e
		change = &c
	})
	prop.SetValue(0)

	require.NotNil(t, change)
	lpc, ok := change.ChangeType().(ListPropertyChange)
	require.True(t, ok, "Outer change type should be ListPropertyChange")
	require.NotNil(t, innerChange)
	assert.Equal(t, *innerChange, lpc.ChangeEvent)
}

func TestDefaultPropertyListProperty_AddedThenChangedFiresProperEvent(t *testing.T) {
	target := ToPropertyListProperty[int, Property[int]](dplpInitial())
	prop := ToProperty(4)
	var change *ObservableValueChanged[[]Property[int]]
	var innerChange *ObservableValueChanged[int]
	prop.OnChange(func(e ObservableValueChanged[int]) {
		c := e
		innerChange = &c
	})
	target.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		c := e
		change = &c
	})
	target.Add(prop)
	prop.SetValue(0)

	require.NotNil(t, change)
	lpc, ok := change.ChangeType().(ListPropertyChange)
	require.True(t, ok, "Outer change type should be ListPropertyChange")
	require.NotNil(t, innerChange)
	assert.Equal(t, *innerChange, lpc.ChangeEvent)
}

func TestDefaultPropertyListProperty_AddedAndRemovedThenChangedNoEvent(t *testing.T) {
	target := ToPropertyListProperty[int, Property[int]](dplpInitial())
	prop := ToProperty(4)
	target.Add(prop)
	target.Remove(prop)
	var change *ObservableValueChanged[[]Property[int]]
	target.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		c := e
		change = &c
	})
	prop.SetValue(0)

	assert.Nil(t, change)
}
