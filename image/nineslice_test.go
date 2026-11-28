package image_test

import (
	"image"
	"image/color"
	"testing"

	eimage "github.com/ebitenui/ebitenui/image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

func TestNineSlice_MinSize(t *testing.T) {
	is := is.New(t)

	n := eimage.NewNineSlice(newImageEmptySize(20, 20, t), [3]int{3, 10, 7}, [3]int{2, 16, 2})
	w, h := n.MinSize()
	is.Equal(w, 10)
	is.Equal(h, 4)

	n = eimage.NewNineSliceColor(color.White)
	w, h = n.MinSize()
	is.Equal(w, 0)
	is.Equal(h, 0)

	n = eimage.NewNineSliceColor(color.Transparent)
	w, h = n.MinSize()
	is.Equal(w, 0)
	is.Equal(h, 0)
}

func newImageEmptySize(width int, height int, t *testing.T) *ebiten.Image {
	t.Helper()
	return ebiten.NewImage(width, height)
}

func Test_NewNineSliceColor(t *testing.T) {
	is := is.New(t)

	n := eimage.NewNineSliceColor(nil)
	is.Equal(n.Transparent(), true)

	n = eimage.NewNineSliceColor(color.RGBA{0, 0, 0, 0})
	is.Equal(n.Transparent(), true)
}

// Each call makes the caller a nine-slice and an image of its own, kept by
// no one else, so two interfaces asking for the same color share nothing.
func TestNewNineSliceColor_OwnEach(t *testing.T) {
	is := is.New(t)

	red := color.NRGBA{255, 0, 0, 255}
	a, b := eimage.NewNineSliceColor(red), eimage.NewNineSliceColor(red)
	is.True(a != b)
	first, second := eimage.NewImageColor(red), eimage.NewImageColor(red)
	is.True(first != second)
	is.Equal(a.FillColor(), color.Color(red))                                            // what it fills with
	is.Equal(eimage.NewNineSliceColor(color.Transparent).FillColor(), color.Transparent) // transparent too
	is.True(eimage.NewNineSliceColor(nil).FillColor() == nil)
	is.True(eimage.NewFixedNineSlice(ebiten.NewImage(2, 2)).FillColor() == nil) // not a color's
}

func Test_NewNineSliceBorder(t *testing.T) {
	is := is.New(t)

	img := newImageEmptySize(20, 20, t)
	n := eimage.NewNineSliceBorder(img, 3)

	is.Equal(n.Widths()[0], 3)
	is.Equal(n.Widths()[1], 20-3*2)
	is.Equal(n.Widths()[2], 3)
	is.Equal(n.Heights()[0], 3)
	is.Equal(n.Heights()[1], 20-3*2)
	is.Equal(n.Heights()[2], 3)
}

// A nine-slice is drawn when any of it lands on the image, and skipped when
// none does; its options are set once to place it, then once for each tile.
func TestNineSlice_SkipsDrawingOffTheImage(t *testing.T) {
	is := is.New(t)

	n := eimage.NewNineSliceSimple(ebiten.NewImage(12, 12), 4, 4)
	screen := ebiten.NewImageWithOptions(image.Rect(100, 100, 200, 200), nil)
	calls := func(opts func(*ebiten.DrawImageOptions)) int {
		count := 0
		n.Draw(screen, 20, 20, func(o *ebiten.DrawImageOptions) {
			count++
			opts(o)
		})
		return count
	}
	at := func(x, y float64) func(*ebiten.DrawImageOptions) {
		return func(o *ebiten.DrawImageOptions) { o.GeoM.Translate(x, y) }
	}

	is.Equal(calls(at(120, 120)), 10) // inside
	is.Equal(calls(at(81, 81)), 10)   // a pixel in, at the top left
	is.Equal(calls(at(199, 199)), 10) // a pixel in, at the bottom right
	is.Equal(calls(at(80, 120)), 1)   // touching the left edge
	is.Equal(calls(at(120, 200)), 1)  // touching the bottom edge
	is.Equal(calls(at(0, 0)), 1)      // outside
	is.Equal(calls(func(o *ebiten.DrawImageOptions) {
		o.GeoM.Scale(-1, 1) // mirrored onto the left of x
		o.GeoM.Translate(110, 120)
	}), 10)
}
