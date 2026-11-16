package uitest

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// A Driver runs an interface frame by frame on a screen of its own, an
// update and then a draw each, as a game's loop does: the draw lays the
// interface out, so after any frame its widgets' rectangles are where they
// show. Its methods play a user through the Devices, which a test may also
// set by hand between frames, to drag for one.
type Driver struct {
	*Devices
	update func()
	draw   func(screen *ebiten.Image)
	screen *ebiten.Image
}

// New makes the input read the driver's devices until the test ends (see
// Use) and drives an interface on a screen width by height: update and draw
// run one frame of it, ui.Update and ui.Draw for an ebitenui.UI, or a game's
// own around one.
func New(t testing.TB, width, height int, update func(), draw func(screen *ebiten.Image)) *Driver {
	t.Helper()
	return &Driver{Devices: Use(t), update: update, draw: draw, screen: ebiten.NewImage(width, height)}
}

// Frames runs n frames.
func (d *Driver) Frames(n int) {
	for range n {
		d.update()
		d.draw(d.screen)
	}
}

// Click moves the pointer to at, presses the left button for a frame and
// lets go of it for another.
func (d *Driver) Click(at image.Point) {
	d.Cursor = at
	d.Mouse[ebiten.MouseButtonLeft] = true
	d.Frames(1)
	d.Mouse[ebiten.MouseButtonLeft] = false
	d.Frames(1)
}

// ClickIn clicks in the middle of r, such as a widget's rectangle.
func (d *Driver) ClickIn(r image.Rectangle) {
	d.Click(image.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2))
}

// Tap puts a finger down at at for a frame and lifts it for another.
func (d *Driver) Tap(at image.Point) {
	d.Touches = append(d.Touches[:0], at)
	d.Frames(1)
	d.Touches = d.Touches[:0]
	d.Frames(1)
}

// Key presses k for a frame and lets go of it for another.
func (d *Driver) Key(k ebiten.Key) {
	d.Keys[k] = true
	d.Frames(1)
	d.Keys[k] = false
	d.Frames(1)
}

// Type types s in one frame, as a fast typist's update can bring.
func (d *Driver) Type(s string) {
	d.Chars = append(d.Chars, []rune(s)...)
	d.Frames(1)
}
