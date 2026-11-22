package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// clicker is a widget at 0,0 to 100,100 that counts its double clicks of
// each button, driven by the stand-in devices.
type clicker struct {
	w       *widget.Widget
	d       *uitest.Devices
	doubles map[ebiten.MouseButton]int
	at      []img.Point // where each double click was, from the widget's corner
}

func newClicker(t *testing.T) *clicker {
	t.Helper()
	c := &clicker{d: uitest.Use(t), doubles: map[ebiten.MouseButton]int{}}
	c.w = widget.NewWidget(widget.WidgetOpts.MouseButtonDoubleClickedHandler(func(a *widget.MouseButtonDoubleClickedEventArgs) {
		c.doubles[a.Button]++
		c.at = append(c.at, img.Pt(a.OffsetX, a.OffsetY))
	}))
	c.w.Rect = img.Rect(0, 0, 100, 100)
	return c
}

// update runs one update of the input and the widget.
func (c *clicker) update() {
	input.Update()
	c.w.Update(&widget.UpdateObject{})
	input.AfterUpdate()
	event.ExecuteDeferred()
}

// wait runs n updates with nothing pressed.
func (c *clicker) wait(n int) {
	for range n {
		c.update()
	}
}

// click presses button at x, y for an update and lets go for another.
func (c *clicker) click(b ebiten.MouseButton, x, y int) {
	c.d.Cursor = img.Pt(x, y)
	c.d.Mouse[b] = true
	c.update()
	c.d.Mouse[b] = false
	c.update()
}

// tap touches x, y for an update and lifts for another.
func (c *clicker) tap(x, y int) {
	c.d.Touches = append(c.d.Touches[:0], img.Pt(x, y))
	c.update()
	c.d.Touches = c.d.Touches[:0]
	c.update()
}

// Two clicks of a button on a widget, the second within the double click's
// interval of the first and its distance, double click it; a third starts
// a new pair. For ebitenui/ebitenui#211.
func TestWidget_DoubleClick(t *testing.T) {
	is := is.New(t)

	c := newClicker(t)
	c.click(ebiten.MouseButtonLeft, 10, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 0) // one click
	c.click(ebiten.MouseButtonLeft, 12, 8)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 1)
	is.Equal(c.at, []img.Point{{12, 8}}) // where the second was
	c.click(ebiten.MouseButtonLeft, 12, 8)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 1) // a third: a new pair
	c.click(ebiten.MouseButtonLeft, 12, 8)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 2)

	c.click(ebiten.MouseButtonRight, 50, 50)
	c.click(ebiten.MouseButtonRight, 50, 50)
	is.Equal(c.doubles[ebiten.MouseButtonRight], 1) // the right button too

	c.click(ebiten.MouseButtonLeft, 50, 50)
	c.click(ebiten.MouseButtonRight, 50, 50)
	is.Equal(c.doubles[ebiten.MouseButtonLeft]+c.doubles[ebiten.MouseButtonRight], 3) // two buttons: none
}

// Clicks too far apart in time or place, or with a press off the widget
// between them, are two clicks; the defaults are 500 ms, as Windows has
// it, counted in updates at the TPS, and 4 pixels along each axis.
func TestWidget_DoubleClickWithin(t *testing.T) {
	is := is.New(t)

	updates := ebiten.TPS() / 2
	c := newClicker(t)
	c.click(ebiten.MouseButtonLeft, 10, 10)
	c.wait(updates - 2) // the second click lets go 500 ms after the first
	c.click(ebiten.MouseButtonLeft, 10, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 1)

	c.click(ebiten.MouseButtonLeft, 10, 10)
	c.wait(updates - 1) // a frame late
	c.click(ebiten.MouseButtonLeft, 10, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 1)
	c.click(ebiten.MouseButtonLeft, 10, 10) // the late one's pair
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 2)

	c.click(ebiten.MouseButtonLeft, 10, 10)
	c.click(ebiten.MouseButtonLeft, 15, 10) // 5 pixels across
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 2)
	c.click(ebiten.MouseButtonLeft, 15, 14) // the far one's pair, 4 down
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 3)
	c.click(ebiten.MouseButtonLeft, 10, 10)
	c.click(ebiten.MouseButtonLeft, 10, 15) // 5 down
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 3)
	c.click(ebiten.MouseButtonLeft, 50, 50) // a new first, far away

	c.click(ebiten.MouseButtonLeft, 99, 10)
	c.click(ebiten.MouseButtonLeft, 101, 10) // off the widget, between
	c.click(ebiten.MouseButtonLeft, 99, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 3)
	c.click(ebiten.MouseButtonLeft, 99, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 4)
}

// Taps double tap as clicks double click, though a finger lands less
// precisely: 24 pixels apart by default. The settings are the game's.
func TestWidget_DoubleTapAndSettings(t *testing.T) {
	is := is.New(t)

	c := newClicker(t)
	c.tap(10, 10)
	c.tap(30, 34)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 1)
	c.tap(10, 10)
	c.tap(35, 10) // 25 across
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 1)

	old := widget.DoubleClick
	t.Cleanup(func() { widget.DoubleClick = old })
	widget.DoubleClick.Distance = 20
	c.click(ebiten.MouseButtonLeft, 10, 10)
	c.click(ebiten.MouseButtonLeft, 30, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 2)
	widget.DoubleClick.Interval = 0 // off
	c.click(ebiten.MouseButtonLeft, 10, 10)
	c.click(ebiten.MouseButtonLeft, 10, 10)
	is.Equal(c.doubles[ebiten.MouseButtonLeft], 2)
}
