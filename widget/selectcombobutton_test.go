package widget_test

import (
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestSelectComboButton_SetSelectedEntry(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.SelectComboButtonEntrySelectedEventArgs
	numEvents := 0

	b := newSelectComboButton(t,
		widget.SelectComboButtonOpts.EntrySelectedHandler(func(args *widget.SelectComboButtonEntrySelectedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	entry := "foo"
	b.SetSelectedEntry(entry)
	event.ExecuteDeferred()

	is.Equal(b.SelectedEntry(), entry)
	is.Equal(eventArgs.Entry, entry)
	is.Equal(b.Label(), "label foo")

	b.SetSelectedEntry(entry)
	event.ExecuteDeferred()

	is.Equal(numEvents, 1)

	entry2 := "bar"
	b.SetSelectedEntry(entry2)
	event.ExecuteDeferred()

	is.Equal(eventArgs.PreviousEntry, entry)
}

func TestSelectComboButton_ContentVisible_Click(t *testing.T) {
	is := is.New(t)

	b := newSelectComboButton(t)

	leftMouseButtonClick(b, t)
	is.True(b.ContentVisible())

	leftMouseButtonClick(b, t)
	is.True(!b.ContentVisible())
}

func TestSelectComboButton_ContentVisible_Programmatic(t *testing.T) {
	is := is.New(t)

	b := newSelectComboButton(t)

	b.SetContentVisible(true)
	event.ExecuteDeferred()

	is.True(b.ContentVisible())

	b.SetContentVisible(false)
	event.ExecuteDeferred()

	is.True(!b.ContentVisible())
}

func newSelectComboButton(t *testing.T, opts ...widget.SelectComboButtonOpt) *widget.SelectComboButton {
	t.Helper()

	b := widget.NewSelectComboButton(append(opts,
		widget.SelectComboButtonOpts.ComboButtonOpts(
			widget.ComboButtonOpts.ButtonOpts(
				widget.ButtonOpts.Image(&widget.ButtonImage{
					Idle:    newNineSliceEmpty(t),
					Pressed: newNineSliceEmpty(t),
				}),
				widget.ButtonOpts.TextAndImage("", loadFont(t), &widget.GraphicImage{
					Idle:     newImageEmpty(t),
					Disabled: newImageEmpty(t),
				}, &widget.ButtonTextColor{
					Idle:     color.Transparent,
					Disabled: color.Transparent,
				}),
			),
			widget.ComboButtonOpts.Content(newButton(t)),
		),
		widget.SelectComboButtonOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return "label " + result
		}),
	)...)

	event.ExecuteDeferred()
	render(b, t)
	return b
}
