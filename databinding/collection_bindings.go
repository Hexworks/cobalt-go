package databinding

import "reflect"

// BindListSize wires a Binding[int] tracking len(source.Value()).
// Slice-shaped counterpart of the Kotlin
// `ObservablePersistentCollection.bindSize` extension.
func BindListSize[T any](source ObservableValue[[]T]) Binding[int] {
	return NewComputedBinding(source, func(s []T) int { return len(s) })
}

// BindMapSize is the map-shaped counterpart of BindListSize.
func BindMapSize[K comparable, V any](source ObservableValue[map[K]V]) Binding[int] {
	return NewComputedBinding(source, func(m map[K]V) int { return len(m) })
}

// BindListIsEmpty wires a Binding[bool] true when source has zero
// elements.
func BindListIsEmpty[T any](source ObservableValue[[]T]) Binding[bool] {
	return NewComputedBinding(source, func(s []T) bool { return len(s) == 0 })
}

// BindMapIsEmpty is the map-shaped counterpart of BindListIsEmpty.
func BindMapIsEmpty[K comparable, V any](source ObservableValue[map[K]V]) Binding[bool] {
	return NewComputedBinding(source, func(m map[K]V) bool { return len(m) == 0 })
}

// BindListContains wires a Binding[bool] true when source contains
// element.Value().
func BindListContains[T any](source ObservableValue[[]T], element ObservableValue[T]) Binding[bool] {
	return BindCompute(source, element, func(s []T, e T) bool {
		for _, v := range s {
			if reflect.DeepEqual(v, e) {
				return true
			}
		}
		return false
	})
}

// BindListContainsAll wires a Binding[bool] true when every element
// of other is present in source.
func BindListContainsAll[T any](source, other ObservableValue[[]T]) Binding[bool] {
	return BindCompute(source, other, func(s, o []T) bool {
		for _, e := range o {
			found := false
			for _, v := range s {
				if reflect.DeepEqual(v, e) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	})
}

// BindListIndexOf wires a Binding[int] tracking the position of
// element in source (-1 when absent).
func BindListIndexOf[T any](source ObservableValue[[]T], element ObservableValue[T]) Binding[int] {
	return BindCompute(source, element, func(s []T, e T) int {
		for i, v := range s {
			if reflect.DeepEqual(v, e) {
				return i
			}
		}
		return -1
	})
}

// BindListLastIndexOf is the BindListIndexOf variant scanning from
// the back.
func BindListLastIndexOf[T any](source ObservableValue[[]T], element ObservableValue[T]) Binding[int] {
	return BindCompute(source, element, func(s []T, e T) int {
		for i := len(s) - 1; i >= 0; i-- {
			if reflect.DeepEqual(s[i], e) {
				return i
			}
		}
		return -1
	})
}

// BindListIsEqualTo wires a Binding[bool] true when a and b have
// equal (DeepEqual) contents.
func BindListIsEqualTo[T any](a, b ObservableValue[[]T]) Binding[bool] {
	return BindCompute(a, b, func(x, y []T) bool {
		return reflect.DeepEqual(x, y)
	})
}

// BindListPlus wires an ObservableListBinding holding the
// concatenation of a and b.
func BindListPlus[T any](a, b ObservableValue[[]T]) ObservableListBinding[T] {
	return NewObservableListBinding[T](
		BindCompute(a, b, func(x, y []T) []T {
			out := make([]T, 0, len(x)+len(y))
			out = append(out, x...)
			out = append(out, y...)
			return out
		}),
		"",
	)
}

// BindListMinus wires an ObservableListBinding holding a with every
// element of b removed (first match wins per occurrence).
func BindListMinus[T any](a, b ObservableValue[[]T]) ObservableListBinding[T] {
	return NewObservableListBinding[T](
		BindCompute(a, b, func(x, y []T) []T {
			out := make([]T, 0, len(x))
			for _, v := range x {
				skip := false
				for _, w := range y {
					if reflect.DeepEqual(v, w) {
						skip = true
						break
					}
				}
				if !skip {
					out = append(out, v)
				}
			}
			return out
		}),
		"",
	)
}

// BindSetPlus wires an ObservableSetBinding holding the union of a
// and b.
func BindSetPlus[T comparable](a, b ObservableValue[[]T]) ObservableSetBinding[T] {
	return NewObservableSetBinding[T](
		BindCompute(a, b, func(x, y []T) []T {
			seen := make(map[T]struct{}, len(x)+len(y))
			out := make([]T, 0, len(x)+len(y))
			for _, v := range x {
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				out = append(out, v)
			}
			for _, v := range y {
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				out = append(out, v)
			}
			return out
		}),
		"",
	)
}

// BindSetMinus wires an ObservableSetBinding holding a with every
// element of b removed.
func BindSetMinus[T comparable](a, b ObservableValue[[]T]) ObservableSetBinding[T] {
	return NewObservableSetBinding[T](
		BindCompute(a, b, func(x, y []T) []T {
			exclude := make(map[T]struct{}, len(y))
			for _, v := range y {
				exclude[v] = struct{}{}
			}
			out := make([]T, 0, len(x))
			for _, v := range x {
				if _, ok := exclude[v]; ok {
					continue
				}
				out = append(out, v)
			}
			return out
		}),
		"",
	)
}

// BindListFlatten flattens a list of observable lists into a single
// list, recomputed whenever the outer list changes. The inner lists'
// own changes are NOT observed — matches the Kotlin implementation,
// which only recomputes on outer-list mutations.
func BindListFlatten[T any](source ObservableValue[[]ObservableValue[[]T]]) ObservableListBinding[T] {
	return NewObservableListBinding[T](
		NewComputedBinding(source, func(s []ObservableValue[[]T]) []T {
			out := []T{}
			for _, inner := range s {
				out = append(out, inner.Value()...)
			}
			return out
		}),
		"",
	)
}

// BindListFlatMap is BindListFlatten plus an element-mapping step.
func BindListFlatMap[S, T any](source ObservableValue[[]ObservableValue[[]S]], converter func(S) T) ObservableListBinding[T] {
	return NewObservableListBinding[T](
		NewComputedBinding(source, func(s []ObservableValue[[]S]) []T {
			out := []T{}
			for _, inner := range s {
				for _, v := range inner.Value() {
					out = append(out, converter(v))
				}
			}
			return out
		}),
		"",
	)
}

// BindSetFlatten / BindSetFlatMap are the set-shaped counterparts.
func BindSetFlatten[T comparable](source ObservableValue[[]ObservableValue[[]T]]) ObservableSetBinding[T] {
	return NewObservableSetBinding[T](
		NewComputedBinding(source, func(s []ObservableValue[[]T]) []T {
			seen := make(map[T]struct{})
			out := []T{}
			for _, inner := range s {
				for _, v := range inner.Value() {
					if _, ok := seen[v]; ok {
						continue
					}
					seen[v] = struct{}{}
					out = append(out, v)
				}
			}
			return out
		}),
		"",
	)
}

func BindSetFlatMap[S any, T comparable](source ObservableValue[[]ObservableValue[[]S]], converter func(S) T) ObservableSetBinding[T] {
	return NewObservableSetBinding[T](
		NewComputedBinding(source, func(s []ObservableValue[[]S]) []T {
			seen := make(map[T]struct{})
			out := []T{}
			for _, inner := range s {
				for _, v := range inner.Value() {
					t := converter(v)
					if _, ok := seen[t]; ok {
						continue
					}
					seen[t] = struct{}{}
					out = append(out, t)
				}
			}
			return out
		}),
		"",
	)
}

// BindMapList element-maps source through transformer into an
// ObservableListBinding[T]. Equivalent to the Kotlin
// `ObservableList.bindMap` extension; lives here so the
// ListBindingDecorator wrapping stays in one place.
func BindMapList[S, T any](source ObservableList[S], transformer func(S) T) ObservableListBinding[T] {
	return NewObservableListBinding[T](NewListBinding(source, transformer), "")
}

// BindMapSet is the set-shaped counterpart of BindMapList.
func BindMapSet[S, T comparable](source ObservableSet[S], transformer func(S) T) ObservableSetBinding[T] {
	return NewObservableSetBinding[T](NewSetBinding(source, transformer), "")
}
