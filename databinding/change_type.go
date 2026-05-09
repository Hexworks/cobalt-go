package databinding

// ChangeType describes what mutation produced an
// ObservableValueChanged event. Sealed: only the variants declared in
// this file (ScalarChange, the ListChange family, the SetChange
// family, the MapChange family) satisfy it.
type ChangeType interface {
	isChangeType()
}

// ListChange marks the list-mutation variants of ChangeType. Stored
// on the Type field of ObservableValueChanged when the source is an
// observable list.
type ListChange interface {
	ChangeType
	isListChange()
}

// SetChange marks the set-mutation variants of ChangeType.
type SetChange interface {
	ChangeType
	isSetChange()
}

// MapChange marks the map-mutation variants of ChangeType.
type MapChange interface {
	ChangeType
	isMapChange()
}

type scalarChange struct{}

func (scalarChange) isChangeType()   {}
func (scalarChange) String() string  { return "ChangeType.ScalarChange" }

// ScalarChange is the singleton ChangeType used for every plain
// scalar mutation (i.e. the result of WritableValue.SetValue or
// UpdateValue on a non-collection property).
var ScalarChange ChangeType = scalarChange{}

// --- ListChange variants ---------------------------------------------------

// ListAdd records that Element was appended to the list.
type ListAdd[T any] struct{ Element T }

func (ListAdd[T]) isChangeType() {}
func (ListAdd[T]) isListChange() {}

// ListAddAt records that Element was inserted at Index.
type ListAddAt[T any] struct {
	Index   int
	Element T
}

func (ListAddAt[T]) isChangeType() {}
func (ListAddAt[T]) isListChange() {}

// ListRemove records that Element was removed.
type ListRemove[T any] struct{ Element T }

func (ListRemove[T]) isChangeType() {}
func (ListRemove[T]) isListChange() {}

// ListRemoveAt records that the element at Index was removed. The
// removed value type isn't part of the variant in the Kotlin
// original, so it stays parameter-free here too.
type ListRemoveAt struct{ Index int }

func (ListRemoveAt) isChangeType() {}
func (ListRemoveAt) isListChange() {}

// ListSet records that the slot at Index was overwritten with
// Element.
type ListSet[T any] struct {
	Index   int
	Element T
}

func (ListSet[T]) isChangeType() {}
func (ListSet[T]) isListChange() {}

// ListAddAll records that Elements were appended.
type ListAddAll[T any] struct{ Elements []T }

func (ListAddAll[T]) isChangeType() {}
func (ListAddAll[T]) isListChange() {}

// ListAddAllAt records that C was inserted starting at Index.
type ListAddAllAt[T any] struct {
	Index int
	C     []T
}

func (ListAddAllAt[T]) isChangeType() {}
func (ListAddAllAt[T]) isListChange() {}

// ListRemoveAll records that every element in Elements was removed.
type ListRemoveAll[T any] struct{ Elements []T }

func (ListRemoveAll[T]) isChangeType() {}
func (ListRemoveAll[T]) isListChange() {}

// ListRemoveAllWhen records that every element matching Predicate
// was removed.
type ListRemoveAllWhen[T any] struct{ Predicate func(T) bool }

func (ListRemoveAllWhen[T]) isChangeType() {}
func (ListRemoveAllWhen[T]) isListChange() {}

// ListRetainAll records that the list was filtered down to Elements.
type ListRetainAll[T any] struct{ Elements []T }

func (ListRetainAll[T]) isChangeType() {}
func (ListRetainAll[T]) isListChange() {}

// ListPropertyChange wraps a change event from a property contained
// in the list. The inner event is type-erased because the outer list
// doesn't know its element-property's value type.
type ListPropertyChange struct {
	ChangeEvent AnyObservableValueChanged
}

func (ListPropertyChange) isChangeType() {}
func (ListPropertyChange) isListChange() {}

type listClear struct{}

func (listClear) isChangeType()  {}
func (listClear) isListChange()  {}
func (listClear) String() string { return "ChangeType.ListChange.ListClear" }

// ListClear is the singleton variant for clearing a list.
var ListClear ListChange = listClear{}

// --- SetChange variants ----------------------------------------------------

// SetAdd records that Element was added to the set.
type SetAdd[T any] struct{ Element T }

func (SetAdd[T]) isChangeType() {}
func (SetAdd[T]) isSetChange()  {}

// SetRemove records that Element was removed from the set.
type SetRemove[T any] struct{ Element T }

func (SetRemove[T]) isChangeType() {}
func (SetRemove[T]) isSetChange()  {}

// SetAddAll records that every element of Elements was added.
type SetAddAll[T any] struct{ Elements []T }

func (SetAddAll[T]) isChangeType() {}
func (SetAddAll[T]) isSetChange()  {}

// SetRemoveAll records that every element of Elements was removed.
type SetRemoveAll[T any] struct{ Elements []T }

func (SetRemoveAll[T]) isChangeType() {}
func (SetRemoveAll[T]) isSetChange()  {}

// SetRemoveAllWhen records that every element matching Predicate was
// removed.
type SetRemoveAllWhen[T any] struct{ Predicate func(T) bool }

func (SetRemoveAllWhen[T]) isChangeType() {}
func (SetRemoveAllWhen[T]) isSetChange()  {}

// SetRetainAll records that the set was filtered down to Elements.
type SetRetainAll[T any] struct{ Elements []T }

func (SetRetainAll[T]) isChangeType() {}
func (SetRetainAll[T]) isSetChange()  {}

// SetPropertyChange wraps a change event from a property contained
// in the set.
type SetPropertyChange struct {
	ChangeEvent AnyObservableValueChanged
}

func (SetPropertyChange) isChangeType() {}
func (SetPropertyChange) isSetChange()  {}

type setClear struct{}

func (setClear) isChangeType()  {}
func (setClear) isSetChange()   {}
func (setClear) String() string { return "ChangeType.SetChange.SetClear" }

// SetClear is the singleton variant for clearing a set.
var SetClear SetChange = setClear{}

// --- MapChange variants ----------------------------------------------------

// MapPut records that (Key, Value) was inserted or replaced. The
// Kotlin signature constrained K to non-null Any; in Go we require
// `comparable` because the underlying storage uses it as a map key.
type MapPut[K comparable, V any] struct {
	Key   K
	Value V
}

func (MapPut[K, V]) isChangeType() {}
func (MapPut[K, V]) isMapChange()  {}

// MapPutAll records a bulk insert.
type MapPutAll[K comparable, V any] struct{ M map[K]V }

func (MapPutAll[K, V]) isChangeType() {}
func (MapPutAll[K, V]) isMapChange()  {}

// MapRemove records that the entry for Key was removed.
type MapRemove[K comparable] struct{ Key K }

func (MapRemove[K]) isChangeType() {}
func (MapRemove[K]) isMapChange()  {}

// MapRemoveWithValue records the removal of (Key, Value).
type MapRemoveWithValue[K comparable, V any] struct {
	Key   K
	Value V
}

func (MapRemoveWithValue[K, V]) isChangeType() {}
func (MapRemoveWithValue[K, V]) isMapChange()  {}

// MapPropertyChange wraps a change event from a property contained
// in the map.
type MapPropertyChange struct {
	ChangeEvent AnyObservableValueChanged
}

func (MapPropertyChange) isChangeType() {}
func (MapPropertyChange) isMapChange()  {}

type mapClear struct{}

func (mapClear) isChangeType()  {}
func (mapClear) isMapChange()   {}
func (mapClear) String() string { return "ChangeType.MapChange.MapClear" }

// MapClear is the singleton variant for clearing a map.
var MapClear MapChange = mapClear{}
