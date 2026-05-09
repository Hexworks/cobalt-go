package databinding

// WritableCollection is a WritableValue exposing collection-shaped
// writes. Mirrors the Kotlin
// WritableCollection<T, C : PersistentCollection<T>> hierarchy.
// Concrete mutators live on the specialised WritableList /
// WritableSet / WritableMap variants below.
type WritableCollection[T any, C any] interface {
	WritableValue[C]
}

// WritableList is the list-shaped writable variant. Every mutator
// returns the resulting slice so callers can chain or inspect the
// new state without re-reading Value(). The returned slice is owned
// by the property — callers must not mutate it.
type WritableList[T any] interface {
	WritableCollection[T, []T]
	// Add appends element. Returns the new list.
	Add(element T) []T
	// AddAt inserts element at index. Panics on out-of-range index.
	AddAt(index int, element T) []T
	// Set overwrites the slot at index with element. Panics on
	// out-of-range index.
	Set(index int, element T) []T
	// Remove removes the first occurrence of element. Returns the
	// new list, unchanged when element was absent.
	Remove(element T) []T
	// RemoveAt removes the element at index. Panics on out-of-range
	// index.
	RemoveAt(index int) []T
	// AddAll appends every element of elements.
	AddAll(elements []T) []T
	// AddAllAt inserts c starting at index. Panics on out-of-range
	// index.
	AddAllAt(index int, c []T) []T
	// RemoveAll removes every occurrence of every element in
	// elements.
	RemoveAll(elements []T) []T
	// RemoveAllWhere removes every element for which predicate
	// returns true.
	RemoveAllWhere(predicate func(T) bool) []T
	// RetainAll filters the list down to entries that appear in
	// elements.
	RetainAll(elements []T) []T
	// Clear empties the list.
	Clear() []T
}

// WritableSet is the set-shaped writable variant. Identical mutator
// set to WritableList except for the position-based operations,
// which a set does not support.
type WritableSet[T comparable] interface {
	WritableCollection[T, []T]
	// Add inserts element. Returns the new set (unchanged when
	// element was already present).
	Add(element T) []T
	// Remove removes element. Returns the new set (unchanged when
	// element was absent).
	Remove(element T) []T
	// AddAll inserts every element of elements.
	AddAll(elements []T) []T
	// RemoveAll removes every element of elements.
	RemoveAll(elements []T) []T
	// RemoveAllWhere removes every element matching predicate.
	RemoveAllWhere(predicate func(T) bool) []T
	// RetainAll filters the set down to elements that appear in
	// elements.
	RetainAll(elements []T) []T
	// Clear empties the set.
	Clear() []T
}

// WritableMap is the map-shaped writable variant.
type WritableMap[K comparable, V any] interface {
	WritableValue[map[K]V]
	// Put inserts or replaces an entry. Returns the new map.
	Put(key K, value V) map[K]V
	// PutAll merges m into the map.
	PutAll(m map[K]V) map[K]V
	// Remove removes key. Returns the new map (unchanged when key
	// was absent).
	Remove(key K) map[K]V
	// RemoveWithValue removes the (key, value) entry only when the
	// current value under key equals value.
	RemoveWithValue(key K, value V) map[K]V
	// Clear empties the map.
	Clear() map[K]V
}
