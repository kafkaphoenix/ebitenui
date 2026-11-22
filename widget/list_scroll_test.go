package widget_test

import (
	"strconv"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// ScrollToIndex scrolls the list at once, as little as shows entry i's row
// whole: one below the view comes to its bottom, one above to its top, one
// in view stays; an index out of range scrolls nothing.
func TestList_ScrollToIndex(t *testing.T) {
	is := is.New(t)

	entries := make([]any, 1000)
	for i := range entries {
		entries[i] = strconv.Itoa(i)
	}
	l := newList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryLabelFunc(label))
	screen := ebiten.NewImage(200, 300)
	draw := func() {
		for range 2 {
			l.Render(screen)
		}
	}
	view := func() (int, int) {
		v := l.ScrollContainer().ViewRect()
		return v.Min.Y, v.Max.Y
	}
	row := func(i int) (int, int) {
		r := l.Rows().RowRect(i)
		return r.Min.Y, r.Max.Y
	}

	l.ScrollToIndex(500)
	draw()
	is.Equal(shownLabel(l, 500), "500")
	_, vBottom := view()
	_, rBottom := row(500)
	is.Equal(rBottom, vBottom) // at the bottom

	l.ScrollToIndex(400)
	draw()
	vTop, _ := view()
	rTop, _ := row(400)
	is.Equal(rTop, vTop) // at the top

	at := l.ScrollContainer().ScrollTop
	l.ScrollToIndex(401)
	l.ScrollToIndex(-1)
	l.ScrollToIndex(1000)
	draw()
	is.Equal(l.ScrollContainer().ScrollTop, at) // in view, or none: no scroll
	h0, h1 := l.Rows().RowSpan(0)
	is.Equal(l.VSlider().Current, 400*(h1-h0)) // the slider at row 400's top

	l.ScrollToIndex(0)
	draw()
	is.Equal(l.ScrollContainer().ScrollTop, 0.0)
}
