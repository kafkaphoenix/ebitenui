package widget_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

func TestColorPicker_HSVRoundTrip(t *testing.T) {
	is := is.New(t)

	for _, c := range []color.NRGBA{
		{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}, {128, 128, 128, 255},
		{255, 255, 255, 255}, {0, 0, 0, 255}, {12, 200, 90, 255}, {255, 128, 0, 255},
	} {
		is.Equal(widget.HsvToRGB(widget.RgbToHSV(c)), c) // the color back from its hue, saturation and value
	}
}

// Pressing the wheel picks the hue and saturation under the pointer: red at
// the right of the rim, white in the middle.
func TestColorPicker_PickOnWheel(t *testing.T) {
	is := is.New(t)

	p, changes := newColorPicker(t, widget.ColorPickerOpts.InitialColor(color.Black))
	for _, c := range []struct {
		x, y int
		want color.NRGBA
	}{
		{99, 50, color.NRGBA{255, 0, 0, 255}},
		{50, 50, color.NRGBA{255, 255, 255, 255}},
	} {
		pressWheel(p, c.x, c.y)
		is.True(near(p.Color(), c.want))                       // picked
		is.Equal((*changes)[len(*changes)-1].Color, p.Color()) // reported
		is.True((*changes)[len(*changes)-1].Dragging)          // while the button is down
	}
}

// The slider sets the brightness and keeps the hue and saturation.
func TestColorPicker_Brightness(t *testing.T) {
	is := is.New(t)

	p, changes := newColorPicker(t, widget.ColorPickerOpts.InitialColor(color.NRGBA{255, 128, 0, 255}))
	p.Slider().Current = 128
	render(p, t)

	is.Equal(p.Color(), color.NRGBA{128, 64, 0, 255}) // half as bright, the same hue
	is.Equal(len(*changes), 1)
	is.Equal((*changes)[0].Color, p.Color()) // reported
}

// SetColor changes the color and the slider without reporting a change.
func TestColorPicker_SetColor(t *testing.T) {
	is := is.New(t)

	p, changes := newColorPicker(t)
	p.SetColor(color.NRGBA{0, 0, 200, 255})
	render(p, t)

	is.Equal(p.Color(), color.NRGBA{0, 0, 200, 255})
	is.Equal(p.Slider().Current, 200)
	is.Equal(len(*changes), 0)
}

// Dragging across the wheel follows the pointer, and releasing the button
// reports the last color, no longer dragging.
func TestColorPicker_Drag(t *testing.T) {
	is := is.New(t)

	p, changes := newColorPicker(t)
	pressWheel(p, 50, 50)

	d := uitest.Use(t)
	point(d, 99, 50, true)
	p.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()
	is.True(near(p.Color(), color.NRGBA{255, 0, 0, 255})) // followed to the rim
	is.True((*changes)[len(*changes)-1].Dragging)         // still held

	point(d, 99, 50, false)
	p.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()
	last := (*changes)[len(*changes)-1]
	is.True(!last.Dragging) // released
	is.Equal(last.Color, p.Color())
}

// Pressing a swatch picks its color, done at once, and the swatch of the
// color picked is the one framed.
func TestColorPicker_Swatches(t *testing.T) {
	is := is.New(t)

	red, green := color.NRGBA{255, 0, 0, 255}, color.NRGBA{0, 255, 0, 255}
	p, changes := newColorPicker(t, widget.ColorPickerOpts.InitialColor(red), widget.ColorPickerOpts.Swatches(20, red, green))
	swatches := p.Container().ChildrenForTest()[2].(*widget.Container).ChildrenForTest()
	redSwatch, greenSwatch := swatches[0].(*widget.ColorSwatch), swatches[1].(*widget.ColorSwatch)
	is.True(redSwatch.Chosen())

	greenSwatch.Widget().MouseButtonPressedEvent.Fire(&widget.MouseButtonPressedEventArgs{Widget: greenSwatch.Widget(), Button: ebiten.MouseButtonLeft})
	event.ExecuteDeferred()
	render(p, t)

	is.Equal(p.Color(), green)
	is.Equal(len(*changes), 1)       // reported once, the slider's move included
	is.True(!(*changes)[0].Dragging) // done at once
	is.True(greenSwatch.Chosen() && !redSwatch.Chosen())
}

func newColorPicker(t *testing.T, opts ...widget.ColorPickerOpt) (*widget.ColorPicker, *[]widget.ColorPickerChangedEventArgs) {
	t.Helper()

	var changes []widget.ColorPickerChangedEventArgs
	ns := newNineSliceEmpty(t)
	p := widget.NewColorPicker(append([]widget.ColorPickerOpt{
		widget.ColorPickerOpts.WheelSize(100),
		widget.ColorPickerOpts.SliderHandle(&widget.ButtonImage{Idle: ns, Pressed: ns}, 10),
		widget.ColorPickerOpts.ChangedHandler(func(args *widget.ColorPickerChangedEventArgs) { changes = append(changes, *args) }),
	}, opts...)...)
	p.Validate()
	p.SetLocation(image.Rect(0, 0, 100, 130))
	render(p, t)
	return p, &changes
}

// pressWheel presses the left button on the wheel, x and y from its top left.
func pressWheel(p *widget.ColorPicker, x, y int) {
	p.Wheel().Widget().MouseButtonPressedEvent.Fire(&widget.MouseButtonPressedEventArgs{
		Widget: p.Wheel().Widget(), Button: ebiten.MouseButtonLeft, OffsetX: x, OffsetY: y,
	})
	event.ExecuteDeferred()
}

// near reports whether two colors are within a few levels in each channel.
func near(a, b color.NRGBA) bool {
	d := func(x, y uint8) bool { return int(x)-int(y) <= 6 && int(y)-int(x) <= 6 }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B) && a.A == b.A
}

// With a radius, swatches draw rounded layers: the frames the picker
// shares, rounded by the radius and 2 pixels less, and each its color,
// rounded by 4 pixels less; without one, none.
func TestColorPicker_SwatchRadius(t *testing.T) {
	is := is.New(t)

	rounded, _ := newColorPicker(t, widget.ColorPickerOpts.Swatches(24, color.White), widget.ColorPickerOpts.SwatchRadius(6))
	w, _ := rounded.SwatchFrames()[0][0].MinSize()
	is.Equal(w, 12) // the outer frame: radius 6
	w, _ = rounded.SwatchFrames()[1][1].MinSize()
	is.Equal(w, 8) // the inner frame: radius 4
	sw := swatchesOf(rounded)
	w, _ = sw[0].Fill().MinSize()
	is.Equal(w, 4) // the color: radius 2

	square, _ := newColorPicker(t, widget.ColorPickerOpts.Swatches(24, color.White))
	is.True(swatchesOf(square)[0].Fill() == nil)
}

// swatchesOf returns a picker's swatches.
func swatchesOf(p *widget.ColorPicker) []*widget.ColorSwatch {
	var out []*widget.ColorSwatch
	for _, ch := range p.Container().Children() {
		if row, ok := ch.(*widget.Container); ok {
			for _, s := range row.Children() {
				if s, ok := s.(*widget.ColorSwatch); ok {
					out = append(out, s)
				}
			}
		}
	}
	return out
}

// The ring that marks the color is white on its radius, edged in black
// that shows 0.75 px past the white on each side, so the pixels either side
// are opaque and mostly black; it is clear at its middle and beyond.
func TestColorPicker_RingPixels(t *testing.T) {
	is := is.New(t)

	pix, size := widget.RingPixels(6)
	is.Equal(size, 17)          // the ring, 1.5 px of black and a pixel of smoothing each side
	at := func(x int) [4]byte { // along the middle row; the center lies at 8.5, the radius at 14.5
		i := 4 * (8*size + x)
		return [4]byte{pix[i], pix[i+1], pix[i+2], pix[i+3]}
	}
	is.Equal(at(14), [4]byte{255, 255, 255, 255}) // white on the radius
	is.Equal(at(15), [4]byte{64, 64, 64, 255})    // edged in black outside it
	is.Equal(at(13), [4]byte{64, 64, 64, 255})    // and inside it
	is.Equal(at(16), [4]byte{})                   // clear beyond
	is.Equal(at(12), [4]byte{})                   // and towards the middle
	is.Equal(at(8), [4]byte{})
}

// pressSquare presses the left button on the square, x and y from its top
// left.
func pressSquare(p *widget.ColorPicker, x, y int) {
	p.Square().Widget().MouseButtonPressedEvent.Fire(&widget.MouseButtonPressedEventArgs{
		Widget: p.Square().Widget(), Button: ebiten.MouseButtonLeft, OffsetX: x, OffsetY: y,
	})
	event.ExecuteDeferred()
}

// With Square, pressing the square picks the saturation, left to right, and
// the brightness, top to bottom, of the hue on the slider: the hue itself at
// the top right, white at the top left, black along the bottom.
func TestColorPicker_SquarePick(t *testing.T) {
	is := is.New(t)

	p, changes := newColorPicker(t, widget.ColorPickerOpts.Square(), widget.ColorPickerOpts.InitialColor(color.NRGBA{0, 0, 255, 255}))
	is.True(p.Wheel() == nil) // a square in its place
	for _, c := range []struct {
		x, y int
		want color.NRGBA
	}{
		{99, 0, color.NRGBA{0, 0, 255, 255}},
		{0, 0, color.NRGBA{255, 255, 255, 255}},
		{50, 99, color.NRGBA{0, 0, 0, 255}},
		{50, 50, color.NRGBA{64, 64, 128, 255}},
	} {
		pressSquare(p, c.x, c.y)
		is.True(near(p.Color(), c.want))                       // picked
		is.Equal((*changes)[len(*changes)-1].Color, p.Color()) // reported
		is.True((*changes)[len(*changes)-1].Dragging)          // while the button is down
	}

	d := uitest.Use(t)
	point(d, 99, 0, true) // dragged to the top right
	p.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()
	point(d, 99, 0, false)
	p.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()
	is.True(near(p.Color(), color.NRGBA{0, 0, 255, 255}))
	is.True(!(*changes)[len(*changes)-1].Dragging) // released
}

// With Square, the slider under the square sets the hue, keeping the
// saturation and the brightness. SetColor and the swatches move it to their
// color's hue, but a gray, which has none, keeps the hue picked, so the
// square stays on it.
func TestColorPicker_SquareHue(t *testing.T) {
	is := is.New(t)

	p, changes := newColorPicker(t, widget.ColorPickerOpts.Square(), widget.ColorPickerOpts.InitialColor(color.NRGBA{255, 0, 0, 255}))
	is.Equal(p.Slider().Current, 0) // red
	p.Slider().Current = 120
	render(p, t)
	is.Equal(p.Color(), color.NRGBA{0, 255, 0, 255}) // green, as saturated and bright
	is.Equal(len(*changes), 1)

	p.SetColor(color.NRGBA{0, 0, 255, 255})
	is.Equal(p.Slider().Current, 240) // blue's hue
	p.SetColor(color.NRGBA{128, 128, 128, 255})
	is.Equal(p.Slider().Current, 240) // a gray keeps it
	pressSquare(p, 99, 0)
	is.True(near(p.Color(), color.NRGBA{0, 0, 255, 255})) // the square still blue's
	pressSquare(p, 150, -20)
	is.Equal(p.Color(), color.NRGBA{0, 0, 255, 255}) // past the corner: the corner's

	magenta := color.NRGBA{255, 0, 255, 255}
	p.SetColor(magenta)
	render(p, t)
	is.Equal(p.Slider().Current, 300) // every degree on the slider
	is.Equal(p.Color(), magenta)
}

// The hue slider's track runs through every hue, a degree a pixel, from
// red at the slider's 0.
func TestColorPicker_HueTrack(t *testing.T) {
	is := is.New(t)

	pix := widget.HueTrack()
	is.Equal(len(pix), 4*360)
	at := func(x int) color.NRGBA { return color.NRGBA{pix[4*x], pix[4*x+1], pix[4*x+2], pix[4*x+3]} }
	is.Equal(at(0), color.NRGBA{255, 0, 0, 255})
	is.Equal(at(120), color.NRGBA{0, 255, 0, 255})
	is.Equal(at(240), color.NRGBA{0, 0, 255, 255})
}

// A disabled wheel or square picks nothing when pressed.
func TestColorPicker_DisabledFieldIgnoresPresses(t *testing.T) {
	is := is.New(t)

	wheel, _ := newColorPicker(t, widget.ColorPickerOpts.InitialColor(color.Black))
	wheel.Wheel().Widget().Disabled = true
	pressWheel(wheel, 99, 50)
	is.Equal(wheel.Color(), color.NRGBA{0, 0, 0, 255})

	square, _ := newColorPicker(t, widget.ColorPickerOpts.Square(), widget.ColorPickerOpts.InitialColor(color.Black))
	square.Square().Widget().Disabled = true
	pressSquare(square, 99, 0)
	is.Equal(square.Color(), color.NRGBA{0, 0, 0, 255})
}
