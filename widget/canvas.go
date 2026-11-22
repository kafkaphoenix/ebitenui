package widget

import (
	img "image"

	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/hajimehoshi/ebiten/v2"
)

// CanvasDrawFunc draws a canvas's content on screen, an image covering
// only rect, the canvas's place, in the screen's coordinates.
type CanvasDrawFunc func(screen *ebiten.Image, rect img.Rectangle)

// Canvas is a widget a game draws into: a minimap, a chart, a preview,
// anything the other widgets don't draw, laid out, clipped and given the
// pointer as any widget is. Its draw function draws on an image covering
// only its place, so nothing it draws spills onto its neighbors, and not
// at all while it is off the screen. The pointer over it reaches the game
// through its widget's events (GetWidget: MouseButtonClickedEvent,
// CursorMoveEvent, ScrolledEvent, and the rest).
//
// It draws every frame. A game that draws only frames that show something
// new tells UI.Changed its picture changed with MarkChanged.
type Canvas struct {
	draw CanvasDrawFunc

	widgetOpts []WidgetOpt
	init       *MultiOnce
	widget     *Widget
}

type CanvasOpt func(c *Canvas)

type CanvasOptions struct{}

// CanvasOpts holds the options of NewCanvas.
var CanvasOpts CanvasOptions

// NewCanvas returns a canvas, as big as its minimum size, or 64 by 64.
func NewCanvas(opts ...CanvasOpt) *Canvas {
	c := &Canvas{init: &MultiOnce{}}
	c.init.Append(c.createWidget)
	for _, o := range opts {
		o(c)
	}
	return c
}

func (o CanvasOptions) WidgetOpts(opts ...WidgetOpt) CanvasOpt {
	return func(c *Canvas) { c.widgetOpts = append(c.widgetOpts, opts...) }
}

// DrawFunc sets what draws the canvas's content.
func (o CanvasOptions) DrawFunc(f CanvasDrawFunc) CanvasOpt {
	return func(c *Canvas) { c.draw = f }
}

// SetDrawFunc changes what draws the canvas's content.
func (c *Canvas) SetDrawFunc(f CanvasDrawFunc) {
	c.draw = f
	changes.Mark()
}

// MarkChanged tells UI.Changed the canvas's picture changed, for a game
// drawing only frames that show something new.
func (c *Canvas) MarkChanged() { changes.Mark() }

func (c *Canvas) GetWidget() *Widget {
	c.init.Do()
	return c.widget
}

func (c *Canvas) SetLocation(rect img.Rectangle) {
	c.init.Do()
	c.widget.Rect = rect
}

func (c *Canvas) Validate() { c.init.Do() }

// PreferredSize is its minimum size, or 64 by 64 without one.
func (c *Canvas) PreferredSize() (int, int) {
	c.init.Do()
	w, h := c.widget.MinWidth, c.widget.MinHeight
	if w == 0 && h == 0 {
		return 64, 64
	}
	return w, h
}

func (c *Canvas) Update(updObj *UpdateObject) {
	c.init.Do()
	c.widget.Update(updObj)
}

// Render calls the draw function on the part of screen the canvas covers.
func (c *Canvas) Render(screen *ebiten.Image) {
	c.init.Do()
	c.widget.Render(screen)
	if c.draw == nil {
		return
	}
	r := c.widget.Rect.Intersect(screen.Bounds())
	if r.Empty() {
		return
	}
	if part, ok := screen.SubImage(r).(*ebiten.Image); ok {
		c.draw(part, c.widget.Rect)
	}
}

func (c *Canvas) createWidget() {
	c.widget = NewWidget(c.widgetOpts...)
	c.widgetOpts = nil
}
