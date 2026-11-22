package widget_test

import (
	"image"
	"image/color"
	"strconv"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// statesInput is a text input holding "text" at 0,0 to 200,20, of an
// image and a color for each state, and the stand-in devices.
func statesInput(t *testing.T, images *widget.TextInputImage, colors *widget.TextInputColor, opts ...widget.TextInputOpt) (*widget.TextInput, *uitest.Devices) {
	t.Helper()
	d := uitest.Use(t)
	ti := widget.NewTextInput(append(opts, widget.TextInputOpts.Face(loadFont(t)),
		widget.TextInputOpts.Image(images), widget.TextInputOpts.Color(colors))...)
	ti.SetText("text")
	ti.GetWidget().Rect = image.Rect(0, 0, 200, 20)
	event.ExecuteDeferred()
	render(ti, t)
	return ti, d
}

// at moves the pointer to x with the left button down or not, runs an
// update of ti and draws it, and returns the image and the text color it
// drew with.
func at(t *testing.T, ti *widget.TextInput, d *uitest.Devices, x int, down bool) (*eimage.NineSlice, color.Color) {
	t.Helper()
	point(d, x, 10, down)
	ti.Update(&widget.UpdateObject{})
	input.AfterUpdate()
	event.ExecuteDeferred()
	ti.Render(ebiten.NewImage(200, 20))
	return ti.StateImage(), ti.Text().ComputedParams().Color
}

func nineSlices(n int) []*eimage.NineSlice {
	s := make([]*eimage.NineSlice, n)
	for i := range s {
		s[i] = eimage.NewNineSliceColor(color.NRGBA{uint8(i), 0, 0, 255})
	}
	return s
}

// A text input shows Hover while hovered or focused, Pressed while the
// left button pressed on it is held, PressedHover for both, Disabled while
// disabled, and PressedDisabled if pressed then; its text takes the Hover
// and Pressed colors. For ebitenui/ebitenui#341.
func TestTextInput_HoverAndPressedImages(t *testing.T) {
	is := is.New(t)

	i := nineSlices(6)
	images := &widget.TextInputImage{Idle: i[0], Disabled: i[1], Hover: i[2], Pressed: i[3], PressedHover: i[4], PressedDisabled: i[5]}
	idleC, hoverC, pressedC := color.NRGBA{1, 1, 1, 255}, color.NRGBA{2, 2, 2, 255}, color.NRGBA{3, 3, 3, 255}
	colors := &widget.TextInputColor{Idle: idleC, Caret: idleC, Hover: hoverC, Pressed: pressedC}
	ti, d := statesInput(t, images, colors)

	img, c := at(t, ti, d, 300, false)
	is.Equal([]any{img, c}, []any{i[0], idleC}) // away
	img, c = at(t, ti, d, 50, false)
	is.Equal([]any{img, c}, []any{i[2], hoverC}) // over it
	img, c = at(t, ti, d, 50, true)
	is.Equal([]any{img, c}, []any{i[4], pressedC}) // pressed on it
	img, c = at(t, ti, d, 50, false)
	is.Equal([]any{img, c}, []any{i[2], hoverC}) // let go: focused and over it
	img, c = at(t, ti, d, 300, false)
	is.Equal([]any{img, c}, []any{i[2], idleC}) // away, still focused

	ti.Focus(false)
	img, _ = at(t, ti, d, 300, false)
	is.Equal(img, i[0])
	ti.GetWidget().Disabled = true
	img, _ = at(t, ti, d, 50, false)
	is.Equal(img, i[1]) // disabled, over it
}

// Without a Hover image a press shows Pressed; disabled with the button
// held from before, PressedDisabled.
func TestTextInput_PressedImages(t *testing.T) {
	is := is.New(t)

	i := nineSlices(4)
	images := &widget.TextInputImage{Idle: i[0], Disabled: i[1], Pressed: i[2], PressedDisabled: i[3]}
	ti, d := statesInput(t, images, &widget.TextInputColor{Idle: color.White, Caret: color.White})

	img, _ := at(t, ti, d, 50, false)
	is.Equal(img, i[0]) // over it, no Hover
	img, _ = at(t, ti, d, 50, true)
	is.Equal(img, i[2])
	ti.GetWidget().Disabled = true
	img, _ = at(t, ti, d, 50, true)
	is.Equal(img, i[3]) // held, now disabled
	img, _ = at(t, ti, d, 50, false)
	is.Equal(img, i[1])
}

// A theme's Hover and Pressed images and colors reach a text input that
// gives none of its own.
func TestTextInput_ThemeStates(t *testing.T) {
	is := is.New(t)

	i := nineSlices(6)
	hoverC, pressedC := color.NRGBA{2, 2, 2, 255}, color.NRGBA{3, 3, 3, 255}
	ti := widget.NewTextInput()
	ti.GetWidget().SetTheme(&widget.Theme{DefaultFace: loadFont(t), TextInputTheme: &widget.TextInputParams{
		Image: &widget.TextInputImage{Idle: i[0], Disabled: i[1], Hover: i[2], Pressed: i[3], PressedHover: i[4], PressedDisabled: i[5]},
		Color: &widget.TextInputColor{Idle: color.White, Caret: color.White, Hover: hoverC, Pressed: pressedC},
	}})
	ti.Validate()
	p := ti.ComputedParams()
	is.Equal([]any{p.Image.Hover, p.Image.Pressed, p.Image.PressedHover, p.Image.PressedDisabled, p.Color.Hover, p.Color.Pressed},
		[]any{i[2], i[3], i[4], i[5], hoverC, pressedC})
}

// A text input tells of the pointer entering, moving over and leaving it,
// and of the left button pressed on it, let go and clicking it, as a
// button does; disabled, it tells of no press.
func TestTextInput_PointerEvents(t *testing.T) {
	is := is.New(t)

	var got []string
	ti, d := statesInput(t, &widget.TextInputImage{Idle: nineSlices(1)[0]}, &widget.TextInputColor{Idle: color.White, Caret: color.White},
		widget.TextInputOpts.CursorEnteredHandler(func(a *widget.TextInputHoverEventArgs) {
			got = append(got, "entered", strconv.Itoa(a.OffsetX))
		}),
		widget.TextInputOpts.CursorMovedHandler(func(a *widget.TextInputHoverEventArgs) {
			got = append(got, "moved", strconv.Itoa(a.DiffX))
		}),
		widget.TextInputOpts.CursorExitedHandler(func(*widget.TextInputHoverEventArgs) { got = append(got, "exited") }),
		widget.TextInputOpts.PressedHandler(func(a *widget.TextInputPressedEventArgs) {
			got = append(got, "pressed", strconv.Itoa(a.OffsetX))
		}),
		widget.TextInputOpts.ReleasedHandler(func(a *widget.TextInputReleasedEventArgs) {
			got = append(got, "released", map[bool]string{true: "in", false: "out"}[a.Inside])
		}),
		widget.TextInputOpts.ClickedHandler(func(*widget.TextInputClickedEventArgs) { got = append(got, "clicked") }),
	)
	at(t, ti, d, 300, false)
	at(t, ti, d, 50, false)
	at(t, ti, d, 70, false)
	at(t, ti, d, 70, true)
	at(t, ti, d, 70, false)
	at(t, ti, d, 300, false)
	is.Equal(got, []string{"entered", "50", "moved", "50", "moved", "20", "pressed", "70", "released", "in", "clicked", "exited"})

	got = nil
	at(t, ti, d, 70, false)
	at(t, ti, d, 70, true)
	at(t, ti, d, 300, true)
	at(t, ti, d, 300, false)
	is.Equal(got, []string{"entered", "70", "pressed", "70", "exited", "released", "out"}) // let go off it, the exit first

	got = nil
	ti.GetWidget().Disabled = true
	at(t, ti, d, 50, false)
	at(t, ti, d, 50, true)
	at(t, ti, d, 50, false)
	is.Equal(got, []string{"moved", "-20"}) // disabled: no hover, press or click, only the move
}
