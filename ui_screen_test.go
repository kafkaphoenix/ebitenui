package ebitenui_test

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// A UI drawn on an image of the game's, which the game draws to the
// screen enlarged and moved (ScreenTransform), takes the pointer and the
// fingers where it shows: on the screen where the button shows, a click
// or a tap clicks it, while the button's place in the UI's own pixels
// clicks nothing. Without a transform, the two are one. For
// ebitenui/ebitenui#14, #285 and #66.
func TestUI_ScreenTransform(t *testing.T) {
	is := is.New(t)

	clicks := 0
	b := focusButton(image.Rect(10, 10, 50, 30), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { clicks++ }))
	ui := focusUI(b)
	var g ebiten.GeoM
	g.Scale(2, 2)
	g.Translate(100, 50)
	ui.ScreenTransform = g
	d := uitest.New(t, 400, 300, ui.Update, ui.Draw)
	d.Frames(1)

	d.Click(image.Pt(30, 20)) // the button's place in the UI
	is.Equal(clicks, 0)
	d.Click(image.Pt(100+2*30, 50+2*20)) // where it shows on the screen
	is.Equal(clicks, 1)
	d.Tap(image.Pt(100+2*12, 50+2*28)) // a finger, near its corner
	is.Equal(clicks, 2)
	d.Tap(image.Pt(100+2*52, 50+2*20)) // past its right edge
	is.Equal(clicks, 2)

	ui.ScreenTransform = ebiten.GeoM{}
	d.Click(image.Pt(30, 20))
	is.Equal(clicks, 3)

	var flat ebiten.GeoM
	flat.Scale(0, 0) // can't be undone: the pointer as is
	ui.ScreenTransform = flat
	d.Click(image.Pt(30, 20))
	is.Equal(clicks, 4)
}
