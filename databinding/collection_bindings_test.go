package databinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	cbtNumbers1To3 = []int{1, 2, 3}
	cbtNumbers4To6 = []int{4, 5, 6}
	cbtNumbers7To9 = []int{7, 8, 9}
)

func TestCollectionBindings_ListPropertiesFlattened(t *testing.T) {
	a := ToListProperty([]int{1, 2, 3})
	b := ToListProperty([]int{4, 5, 6})

	fm := ToListProperty[ObservableValue[[]int]](nil)

	result := BindListFlatten[int](fm)

	fm.Add(a)
	fm.Add(b)

	expected := append(append([]int{}, a.Value()...), b.Value()...)
	assert.Equal(t, expected, result.Value())
}

func TestCollectionBindings_ListPropertiesFlattenedAfterRemove(t *testing.T) {
	a := ToListProperty([]int{1, 2, 3})
	b := ToListProperty([]int{4, 5, 6})

	fm := ToListProperty[ObservableValue[[]int]](nil)

	result := BindListFlatten[int](fm)

	fm.Add(a)
	fm.Add(b)

	fm.Remove(a)

	assert.Equal(t, b.Value(), result.Value())
}

func TestCollectionBindings_PlusGivesConcatenatedList(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop4To6 := ToListProperty(cbtNumbers4To6)

	binding := BindListPlus[int](prop1To3, prop4To6)

	assert.Equal(t, append(append([]int{}, cbtNumbers1To3...), cbtNumbers4To6...), binding.Value())
}

func TestCollectionBindings_PlusWithDerivedBinding(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop4To6 := ToListProperty(cbtNumbers4To6)
	prop7To9 := ToListProperty(cbtNumbers7To9)

	binding := BindListPlus[int](prop1To3, prop4To6)
	derived := BindListPlus[int](binding, prop7To9)

	expected := append(append(append([]int{}, cbtNumbers1To3...), cbtNumbers4To6...), cbtNumbers7To9...)
	assert.Equal(t, expected, derived.Value())
}

func TestCollectionBindings_MinusGivesDifference(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop4To6 := ToListProperty(cbtNumbers4To6)

	binding1To6 := BindListPlus[int](prop1To3, prop4To6)
	subtracted := BindListMinus[int](binding1To6, prop4To6)

	assert.Equal(t, cbtNumbers1To3, subtracted.Value())
}

func TestCollectionBindings_MinusWithDerivedBinding(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop4To6 := ToListProperty(cbtNumbers4To6)
	prop7To9 := ToListProperty(cbtNumbers7To9)

	binding := BindListPlus[int](prop1To3, prop4To6)
	derived := BindListPlus[int](binding, prop7To9)
	subtracted := BindListMinus[int](derived, prop7To9)

	expected := append(append([]int{}, cbtNumbers1To3...), cbtNumbers4To6...)
	assert.Equal(t, expected, subtracted.Value())
}

func TestCollectionBindings_EqualPropertiesIsEqualBindingTrue(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)

	binding := BindListIsEqualTo[int](prop1To3, prop1To3)

	assert.Equal(t, true, binding.Value())
}

func TestCollectionBindings_UnequalPropertiesIsEqualBindingFalse(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop4To6 := ToListProperty(cbtNumbers4To6)

	binding := BindListIsEqualTo[int](prop1To3, prop4To6)

	assert.Equal(t, false, binding.Value())
}

func TestCollectionBindings_SizeBindingCorrect(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)

	binding := BindListSize[int](prop1To3)

	assert.Equal(t, 3, binding.Value())
}

func TestCollectionBindings_SizeBindingUpdatesOnAdd(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)

	binding := BindListSize[int](prop1To3)

	prop1To3.Add(4)

	assert.Equal(t, 4, binding.Value())
}

func TestCollectionBindings_EmptyBindingCorrect(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)

	binding := BindListIsEmpty[int](prop1To3)

	assert.Equal(t, false, binding.Value())
}

func TestCollectionBindings_EmptyBindingChangesOnClear(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)

	binding := BindListIsEmpty[int](prop1To3)

	prop1To3.Clear()

	assert.Equal(t, true, binding.Value())
}

func TestCollectionBindings_ContainsBindingCorrect(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop1 := ToProperty(1)

	binding := BindListContains[int](prop1To3, prop1)

	assert.Equal(t, true, binding.Value())
}

func TestCollectionBindings_ContainsBindingChangesOnElementChange(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	prop1 := ToProperty(1)

	binding := BindListContains[int](prop1To3, prop1)

	prop1.SetValue(4)

	assert.Equal(t, false, binding.Value())
}

func TestCollectionBindings_ContainsAllBindingCorrect(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	listProp := ToListProperty([]int{1, 2, 3})

	binding := BindListContainsAll[int](prop1To3, listProp)

	assert.Equal(t, true, binding.Value())
}

func TestCollectionBindings_ContainsAllBindingChangesOnListChange(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	listProp := ToListProperty([]int{1, 2, 3})

	binding := BindListContainsAll[int](prop1To3, listProp)

	listProp.Add(4)

	assert.Equal(t, false, binding.Value())
}

func TestCollectionBindings_IndexOfBindingCorrect(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	numProp := ToProperty(1)

	binding := BindListIndexOf[int](prop1To3, numProp)

	assert.Equal(t, 0, binding.Value())
}

func TestCollectionBindings_IndexOfBindingChangesOnElementChange(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	numProp := ToProperty(1)

	binding := BindListIndexOf[int](prop1To3, numProp)

	numProp.SetValue(2)

	assert.Equal(t, 1, binding.Value())
}

func TestCollectionBindings_LastIndexOfBindingCorrect(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	numProp := ToProperty(1)

	binding := BindListLastIndexOf[int](prop1To3, numProp)

	assert.Equal(t, 0, binding.Value())
}

func TestCollectionBindings_LastIndexOfBindingChangesOnListChange(t *testing.T) {
	prop1To3 := ToListProperty(cbtNumbers1To3)
	numProp := ToProperty(1)

	binding := BindListLastIndexOf[int](prop1To3, numProp)

	prop1To3.Add(1)

	assert.Equal(t, 3, binding.Value())
}
