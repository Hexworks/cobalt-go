package databinding

// CollectionProperty unions the observable, writable and Property
// surfaces over a collection-shaped underlying value. Mirrors the
// Kotlin CollectionProperty<T, C : PersistentCollection<T>>.
type CollectionProperty[T any, C any] interface {
	ObservableCollection[T, C]
	WritableCollection[T, C]
	Property[C]
}

// ListProperty is the list-shaped CollectionProperty. The underlying
// value type is `[]T`.
type ListProperty[T any] interface {
	ObservableList[T]
	WritableList[T]
	Property[[]T]
}

// SetProperty is the set-shaped CollectionProperty. The underlying
// value type is `[]T`; uniqueness is enforced by the mutators.
type SetProperty[T comparable] interface {
	ObservableSet[T]
	WritableSet[T]
	Property[[]T]
}

// MapProperty is the map-shaped CollectionProperty. Underlying value
// type is `map[K]V`.
type MapProperty[K comparable, V any] interface {
	ObservableMap[K, V]
	WritableMap[K, V]
	Property[map[K]V]
}
