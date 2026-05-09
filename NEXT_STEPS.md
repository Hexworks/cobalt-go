# cobalt-go — Port Handoff

This document is the running handoff for porting the Kotlin
[`cobalt`](https://github.com/Hexworks/cobalt) library to Go. It is
intentionally written so any future session can pick up the port with
zero prior context.

Original sources live at `/home/addamsson/projects/cobalt/cobalt.core`.
Always cross-reference them rather than guessing.

---

## Project ground rules (from the user)

1. **Module path**: `github.com/hexworks/cobalt-go`.
2. **Go version**: `1.26.1` (installed). `go.mod` pinned to it.
3. **Layout**: single Go module, multiple packages, mirroring the
   original top-level layout (`core`, `events`, `databinding`). Skip
   `logging` — use `log/slog` where logging is needed.
4. **Generics**: use them where they make the API cleaner, but stay
   idiomatic.
5. **Single-threaded**: no goroutines, no mutexes, no atomics. Port
   anything that exists for thread safety as a plain value/holder.
6. **No persistent collections**: regular Go `[]T`, `map[K]V`, and
   `map[T]struct{}` for sets. The Kotlin original uses
   `kotlinx.collections.immutable` everywhere — translate by
   defensively copying on read/write where the original relied on
   value semantics.
7. **Tests**: port every Kotlin test verbatim into Go. We use
   `github.com/stretchr/testify` for assertions (`assert`/`require`)
   to keep the ported tests readable.
8. **Workflow**: port one package at a time, commit after each
   package, then update this document with the next package's plan so
   the following session starts with a clean context.
9. **Maybe / Optional**: Cobalt does not actually have a `Maybe`
   type. It uses nullable `T?` for absence and sealed result types
   (e.g. `ValueValidationResult<T>`) for richer outcomes. In Go we
   model absence with `*T` (pointer) when the value type is large or
   when nil is meaningful, and with `(T, bool)` / `(T, error)` at
   call boundaries. If a richer `Maybe[T]`-style type is genuinely
   needed in databinding, introduce it locally then; do not invent
   it upfront.
10. **Dependencies**: stdlib first. Google packages allowed. Currently
    pulled in:
    - `github.com/google/uuid`
    - `github.com/stretchr/testify`

---

## Design decisions worth preserving

### Sealed Kotlin types → tagged interface in Go

For each `sealed class` we use a Go interface with an unexported
marker method so only types in the same package can implement it.
Example (see `core/dispose_state.go`):

```go
type DisposeState interface {
    IsDisposed() bool
    isDisposeState()
}
```

Variants are concrete types in the same package. Parameterless
variants (`NotDisposed`, `DisposedManually`) are exposed as package
vars holding singleton values.

### Kotlin `<T : Any>` data classes with generics

When the type parameter is purely informational (e.g. the Kotlin
`DisposedByEvent<T : Any>` only stored the value), we drop the
generic and use `any` for the field. Callers do a type assertion at
the point of consumption. This keeps the sealed interface itself
parameter-free, which would otherwise infect every signature that
mentions `DisposeState`.

If a future case genuinely needs type-safe sum variants, revisit; Go
sum types with generics quickly become awkward.

### Extension functions

Go has no extension methods. Three idioms in this port:

- Stand-alone functions in the same package (e.g.
  `core.Abbreviate(u UUID)` replaces `UUID.abbreviate()`).
- Wrapper types when behaviour is heavy enough to deserve them.
- Package-level helpers for cross-package extensions (e.g. the
  Kotlin `Disposable.disposeWhen(ObservableValue<Boolean>)` infix
  fn must live in `databinding`, not `core`, to avoid an import
  cycle — see "Circular import" below).

### Default arguments

Go has none. Where Kotlin used a sensible default we either:

- Force the caller to pass it explicitly (cleanest).
- Add a paired no-arg helper function alongside the main method.

For `Disposable.Dispose`, we kept a single method
`Dispose(state DisposeState)` and require callers to pass
`core.DisposedManually` explicitly. Less ergonomic than Kotlin's
default arg but it keeps the interface tight.

### `internal` packages

Kotlin's `internal` visibility lives at the module level. Go's
`internal/` is path-based:

- `cobalt-go/internal/...` is reachable by every package under
  `cobalt-go/`.
- `cobalt-go/foo/internal/...` is reachable only under
  `cobalt-go/foo/`.

We use `cobalt-go/internal/atom/` for `Atom[T]` because databinding
(a sibling of `core`) needs it. The databinding-only singleton bus
lives at `cobalt-go/databinding/internal/cobalt/` so external code
cannot publish on or close it.

### Circular import: `Disposable` ↔ `ObservableValue`

In Kotlin, `Disposable` (`core`) has infix helpers that reference
`ObservableValue` (`databinding`). Go forbids the resulting cycle.
Fix: keep `core.Disposable` minimal (just `DisposeState()` /
`Dispose`) and re-add `DisposeWhen(d, cond)` and `KeepWhile(d, cond)`
as **package-level functions in `databinding`** when stage 5 lands.

---

## Layout so far

```
cobalt-go/
├── go.mod                       # module github.com/hexworks/cobalt-go (Go 1.26.1)
├── go.sum
├── NEXT_STEPS.md                # this file
├── LICENSE
├── README.md
├── core/
│   ├── disposable.go            # Disposable interface, Disposed(d) helper
│   ├── dispose_state.go         # sealed DisposeState + variants
│   ├── uuid.go                  # UUID type alias of google/uuid + helpers
│   ├── predicate.go             # Predicate[T], And/Or/Not
│   └── function.go              # Identity[T]
├── events/
│   ├── event.go                 # Event, EventSource, EventScope, EventDescriptor[E]
│   ├── subscription.go          # Subscription, CallbackResult (sealed)
│   ├── bus.go                   # EventBus, NewEventBus, ApplicationScope,
│   │                            # SubscribeTo / SimpleSubscribeTo helpers
│   ├── default_bus.go           # defaultEventBus impl + busSubscription
│   └── bus_test.go              # port of EventBusTest.kt
├── databinding/                 # stages 1-6 (foundations + properties + bindings + expressions + collections + extensions + tests)
│   ├── value.go                 # Value / WritableValue / ObservableValue interfaces
│   ├── binding.go               # Binding[T] interface
│   ├── binding_action.go        # BindingAction sealed (UpdateOnBind / NoActionOnBind)
│   ├── value_validation.go      # ValueValidationResult sealed + ValueValidationFailedError
│   ├── observable_value_changed.go # ObservableValueChanged event + AnyObservableValueChanged
│   ├── change_type.go           # ChangeType sealed: Scalar/List/Set/Map variants
│   ├── converter.go             # Converter / IsomorphicConverter / IdentityConverter
│   ├── property_scope.go        # PropertyScope (EventScope wrapping a UUID)
│   ├── circular_binding_error.go # unexported circularBindingError
│   ├── property.go              # Property / PropertyDelegate / PropertyValidator + internalProperty
│   ├── base_property.go         # basePropertyState[T] — shared storage + event routing
│   ├── default_property.go      # defaultProperty[T] (scalar) + NewProperty constructor
│   ├── default_property_delegate.go
│   ├── internal_helpers.go      # disposeSubscriptions, runWithDisposeOnFailure
│   ├── base_binding.go          # baseBinding[T] embeddable shared state
│   ├── unidirectional_binding.go
│   ├── bidirectional_binding.go
│   ├── computed_binding.go      # ComputedBinding[S,T] (exported)
│   ├── computed_dual_binding.go # computedDualBinding[S0,S1,T]
│   ├── generic_bindings.go      # BindTransform / BindCompute / BindWithConverter / UpdateFromConverter
│   ├── numeric_bindings.go      # BindNegate, BindPlus/Minus/Times/Div, Bind{Greater,Less}{Than,Equal}, BindEquals, BindToString
│   ├── boolean_bindings.go      # BindBool{Not,And,Or,Xor}
│   ├── string_bindings.go       # BindString{IsEmpty,IsBlank,Concat,EqualsIgnoreCase,Length}
│   ├── observable_collection.go # ObservableCollection / ObservableList / ObservableSet / ObservableMap / *Binding
│   ├── writable_collection.go   # WritableCollection / WritableList / WritableSet / WritableMap
│   ├── collection_property.go   # CollectionProperty / ListProperty / SetProperty / MapProperty
│   ├── default_list_property.go # defaultListProperty[T] + NewListProperty
│   ├── default_set_property.go  # defaultSetProperty[T] + NewSetProperty
│   ├── default_map_property.go  # defaultMapProperty[K,V] + NewMapProperty (+ cloneMap helper)
│   ├── default_property_list_property.go  # list-of-properties variant
│   ├── default_property_set_property.go
│   ├── default_property_map_property.go
│   ├── list_binding.go          # ListBinding[S,T] + applyListChange dispatcher
│   ├── set_binding.go           # SetBinding[S,T] + applySetChange dispatcher
│   ├── list_binding_decorator.go
│   ├── set_binding_decorator.go
│   ├── collection_bindings.go   # BindListSize/Plus/Minus/.../ BindMapList / BindMapSet / Flatten / FlatMap
│   ├── disposable_extensions.go # DisposeWhen / KeepWhile (stage 5)
│   ├── extensions.go            # ToProperty family + asInternalProperty / toInternalProperty (stage 5)
│   ├── default_property_test.go            # stage 6 — port of DefaultPropertyTest.kt
│   ├── default_property_delegate_test.go   # stage 6 — port of DefaultPropertyDelegateTest.kt
│   ├── computed_dual_binding_test.go       # stage 6 — port of ComputedDualBindingTest.kt
│   ├── bidirectional_binding_test.go       # stage 6 — port of BidirectionalConverterBindingTest.kt
│   ├── collection_bindings_test.go         # stage 6 — port of CollectionBindingsTest.kt
│   ├── default_list_property_test.go       # stage 6 — port of DefaultListPropertyTest.kt
│   ├── default_set_property_test.go        # stage 6 — port of DefaultSetPropertyTest.kt
│   ├── default_map_property_test.go        # stage 6 — port of DefaultMapPropertyTest.kt
│   ├── default_property_list_property_test.go  # stage 6 — port of DefaultPropertyListPropertyTest.kt
│   ├── default_property_set_property_test.go   # stage 6 — port of DefaultPropertySetPropertyTest.kt
│   ├── numeric_bindings_test.go            # stage 6 — port of LongExpressionsTest.kt
│   ├── string_bindings_test.go             # stage 6 — port of StringExpressionsTest.kt
│   └── internal/cobalt/
│       └── cobalt.go            # singleton EventBus (with ResetForTest)
└── internal/
    └── atom/
        ├── atom.go              # Atom[T] mutable holder
        └── atom_test.go         # port of AtomTest.kt
```

### Mapping core/api → core/

| Kotlin file (`org.hexworks.cobalt.core.*`)            | Go file                              |
| ----------------------------------------------------- | ------------------------------------ |
| `api/UUID.kt`                                         | `core/uuid.go`                       |
| `api/behavior/Disposable.kt`                          | `core/disposable.go` (trimmed)       |
| `api/behavior/DisposeState.kt`                        | `core/dispose_state.go`              |
| `api/extensions/Functions.kt` (`identity`)            | `core/function.go`                   |
| `api/extensions/IdentifierExtensions.kt` (`abbreviate`)| `core/uuid.go`                      |
| `api/extensions/Predicates.kt`                        | `core/predicate.go`                  |
| `internal/Atom.kt` + `internal/impl/DefaultAtom.kt`   | `internal/atom/atom.go`              |
| `internal/DefaultUUID.kt`                             | replaced by `github.com/google/uuid` |

### Tests ported

- `commonTest/.../core/internal/AtomTest.kt` →
  `internal/atom/atom_test.go`.
- `commonTest/.../events/EventBusTest.kt` → `events/bus_test.go`.
- `commonTest/.../databinding/**` → **ported in stage 6** (see databinding stage 6 design notes below).

---

## events: notes for future readers (unchanged from prior sessions)

The `events` port is done; the design choices below differ from the
Kotlin original and from the original NEXT_STEPS guesses, so they
deserve calling out before anyone touches the package again.

### Generic methods → free functions

Go forbids generic methods on interfaces. The bus interface therefore
exposes a non-generic `Subscribe(key string, scope EventScope,
fn func(Event) CallbackResult)`. Type-safe wiring lives in two
package-level helpers:

- `SubscribeTo[E Event](bus, descriptor, scope, fn)` wraps the
  callback in a `func(Event) CallbackResult` that asserts the event
  to `E` before invoking `fn`.
- `SimpleSubscribeTo[E Event](bus, descriptor, scope, fn)` is the
  Kotlin `simpleSubscribeTo` analogue — returns `KeepSubscription`
  for every call.

### `EventDescriptor[E Event]` is just a typed key

A struct with one field is enough:

```go
type EventDescriptor[E Event] struct { Key string }
```

### `EventScope` is open

Defined as `type EventScope any`. Must be comparable (the bus uses
it as a map key); empty struct singletons are the idiomatic choice.
`ApplicationScope` is one such singleton.

### No cycle detection in the bus

The cycle-detection algorithm lives in
`databinding.basePropertyState.updateWithEvent` (and now in
`list_binding.go` / `set_binding.go` via `traceContainsID`), not in
the bus. `Event.trace` is metadata that downstream code inspects.

### Default scope at call sites

Every method (`Subscribe`, `Publish`, `FetchSubscribers`,
`CancelScope`, plus both generic helpers) requires the caller to
pass a scope explicitly.

### Exception handling becomes panic recovery

`recover()` produces an `error`; the disposal itself is wrapped in
another recover so a panic from `Dispose` doesn't take down the
publish loop.

### Closed bus

Calls into a closed bus `panic("This Event Bus is already closed.")`.

### Logging

`slog.Default()` for both debug and warn lines.

---

## databinding: stages 1-4 design notes

### Single flat package

`databinding/` is one Go package, not the multi-layer Kotlin
`api/` + `internal/` tree. Splitting would have created cycles
(`ObservableValueChanged` references `ObservableValue`,
`ListPropertyChange` references `ObservableValueChanged`, etc.) and
Go forbids them. Path-internal helpers live under
`databinding/internal/`.

### `AnyObservableValueChanged` for type erasure

Kotlin uses `ObservableValueChanged<*>` and `<out T>` covariance to
pass change events without knowing T. Go has no variance. We add a
non-generic interface:

```go
type AnyObservableValueChanged interface {
    events.Event
    ChangeType() ChangeType
    isObservableValueChanged()
}
```

`ObservableValueChanged[T]` satisfies it via the marker method.
`updateWithEvent` and `ListPropertyChange` / `SetPropertyChange` /
`MapPropertyChange` use the erased view.

### `basePropertyState[T]` is the shared base

Stage 4 extracted a `basePropertyState[T]` from `defaultProperty[T]`
so every concrete property — scalar, list, set, map, and their
property-of-property variants — embeds the same storage and event
routing. The crucial trick is the `self ObservableValue[T]` field:
every constructor that builds a property sets it to the outer
wrapping struct so events published from inside the embedded state
correctly reference the wrapper, not the embedded base.

`defaultProperty[T]` is now a thin shell adding only `Bind`,
`UpdateFrom`, and `AsDelegate`. All other Property methods live on
`basePropertyState[T]` and are promoted into every embedder.

### `internalProperty[T]` interface (unexported)

Privileged interface bindings cast to so they can call
`updateWithEvent` (bypasses the validating public setters) and read
`propertyScope()`. Lives unexported because outside packages have no
reason to drive that path. `basePropertyState[T]` is the canonical
implementer; promoted into every concrete property type.

### Generic methods → free functions in databinding too

Same constraint as events. The cross-type variants of `Bind` and
`UpdateFrom` are package-level functions:

- `BindWithConverter[S, T](target, other, action, converter)`
- `UpdateFromConverter[S, T](target, observable, action, converter)`

Same-type variants stay on the `Property` / `WritableValue`
interface methods (`Bind`, `UpdateFrom`). Stage 4's collection
property types each declare these explicitly because Go cannot
infer them via embedding (the generic method on `defaultProperty[T]`
doesn't transfer to `defaultListProperty[T]`).

### Binding inheritance via embedded state

`baseBinding[T]` holds id/name/target/subscriptions/scope/disposeState
and implements all the `ObservableValue[T]` + `core.Disposable`
methods. Concrete bindings (`unidirectionalBinding[S, T]`,
`bidirectionalBinding[S, T]`, `ComputedBinding[S, T]`,
`computedDualBinding[S0, S1, T]`, `ListBinding[S, T]`,
`SetBinding[S, T]`) embed it and only write the subscription
wiring.

`baseBinding.publishChange` takes the concrete binding as
`asSource ObservableValue[T]` so the published event's
`ObservableValue` / `Emitter` fields point at the outer struct, not
the embedded base. `changeType` is parameterised because
`ComputedBinding` always emits `ScalarChange` regardless of the
upstream change kind (matches Kotlin), and `ListBinding` /
`SetBinding` pass through the source's change kind so downstream
listeners see element-level mutations.

### Cycle detection by ID

`updateWithEvent` walks `event.Trace()` and checks
`emitter.ID() == p.id`. Same algorithm is shared with
`ListBinding` / `SetBinding` via the package-level
`traceContainsID(trace, id)` helper.

### Validation as panic at the write boundary

`SetValue` panics with `*ValueValidationFailedError` on rejection
(matches Kotlin's setter exception); `UpdateValue` /
`TransformValue` return `ValueValidationFailed[T]` instead.
`updateWithEvent` panics on validator failure too — the binding's
`runWithDisposeOnFailure` catches it and disposes the binding with
`DisposedByException`, matching the Kotlin
`DisposedByException(e)` flow.

### Equality

`reflect.DeepEqual(oldValue, newValue)` is used to decide whether
to fire a change event for `ScalarChange`. List/Set/Map changes
always fire because their callers know mutation happened. For
collection element comparisons, `reflect.DeepEqual` is used when T
is unconstrained; sets constrain T to `comparable` and use `==`.

### Singleton bus in `internal/cobalt/`

`cobalt.EventBus()` returns the process-wide singleton.
`cobalt.ResetForTest()` is available for stage 6 tests that close
the bus or want a clean slate per case. Production code must never
call `ResetForTest`.

### Numeric / boolean / string expression bindings

Kotlin had one file per type with infix extension functions. Go
collapses them via constraints (`Numeric`, `Signed`, `cmp.Ordered`,
`comparable`) into three files:

- `numeric_bindings.go` — `BindNegate[Signed]`,
  `BindPlus/Minus/Times/Div[Numeric]`,
  `Bind{Greater,Less}{Than,OrEqual}[cmp.Ordered]`,
  `BindEquals[comparable]`, `BindToString[any]`
- `boolean_bindings.go` — `BindBool{Not,And,Or,Xor}`
- `string_bindings.go` — `BindString{IsEmpty,IsBlank,Concat,EqualsIgnoreCase,Length}`

Kotlin's cross-type Number overloads (e.g.
`ObservableValue<Int>.bindPlusWith(ObservableValue<Number>)`) are
not directly portable because Go has no `Number` supertype.
Same-type operands only; callers cast explicitly if they need
cross-type arithmetic. The Kotlin tests never exercise the
cross-type path so nothing is lost.

`BindStringLength` returns byte count (Go convention), not UTF-16
code-unit count. Stage 6 tests may need to account for this if any
assertion checks `length()` on multi-byte strings.

`BindStringIsBlank` uses `strings.TrimSpace(s) == ""` —
whitespace-only matches Kotlin's `String.isBlank()` for ASCII
input; Unicode-only spaces behave the same because `TrimSpace`
strips Unicode space.

---

## databinding stage 4 design notes (new)

### Collection representations

- **Lists**: plain `[]T`. Every mutator returns a freshly allocated
  slice, no aliasing with the property's backing storage.
- **Sets**: also `[]T` for the public Value type so the same
  Property[[]T] plumbing carries them through bindings. Uniqueness
  is enforced inside every mutator; T constrained to `comparable`
  so internal dedup uses a `map[T]struct{}`. Insertion order is
  generally preserved by the implementations but is **not
  guaranteed** — leave us free to swap storage later.
- **Maps**: plain `map[K]V`. `cloneMap` clones on every mutation,
  so old and new values in change events are independent.

### Collection property type hierarchy

- Read-only views: `ObservableCollection[T,C]` →
  `ObservableList[T]`, `ObservableSet[T comparable]`,
  `ObservableMap[K comparable, V]`.
- Writable views: `WritableCollection[T,C]` → `WritableList[T]`,
  `WritableSet[T comparable]`, `WritableMap[K,V]`.
- Property aggregates: `CollectionProperty[T,C]` →
  `ListProperty[T]`, `SetProperty[T comparable]`,
  `MapProperty[K,V]`. Each unions the matching observable +
  writable + scalar `Property[C]` surfaces.

### `T comparable` for sets

The internal dedup uses a map lookup, which requires
`comparable`. ObservableList stays `T any` because slices don't
need element-level equality except for `Contains`/`IndexOf`, which
use `reflect.DeepEqual`.

For `defaultPropertySetProperty[T any, P comparable+ObservableValue[T]]`,
P must be `comparable` because the set's element type is P. Pointer
property types satisfy `comparable` in practice (interface values
backed by pointers are comparable).

### Composition over inheritance: property-of-property

`defaultPropertyListProperty[T, P]` embeds a `*defaultListProperty[P]`
as a field (composition). After constructing the inner property,
the outer constructor sets `inner.self = outer` so the inner's
embedded `basePropertyState.self` references the outer. The outer
then overrides every mutator method to add subscribe / unsubscribe
side effects before delegating to the inner's implementation. Same
pattern for `defaultPropertySetProperty` and
`defaultPropertyMapProperty`.

Subscription bookkeeping is `map[core.UUID]propertySubscription[P]`
keyed by the child property's id. Idempotent on duplicate adds;
disposes the subscription on remove.

`SetValue` / `UpdateValue` / `TransformValue` on a
property-of-property bypass the per-element bookkeeping — matches
the Kotlin behaviour where `value.setter` bypasses
`add`/`remove`. Production code must use the collection mutators if
it wants the inner-change forwarding.

### `ListBinding` / `SetBinding`

These observe an `ObservableList[S]` / `ObservableSet[S]`, map each
element through a converter, and replay every source change onto
an internal `defaultListProperty[T]` / `defaultSetProperty[T]`
target.

The dispatch on change type lives in `applyListChange[S,T]` /
`applySetChange[S,T]`. Both are typed switches over the variants
declared in `change_type.go`. `ListClear` and `SetClear` map to
empty slices; `ListPropertyChange` / `SetPropertyChange` leave the
target untouched (the outer source list/set didn't change — only
an inner property did).

`ListRemoveAllWhen[S]` carries a `func(S) bool`. The Kotlin code
unsafe-casts it to `func(T) bool`; we attempt the same cast via
`any(c.Predicate).(func(T) bool)` and leave the target untouched
on failure. In practice this variant only flows when S == T (the
identity-converter case).

The change-type unsafe-cast pattern is the only "type ladder" the
binding has to walk. If the upstream `ChangeType` interface adds
new collection variants, both `applyListChange` and `applySetChange`
need new cases.

### List / Set binding decorators

`listBindingDecorator[T]` and `setBindingDecorator[T comparable]`
wrap any `Binding[[]T]` to satisfy the full
`ObservableListBinding[T]` / `ObservableSetBinding[T]` surface.
Used by the helpers in `collection_bindings.go` (BindListPlus,
BindListFlatten, etc.) so callers get a list/set-shaped read API
on top of a generic ComputedBinding.

### Cross-type collection helpers

`collection_bindings.go` provides:

- `BindListSize` / `BindMapSize` / `BindListIsEmpty` /
  `BindMapIsEmpty`
- `BindListContains` / `BindListContainsAll` /
  `BindListIndexOf` / `BindListLastIndexOf` /
  `BindListIsEqualTo`
- `BindListPlus` / `BindListMinus` / `BindSetPlus` /
  `BindSetMinus`
- `BindListFlatten` / `BindListFlatMap` /
  `BindSetFlatten` / `BindSetFlatMap`
- `BindMapList` / `BindMapSet` (the Kotlin
  `ObservableList.bindMap` / `ObservableSet.bindMap` helpers)

Kotlin's `bindContainsWith` etc. all collapse onto `BindCompute`
under the hood, same as our scalar expression bindings.

`BindListFlatten` / `BindSetFlatten` re-compute on **outer**
list/set changes only — the inner collections' own changes are
not observed (matches Kotlin's ComputedBinding-only
implementation). Listeners that need transitive change forwarding
should use a `PropertyListProperty` instead.

---

## databinding stage 5 design notes (new)

### What landed

- `disposable_extensions.go` — `DisposeWhen(d core.Disposable, condition ObservableValue[bool])`
  and `KeepWhile(d core.Disposable, condition ObservableValue[bool])`.
  Lives in `databinding` (not `core`) to keep core free of an
  `ObservableValue` import — see the "Circular import" note above.
  Both check the condition's current value first; if the trigger
  state already holds they dispose immediately, otherwise they
  install an `OnChange` subscription that disposes itself in the
  same callback it fires from. The subscription variable is captured
  via `var sub events.Subscription` declared before the OnChange
  call so the closure can reach it.
- `extensions.go` — `ToProperty` family of zero-validator,
  default-name convenience constructors:
  - `ToProperty[T](v T) Property[T]`
  - `ToListProperty[T any]([]T) ListProperty[T]`
  - `ToSetProperty[T comparable]([]T) SetProperty[T]`
  - `ToMapProperty[K comparable, V any](map[K]V) MapProperty[K,V]`
  - `ToPropertyListProperty[T any, P ObservableValue[T]]([]P) ListProperty[P]`
  - `ToPropertySetProperty[T any, P ObservableValue[T]+comparable]([]P) SetProperty[P]`
  - `ToPropertyMapProperty[K comparable, T any, P ObservableValue[T]](map[K]P) MapProperty[K,P]`
- `extensions.go` (unexported) — `toInternalProperty[T](v T) internalProperty[T]`
  and `asInternalProperty[T any](p Property[T]) internalProperty[T]`.
  The Kotlin originals are `internal` extensions used inside
  bindings; existing Go bindings reach the same surface via
  `newDefaultProperty` + embedded `basePropertyState`, but the
  helpers exist for API parity and future binding code.

### What was intentionally skipped

- **`AnyExtensions.kt`** (`fold`, `orElse`, `orElseGet`,
  `orElseThrow` for nullable `T?`). Per project rule 9 we model
  absence with `*T` or `(T, bool)`. These helpers are unused
  upstream outside the file itself; skipped to avoid inventing a
  `Maybe`-flavoured API. Add later if a real call site demands it.
- **`ListExtensions.kt`** (`PersistentList.map`). Go has
  `slices` / range loops; nothing to port.
- **`TypeAliases.kt`** (`ObservablePersistentCollection` etc.).
  Used only inside Kotlin `CollectionBindings.kt`; our stage-4
  `collection_bindings.go` already spells everything out without
  aliases. No call sites need them.
- **`Subscriptions.kt`** (`disposeSubscriptions`) and
  **`BindingExtensions.kt`** (`runWithDisposeOnFailure`) already
  live in `internal_helpers.go` from stage 2.

---

## databinding stage 6 design notes (new)

### Test package layout

All databinding tests live in `package databinding` (whitebox), not
`databinding_test`. Two reasons:

1. `assert.Equal` on `ObservableValueChanged[T]` needs to compare
   structs that carry unexported `emitter` / `trace` fields. With a
   whitebox test we can pass `change.Emitter()` / `change.Trace()`
   into the `NewObservableValueChanged` constructor and let testify's
   `reflect.DeepEqual` compare the resulting structs cleanly.
2. Future tests that want to reach into unexported helpers
   (`asInternalProperty`, the binding internals, etc.) can do so
   without an additional bridge file.

The `events/bus_test.go` port used the external `events_test` package
because it never compared structs with unexported fields and the
generic helpers it exercises are all exported. Both conventions are
fine; pick whichever the test needs.

### No bus reset between tests

Every property / binding constructed in tests owns a unique
`PropertyScope{ID: uuid}`, so events from different cases route
through disjoint scope keys on the singleton bus. None of the ported
tests exercise teardown semantics (close, cancel scope), so
`cobalt.ResetForTest()` is **not** called anywhere in stage 6. If a
future test needs a clean bus (e.g. asserting subscriber counts) it
should call `cobalt.ResetForTest()` from `internal/cobalt`.

### Set iteration order in tests

`defaultSetProperty` preserves insertion order in practice but the
contract doesn't promise it (see stage 4 notes). Tests that need a
deterministic comparison sort the slice before asserting — see
`sortedInts` in `default_set_property_test.go` and `sortStrings` in
`default_property_set_property_test.go`. The Kotlin originals relied
on `PersistentSet`'s order guarantee; Go set tests must sort.

### Property-list / Property-set event-equality tests

`DefaultPropertyListPropertyTest` / `DefaultPropertySetPropertyTest`
assert that the outer container republishes an inner change as
`ListPropertyChange` / `SetPropertyChange` wrapping the original
event. The Go port preserves struct identity through the wrapper —
both the inner subscriber and `change.ChangeType().ChangeEvent`
receive the same `ObservableValueChanged[int]` value, so
`reflect.DeepEqual` matches. Done via direct type assertion
`change.ChangeType().(ListPropertyChange)` / `(SetPropertyChange)`
followed by `assert.Equal(t, *innerChange, lpc.ChangeEvent)`.

### Cross-type bind exception path

`BidirectionalConverterBindingTest.kt` asserts that setting an
invalid value on the target (the converter panics on the way back)
ends with `binding.disposeState == DisposedByException`. In Go the
`Atoi("foo")` panic is caught by `runWithDisposeOnFailure` and the
binding gets `core.DisposedByException{Err: err}`. The test type-
asserts to `core.DisposedByException` rather than comparing on
`reflect.DeepEqual` because the wrapped error message format isn't
load-bearing.

### Numeric expression test uses int64

Kotlin's `LongExpressionsTest` exercises `Long`; the Go port uses
`int64` directly via `ToProperty[int64](2)` and `BindNegate[int64]`.
No `Long` type alias exists in the port — the `Numeric` constraint
covers every integer / float built-in and tests pick the one they
need.

### Hierarchical binding test reduce

`DefaultListPropertyTest.Given_a_array_of_lists_to_bind_When_they_are_bound`
uses Kotlin's `List.reduce` to chain `bindPlusWith`. Go has no
generic `reduce` for slices; the test fold uses an explicit for loop
seeded with the first property and accumulating through
`BindListPlus`. Same shape, different syntax.

### Flatten test on an empty list-of-lists

`CollectionBindingsTest.When_list_properties_are_flattened` starts
from `mutableListOf<ObservablePersistentCollection<Int>>().toProperty()`.
Direct Go translation: `ToListProperty[ObservableValue[[]int]](nil)`
(`nil` slice; the constructor's defensive copy turns it into `[]T{}`).
The test then adds two `ListProperty[int]` values which satisfy
`ObservableValue[[]int]` via the embedded ObservableList chain.

### BooleanExpressionsTest skipped

The Kotlin file is an empty class (no `@Test` methods). Stage 6 skips
it — there's nothing to port. The original NEXT_STEPS flagged this.

---

## What's left

The whole `cobalt.core` Kotlin module is now ported, including every
common test except the empty `BooleanExpressionsTest`. There is no
`commonMain` code left behind, and the upstream repo has no other
modules (`/home/addamsson/projects/cobalt/` only contains
`cobalt.core/`).

Possible follow-up work (not yet planned by the user):

- **Logging package**: skipped per rule 3. If anything in downstream
  consumers references the Kotlin `cobalt.logging` types, port at
  that point.
- **API smoothing**: some constructors are noisy (`ToProperty[T](v)`
  vs `NewProperty[T](v, "", nil)`); consider whether functional
  options would help once a real consumer surfaces the pain.
- **Doc comments → godoc** sweep: the port preserves the Kotlin
  KDoc-style commentary in most files. A pass to align with idiomatic
  godoc (lead with the identifier name in the doc comment) would
  improve `go doc` output. Style-only.
- **Cross-type numeric bindings**: the Kotlin originals supported a
  `Number` supertype on the RHS of `bindPlusWith` etc. Go has no such
  bound; same-type only. Add specific overloads (e.g.
  `BindPlusIntFloat`) if a real call site needs them.
- **Set iteration order**: contract says no guarantee but tests sort
  before asserting. If downstream consumers need deterministic order,
  swap `defaultSetProperty` to a linked-hash-style storage and
  promote the order guarantee into the doc.

### Known limitations

- **Set iteration order is not guaranteed.** `defaultSetProperty`
  preserves insertion-ish order incidentally but the contract
  doesn't promise it. Stage 6 tests sort before comparing.
- **`BindStringLength`** returns byte count, not character count.
- **`ListRemoveAllWhen` cross-type filter** is a no-op unless the
  predicate is castable to `func(T) bool` (i.e. S == T at runtime).
- **`SetValue` on a property-of-property** does not maintain child
  subscriptions. Matches Kotlin; can leak subscriptions on raw
  setters.

---

## Verification commands

```bash
go build ./...
go vet ./...
go test ./...
```

All three pass on the current tree (111 tests across 5 packages —
`core`, `events`, `internal/atom`, `databinding`, and
`databinding/internal/cobalt` which has no tests yet).

---

## Pointer for the user (reminder)

User prefers commits per package (and per stage within databinding)
and a fresh NEXT_STEPS handoff after each. Update this document
**before** committing each stage's work so the commit includes the
updated plan.
