package widget_test

import (
	img "image"
	"image/color"
	"testing"
	"time"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// A carousel's arrows show the game's icons, and its dots the game's
// image, given as options or by a theme; a dot as wide as its image.
func TestCarousel_OwnImages(t *testing.T) {
	is := is.New(t)

	prev := &widget.GraphicImage{Idle: ebiten.NewImage(8, 8), Disabled: ebiten.NewImage(8, 8)}
	next := &widget.GraphicImage{Idle: ebiten.NewImage(8, 8), Disabled: ebiten.NewImage(8, 8)}
	dot := ebiten.NewImage(12, 12)
	c, _ := newCarousel(t, 3, widget.CarouselOpts.ArrowGraphics(prev, next), widget.CarouselOpts.DotImage(dot))
	is.True(c.PrevButton().GraphicImage() == prev)
	is.True(c.NextButton().GraphicImage() == next)
	is.True(c.Dot() == dot)
	is.Equal(c.DotRect(0).Dx(), 12)

	themed := widget.NewCarousel(widget.CarouselOpts.Pages(newSimpleWidget(10, 10, nil)))
	themed.GetWidget().SetTheme(&widget.Theme{CarouselTheme: &widget.CarouselParams{PrevGraphic: prev, NextGraphic: next, DotImage: dot}})
	themed.Validate()
	is.True(themed.PrevButton().GraphicImage() == prev)
	is.True(themed.Dot() == dot)

	plain, _ := newCarousel(t, 3)
	is.True(plain.PrevButton().GraphicImage() != nil) // its own arrows
	is.Equal(plain.DotRect(0).Dx(), 8)
}

// A loading spinner's dots are the game's image, or with SpinnerImage one
// image turns, 24 steps a turn, telling UI.Changed at each.
func TestLoading_OwnImages(t *testing.T) {
	is := is.New(t)

	dot := ebiten.NewImage(6, 6)
	l := newLoading(t, img.Rect(0, 0, 32, 32), widget.LoadingOpts.Spinner(), widget.LoadingOpts.DotImage(dot))
	is.True(l.Dot() == dot)

	icon := ebiten.NewImage(32, 32)
	s := newLoading(t, img.Rect(0, 0, 32, 32), widget.LoadingOpts.Spinner(), widget.LoadingOpts.SpinnerImage(icon), widget.LoadingOpts.Period(time.Second))
	is.True(s.ComputedParams().SpinnerImage == icon)
	is.Equal(marks(s, ebiten.TPS()), 24)

	themed := widget.NewLoading(widget.LoadingOpts.Spinner())
	themed.GetWidget().SetTheme(&widget.Theme{LoadingTheme: &widget.LoadingParams{DotImage: dot}})
	themed.Validate()
	is.True(themed.Dot() == dot)
}

// A color picker's marker is the game's image, and its swatches' frames
// the game's, plain and for the color picked.
func TestColorPicker_OwnImages(t *testing.T) {
	is := is.New(t)

	marker := ebiten.NewImage(10, 10)
	plain, chosen := eimage.NewNineSliceColor(color.Black), eimage.NewNineSliceColor(color.White)
	p := widget.NewColorPicker(widget.ColorPickerOpts.Marker(marker),
		widget.ColorPickerOpts.Swatches(20, color.NRGBA{255, 0, 0, 255}),
		widget.ColorPickerOpts.SwatchFrames(plain, chosen))
	p.GetWidget()
	is.True(p.Ring() == marker)
	f := p.SwatchFrames()
	is.True(f[0][0] == plain && f[1][0] == chosen && f[0][1] == nil && f[1][1] == nil)

	sw := p.Container().Children()[2].(*widget.Container).Children()[0].(*widget.ColorSwatch)
	is.True(sw.Fill() != nil) // the color drawn inside the frame
	sw.Widget().Rect = img.Rect(0, 0, 20, 20)
	sw.Render(ebiten.NewImage(20, 20))
}
