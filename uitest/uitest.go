// Package uitest plays a user on an interface in tests: Devices stands in
// for the mouse, the keyboard, the gamepads and the touch screen, and a
// Driver runs the interface frame by frame while it clicks, taps and types,
// so a test presses where a player would instead of firing widget events.
//
// The stand-in replaces what the input package reads from Ebitengine; a
// cursor updater set with input.SetCursorUpdater still takes the pointer's
// place, as it does in a game. The input's state, kept from one interface
// to the next in package variables, starts over with each test as with a
// new window: nothing hovered, no input layers.
package uitest

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/input"
	internalinput "github.com/ebitenui/ebitenui/internal/input"
	"github.com/hajimehoshi/ebiten/v2"
)

// Devices is a mouse, a keyboard, gamepads and a screen a test sets. The
// input reads them in each update as it would Ebitengine.
type Devices struct {
	// Cursor is where the mouse pointer is; Mouse is each button held.
	Cursor image.Point
	Mouse  [ebiten.MouseButtonMax + 1]bool
	// WheelX and WheelY are how far the wheel turns, in its units,
	// ScrollX and ScrollY how far that scrolls, in pixels, and Chars the
	// text typed, all by the next update only, which takes them.
	WheelX, WheelY   float64
	ScrollX, ScrollY float64
	Chars            []rune

	Keys [ebiten.KeyMax + 1]bool
	// Pads are the gamepads connected, all with the standard layout; pad i
	// has the ID i.
	Pads []Pad
	// Touches are the fingers on the screen, where each is; their IDs count
	// up from FirstTouchID, so changing it puts down other fingers.
	// Touchscreen makes the platform one without a mouse pointer.
	Touches      []image.Point
	FirstTouchID ebiten.TouchID
	Touchscreen  bool
}

// Pad is a gamepad with the standard layout.
type Pad struct {
	Buttons [ebiten.StandardGamepadButtonMax + 1]bool
	Axes    [ebiten.StandardGamepadAxisMax + 1]float64
}

// Use makes the input read the returned devices until the test ends, and
// then updates it once with nothing pressed. The input's state starts over
// as Use is called and again as the test ends (see the package's doc).
func Use(t testing.TB) *Devices {
	t.Helper()
	startOver()
	d := &Devices{}
	h := internalinput.InputHandler
	old := h.Devices
	h.Devices = d
	t.Cleanup(func() {
		*d = Devices{}
		h.Update()
		h.AfterUpdate()
		h.Devices = old
		startOver()
	})
	return d
}

// startOver clears what an interface leaves in the input's package state
// between frames: whether it was hovered, its input layers and those
// deferred, the transform it is drawn to the screen with, and a gesture of
// fingers just lifted. Each update starts the cursor's shape over itself.
func startOver() {
	input.UIHovered = false
	input.SetupInputLayersWithDeferred(nil)
	internalinput.InputHandler.SetScreenTransform(ebiten.GeoM{})
	internalinput.InputHandler.TouchGesture = false
}

func (d *Devices) IsMouseButtonPressed(b ebiten.MouseButton) bool { return d.Mouse[b] }

func (d *Devices) CursorPosition() (int, int) { return d.Cursor.X, d.Cursor.Y }

func (d *Devices) Wheel() (float64, float64) {
	x, y := d.WheelX, d.WheelY
	d.WheelX, d.WheelY = 0, 0
	return x, y
}

func (d *Devices) ScrollDelta() (float64, float64) {
	x, y := d.ScrollX, d.ScrollY
	d.ScrollX, d.ScrollY = 0, 0
	return x, y
}

func (d *Devices) AppendInputChars(chars []rune) []rune {
	chars = append(chars, d.Chars...)
	d.Chars = d.Chars[:0]
	return chars
}

func (d *Devices) IsKeyPressed(k ebiten.Key) bool { return d.Keys[k] }

func (d *Devices) AppendGamepadIDs(ids []ebiten.GamepadID) []ebiten.GamepadID {
	for i := range d.Pads {
		ids = append(ids, ebiten.GamepadID(i))
	}
	return ids
}

func (d *Devices) IsStandardGamepadButtonPressed(id ebiten.GamepadID, b ebiten.StandardGamepadButton) bool {
	return int(id) < len(d.Pads) && d.Pads[id].Buttons[b]
}

func (d *Devices) StandardGamepadAxisValue(id ebiten.GamepadID, a ebiten.StandardGamepadAxis) float64 {
	if int(id) >= len(d.Pads) {
		return 0
	}
	return d.Pads[id].Axes[a]
}

func (d *Devices) AppendTouchIDs(ids []ebiten.TouchID) []ebiten.TouchID {
	for i := range d.Touches {
		ids = append(ids, d.FirstTouchID+ebiten.TouchID(i))
	}
	return ids
}

func (d *Devices) TouchPosition(id ebiten.TouchID) (int, int) {
	i := int(id - d.FirstTouchID)
	if i < 0 || i >= len(d.Touches) {
		return 0, 0
	}
	return d.Touches[i].X, d.Touches[i].Y
}

func (d *Devices) IsTouchscreen() bool { return d.Touchscreen }
