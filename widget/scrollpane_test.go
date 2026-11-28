package widget_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/matryer/is"
)

// paneRows returns a column of n buttons 100 by 20 pixels, which take the
// focus, and the column.
func paneRows(n int) (*widget.Container, []*widget.Button) {
	column := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	rows := make([]*widget.Button, n)
	for i := range rows {
		rows[i] = widget.NewButton(widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 20)))
		column.AddChild(rows[i])
	}
	return column, rows
}

// showPane shows a pane over content, its view at most 200 pixels tall, in
// the dark theme on a 640 by 480 screen, with more options, and draws two
// frames of it.
func showPane(t *testing.T, content widget.PreferredSizeLocateableWidget, opts ...widget.ScrollPaneOpt) (*widget.ScrollPane, *uitest.Driver) {
	t.Helper()
	t.Cleanup(func() { widget.SetUITheme(nil) })
	pane := widget.NewScrollPane(append([]widget.ScrollPaneOpt{
		widget.ScrollPaneOpts.Content(content),
		widget.ScrollPaneOpts.MaxHeight(200),
		widget.ScrollPaneOpts.Spacing(6),
		widget.ScrollPaneOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: eimage.NewNineSliceColor(color.Transparent),
			Mask: eimage.NewNineSliceColor(color.White),
		})),
	}, opts...)...)
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	root.AddChild(pane)
	ui := &ebitenui.UI{Container: root, PrimaryTheme: themes.GetBasicDarkTheme()}
	drv := uitest.New(t, 640, 480, ui.Update, ui.Draw)
	drv.Frames(2)
	return pane, drv
}

// Content taller than the cap shows in a view the cap's height, with the
// scrollbar beside it; content that fits shows whole, the scrollbar hidden
// but keeping its room, so the pane is as wide either way.
func TestScrollPane_ShowsTheScrollbarWhileNeeded(t *testing.T) {
	is := is.New(t)
	long, _ := paneRows(30)
	tall, _ := showPane(t, long)
	short, _ := paneRows(5)
	fits, _ := showPane(t, short)

	is.Equal(tall.View().ViewRect().Dy(), 200)                                         // the view at the cap
	is.True(tall.Scrollbar().GetWidget().IsVisible())                                  // the scrollbar shown
	is.Equal(fits.View().ViewRect().Dy(), 100)                                         // the content whole
	is.True(!fits.Scrollbar().GetWidget().IsVisible())                                 // the scrollbar hidden
	is.Equal(fits.GetWidget().Rect.Dx(), tall.GetWidget().Rect.Dx())                   // as wide
	is.Equal(tall.GetWidget().Rect.Dx(), 100+6+tall.Scrollbar().GetWidget().Rect.Dx()) // the view, the gap and the bar
}

// In a pane wider than its content, the content is as wide as the view.
func TestScrollPane_StretchesTheContent(t *testing.T) {
	is := is.New(t)
	long, _ := paneRows(30)
	pane, _ := showPane(t, long, widget.ScrollPaneOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
		widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
	)))
	view := pane.View().ViewRect()
	is.Equal(pane.GetWidget().Rect.Dx(), 640) // stretched
	is.Equal(long.GetWidget().Rect.Dx(), view.Dx())
}

// The wheel over the view scrolls it the pixels the platform gives, and
// the scrollbar follows; a click on the track scrolls a view's height, and
// moving the scrollbar scrolls the view, to the content's end at the
// bar's.
func TestScrollPane_WheelAndScrollbar(t *testing.T) {
	is := is.New(t)
	long, rows := paneRows(30)
	pane, drv := showPane(t, long)
	view := pane.View().ViewRect()

	drv.Cursor = view.Min.Add(view.Size().Div(2))
	drv.WheelY, drv.ScrollY = -1, -40 // a notch down, 40 pixels
	drv.Frames(2)
	is.Equal(pane.View().ScrollTop, 40.0/400) // 40 of the 400 pixels below the view
	is.Equal(pane.Scrollbar().Current, 40)    // the bar with it

	drv.WheelY = -1 // a notch from a wheel that gives no pixels
	drv.Frames(2)
	is.Equal(pane.Scrollbar().Current, 40+66) // a third of the view, the rest of a pixel waiting

	bar, handle := pane.Scrollbar().GetWidget().Rect, pane.Scrollbar().Handle().GetWidget().Rect
	drv.Click(image.Pt(bar.Min.X+bar.Dx()/2, (handle.Max.Y+bar.Max.Y)/2)) // on the track, under the handle
	drv.Frames(1)
	is.Equal(pane.Scrollbar().Current, 40+66+200) // a view down

	pane.Scrollbar().Current = pane.Scrollbar().Max
	drv.Frames(2)
	is.Equal(pane.View().ScrollTop, 1.0)                                    // at the end
	is.Equal(rows[29].GetWidget().Rect.Max.Y, pane.View().ViewRect().Max.Y) // the last row at the view's bottom
}

// A widget inside given the focus is scrolled into view, halfway down it
// or where FocusAt says, and the scrollbar follows; one losing the focus
// leaves the view where it is.
func TestScrollPane_FollowsTheFocus(t *testing.T) {
	is := is.New(t)
	long, rows := paneRows(30)
	pane, drv := showPane(t, long)

	rows[20].Focus(true)
	drv.Frames(2)
	view, row := pane.View().ViewRect(), rows[20].GetWidget().Rect
	is.Equal(row.Min.Y-view.Min.Y, (view.Dy()-row.Dy())/2)                 // halfway down
	is.Equal(pane.Scrollbar().Current, int(pane.View().ScrollTop*400+0.5)) // the bar with it

	pane.Scrollbar().Current = 0
	drv.Frames(2)
	rows[20].Focus(false)
	drv.Frames(2)
	is.Equal(pane.View().ScrollTop, 0.0) // left at the top

	other, others := paneRows(30)
	atTop, drv := showPane(t, other, widget.ScrollPaneOpts.FocusAt(0))
	others[20].Focus(true)
	drv.Frames(2)
	is.Equal(others[20].GetWidget().Rect.Min.Y, atTop.View().ViewRect().Min.Y) // at the top
}

// A new cap takes effect at the next layout; content that shrinks to fit
// is scrolled back to its top, the scrollbar hidden, and growing again
// shows it.
func TestScrollPane_FitsAgain(t *testing.T) {
	is := is.New(t)
	long, rows := paneRows(30)
	pane, drv := showPane(t, long)
	pane.SetMaxHeight(300)
	drv.Frames(2)
	is.Equal(pane.View().ViewRect().Dy(), 300) // the new cap
	is.Equal(pane.Scrollbar().Max, 300)        // the 600 pixels' 300 below the view

	pane.Scrollbar().Current = 200
	drv.Frames(2)
	is.True(pane.View().ScrollTop > 0) // scrolled

	for _, r := range rows[5:] {
		long.RemoveChild(r)
	}
	drv.Frames(3)
	is.Equal(pane.View().ScrollTop, 0.0)                                   // back at the top
	is.True(!pane.Scrollbar().GetWidget().IsVisible())                     // the bar hidden
	is.Equal(rows[0].GetWidget().Rect.Min.Y, pane.View().ViewRect().Min.Y) // the first row at the top

	for _, r := range rows[5:] {
		long.AddChild(r)
	}
	drv.Frames(3)
	is.True(pane.Scrollbar().GetWidget().IsVisible()) // the bar back
}
