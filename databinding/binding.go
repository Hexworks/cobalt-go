package databinding

import "github.com/hexworks/cobalt-go/core"

// Binding computes its value from one or more ObservableValue
// dependencies and refreshes itself whenever any of them changes. A
// Binding is read-only — the difference from Property is the absence
// of SetValue/UpdateValue. Disposing a Binding stops it from
// observing its dependencies.
//
// The concrete unidirectional/bidirectional/computed implementations
// land in stage 2; this interface is declared here so the stage 1
// WritableValue contract can mention it.
type Binding[T any] interface {
	ObservableValue[T]
	core.Disposable
}
