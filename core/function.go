package core

// Identity returns the identity function for T: a function that
// returns its argument unchanged.
func Identity[T any]() func(T) T {
	return func(v T) T { return v }
}
