package widget_test

import (
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
)

// The caret blinks in updates, not wall-clock time: it shows for 450 ms,
// 27 updates at 60 a second, hides as long, and shows again at once when
// reset, as typing does. It allocates nothing as it blinks.
func TestCaret_BlinksInUpdates(t *testing.T) {
	c := widget.NewCaret(widget.CaretOpts.Size(10, 2), widget.CaretOpts.Color(color.White))
	c.ResetBlinking()
	update := func(n int) bool {
		for range n {
			c.Update(&widget.UpdateObject{})
		}
		return c.Visible()
	}
	for i, s := range []struct {
		updates int
		visible bool
	}{
		{0, true},
		{26, true},
		{1, false}, // 27 updates in
		{26, false},
		{1, true},
		{30, false},
	} {
		if got := update(s.updates); got != s.visible {
			t.Fatalf("step %d, %d more updates: the caret shows %v, want %v", i, s.updates, got, s.visible)
		}
	}
	c.ResetBlinking()
	if !c.Visible() || !update(26) || update(1) {
		t.Error("after a reset the caret doesn't show for 27 updates, then hide")
	}
	if n := testing.AllocsPerRun(100, func() { c.Update(&widget.UpdateObject{}) }); n != 0 {
		t.Errorf("an update of the caret allocated %v times, want 0", n)
	}
}
