package ebitenui_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// box is a leaf w by h pixels.
func box(w, h int) *widget.Container {
	return widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(w, h)))
}

// column returns a container laying widgets out one under the other.
func column(widgets ...widget.PreferredSizeLocateableWidget) *widget.Container {
	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(5))))
	c.AddChild(widgets...)
	return c
}

// placed is a widget of a game's that places a child of its own.
type placed struct {
	*widget.Container
	child  *widget.Container
	placed int
}

func (p *placed) PlaceChildren() {
	p.placed++
	r := p.GetWidget().Rect
	p.child.SetLocation(image.Rect(r.Min.X+1, r.Min.Y+1, r.Min.X+11, r.Min.Y+11))
}

// UI.Layout lays the interface out without drawing: every widget in its
// containers, scroll containers, lists, tab books, flip books, text areas
// and windows, and in a game's widgets placing children (ChildPlacer), is
// where Draw then puts it. For ebitenui/ebitenui#233.
func TestUI_Layout(t *testing.T) {
	is := is.New(t)

	inScroll := box(50, 30)
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(column(box(50, 30), inScroll)),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{Idle: eimage.NewNineSliceColor(color.Black), Mask: eimage.NewNineSliceColor(color.White)}),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(200, 40)),
	)
	inTab := box(40, 20)
	tab := widget.NewTabBookTab(widget.TabBookTabOpts.Label("Tab"),
		widget.TabBookTabOpts.ContainerOpts(widget.ContainerOpts.Layout(widget.NewRowLayout())))
	tab.AddChild(inTab)
	tabs := widget.NewTabBook(widget.TabBookOpts.Tabs(tab))
	inFlip := box(30, 30)
	flip := widget.NewFlipBook()
	flip.SetPage(column(inFlip))
	entryWidget := box(80, 12)
	list := widget.NewList(widget.ListOpts.Entries([]any{"a"}),
		widget.ListOpts.EntryWidgetFunc(func(any) widget.PreferredSizeLocateableWidget { return entryWidget }),
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(150, 60))))
	area := widget.NewTextArea(widget.TextAreaOpts.Text("some text"), widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(150, 40))))
	own := &placed{Container: box(20, 20), child: box(10, 10)}

	ui := &ebitenui.UI{Container: column(box(60, 20), scroll, tabs, flip, list, area, own), PrimaryTheme: themes.GetBasicDarkTheme()}
	inWindow := box(25, 25)
	window := widget.NewWindow(widget.WindowOpts.Contents(column(inWindow)), widget.WindowOpts.Location(image.Rect(300, 200, 360, 260)))
	ui.AddWindow(window)

	ui.Layout(400, 600)
	widgets := []widget.PreferredSizeLocateableWidget{inScroll, inTab, inFlip, entryWidget, area, own, own.child, inWindow}
	laidOut := make([]image.Rectangle, len(widgets))
	for i, w := range widgets {
		laidOut[i] = w.GetWidget().Rect
		if laidOut[i].Empty() {
			t.Errorf("widget %d not placed: %v", i, laidOut[i])
		}
	}
	is.Equal(own.placed, 1)
	is.Equal(ui.Container.GetWidget().Rect, image.Rect(0, 0, 400, 600)) // on the screen given
	is.True(inWindow.GetWidget().Rect.In(image.Rect(300, 200, 360, 260)))

	ui.Draw(ebiten.NewImage(400, 600))
	for i, w := range widgets {
		is.Equal(w.GetWidget().Rect, laidOut[i]) // where Draw puts it
	}
}
