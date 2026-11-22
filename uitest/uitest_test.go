package uitest_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/matryer/is"
	"golang.org/x/image/font/basicfont"
)

// The mouse reads as the test sets it: where the cursor is, and each
// button pressed, just pressed in the update it goes down and just
// released in the update it comes up.
func TestMouse(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	d.Cursor = image.Pt(5, 6)
	b := ebiten.MouseButtonRight
	for i, s := range []struct{ down, pressed, just, released bool }{
		{false, false, false, false},
		{true, true, true, false},  // goes down
		{true, true, false, false}, // held
		{false, false, false, true},
	} {
		d.Mouse[b] = s.down
		input.Update()
		x, y := input.CursorPosition()
		is.Equal([]any{i, x, y, input.MouseButtonPressed(b), input.MouseButtonJustPressed(b), input.MouseButtonJustReleased(b)},
			[]any{i, 5, 6, s.pressed, s.just, s.released})
		input.AfterUpdate()
	}
}

// The wheel turns and the text typed count in the next update only, as
// Ebitengine reports each update's.
func TestWheelAndTextOnce(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	d.WheelX, d.WheelY = 1, -2
	d.Chars = []rune("hé")
	input.Update()
	wx, wy := input.Wheel()
	is.Equal([]any{wx, wy, string(input.Chars())}, []any{1.0, -2.0, "hé"})
	input.AfterUpdate()

	input.Update()
	wx, wy = input.Wheel()
	is.Equal([]any{wx, wy, len(input.Chars())}, []any{0.0, 0.0, 0})
	input.AfterUpdate()
}

// When the test ends, the input reads Ebitengine again, and nothing the
// test held stays pressed.
func TestUseEnds(t *testing.T) {
	t.Run("holding", func(t *testing.T) {
		d := uitest.Use(t)
		d.Mouse[ebiten.MouseButtonLeft], d.Keys[ebiten.KeyA] = true, true
		input.Update()
		input.AfterUpdate()
	})
	input.Update()
	defer input.AfterUpdate()
	if input.MouseButtonPressed(ebiten.MouseButtonLeft) || input.KeyPressed(ebiten.KeyA) {
		t.Error("the button or key the test held is still pressed")
	}
}

// newUI is a window of a button and a text input, one above the other.
func newUI(t *testing.T) (*uitest.Driver, *widget.Button, *widget.TextInput, *int) {
	t.Helper()
	clicks := new(int)
	look := eimage.NewNineSliceColor(color.White)
	button := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 30)),
		widget.ButtonOpts.Image(&widget.ButtonImage{Idle: look, Pressed: look}),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { *clicks++ }),
	)
	face := text.Face(text.NewGoXFace(basicfont.Face7x13))
	field := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 30)),
		widget.TextInputOpts.Image(&widget.TextInputImage{Idle: look}),
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.Black, Caret: color.Black}),
	)
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	root.AddChild(button, field)
	ui := &ebitenui.UI{Container: root}
	drv := uitest.New(t, 200, 100, ui.Update, ui.Draw)
	drv.Frames(1)
	return drv, button, field, clicks
}

// A click in a button clicks it, once; a click beside it does not.
func TestDriverClick(t *testing.T) {
	is := is.New(t)

	drv, button, _, clicks := newUI(t)
	r := button.GetWidget().Rect
	drv.Click(image.Pt(r.Max.X+20, r.Min.Y+5))
	is.Equal(*clicks, 0)
	drv.ClickIn(r)
	is.Equal(*clicks, 1)
	is.True(!drv.Mouse[ebiten.MouseButtonLeft]) // let go
}

// A tap on a touch screen clicks the button under it.
func TestDriverTap(t *testing.T) {
	is := is.New(t)

	drv, button, _, clicks := newUI(t)
	drv.Touchscreen = true
	r := button.GetWidget().Rect
	drv.Tap(image.Pt(r.Min.X+5, r.Min.Y+5))
	is.Equal(*clicks, 1)
	is.Equal(len(drv.Touches), 0) // lifted
}

// Text typed into a focused input lands in it, and a key pressed acts on
// it: Backspace takes the last letter back.
func TestDriverTypeAndKey(t *testing.T) {
	is := is.New(t)

	drv, _, field, _ := newUI(t)
	drv.ClickIn(field.GetWidget().Rect)
	drv.Type("abc")
	is.Equal(field.GetText(), "abc")
	drv.Key(ebiten.KeyBackspace)
	is.Equal(field.GetText(), "ab")
	is.True(!drv.Keys[ebiten.KeyBackspace]) // let go
}

// Frames runs the interface's update and draw, as many times as asked.
func TestDriverFrames(t *testing.T) {
	is := is.New(t)

	var updates, draws int
	drv := uitest.New(t, 10, 10, func() { updates++ }, func(screen *ebiten.Image) {
		is.Equal(screen.Bounds().Size(), image.Pt(10, 10))
		draws++
	})
	drv.Frames(3)
	is.Equal([]int{updates, draws}, []int{3, 3})
}

// Each test starts with the input as a new window does, nothing hovered
// and no input layers, whatever the test before it left, and leaves it so.
func TestUseStartsOver(t *testing.T) {
	left := &input.Layer{EventTypes: input.LayerEventTypeAll, FullScreen: true}
	t.Run("a test that leaves state", func(t *testing.T) {
		uitest.Use(t)
		input.UIHovered = true
		input.AddLayer(left)
	})
	if input.UIHovered || left.Valid() {
		t.Errorf("after a test, hovered %v and its layer valid %v, want neither", input.UIHovered, left.Valid())
	}

	input.UIHovered = true
	input.AddLayer(left)
	uitest.Use(t)
	if input.UIHovered || left.Valid() {
		t.Errorf("as a test starts, hovered %v and an earlier layer valid %v, want neither", input.UIHovered, left.Valid())
	}
}
