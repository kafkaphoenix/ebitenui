package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestGraphic_PreferredSize(t *testing.T) {
	is := is.New(t)

	i := newImageEmptySize(47, 11, t)
	g := newGraphic(t, widget.GraphicOpts.Image(i))
	w, h := g.PreferredSize()
	is.Equal(w, i.Bounds().Dx())
	is.Equal(h, i.Bounds().Dy())
}

// A graphic of images is as big as its idle image.
func TestGraphic_PreferredSizeOfImages(t *testing.T) {
	is := is.New(t)

	g := newGraphic(t, widget.GraphicOpts.Images(&widget.GraphicImage{
		Idle: newImageEmptySize(47, 11, t), Hover: newImageEmptySize(60, 60, t)}))
	w, h := g.PreferredSize()
	is.Equal([]int{w, h}, []int{47, 11})
}

// A graphic draws Image while it is set, otherwise its images for its
// state, and picking one allocates nothing.
func TestGraphic_DrawsItsImage(t *testing.T) {
	is := is.New(t)

	idle, disabled, single := newImageEmptySize(2, 2, t), newImageEmptySize(2, 2, t), newImageEmptySize(3, 3, t)
	g := newGraphic(t, widget.GraphicOpts.Images(&widget.GraphicImage{Idle: idle, Disabled: disabled}))
	is.Equal(g.ImageForTest(), idle)
	g.GetWidget().Disabled = true
	is.Equal(g.ImageForTest(), disabled)

	g.Image = single
	is.Equal(g.ImageForTest(), single)
	is.Equal(testing.AllocsPerRun(10, func() { g.ImageForTest() }), 0.0)
	render(g, t)
	g.Image = nil
	is.Equal(g.ImageForTest(), disabled) // its images again
}

// A graphic of images shows Hover while the pointer is over it, and
// Pressed while the left button, pressed on it, is held; disabled, it
// shows Disabled, and without an image for a state, Idle. For
// ebitenui/ebitenui#218.
func TestGraphic_HoverAndPressed(t *testing.T) {
	is := is.New(t)

	idle, hover, pressed, disabled := newImageEmptySize(2, 2, t), newImageEmptySize(2, 2, t), newImageEmptySize(2, 2, t), newImageEmptySize(2, 2, t)
	d := uitest.Use(t)
	shown := func(g *widget.Graphic, x int, down bool) *ebiten.Image {
		point(d, x, 10, down)
		g.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
		return g.ImageForTest()
	}

	g := newGraphic(t, widget.GraphicOpts.Images(&widget.GraphicImage{Idle: idle, Hover: hover, Pressed: pressed, Disabled: disabled}))
	g.SetLocation(img.Rect(0, 0, 20, 20))
	is.Equal(shown(g, 50, false), idle)   // away
	is.Equal(shown(g, 10, false), hover)  // over it
	is.Equal(shown(g, 10, true), pressed) // pressed on it
	is.Equal(shown(g, 10, true), pressed) // held
	is.Equal(shown(g, 10, false), hover)  // let go
	is.Equal(shown(g, 50, false), idle)   // away
	is.Equal(shown(g, 50, true), idle)    // pressed elsewhere
	is.Equal(shown(g, 50, false), idle)

	d.Mouse[ebiten.MouseButtonRight] = true
	is.Equal(shown(g, 10, false), hover) // the right button pressed on it
	d.Mouse[ebiten.MouseButtonRight] = false
	is.Equal(shown(g, 10, true), pressed)
	d.Mouse[ebiten.MouseButtonRight] = true
	is.Equal(shown(g, 10, true), pressed)
	d.Mouse[ebiten.MouseButtonRight] = false
	is.Equal(shown(g, 10, true), pressed) // the right button let go, the left held
	is.Equal(shown(g, 10, false), hover)

	g.GetWidget().Disabled = true
	is.Equal(shown(g, 10, false), disabled) // over it, disabled
	is.Equal(shown(g, 10, true), disabled)

	plain := newGraphic(t, widget.GraphicOpts.Images(&widget.GraphicImage{Idle: idle}))
	plain.SetLocation(img.Rect(0, 0, 20, 20))
	is.Equal(shown(plain, 10, false), idle) // no Hover
	is.Equal(shown(plain, 10, true), idle)  // no Pressed

	noDisabled := newGraphic(t, widget.GraphicOpts.Images(&widget.GraphicImage{Idle: idle, Hover: hover}))
	noDisabled.SetLocation(img.Rect(0, 0, 20, 20))
	noDisabled.GetWidget().Disabled = true
	is.Equal(shown(noDisabled, 10, false), idle) // disabled without Disabled: Idle, not Hover
}

// A finger shows Pressed while it touches a graphic of images: it goes
// down without the pointer entering first.
func TestGraphic_PressedByATouch(t *testing.T) {
	is := is.New(t)

	idle, pressed := newImageEmptySize(2, 2, t), newImageEmptySize(2, 2, t)
	g := newGraphic(t, widget.GraphicOpts.Images(&widget.GraphicImage{Idle: idle, Pressed: pressed}))
	g.SetLocation(img.Rect(0, 0, 20, 20))
	d := uitest.Use(t)
	touch := func(on bool) *ebiten.Image {
		d.Touches = d.Touches[:0]
		if on {
			d.Touches = append(d.Touches, img.Pt(10, 10))
		}
		input.Update()
		g.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
		return g.ImageForTest()
	}
	is.Equal(touch(true), pressed)
	is.Equal(touch(true), pressed)
	is.Equal(touch(false), idle)
}

func newGraphic(t *testing.T, opts ...widget.GraphicOpt) *widget.Graphic {
	t.Helper()

	g := widget.NewGraphic(opts...)
	event.ExecuteDeferred()
	render(g, t)
	return g
}
