package event

import (
	"slices"
	"testing"
)

// action is a deferred action that runs a function.
type action func()

func (a action) Do() { a() }

// An action deferred while deferred actions run runs in the same
// ExecuteDeferred, after the one that deferred it.
func TestAddDeferred(t *testing.T) {
	var ran []string
	second := action(func() { ran = append(ran, "second") })
	first := action(func() {
		ran = append(ran, "first")
		AddDeferred(second)
	})

	AddDeferred(first)
	ExecuteDeferred()

	if want := []string{"first", "second"}; !slices.Equal(ran, want) {
		t.Errorf("ran %v, want %v", ran, want)
	}
}
