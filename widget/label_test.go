package widget_test

import (
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestLabel_SetLabel(t *testing.T) {
	is := is.New(t)

	l := newLabel(t)

	l.Label = "foo"
	render(l, t)

	is.Equal(labelText(l).Label, "foo")
}

func newLabel(t *testing.T, opts ...widget.LabelOpt) *widget.Label {
	t.Helper()

	l := widget.NewLabel(append(opts, widget.LabelOpts.Text("", loadFont(t), &widget.LabelColor{
		Idle:     color.White,
		Disabled: color.Black,
	}))...)
	event.ExecuteDeferred()
	render(l, t)
	return l
}

func labelText(l *widget.Label) *widget.Text {
	return l.Text()
}
