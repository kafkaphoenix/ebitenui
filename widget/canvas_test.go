package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// A canvas calls its draw function with an image covering only its place,
// so what a game draws there, a minimap, a chart, stays inside it, and the
// place itself.
func TestCanvas_DrawsInItsPlace(t *testing.T) {
	is := is.New(t)

	var bounds, at img.Rectangle
	calls := 0
	c := widget.NewCanvas(widget.CanvasOpts.DrawFunc(func(screen *ebiten.Image, rect img.Rectangle) {
		calls++
		bounds, at = screen.Bounds(), rect
	}))
	c.Validate()
	c.SetLocation(img.Rect(20, 30, 120, 90))
	c.Render(ebiten.NewImage(200, 200))
	is.Equal(calls, 1)
	is.Equal(at, img.Rect(20, 30, 120, 90))
	is.Equal(bounds, at) // drawing outside it draws nothing

	c.SetLocation(img.Rect(150, 150, 250, 250)) // partly off the screen
	c.Render(ebiten.NewImage(200, 200))
	is.Equal(bounds, img.Rect(150, 150, 200, 200))
	is.Equal(at, img.Rect(150, 150, 250, 250))  // its whole place, so what it draws keeps its scale
	c.SetLocation(img.Rect(300, 300, 400, 400)) // all off it
	c.Render(ebiten.NewImage(200, 200))
	is.Equal(calls, 2) // not drawn

	widget.NewCanvas().Render(ebiten.NewImage(10, 10)) // no draw function: nothing
}

// Its preferred size is its minimum size, 64 by 64 without one.
func TestCanvas_PreferredSize(t *testing.T) {
	is := is.New(t)

	w, h := widget.NewCanvas().PreferredSize()
	is.Equal([]int{w, h}, []int{64, 64})
	w, h = widget.NewCanvas(widget.CanvasOpts.WidgetOpts(widget.WidgetOpts.MinSize(160, 90))).PreferredSize()
	is.Equal([]int{w, h}, []int{160, 90})
}

// The pointer over it reaches the game through its widget's events, as a
// click to jump a minimap's view; MarkChanged tells UI.Changed the game's
// picture changed.
func TestCanvas_InputAndChanges(t *testing.T) {
	is := is.New(t)

	var clicked img.Point
	c := widget.NewCanvas(widget.CanvasOpts.WidgetOpts(widget.WidgetOpts.MouseButtonClickedHandler(func(a *widget.MouseButtonClickedEventArgs) {
		clicked = img.Pt(a.OffsetX, a.OffsetY)
	})))
	c.Validate()
	c.SetLocation(img.Rect(20, 30, 120, 90))
	d := uitest.Use(t)
	for _, down := range []bool{true, false} {
		d.Cursor, d.Mouse[ebiten.MouseButtonLeft] = img.Pt(50, 40), down
		input.Update()
		c.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
	}
	is.Equal(clicked, img.Pt(30, 10))

	n := changes.Count()
	c.MarkChanged()
	is.Equal(changes.Count(), n+1)

	drawn := false
	c.SetDrawFunc(func(*ebiten.Image, img.Rectangle) { drawn = true })
	is.Equal(changes.Count(), n+2) // a new picture
	c.Render(ebiten.NewImage(200, 200))
	is.True(drawn)
}
