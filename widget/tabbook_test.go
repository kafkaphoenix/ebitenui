package widget

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestTabBook_Tab_Initial(t *testing.T) {
	is := is.New(t)

	tab1 := NewTabBookTab(TabBookTabOpts.Label("Tab 1"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))
	tab2 := NewTabBookTab(TabBookTabOpts.Label("Tab 2"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))

	tb := newTabBook(t,
		TabBookOpts.Tabs(tab1, tab2),
		TabBookOpts.TabSelectedHandler(func(e *TabBookTabSelectedEventArgs) {
			is.Equal(e.Tab, tab1)
			is.Equal(e.PreviousTab, nil)
		}))

	is.Equal(tb.Tab(), tab1)
}

func TestTabBook_SetTab(t *testing.T) {
	is := is.New(t)

	var eventArgs *TabBookTabSelectedEventArgs
	numEvents := 0

	tab1 := NewTabBookTab(TabBookTabOpts.Label("Tab 1"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))
	tab2 := NewTabBookTab(TabBookTabOpts.Label("Tab 2"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))

	tb := newTabBook(t,
		TabBookOpts.Tabs(tab1, tab2),
		TabBookOpts.TabSelectedHandler(func(args *TabBookTabSelectedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	tb.SetTab(tab2)
	event.ExecuteDeferred()

	is.Equal(tb.Tab(), tab2)
	is.Equal(eventArgs.Tab, tab2)
	is.Equal(eventArgs.PreviousTab, tab1)

	tb.SetTab(tab2)
	event.ExecuteDeferred()
	is.Equal(numEvents, 2)
}

func TestTabBook_TabSelectedEvent_User(t *testing.T) {
	is := is.New(t)

	var eventArgs *TabBookTabSelectedEventArgs
	numEvents := 0

	tab1 := NewTabBookTab(TabBookTabOpts.Label("Tab 1"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))
	tab2 := NewTabBookTab(TabBookTabOpts.Label("Tab 2"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))

	tb := newTabBook(t,
		TabBookOpts.Tabs(tab1, tab2),
		TabBookOpts.TabSelectedHandler(func(args *TabBookTabSelectedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	leftMouseButtonClick(tabBookButtons(tb)[1], t)

	is.Equal(tb.Tab(), tab2)
	is.Equal(eventArgs.Tab, tab2)
	is.Equal(eventArgs.PreviousTab, tab1)

	leftMouseButtonClick(tabBookButtons(tb)[1], t)
	is.Equal(numEvents, 2)
}

func TestTabBook_GetTabButton(t *testing.T) {
	is := is.New(t)

	tab1 := NewTabBookTab(TabBookTabOpts.Label("Tab 1"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))
	tab2 := NewTabBookTab(TabBookTabOpts.Label("Tab 2"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))
	tab3 := NewTabBookTab(TabBookTabOpts.Label("Tab 3"), TabBookTabOpts.ContainerOpts(newTabContainerOpts()...))

	tb := newTabBook(t,
		TabBookOpts.Tabs(tab1, tab2),
	)
	// Test Nil Pointer
	is.True(tb.GetTabButton(nil) == nil)

	// Test existing tab
	is.True(tb.GetTabButton(tab1) != nil)
	is.True(tb.GetTabButton(tab1).Text() != nil)
	is.Equal(tb.GetTabButton(tab1).Text().Label, "Tab 1")

	// Test existing tab
	is.True(tb.GetTabButton(tab2) != nil)
	is.True(tb.GetTabButton(tab2).Text() != nil)
	is.Equal(tb.GetTabButton(tab2).Text().Label, "Tab 2")

	// Test non-existing tab
	is.True(tb.GetTabButton(tab3) == nil)
}

func newTabBook(t *testing.T, opts ...TabBookOpt) *TabBook {
	t.Helper()

	tb := NewTabBook(append(opts, []TabBookOpt{
		TabBookOpts.TabButtonImage(&ButtonImage{
			Idle:    newNineSliceEmpty(t),
			Pressed: newNineSliceEmpty(t),
		}),
		TabBookOpts.TabButtonText(loadFont(t), &ButtonTextColor{
			Idle:     color.Transparent,
			Disabled: color.Transparent,
		}),
	}...)...)

	event.ExecuteDeferred()
	render(tb, t)
	return tb
}

func tabBookButtons(t *TabBook) []*Button {
	buttons := make([]*Button, 0, len(t.tabs))
	for _, tab := range t.tabs {
		buttons = append(buttons, t.tabToButton[tab])
	}
	return buttons
}

func newTabContainerOpts() []ContainerOpt {
	result := []ContainerOpt{}

	return result
}

// newPlacedTabBook returns a tab book of two tabs, "A" and "Longer", whose
// pages ask for 60 by 80 and 100 by 120 px, so that the one not shown is the
// larger, with opts, laid out in 300 by 200 px.
func newPlacedTabBook(t *testing.T, opts ...TabBookOpt) *TabBook {
	t.Helper()
	page := func(label string, w, h int) *TabBookTab {
		return NewTabBookTab(TabBookTabOpts.Label(label), TabBookTabOpts.ContainerOpts(ContainerOpts.WidgetOpts(WidgetOpts.MinSize(w, h))))
	}
	tb := newTabBook(t, append([]TabBookOpt{TabBookOpts.Tabs(page("A", 60, 80), page("Longer", 100, 120)), TabBookOpts.ContentSpacing(5)}, opts...)...)
	tb.SetLocation(img.Rect(0, 0, 300, 200))
	render(tb, t)
	return tb
}

// The tabs' buttons run where TabPosition puts them: across the top, the
// default, or the bottom of the pages, or down their left or right side,
// each button then as wide as the widest. The pages take the rest,
// ContentSpacing apart.
func TestTabBook_TabPosition(t *testing.T) {
	is := is.New(t)

	for _, pos := range []TabPosition{TabPositionTop, TabPositionBottom, TabPositionLeft, TabPositionRight} {
		tb := newPlacedTabBook(t, TabBookOpts.TabPosition(pos))
		b := tabBookButtons(tb)
		b0, b1, page := b[0].GetWidget().Rect, b[1].GetWidget().Rect, tb.flipBook.GetWidget().Rect
		switch pos {
		case TabPositionTop:
			is.True(b0.Min.Y == 0 && b1.Min.Y == 0 && b1.Min.X >= b0.Max.X) // a row at the top
			is.Equal(page, img.Rect(0, b0.Max.Y+5, 300, 200))
		case TabPositionBottom:
			is.True(b0.Max.Y == 200 && b1.Max.Y == 200 && b1.Min.X >= b0.Max.X) // a row at the bottom
			is.Equal(page, img.Rect(0, 0, 300, b0.Min.Y-5))
		case TabPositionLeft:
			is.True(b0.Min.X == 0 && b1.Min.Y >= b0.Max.Y) // a column on the left
			is.Equal(b0.Dx(), b1.Dx())                     // as wide as the widest
			is.Equal(page, img.Rect(b0.Max.X+5, 0, 300, 200))
		case TabPositionRight:
			is.True(b0.Max.X == 300 && b1.Min.Y >= b0.Max.Y) // a column on the right
			is.Equal(b0.Dx(), b1.Dx())
			is.Equal(page, img.Rect(0, 0, b0.Min.X-5, 200))
		}
	}
}

// A tab book asks for room for its largest page and its buttons, beside the
// pages or above them, ContentSpacing apart.
func TestTabBook_PreferredSizeByPosition(t *testing.T) {
	is := is.New(t)

	for _, pos := range []TabPosition{TabPositionTop, TabPositionLeft} {
		tb := newPlacedTabBook(t, TabBookOpts.TabPosition(pos))
		bw, bh := tb.btnContainer.PreferredSize()
		w, h := tb.PreferredSize()
		if pos == TabPositionTop {
			is.Equal([]int{w, h}, []int{max(100, bw), bh + 5 + 120})
		} else {
			is.Equal([]int{w, h}, []int{bw + 5 + 100, max(120, bh)})
		}
	}
}
