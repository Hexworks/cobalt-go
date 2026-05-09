# cobalt-go

A Go port of the Kotlin [`cobalt`](https://github.com/Hexworks/cobalt) library:
a small, dependency-light toolkit for reactive value bindings, a typed event
bus, and the foundational primitives both build on (UUID, `Disposable`,
predicates).

The port is **single-threaded by design** — no goroutines, no mutexes, no
atomics. Use it inside one goroutine (typically a game / UI main loop) and let
the surrounding application decide how to dispatch work between threads.

## Status

The original `cobalt.core` Kotlin module is fully ported, including every
common test. See `NEXT_STEPS.md` for the per-stage port notes, design
trade-offs, and known limitations.

## Install

```bash
go get github.com/hexworks/cobalt-go
```

Requires Go 1.26.1 or newer (the module pins the toolchain explicitly because
the port uses recent generics features).

## Packages

| Import path                              | What it offers |
| ---------------------------------------- | -------------- |
| `github.com/hexworks/cobalt-go/core`        | `UUID`, `Disposable` / `DisposeState`, `Predicate`, `Identity` |
| `github.com/hexworks/cobalt-go/events`      | `EventBus`, `EventScope`, `EventDescriptor[E]`, `SubscribeTo` / `SimpleSubscribeTo` |
| `github.com/hexworks/cobalt-go/databinding` | `Property[T]`, `Binding[T]`, `List`/`Set`/`MapProperty`, expression bindings |

Two internal-only packages back the public API and should not be imported by
consumers:

| Internal path                                          | Purpose |
| ------------------------------------------------------ | ------- |
| `internal/atom`                                        | Minimal mutable cell used by `basePropertyState`. |
| `databinding/internal/cobalt`                          | Singleton `EventBus` that backs every property's change notifications. |

## What does it do?

`cobalt-go` lets you express **values that observe other values**. A
`Property[T]` holds a T and notifies subscribers when it changes. A
`Binding[T]` derives its value from one or more other observables and updates
automatically as they change. Properties and bindings can be chained,
converted, transformed, validated, and disposed.

The library also ships a typed event bus that the bindings use internally but
that you can use on its own to wire publish/subscribe communication between
parts of your program without leaking concrete types through interface{}-typed
callbacks.

## Usage

The snippets below assume you have already imported the relevant packages:

```go
import (
    "github.com/hexworks/cobalt-go/core"
    "github.com/hexworks/cobalt-go/databinding"
    "github.com/hexworks/cobalt-go/events"
)
```

### A property that you can observe

```go
name := databinding.ToProperty("Anonymous")

sub := name.OnChange(func(c databinding.ObservableValueChanged[string]) {
    fmt.Printf("name changed: %q -> %q\n", c.OldValue, c.NewValue)
})

name.SetValue("Ada")   // prints: name changed: "Anonymous" -> "Ada"
name.SetValue("Ada")   // no event — value didn't change
sub.Dispose(core.DisposedManually)
name.SetValue("Linus") // no event — subscription is gone
```

`ToProperty` is a convenience constructor with a default name and no
validator. Use `databinding.NewProperty(initial, name, validator)` when you
need either.

### Validation

A validator rejects state transitions. Failed writes through `SetValue` panic
with `*ValueValidationFailedError`; `UpdateValue` / `TransformValue` return a
`ValueValidationResult[T]` instead.

```go
positive := func(_, next int) bool { return next > 0 }
count := databinding.NewProperty(1, "count", positive)

result := count.UpdateValue(-3)
if !result.IsSuccessful() {
    // ValueValidationFailed[int] — count.Value() is still 1
}
```

### One-way binding (transform / compute)

`BindTransform` derives a value from a single source. `BindCompute` derives
from two sources at once.

```go
celsius := databinding.ToProperty[float64](20)
fahrenheit := databinding.BindTransform(celsius, func(c float64) float64 {
    return c*9.0/5.0 + 32
})

fmt.Println(fahrenheit.Value()) // 68

celsius.SetValue(100)
fmt.Println(fahrenheit.Value()) // 212
```

```go
a := databinding.ToProperty[int](2)
b := databinding.ToProperty[int](3)
sum := databinding.BindCompute(a, b, func(x, y int) int { return x + y })

fmt.Println(sum.Value()) // 5
a.SetValue(10)
fmt.Println(sum.Value()) // 13
```

When the dependency relationship is simple arithmetic, the typed expression
bindings are usually shorter:

```go
import "cmp"

price := databinding.ToProperty[int64](100)
qty   := databinding.ToProperty[int64](3)

total   := databinding.BindTimes(price, qty)             // Binding[int64]
inStock := databinding.ToProperty(true)
canBuy  := databinding.BindBoolAnd(
    inStock,
    databinding.BindGreaterThan[int64](qty, databinding.ToProperty[int64](0)),
)

_ = cmp.Compare(total.Value(), 0) // just to silence the unused import
```

The full expression library covers:

- numeric: `BindNegate`, `BindPlus`, `BindMinus`, `BindTimes`, `BindDiv`,
  `BindGreaterThan` / `BindLessThan` / `…OrEqual`, `BindEquals`,
  `BindToString`
- boolean: `BindBoolNot`, `BindBoolAnd`, `BindBoolOr`, `BindBoolXor`
- string: `BindStringIsEmpty`, `BindStringIsBlank`, `BindStringConcat`,
  `BindStringEqualsIgnoreCase`, `BindStringLength`

### Two-way binding (`Bind` and `UpdateFrom`)

`UpdateFrom` is one-way: the target follows the source.

```go
src := databinding.ToProperty("hello")
dst := databinding.ToProperty("")

dst.UpdateFrom(src, databinding.UpdateOnBind)
fmt.Println(dst.Value()) // hello
src.SetValue("world")
fmt.Println(dst.Value()) // world
```

`Bind` is two-way: both sides stay in sync. Pass `UpdateOnBind` to seed the
target from the source immediately, or `NoActionOnBind` to wait for the next
change.

```go
left  := databinding.ToProperty(1)
right := databinding.ToProperty(0)

binding := left.Bind(right, databinding.UpdateOnBind)
right.SetValue(42)
fmt.Println(left.Value()) // 42
left.SetValue(7)
fmt.Println(right.Value()) // 7

binding.Dispose(core.DisposedManually) // breaks the link
```

For cross-type bindings (different `S` and `T`) use `BindWithConverter` or
`UpdateFromConverter` with a `Converter[S, T]` (or `IsomorphicConverter[S, T]`
for the bidirectional case).

### Collection properties

```go
items := databinding.ToListProperty([]string{"a", "b"})

items.Add("c")
items.AddAt(0, "z")
items.RemoveAt(2)
fmt.Println(items.Value()) // [z a c]

size  := databinding.BindListSize(items)       // Binding[int]
empty := databinding.BindListIsEmpty(items)    // Binding[bool]

sub := items.OnChange(func(c databinding.ObservableValueChanged[[]string]) {
    fmt.Printf("list now: %v (change=%T)\n", c.NewValue, c.ChangeType())
})
defer sub.Dispose(core.DisposedManually)
```

`ToSetProperty` (requires `comparable` T) and `ToMapProperty` follow the same
shape. `BindListPlus` / `BindListMinus` / `BindListFlatten` / `BindListFlatMap`
(and their set counterparts) compose collection-shaped bindings without manual
wiring.

### A property of properties (transitive change notifications)

A regular `ListProperty[Property[int]]` only fires events when the outer list
changes. If you want changes from inner elements to propagate through the
outer collection, use the `*PropertyListProperty` family:

```go
a := databinding.ToProperty(1)
b := databinding.ToProperty(2)

list := databinding.ToPropertyListProperty[int, databinding.Property[int]](
    []databinding.Property[int]{a, b},
)

list.OnChange(func(c databinding.ObservableValueChanged[[]databinding.Property[int]]) {
    fmt.Printf("change kind: %T\n", c.ChangeType())
})

a.SetValue(99) // fires ListPropertyChange wrapping the inner ObservableValueChanged
```

### The event bus

Properties use a singleton bus internally; for application-level pub/sub
construct your own:

```go
bus := events.NewEventBus()

type UserLoggedIn struct {
    emitter events.EventSource
    Name    string
}

func (e UserLoggedIn) Key() string                 { return userLoggedInDescriptor.Key }
func (e UserLoggedIn) Emitter() events.EventSource { return e.emitter }
func (e UserLoggedIn) Trace() []events.Event       { return nil }

var userLoggedInDescriptor = events.EventDescriptor[UserLoggedIn]{Key: "UserLoggedIn"}

sub := events.SimpleSubscribeTo(bus, userLoggedInDescriptor, events.ApplicationScope,
    func(e UserLoggedIn) {
        fmt.Println("welcome,", e.Name)
    })
defer sub.Dispose(core.DisposedManually)

bus.Publish(UserLoggedIn{Name: "Ada"}, events.ApplicationScope) // prints: welcome, Ada
```

`SimpleSubscribeTo` always keeps the subscription after the callback runs. Use
`SubscribeTo` directly when the callback should decide per-event by returning
`events.KeepSubscription` or `events.DisposeSubscription`.

Scopes partition the bus: events published under one `EventScope` are
invisible to subscribers in another. Pass `events.ApplicationScope` when you
don't need partitioning; pass your own comparable singleton (typically an
empty struct value) otherwise.

### Disposal helpers

`DisposeWhen` disposes any `Disposable` the moment an `ObservableValue[bool]`
becomes true; `KeepWhile` is the inverse. The helper installs a self-cleaning
subscription so it never leaks past the first trigger:

```go
done    := databinding.ToProperty(false)
binding := celsius.Bind(other, databinding.UpdateOnBind)

databinding.DisposeWhen(binding, done)
done.SetValue(true) // binding is now disposed
```

## Running the tests

```bash
go build ./...
go vet ./...
go test ./...
```

All three pass on the current tree.

## License

Apache License, Version 2.0. See `LICENSE`.
