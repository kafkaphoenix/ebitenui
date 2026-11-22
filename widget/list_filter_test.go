package widget_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// numbers returns the entries "0" to n-1.
func numbers(n int) []any {
	e := make([]any, n)
	for i := range e {
		e[i] = strconv.Itoa(i)
	}
	return e
}

func containing(s string) func(any) bool {
	return func(e any) bool { return strings.Contains(e.(string), s) }
}

// A filter shows only the entries it keeps, as a search does; Entries
// stays every entry, and entries added, removed or set are filtered too.
// Without a filter, every entry shows.
func TestList_Filter(t *testing.T) {
	is := is.New(t)

	l := newList(t, widget.ListOpts.Entries(numbers(100)), widget.ListOpts.EntryLabelFunc(label))
	l.SetFilter(containing("1"))
	render(l, t)
	l.Render(ebiten.NewImage(200, 300))
	is.Equal(len(l.ShownEntries()), 19) // 1, 10 to 19, 21, 31 ... 91
	is.Equal(len(l.Entries()), 100)
	is.Equal(shownLabel(l, 0), "1")
	is.Equal(shownLabel(l, 1), "10")

	l.AddEntry("100")
	l.AddEntry("200") // kept out
	is.Equal(len(l.ShownEntries()), 20)
	is.Equal(len(l.Entries()), 102)
	l.RemoveEntry("10")
	is.Equal(len(l.ShownEntries()), 19)
	is.Equal(len(l.Entries()), 101)

	l.SetEntries([]any{"a1", "b", "c1"})
	is.Equal(l.ShownEntries(), []any{"a1", "c1"})

	l.SetFilter(nil)
	is.Equal(l.ShownEntries(), []any{"a1", "b", "c1"})
}

// The entries shown are the ones chosen from: a click on the second row
// chooses the second entry shown, and the selection outlives a filter
// that hides it.
func TestList_FilterChooses(t *testing.T) {
	is := is.New(t)

	var chosen any
	l := newList(t, widget.ListOpts.Entries(numbers(30)), widget.ListOpts.EntryLabelFunc(label),
		widget.ListOpts.EntrySelectedHandler(func(a *widget.ListEntrySelectedEventArgs) { chosen = a.Entry }))
	l.SetFilter(containing("2"))
	screen := ebiten.NewImage(200, 300)
	for range 2 {
		l.Render(screen)
	}
	is.Equal(shownLabel(l, 1), "12") // the rows show the entries shown
	clickListEntry(t, l, 1)
	event.ExecuteDeferred()
	is.Equal(chosen, "12")
	l.SetFilter(containing("3"))
	is.Equal(l.SelectedEntry(), "12") // hidden, still chosen
}

// A list made with a filter shows only what it keeps from the start; a
// list measuring its labels counts every entry against its limit.
func TestList_FilterOption(t *testing.T) {
	is := is.New(t)

	l := newList(t, widget.ListOpts.Entries(numbers(30)), widget.ListOpts.EntryLabelFunc(label), widget.ListOpts.EntryFilter(containing("2")))
	is.Equal(len(l.ShownEntries()), 12) // 2, 12, 20 to 29

	defer func() {
		if recover() == nil {
			t.Error("a measuring list of too many entries, most filtered out, didn't panic")
		}
	}()
	widget.NewList(widget.ListOpts.Entries(numbers(widget.ListMeasureLabelsMax+1)), widget.ListOpts.EntryLabelFunc(label),
		widget.ListOpts.MeasureLabels(), widget.ListOpts.EntryFilter(containing("x")))
}
