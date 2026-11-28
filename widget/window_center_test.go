package widget_test

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// openWindow opens w in a UI on a 640 by 480 screen and draws a frame.
func openWindow(t *testing.T, w *widget.Window) (*ebitenui.UI, *uitest.Driver) {
	t.Helper()
	ui := &ebitenui.UI{Container: widget.NewContainer()}
	drv := uitest.New(t, 640, 480, ui.Update, ui.Draw)
	ui.AddWindow(w)
	drv.Frames(1)
	return ui, drv
}

// A centered window shows in the middle of the screen at its size, and is
// placed there again when its size changes or the screen's.
func TestWindow_Centered(t *testing.T) {
	is := is.New(t)
	w := widget.NewWindow(
		widget.WindowOpts.Contents(widget.NewContainer()),
		widget.WindowOpts.Location(image.Rect(0, 0, 200, 100)),
		widget.WindowOpts.Centered(),
	)
	ui, drv := openWindow(t, w)
	is.Equal(shownRect(w), image.Rect(220, 190, 420, 290)) // in the middle

	w.SetLocation(image.Rect(0, 0, 300, 200))
	drv.Frames(1)
	is.Equal(shownRect(w), image.Rect(170, 140, 470, 340)) // resized, in the middle again

	ui.Update()
	ui.Draw(ebiten.NewImage(800, 600))
	is.Equal(shownRect(w), image.Rect(250, 200, 550, 400)) // on a larger screen
}

// A centered window given no size takes its contents' size; dynamic, it
// follows its contents' size and stays in the middle.
func TestWindow_CenteredAtItsContentsSize(t *testing.T) {
	is := is.New(t)
	fixed := widget.NewWindow(
		widget.WindowOpts.Contents(widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(120, 80)))),
		widget.WindowOpts.Centered(),
	)
	openWindow(t, fixed)
	is.Equal(shownRect(fixed), image.Rect(260, 200, 380, 280)) // its contents' size

	contents := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(120, 80)))
	w := widget.NewWindow(widget.WindowOpts.Contents(contents), widget.WindowOpts.Centered(), widget.WindowOpts.Dynamic())
	_, drv := openWindow(t, w)
	is.Equal(shownRect(w), image.Rect(260, 200, 380, 280))

	contents.GetWidget().MinWidth, contents.GetWidget().MinHeight = 200, 100
	drv.Frames(1)
	is.Equal(shownRect(w), image.Rect(220, 190, 420, 290)) // grown with them, in the middle
}

// A window not centered stays where it is placed.
func TestWindow_NotCentered(t *testing.T) {
	is := is.New(t)
	w := widget.NewWindow(widget.WindowOpts.Contents(widget.NewContainer()), widget.WindowOpts.Location(image.Rect(10, 20, 210, 120)))
	_, drv := openWindow(t, w)
	drv.Frames(1)
	is.Equal(shownRect(w), image.Rect(10, 20, 210, 120))
}
