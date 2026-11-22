package widget_test

import (
	"image"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2"
)

// holdKey holds key down over n updates of a text input, as frames do.
func holdKey(ti *widget.TextInput, d *uitest.Devices, key ebiten.Key, n int) {
	d.Keys[key] = true
	for range n {
		input.Update()
		ti.Update(&widget.UpdateObject{})
		input.AfterUpdate()
	}
	d.Keys[key] = false
	input.Update()
	ti.Update(&widget.UpdateObject{})
	input.AfterUpdate()
}

// A held key's command runs as it goes down, again after the repeat delay
// and then every repeat interval, counted in updates: by default 300 ms,
// 18 updates at 60 a second, then 35 ms, 2 updates. A key let go and held
// again starts over.
func TestTextInput_HeldKeyRepeatsInUpdates(t *testing.T) {
	for name, tc := range map[string]struct {
		opts    []widget.TextInputOpt
		updates int
		want    string
	}{
		"pressed":              {nil, 1, "abcdefghijk"},
		"held to the delay":    {nil, 18, "abcdefghijk"},
		"past the delay":       {nil, 19, "abcdefghij"},
		"three intervals more": {nil, 25, "abcdefg"},
		"a slower interval":    {[]widget.TextInputOpt{widget.TextInputOpts.RepeatInterval(100 * time.Millisecond)}, 31, "abcdefgh"},
	} {
		t.Run(name, func(t *testing.T) {
			d := uitest.Use(t)
			ti := newTextInput(t, tc.opts...)
			ti.SetText("abcdefghijkl")
			ti.Focus(true)
			holdKey(ti, d, ebiten.KeyBackspace, tc.updates)
			if got := ti.GetText(); got != tc.want {
				t.Errorf("Backspace held %d updates left %q, want %q", tc.updates, got, tc.want)
			}
			holdKey(ti, d, ebiten.KeyBackspace, 18)
			if got, want := ti.GetText(), tc.want[:len(tc.want)-1]; got != want {
				t.Errorf("Backspace held again to the delay left %q, want %q", got, want)
			}
		})
	}
}

// An idle text input's update allocates nothing, focused or not, with the
// pointer resting over it.
func TestTextInput_IdleUpdateAllocatesNothing(t *testing.T) {
	d := uitest.Use(t)
	ti := newTextInput(t)
	ti.SetText("some text")
	ti.GetWidget().Rect = image.Rect(0, 0, 200, 30)
	d.Cursor = image.Pt(20, 15)
	for _, focused := range []bool{false, true} {
		ti.Focus(focused)
		n := testing.AllocsPerRun(100, func() {
			input.Update()
			ti.Update(&widget.UpdateObject{})
			input.AfterUpdate()
		})
		if n != 0 {
			t.Errorf("focused %v: an update allocated %v times, want 0", focused, n)
		}
	}
}
