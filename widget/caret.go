package widget

import (
	img "image"
	"image/color"
	"time"

	"github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
)

type Caret struct {
	Width         int
	Height        int
	Color         color.Color
	blinkInterval time.Duration

	init   *MultiOnce
	widget *Widget

	// The color's image, made again only as the color changes, and how it
	// is placed, made once.
	image      *image.NineSlice
	imageColor color.Color
	place      image.DrawImageOptionsFunc

	// The updates since the caret last showed after a reset, and whether
	// it shows: it blinks by counting updates.
	updates int
	visible bool
}

type CaretOpt func(c *Caret)

type CaretOptions struct {
}

var CaretOpts CaretOptions

func NewCaret(opts ...CaretOpt) *Caret {
	c := &Caret{
		blinkInterval: 450 * time.Millisecond,

		init: &MultiOnce{},
	}
	c.resetBlinking()

	c.init.Append(c.createWidget)

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Caret) Validate() {

}

func (o CaretOptions) Color(c color.Color) CaretOpt {
	return func(ca *Caret) {
		ca.Color = c
	}
}

func (o CaretOptions) Size(height int, width int) CaretOpt {
	return func(c *Caret) {
		c.Height = height
		c.Width = width
	}
}

func (c *Caret) GetWidget() *Widget {
	c.init.Do()
	return c.widget
}

func (c *Caret) SetLocation(rect img.Rectangle) {
	c.init.Do()
	c.widget.Rect = rect
}

func (c *Caret) PreferredSize() (int, int) {
	c.init.Do()
	return c.Width, c.Height
}

func (c *Caret) Render(screen *ebiten.Image) {
	c.init.Do()

	c.widget.Render(screen)

	if !c.visible {
		return
	}

	if c.image == nil || c.Color != c.imageColor {
		c.image, c.imageColor = image.NewNineSliceColor(c.Color), c.Color
	}
	c.image.Draw(screen, c.Width, c.Height, c.place)
}

// Update blinks the caret: it shows for blinkInterval's updates, then
// hides as long.
func (c *Caret) Update(updObj *UpdateObject) {
	c.init.Do()

	c.updates++
	c.visible = c.updates/updatesIn(c.blinkInterval)%2 == 0

	c.widget.Update(updObj)
}

func (c *Caret) ResetBlinking() {
	c.init.Do()
	c.resetBlinking()
}

// resetBlinking shows the caret and starts its blink over.
func (c *Caret) resetBlinking() {
	c.updates, c.visible = 0, true
}

func (c *Caret) createWidget() {
	c.widget = NewWidget()
	c.place = func(opts *ebiten.DrawImageOptions) {
		p := c.widget.Rect.Min
		opts.GeoM.Translate(float64(p.X), float64(p.Y))
	}
}
