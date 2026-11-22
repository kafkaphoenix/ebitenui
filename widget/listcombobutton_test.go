package widget_test

import (
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestListComboButton_SelectedEntry_Initial(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	l := newListComboButton(t,
		widget.ListComboButtonOpts.Entries(entries),

		widget.ListComboButtonOpts.EntryLabelFunc(
			func(e any) string {
				result, _ := e.(string)
				return "label " + result
			}, func(e any) string {
				result, _ := e.(string)
				return result
			}))

	is.Equal(l.SelectedEntry(), entries[0])
	is.Equal(l.Label(), "label first")
}

func TestListComboButton_SetSelectedEntry(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	var eventArgs *widget.ListComboButtonEntrySelectedEventArgs
	numEvents := 0

	l := newListComboButton(t,
		widget.ListComboButtonOpts.Entries(entries),

		widget.ListComboButtonOpts.EntrySelectedHandler(func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			eventArgs = args
			numEvents++
		}),

		widget.ListComboButtonOpts.EntryLabelFunc(
			func(e any) string {
				result, _ := e.(string)
				return "label " + result
			}, func(e any) string {
				result, _ := e.(string)
				return result
			}))

	l.SetSelectedEntry(entries[1])
	event.ExecuteDeferred()

	is.Equal(l.SelectedEntry(), entries[1])
	is.Equal(eventArgs.Entry, entries[1])
	is.Equal(eventArgs.PreviousEntry, entries[0])
	is.Equal(l.Label(), "label second")

	l.SetSelectedEntry(entries[1])
	event.ExecuteDeferred()
	is.Equal(numEvents, 2)
}

func TestListComboButton_EntrySelectedEvent_User(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	var eventArgs *widget.ListComboButtonEntrySelectedEventArgs
	numEvents := 0

	l := newListComboButton(t,
		widget.ListComboButtonOpts.Entries(entries),

		widget.ListComboButtonOpts.EntrySelectedHandler(func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			eventArgs = args
			numEvents++
		}),

		widget.ListComboButtonOpts.EntryLabelFunc(
			func(e any) string {
				result, _ := e.(string)
				return "label " + result
			}, func(e any) string {
				result, _ := e.(string)
				return result
			}))

	render(l, t)

	l.SetContentVisible(true)
	clickListEntry(t, listComboButtonContentList(l), 1)

	is.Equal(l.SelectedEntry(), entries[1])
	is.Equal(eventArgs.Entry, entries[1])
	is.Equal(eventArgs.PreviousEntry, entries[0])
	is.Equal(l.Label(), "label second")

	l.SetContentVisible(true)
	clickListEntry(t, listComboButtonContentList(l), 1)

	is.Equal(numEvents, 3)
}

func TestListComboButton_ContentVisible_Click(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	l := newListComboButton(t,
		widget.ListComboButtonOpts.Entries(entries),

		widget.ListComboButtonOpts.EntryLabelFunc(
			func(e any) string {
				result, _ := e.(string)
				return result
			}, func(e any) string {
				result, _ := e.(string)
				return result
			}))

	render(l, t)
	leftMouseButtonClick(l.Button(), t)
	is.True(l.ContentVisible())

	leftMouseButtonClick(l.Button(), t)
	is.True(!l.ContentVisible())
}

func TestListComboButton_ContentVisible_Programmatic(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	l := newListComboButton(t,
		widget.ListComboButtonOpts.Entries(entries),

		widget.ListComboButtonOpts.EntryLabelFunc(
			func(e any) string {
				result, _ := e.(string)
				return result
			}, func(e any) string {
				result, _ := e.(string)
				return result
			}))

	l.SetContentVisible(true)
	is.True(l.ContentVisible())

	l.SetContentVisible(false)
	is.True(!l.ContentVisible())
}

func newListComboButton(t *testing.T, opts ...widget.ListComboButtonOpt) *widget.ListComboButton {
	t.Helper()

	l := widget.NewListComboButton(append(opts, []widget.ListComboButtonOpt{
		widget.ListComboButtonOpts.ButtonParams(&widget.ButtonParams{
			Image: &widget.ButtonImage{
				Idle:    newNineSliceEmpty(t),
				Pressed: newNineSliceEmpty(t),
			}},
		),
		widget.ListComboButtonOpts.ListParams(&widget.ListParams{
			ScrollContainerImage: &widget.ScrollContainerImage{
				Idle:     newNineSliceEmpty(t),
				Disabled: newNineSliceEmpty(t),
				Mask:     newNineSliceEmpty(t),
			},
			Slider: &widget.SliderParams{
				TrackImage: &widget.SliderTrackImage{},
				HandleImage: &widget.ButtonImage{
					Idle:    newNineSliceEmpty(t),
					Pressed: newNineSliceEmpty(t),
				},
			},
			EntryColor: &widget.ListEntryColor{
				Unselected:                 color.Transparent,
				Selected:                   color.Transparent,
				DisabledUnselected:         color.Transparent,
				DisabledSelected:           color.Transparent,
				SelectedBackground:         color.Transparent,
				DisabledSelectedBackground: color.Transparent,
			},
			EntryFace: loadFont(t),
		}),
		widget.ListComboButtonOpts.Text(loadFont(t), &widget.GraphicImage{
			Idle:     newImageEmpty(t),
			Disabled: newImageEmpty(t),
		}, &widget.ButtonTextColor{
			Idle:     color.Transparent,
			Disabled: color.Transparent,
		}),
	}...)...)

	event.ExecuteDeferred()
	render(l, t)
	return l
}

func listComboButtonContentList(l *widget.ListComboButton) *widget.List {
	result, _ := l.Button().Button().Content().(*widget.List)
	return result
}

// The open list of a ListComboButton shows its entries as the widgets
// EntryWidgetFunc builds.
func TestListComboButton_EntryWidgets(t *testing.T) {
	is := is.New(t)

	var built []any
	l := newListComboButton(t,
		widget.ListComboButtonOpts.Entries([]any{"a", "b"}),
		widget.ListComboButtonOpts.EntryLabelFunc(func(e any) string { return e.(string) }, nil),
		widget.ListComboButtonOpts.EntryWidgetFunc(func(e any) widget.PreferredSizeLocateableWidget {
			built = append(built, e)
			return newSimpleWidget(30, 15, nil)
		}),
	)

	is.Equal(built, []any{"a", "b"})
	is.Equal(len(l.List().Rows().Widgets()), 2) // widgets in place of labels
}
