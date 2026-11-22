package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

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

func newGraphic(t *testing.T, opts ...widget.GraphicOpt) *widget.Graphic {
	t.Helper()

	g := widget.NewGraphic(opts...)
	event.ExecuteDeferred()
	render(g, t)
	return g
}
