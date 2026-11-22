package event_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/internal/event"
)

type nopAction struct {
}

func BenchmarkExecuteDeferred(b *testing.B) {
	a := &nopAction{}
	for b.Loop() {
		for range 100 {
			event.AddDeferred(a)
		}
		event.ExecuteDeferred()
	}
}

func (a *nopAction) Do() {
	// empty
}
