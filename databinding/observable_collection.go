package databinding

// ObservableCollection is an ObservableValue wrapping a collection
// of T-typed elements. Mirrors the Kotlin
// ObservableCollection<T, C : PersistentCollection<T>> hierarchy: T
// is the element type, C the underlying collection representation
// (`[]T` for lists and sets, `map[K]V` for maps). Reading C through
// Value() returns the live underlying view — callers must not mutate
// it. All mutation goes through the writable interfaces.
type ObservableCollection[T any, C any] interface {
	ObservableValue[C]
	// Size returns the number of elements currently stored.
	Size() int
	// IsEmpty reports whether the collection has zero elements.
	IsEmpty() bool
	// Contains reports whether element is present. Equality follows
	// reflect.DeepEqual to match the Kotlin Object.equals semantics
	// for arbitrary T.
	Contains(element T) bool
	// ContainsAll reports whether every element of elements is
	// present in the collection.
	ContainsAll(elements []T) bool
}

// ObservableList is the list-shaped variant of ObservableCollection.
// Adds position-based reads.
type ObservableList[T any] interface {
	ObservableCollection[T, []T]
	// Get returns the element at index. Panics on out-of-range
	// access — same shape as a Kotlin list `get`.
	Get(index int) T
	// IndexOf returns the position of the first element equal to
	// element, or -1 if absent.
	IndexOf(element T) int
	// LastIndexOf returns the position of the last element equal to
	// element, or -1 if absent.
	LastIndexOf(element T) int
	// SubList returns the inclusive-exclusive slice of [fromIndex,
	// toIndex). The returned slice is freshly allocated; callers
	// can mutate it without affecting the property.
	SubList(fromIndex, toIndex int) []T
}

// ObservableSet is the set-shaped variant of ObservableCollection.
// Set semantics (no duplicates) are enforced by the mutator methods;
// the underlying representation is still `[]T` so the same Property
// abstractions can carry it through the binding plumbing.
//
// T is constrained to `comparable` because the set's internal
// deduplication relies on equality, and `comparable` is required to
// build the `map[T]struct{}` lookup table backing it.
type ObservableSet[T comparable] interface {
	ObservableCollection[T, []T]
}

// ObservableMap is the map-shaped variant. Key type K must be
// comparable (Go map-key constraint); V is unconstrained.
type ObservableMap[K comparable, V any] interface {
	ObservableValue[map[K]V]
	// Size returns the number of entries.
	Size() int
	// IsEmpty reports whether the map has zero entries.
	IsEmpty() bool
	// ContainsKey reports whether key has an entry.
	ContainsKey(key K) bool
	// ContainsValue reports whether any entry has the given value.
	ContainsValue(value V) bool
	// Get returns the value stored under key and whether it was
	// present. Matches the (V, bool) idiom from Go maps so callers
	// can distinguish "missing" from "present-but-zero".
	Get(key K) (V, bool)
	// Keys returns a freshly allocated slice of every key. Order is
	// not guaranteed.
	Keys() []K
	// Values returns a freshly allocated slice of every value.
	// Order is not guaranteed.
	Values() []V
}

// ObservableListBinding is a Binding that also exposes the
// ObservableList[T] surface. Stage 4 wraps ListBinding instances in
// a ListBindingDecorator to satisfy this combined interface.
type ObservableListBinding[T any] interface {
	ObservableList[T]
	Binding[[]T]
}

// ObservableSetBinding is the set-shaped analogue of
// ObservableListBinding.
type ObservableSetBinding[T comparable] interface {
	ObservableSet[T]
	Binding[[]T]
}
