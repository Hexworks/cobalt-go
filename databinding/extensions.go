package databinding

// ToProperty wraps v in a Property[T] with default name and no
// validator. Convenience over NewProperty for the common case where
// callers don't care about either knob — mirrors the Kotlin
// `T.toProperty()` extension.
func ToProperty[T any](v T) Property[T] {
	return NewProperty[T](v, "", nil)
}

// ToListProperty wraps xs in a ListProperty[T] with default name and
// no validator. Mirrors the Kotlin `List<T>.toProperty()` /
// `Collection<T>.toProperty()` / `Iterable<T>.toProperty()` overloads
// that all funnel into DefaultListProperty in the original code.
func ToListProperty[T any](xs []T) ListProperty[T] {
	return NewListProperty[T](xs, "", nil)
}

// ToSetProperty wraps xs in a SetProperty[T] with default name and no
// validator. T must be comparable for the same reason as
// NewSetProperty. Mirrors the Kotlin `Set<T>.toProperty()` extension.
func ToSetProperty[T comparable](xs []T) SetProperty[T] {
	return NewSetProperty[T](xs, "", nil)
}

// ToMapProperty wraps m in a MapProperty[K,V] with default name and
// no validator. Mirrors the Kotlin `Map<K, V>.toProperty()` extension.
func ToMapProperty[K comparable, V any](m map[K]V) MapProperty[K, V] {
	return NewMapProperty[K, V](m, "", nil)
}

// ToPropertyListProperty wraps xs in a ListProperty[P] that observes
// each contained P and republishes inner changes as
// ListPropertyChange events. Mirrors the Kotlin
// `List<V>.toProperty()` overload where V : ObservableValue<T>.
func ToPropertyListProperty[T any, P ObservableValue[T]](xs []P) ListProperty[P] {
	return NewPropertyListProperty[T, P](xs, "", nil)
}

// ToPropertySetProperty wraps xs in a SetProperty[P] that observes
// each contained P and republishes inner changes as
// SetPropertyChange events. Mirrors the Kotlin
// `Set<V>.toProperty()` overload where V : ObservableValue<T>.
func ToPropertySetProperty[T any, P interface {
	ObservableValue[T]
	comparable
}](xs []P) SetProperty[P] {
	return NewPropertySetProperty[T, P](xs, "", nil)
}

// ToPropertyMapProperty wraps m in a MapProperty[K,P] that observes
// each contained P and republishes inner changes as
// MapPropertyChange events. Mirrors the Kotlin
// `Map<K, P>.toProperty()` overload where P : Property<V>.
func ToPropertyMapProperty[K comparable, T any, P ObservableValue[T]](m map[K]P) MapProperty[K, P] {
	return NewPropertyMapProperty[K, T, P](m, "", nil)
}

// toInternalProperty wraps v in a privileged internalProperty[T]. The
// Kotlin internal `T.toInternalProperty()` extension was used by
// bindings to stand up their target storage; here it is unexported
// because external code has no reason to drive the bypass path.
//
// Note: existing binding implementations call newDefaultProperty
// directly because the returned *defaultProperty already satisfies
// internalProperty[T] via embedded basePropertyState. This helper
// exists for API parity and future binding code that prefers the
// extension-style spelling.
func toInternalProperty[T any](v T) internalProperty[T] {
	return newDefaultProperty[T](v)
}

// asInternalProperty narrows a Property[T] to internalProperty[T].
// Mirrors the Kotlin internal `Property<T>.asInternalProperty()` cast
// helper. Every Property type defined in this package embeds
// basePropertyState[T] and thus satisfies internalProperty[T]; the
// assertion exists so binding code can pivot on the privileged
// interface without naming the embedded base directly.
//
// Panics if p was constructed outside this package and does not
// implement internalProperty[T] — same failure mode as the Kotlin
// unchecked cast.
func asInternalProperty[T any](p Property[T]) internalProperty[T] {
	return p.(internalProperty[T])
}
