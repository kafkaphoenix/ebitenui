package event

import "testing"

type nopAction struct {
}

func BenchmarkExecuteDeferred(b *testing.B) {
	a := &nopAction{}
	for b.Loop() {
		for range 100 {
			AddDeferred(a)
		}
		ExecuteDeferred()
	}
}

func (a *nopAction) Do() {
	// empty
}
