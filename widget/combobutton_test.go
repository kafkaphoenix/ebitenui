package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestComboButton_ContentVisible_Click(t *testing.T) {
	is := is.New(t)

	b := newComboButton(t)

	leftMouseButtonClick(b, t)
	is.True(b.ContentVisible)

	leftMouseButtonClick(b, t)
	is.True(!b.ContentVisible)
}

func newComboButton(t *testing.T, opts ...widget.ComboButtonOpt) *widget.ComboButton {
	t.Helper()

	b := widget.NewComboButton(append(opts, []widget.ComboButtonOpt{
		widget.ComboButtonOpts.ButtonOpts(widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    newNineSliceEmpty(t),
			Pressed: newNineSliceEmpty(t),
		})),
		widget.ComboButtonOpts.Content(newButton(t)),
	}...)...)
	event.ExecuteDeferred()
	render(b, t)
	return b
}
