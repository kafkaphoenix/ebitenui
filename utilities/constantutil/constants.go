// Package constantutil makes pointers to values, which Go does itself
// since 1.26.
package constantutil

// ConstantToPointer returns a pointer to a copy of input.
//
// Deprecated: new(input) does the same since Go 1.26; go fix rewrites
// calls to it.
//
//go:fix inline
func ConstantToPointer[T any](input T) *T {
	return new(input)
}
