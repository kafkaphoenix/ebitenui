package image_test

import (
	img "image"
	"testing"

	eimage "github.com/ebitenui/ebitenui/image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

func TestBufferedImage_Image(t *testing.T) {
	is := is.New(t)

	b := &eimage.BufferedImage{}
	b.Width, b.Height = 100, 100
	i := b.Image()
	w, h := i.Bounds().Dx(), i.Bounds().Dy()
	is.Equal(w, 100)
	is.Equal(h, 100)

	b.Width, b.Height = 150, 70
	i = b.Image()
	w, h = i.Bounds().Dx(), i.Bounds().Dy()
	is.Equal(w, 150)
	is.Equal(h, 70)
}

func TestMaskedRenderBuffer_Draw(t *testing.T) {
	is := is.New(t)

	b := eimage.NewMaskedRenderBuffer()
	screen := newImageEmptySize(100, 100, t)

	draw := false
	drawMask := false

	b.Draw(screen, func(_ *ebiten.Image) {
		draw = true
	}, func(_ *ebiten.Image) {
		drawMask = true
	})

	is.True(draw)
	is.True(drawMask)
}

// DrawIn draws onto images covering only the area masked, in the screen's
// coordinates, kept while the area stays the same and replaced when it
// moves.
func TestMaskedRenderBuffer_DrawIn(t *testing.T) {
	is := is.New(t)

	screen := ebiten.NewImage(200, 200)
	m := eimage.NewMaskedRenderBuffer()
	var drawnOn, maskedOn *ebiten.Image
	draw := func(rect img.Rectangle) {
		m.DrawIn(screen, rect, func(buf *ebiten.Image) { drawnOn = buf }, func(buf *ebiten.Image) { maskedOn = buf })
	}

	draw(img.Rect(20, 30, 70, 90))
	is.Equal(drawnOn.Bounds(), img.Rect(20, 30, 70, 90)) // the area, where it is
	is.Equal(maskedOn.Bounds(), img.Rect(20, 30, 70, 90))
	first := drawnOn

	draw(img.Rect(20, 30, 70, 90))
	is.True(drawnOn == first) // kept, not another image of the same bounds

	draw(img.Rect(150, 150, 250, 250))
	is.Equal(drawnOn.Bounds(), img.Rect(150, 150, 200, 200)) // moved, and cut to the screen
}

// An area that stays has images of its size; once it moves, they get a
// margin, so that it can slide, as a window easing in, keeping them.
func TestMaskedRenderBuffer_DrawInMovingArea(t *testing.T) {
	is := is.New(t)

	screen := ebiten.NewImage(800, 600)
	m := eimage.NewMaskedRenderBuffer()
	draw := func(r img.Rectangle) {
		m.DrawIn(screen, r, func(*ebiten.Image) {}, func(*ebiten.Image) {})
	}

	draw(img.Rect(100, 100, 300, 200))
	is.Equal(m.RenderIn().Bounds(), img.Rect(100, 100, 300, 200)) // the area alone while it stays
	is.Equal(m.MaskedIn().Bounds(), img.Rect(100, 100, 300, 200))

	draw(img.Rect(105, 100, 305, 200))
	moved, movedMask := m.RenderIn(), m.MaskedIn()
	is.Equal(moved.Bounds(), img.Rect(41, 36, 369, 264)) // 64 px around it from then on

	for x := 110; x <= 165; x += 5 {
		draw(img.Rect(x, 100, x+200, 200))
	}
	is.True(m.RenderIn() == moved && m.MaskedIn() == movedMask) // kept while sliding inside them

	draw(img.Rect(400, 300, 600, 400))
	is.True(m.RenderIn() != moved) // past them: new images

	draw(img.Rect(650, 500, 850, 650))
	is.Equal(m.RenderIn().Bounds(), img.Rect(586, 436, 800, 600)) // the margin stops at the screen's edge
}

// Release gives a buffer's images back, those of DrawIn and of Draw; a
// draw after makes them anew.
func TestMaskedRenderBuffer_Release(t *testing.T) {
	is := is.New(t)

	screen := ebiten.NewImage(200, 200)
	nothing := func(*ebiten.Image) {}
	m, whole := eimage.NewMaskedRenderBuffer(), eimage.NewMaskedRenderBuffer()
	is.True(!m.Allocated()) // nothing drawn yet
	m.DrawIn(screen, img.Rect(10, 10, 60, 60), nothing, nothing)
	whole.Draw(screen, nothing, nothing)
	is.True(m.Allocated() && whole.Allocated())

	m.Release()
	whole.Release()
	is.True(!m.Allocated() && !whole.Allocated())
	is.True(m.RenderIn() == nil && m.MaskedIn() == nil)
	m.Release() // nothing left: nothing happens

	m.DrawIn(screen, img.Rect(10, 10, 60, 60), nothing, nothing)
	is.True(m.RenderIn() != nil && m.MaskedIn() != nil) // anew
	whole.Draw(screen, nothing, nothing)
	is.True(whole.Allocated())
}
