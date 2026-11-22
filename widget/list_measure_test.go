package widget_test

import (
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/matryer/is"
)

// named is an entry whose label changes.
type named struct{ name string }

func nameOf(e any) string { return e.(*named).name }

// entriesNamed returns n entries named "a", the one at wide named a long
// name.
func entriesNamed(n, wide int) []any {
	e := make([]any, n)
	for i := range e {
		e[i] = &named{"a"}
	}
	e[wide].(*named).name = strings.Repeat("wide ", 20)
	return e
}

// A list measures its labels as their rows come into view, so a wide label
// further down widens it only then; with MeasureLabels it is as wide as
// its widest label from the start, and stays so as entries come, go and
// change.
func TestList_MeasureLabels(t *testing.T) {
	is := is.New(t)

	lazy := newList(t, widget.ListOpts.Entries(entriesNamed(100, 90)), widget.ListOpts.EntryLabelFunc(nameOf))
	entries := entriesNamed(100, 90)
	all := newList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryLabelFunc(nameOf), widget.ListOpts.MeasureLabels())
	lazyW, _ := lazy.Rows().PreferredSize()
	wide, _ := all.Rows().PreferredSize()
	is.True(wide > 2*lazyW) // the wide label, out of view, counted

	all.RemoveEntry(entries[90])
	narrow, _ := all.Rows().PreferredSize()
	is.Equal(narrow, lazyW) // it narrows as the widest goes

	entries[5].(*named).name = strings.Repeat("wide ", 20)
	all.UpdateEntry(entries[5])
	w, _ := all.Rows().PreferredSize()
	is.Equal(w, wide) // and widens as one changes

	all.SetEntries(entriesNamed(3, 1))
	w, _ = all.Rows().PreferredSize()
	is.Equal(w, wide)
	all.AddEntry(&named{strings.Repeat("wider ", 20)})
	w, _ = all.Rows().PreferredSize()
	is.True(w > wide)
}

// A list measuring its labels may have at most ListMeasureLabelsMax
// entries: one made with more, or given more, panics.
func TestList_MeasureLabelsAtMost(t *testing.T) {
	panics := func(name string, f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s: no panic", name)
			}
		}()
		f()
	}
	over := entriesNamed(widget.ListMeasureLabelsMax+1, 0)
	opts := []widget.ListOpt{widget.ListOpts.EntryLabelFunc(nameOf), widget.ListOpts.MeasureLabels()}
	panics("made with more", func() { widget.NewList(append(opts, widget.ListOpts.Entries(over))...) })

	l := newList(t, append(opts, widget.ListOpts.Entries(over[:widget.ListMeasureLabelsMax]))...)
	panics("given one more", func() { l.AddEntry(over[widget.ListMeasureLabelsMax]) })
	panics("set to more", func() { l.SetEntries(over) })

	// Without it, as many as wanted.
	newList(t, widget.ListOpts.EntryLabelFunc(nameOf), widget.ListOpts.Entries(over))
}
