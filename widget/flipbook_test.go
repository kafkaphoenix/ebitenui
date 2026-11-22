package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestFlipBook_SetPage_AlwaysRenderSingleWidget(t *testing.T) {
	is := is.New(t)

	f := newFlipBook(t)

	pages := make([]*control, 0, 3)
	var order []*control

	for range 3 {
		p := newControl()
		p.onRender = func() { order = append(order, p) }
		pages = append(pages, p)
	}

	expectedOrder := []*control{pages[0], pages[1], pages[2], pages[0]}

	for _, p := range expectedOrder {
		f.SetPage(p)
		render(f, t)
	}

	is.Equal([]int{pages[0].renders, pages[1].renders, pages[2].renders}, []int{2, 1, 1})

	is.Equal(order, expectedOrder)
}

func newFlipBook(t *testing.T, opts ...widget.FlipBookOpt) *widget.FlipBook {
	t.Helper()

	f := widget.NewFlipBook(opts...)
	event.ExecuteDeferred()
	render(f, t)
	return f
}
