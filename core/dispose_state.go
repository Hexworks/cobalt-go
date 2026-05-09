package core

// DisposeState describes the disposed state of a Disposable.
//
// Implementations are restricted to the variants declared in this
// package (sealed): NotDisposed, DisposedManually, DisposedByEvent and
// DisposedByException.
type DisposeState interface {
	// IsDisposed reports whether the disposable has been disposed.
	IsDisposed() bool
	// isDisposeState is an unexported marker that prevents external
	// packages from creating their own DisposeState variants.
	isDisposeState()
}

type notDisposed struct{}

func (notDisposed) IsDisposed() bool { return false }
func (notDisposed) isDisposeState()  {}

type disposedManually struct{}

func (disposedManually) IsDisposed() bool { return true }
func (disposedManually) isDisposeState()  {}

// DisposedByEvent indicates the disposable was disposed as a reaction
// to some event. Event carries the originating value.
type DisposedByEvent struct {
	Event any
}

func (DisposedByEvent) IsDisposed() bool { return true }
func (DisposedByEvent) isDisposeState()  {}

// DisposedByException indicates the disposable was disposed because an
// error occurred while it was running.
type DisposedByException struct {
	Err error
}

func (DisposedByException) IsDisposed() bool { return true }
func (DisposedByException) isDisposeState()  {}

// Singleton values for the parameterless variants. Treating them as
// values (not pointers) keeps equality comparisons trivial.
var (
	NotDisposed      DisposeState = notDisposed{}
	DisposedManually DisposeState = disposedManually{}
)
