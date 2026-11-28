package widget_test

import (
	img "image"
	"image/color"
	"image/gif"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// Widgets that animate by themselves pause while hidden: they step nothing
// and tell UI.Changed of nothing, so a game that skips still frames stays
// still while they don't show. Shown again, they go on from where they
// were. A GIF's frames, a spinner's dot, a carousel moving on by itself.
func TestHiddenAnimationsPause(t *testing.T) {
	is := is.New(t)

	frame := func() *img.Paletted {
		return img.NewPaletted(img.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	}
	g := newGraphic(t, widget.GraphicOpts.GIF(&gif.GIF{Image: []*img.Paletted{frame(), frame()}, Delay: []int{5, 5}}))
	g.GetWidget().SetVisibility(widget.VisibilityHide)
	from := changes.Count()
	for range ebiten.TPS() {
		g.Update(&widget.UpdateObject{})
	}
	is.Equal(g.GifCurrentImage(), 0)
	is.Equal(changes.Count(), from) // hidden: still
	g.GetWidget().SetVisibility(widget.VisibilityShow)
	for range 3 {
		g.Update(&widget.UpdateObject{})
	}
	is.Equal(g.GifCurrentImage(), 1) // shown: on from its first frame

	l := newLoading(t, img.Rect(0, 0, 32, 32), widget.LoadingOpts.Spinner())
	l.GetWidget().SetVisibility(widget.VisibilityHideBlocking)
	is.Equal(marks(l, ebiten.TPS()), 0)
	l.GetWidget().SetVisibility(widget.VisibilityShow)
	is.Equal(marks(l, ebiten.TPS()), 8) // a turn, as if never paused

	d := uitest.Use(t)
	d.Cursor = img.Pt(400, 400) // away, so hovering doesn't pause it
	c, _ := newCarousel(t, 2, widget.CarouselOpts.AutoAdvance(time.Second), widget.CarouselOpts.SlideTime(0))
	c.GetWidget().SetVisibility(widget.VisibilityHide)
	from = changes.Count()
	updateCarousel(c, 2*ebiten.TPS())
	is.Equal(c.Page(), 0)
	is.Equal(changes.Count(), from)
	c.GetWidget().SetVisibility(widget.VisibilityShow)
	updateCarousel(c, ebiten.TPS()+1)
	is.Equal(c.Page(), 1)
}
