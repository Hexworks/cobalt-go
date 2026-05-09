package databinding

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dpspInitial() []Property[int] {
	return []Property[int]{ToProperty(1), ToProperty(2), ToProperty(3)}
}

var dpspNumbers1To4 = []int{1, 2, 3, 4}

func sortStrings(xs []string) []string {
	out := append([]string{}, xs...)
	sort.Strings(out)
	return out
}

func TestDefaultPropertySetProperty_AddChangesValue(t *testing.T) {
	target := ToPropertySetProperty[int, Property[int]](dpspInitial())

	target.Add(ToProperty(4))

	assert.Equal(t, dpspNumbers1To4, sortedInts(plpPropValues(target.Value())))
}

func TestDefaultPropertySetProperty_ChangeEmittedAsEvent(t *testing.T) {
	target := ToPropertySetProperty[int, Property[int]](dpspInitial())
	var newValue []Property[int]
	target.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		newValue = e.NewValue
	})

	target.Add(ToProperty(4))

	assert.Equal(t, dpspNumbers1To4, sortedInts(plpPropValues(newValue)))
}

func TestDefaultPropertySetProperty_HierarchicalListBinding(t *testing.T) {
	list0 := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(1)})
	list1 := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(2)})
	list2 := ToPropertyListProperty[int, Property[int]]([]Property[int]{ToProperty(3)})

	step := BindListPlus[Property[int]](list0, list1)
	binding := BindListPlus[Property[int]](step, list2)

	assert.Equal(t, []int{1, 2, 3}, plpPropValues(binding.Value()))
}

func TestDefaultPropertySetProperty_HierarchicalSetBindingStaysCorrectOnModify(t *testing.T) {
	set0 := ToSetProperty([]int{1})
	set1 := ToSetProperty([]int{3})
	set2 := ToSetProperty([]int{5})

	step := BindSetPlus[int](set0, set1)
	binding := BindSetPlus[int](step, set2)

	set0.Add(2)
	set1.Add(4)
	set2.Add(6)

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, sortedInts(binding.Value()))
}

func TestDefaultPropertySetProperty_BoundTakesOtherValue(t *testing.T) {
	target := ToPropertySetProperty[int, Property[int]](dpspInitial())
	other := ToPropertySetProperty[int, Property[int]]([]Property[int]{
		ToProperty(1), ToProperty(2), ToProperty(3), ToProperty(4),
	})

	target.Bind(other, UpdateOnBind)

	assert.Equal(t, other.Value(), target.Value())
}

func TestDefaultPropertySetProperty_TransformBindingValueCorrect(t *testing.T) {
	prop := ToPropertySetProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapSet[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	assert.Equal(t, []string{"1", "2"}, sortStrings(binding.Value()))
}

func TestDefaultPropertySetProperty_TransformBindingTracksSourceAdds(t *testing.T) {
	prop := ToPropertySetProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapSet[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	prop.Add(ToProperty(3))

	assert.Equal(t, []string{"1", "2", "3"}, sortStrings(binding.Value()))
}

func TestDefaultPropertySetProperty_TransformBindingEmitsCorrectEvent(t *testing.T) {
	prop := ToPropertySetProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapSet[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	var setChanges []ObservableValueChanged[[]Property[int]]
	var changes []ObservableValueChanged[[]string]

	prop.OnChange(func(e ObservableValueChanged[[]Property[int]]) {
		setChanges = append(setChanges, e)
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
		setChanges[0].ChangeType(),
		change.Emitter(),
		change.Trace(),
	)
	assert.Equal(t, expected, change)
}

func TestDefaultPropertySetProperty_HierarchicalTransformBindingTracksChanges(t *testing.T) {
	prop := ToPropertySetProperty[int, Property[int]]([]Property[int]{ToProperty(1), ToProperty(2)})

	binding := BindMapSet[Property[int], string](prop, func(p Property[int]) string {
		return strconv.Itoa(p.Value())
	})

	derived := BindSetPlus[string](binding, ToSetProperty([]string{"foo"}))

	prop.Add(ToProperty(3))

	assert.Equal(t, []string{"1", "2", "3", "foo"}, sortStrings(derived.Value()))
}

func TestDefaultPropertySetProperty_InnerPropertyChangeFiresProperEvent(t *testing.T) {
	target := ToPropertySetProperty[int, Property[int]](dpspInitial())
	prop := target.Value()[0]
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
	spc, ok := change.ChangeType().(SetPropertyChange)
	require.True(t, ok, "Outer change type should be SetPropertyChange")
	require.NotNil(t, innerChange)
	assert.Equal(t, *innerChange, spc.ChangeEvent)
}

func TestDefaultPropertySetProperty_AddedThenChangedFiresProperEvent(t *testing.T) {
	target := ToPropertySetProperty[int, Property[int]](dpspInitial())
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
	spc, ok := change.ChangeType().(SetPropertyChange)
	require.True(t, ok, "Outer change type should be SetPropertyChange")
	require.NotNil(t, innerChange)
	assert.Equal(t, *innerChange, spc.ChangeEvent)
}

func TestDefaultPropertySetProperty_AddedAndRemovedThenChangedNoEvent(t *testing.T) {
	target := ToPropertySetProperty[int, Property[int]](dpspInitial())
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
