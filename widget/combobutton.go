package widget

import (
	"image"

	"github.com/ebitenui/ebitenui/input"

	"github.com/hajimehoshi/ebiten/v2"
)

type ComboButton struct {
	ContentVisible bool

	buttonOpts       []ButtonOpt
	maxContentHeight int

	init    *MultiOnce
	button  *Button
	content HasWidget

	// The open content's input layer, and the function that sets it up,
	// kept from frame to frame.
	contentLayer      input.Layer
	setupContentLayer input.SetupInputLayerFunc
}

type ComboButtonOpt func(c *ComboButton)

type ComboButtonOptions struct {
}

var ComboButtonOpts ComboButtonOptions

func NewComboButton(opts ...ComboButtonOpt) *ComboButton {
	c := &ComboButton{
		init: &MultiOnce{},
	}

	c.init.Append(c.createWidget)

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *ComboButton) Validate() {
	c.init.Do()
	if c.content == nil {
		panic("ComboButton: Content is required.")
	}
	if len(c.buttonOpts) == 0 {
		panic("ComboButton: ButtonOpts are required.")
	}
	c.button.Validate()
}
func (o ComboButtonOptions) ButtonOpts(opts ...ButtonOpt) ComboButtonOpt {
	return func(c *ComboButton) {
		c.buttonOpts = append(c.buttonOpts, opts...)
	}
}

func (o ComboButtonOptions) Content(c HasWidget) ComboButtonOpt {
	return func(cb *ComboButton) {
		cb.content = c
	}
}

func (o ComboButtonOptions) MaxContentHeight(h int) ComboButtonOpt {
	return func(c *ComboButton) {
		c.maxContentHeight = h
	}
}

func (c *ComboButton) GetWidget() *Widget {
	c.init.Do()
	return c.button.GetWidget()
}

func (c *ComboButton) SetLocation(rect image.Rectangle) {
	c.init.Do()
	c.button.GetWidget().Rect = rect
}

func (c *ComboButton) PreferredSize() (int, int) {
	c.init.Do()
	return c.button.PreferredSize()
}

func (c *ComboButton) SetLabel(l string) {
	c.init.Do()
	c.button.Text().Label = l
	c.button.RequestRelayout()
}

func (c *ComboButton) Label() string {
	c.init.Do()
	return c.button.Text().Label
}

func (c *ComboButton) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	c.init.Do()

	c.button.SetupInputLayer(def)

	if c.content != nil && c.ContentVisible {
		if c.setupContentLayer == nil { // made once, kept from frame to frame
			c.contentLayer = input.Layer{
				DebugLabel: "combo button content visible",
				EventTypes: input.LayerEventTypeAll,
				BlockLower: true,
				FullScreen: false,
				RectFunc: func() image.Rectangle {
					return c.content.GetWidget().Rect
				},
			}
			c.setupContentLayer = c.setupContentInputLayer
		}
		def(c.setupContentLayer)
	}
}

// setupContentInputLayer raises the open content to its own input layer,
// above the widgets it covers.
func (c *ComboButton) setupContentInputLayer(def input.DeferredSetupInputLayerFunc) {
	c.content.GetWidget().ElevateToNewInputLayer(&c.contentLayer)
	if il, ok := c.content.(input.Layerer); ok {
		il.SetupInputLayer(def)
	}
}

func (c *ComboButton) Render(screen *ebiten.Image) {
	c.init.Do()

	c.button.Render(screen)

	if c.content != nil && c.ContentVisible {
		c.relayoutContent()

		r, ok := c.content.(Renderer)
		if !ok {
			return
		}
		AppendToDeferredRenderQueue(r.Render)
	}
}

func (c *ComboButton) Update(updObj *UpdateObject) {
	c.init.Do()

	c.button.Update(updObj)
	c.handleClick()

	if c.content != nil && c.ContentVisible {
		if cu, ok := c.content.(Updater); ok {
			cu.Update(updObj)
		}
	}
}

func (c *ComboButton) handleClick() {
	if input.MouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := input.CursorPosition()
		p := image.Point{x, y}
		if !p.In(c.button.GetWidget().Rect) && !p.In(c.content.GetWidget().Rect) {
			c.ContentVisible = false
		}
	}
}

func (c *ComboButton) relayoutContent() {
	l, ok := c.content.(Locateable)
	if !ok {
		return
	}

	rect := c.button.GetWidget().Rect
	x, y := rect.Min.X, rect.Max.Y+2

	var w int
	var h int
	if p, ok := c.content.(PreferredSizer); ok {
		w, h = p.PreferredSize()
	} else {
		w, h = 50, 50
	}

	if c.maxContentHeight > 0 && h > c.maxContentHeight {
		h = c.maxContentHeight
	}

	cr := image.Rect(0, 0, w, h)
	cr = cr.Add(image.Point{x, y})

	if cr == c.content.GetWidget().Rect {
		return
	}

	l.SetLocation(cr)

	r, ok := c.content.(Relayoutable)
	if !ok {
		return
	}

	r.RequestRelayout()
}

func (c *ComboButton) createWidget() {
	c.button = NewButton(append(c.buttonOpts, ButtonOpts.ClickedHandler(func(_ *ButtonClickedEventArgs) {
		c.ContentVisible = !c.ContentVisible
	}))...)
}
