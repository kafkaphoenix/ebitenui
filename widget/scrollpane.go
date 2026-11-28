package widget

import (
	img "image"
	"math"

	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
)

// ScrollPane shows content in a view with a vertical scrollbar beside it
// while the content is taller than the view: the wheel over the view, the
// scrollbar and dragging (with DragToScroll among its ScrollContainerOpts)
// scroll it, and a widget inside given the focus is scrolled into view.
// MaxHeight caps the view's height; without it, the view is as tall as the
// pane's layout lets it be. The scrollbar keeps its room while hidden, so
// the pane is as wide either way, and content that comes to fit is
// scrolled back to its top.
//
// The view's scroll and the scrollbar are kept together once an update,
// after the pane's widgets have updated: the scrollbar moved since, by a
// drag or by the game, scrolls the view; otherwise it follows the view,
// which the wheel and the focus scroll. The pane works from its last
// layout, so a change of the content's size shows on the scrollbar a frame
// after the content is laid out anew.
type ScrollPane struct {
	content       PreferredSizeLocateableWidget
	maxHeight     int
	spacing       int
	focusAt       float64
	viewOpts      []ScrollContainerOpt
	sliderOpts    []SliderOpt
	containerOpts []ContainerOpt

	init      *MultiOnce
	container *Container
	view      *ScrollContainer
	bar       *Slider
	rest      scrollRest
	span      int // the pixels of content below the view, as last fitted
	barAt     int // where the pane last put the scrollbar
}

type ScrollPaneOpt func(p *ScrollPane)

type ScrollPaneOptions struct{}

// ScrollPaneOpts holds the options of NewScrollPane.
var ScrollPaneOpts ScrollPaneOptions

// NewScrollPane returns a scroll pane; its content is required.
func NewScrollPane(opts ...ScrollPaneOpt) *ScrollPane {
	p := &ScrollPane{focusAt: 0.5, init: &MultiOnce{}}
	p.init.Append(p.createWidget)
	for _, o := range opts {
		o(p)
	}
	return p
}

// Content sets what the pane shows.
func (o ScrollPaneOptions) Content(c PreferredSizeLocateableWidget) ScrollPaneOpt {
	return func(p *ScrollPane) { p.content = c }
}

// MaxHeight caps the view's height, in pixels; 0 leaves it uncapped.
func (o ScrollPaneOptions) MaxHeight(h int) ScrollPaneOpt {
	return func(p *ScrollPane) { p.maxHeight = h }
}

// Spacing sets the gap between the view and the scrollbar, in pixels.
func (o ScrollPaneOptions) Spacing(s int) ScrollPaneOpt {
	return func(p *ScrollPane) { p.spacing = s }
}

// FocusAt sets where in the view a widget given the focus is scrolled to:
// a fraction of the way down, 0 at its top, 1 at its bottom; by default
// 0.5, halfway.
func (o ScrollPaneOptions) FocusAt(f float64) ScrollPaneOpt {
	return func(p *ScrollPane) { p.focusAt = f }
}

// ScrollContainerOpts configures the view: its images, padding, dragging.
// The pane sets its content.
func (o ScrollPaneOptions) ScrollContainerOpts(opts ...ScrollContainerOpt) ScrollPaneOpt {
	return func(p *ScrollPane) { p.viewOpts = append(p.viewOpts, opts...) }
}

// SliderOpts configures the scrollbar: its images and sizes. The pane sets
// its direction, range and handler.
func (o ScrollPaneOptions) SliderOpts(opts ...SliderOpt) ScrollPaneOpt {
	return func(p *ScrollPane) { p.sliderOpts = append(p.sliderOpts, opts...) }
}

// ContainerOpts configures the pane's own container, as its widget options
// and layout data.
func (o ScrollPaneOptions) ContainerOpts(opts ...ContainerOpt) ScrollPaneOpt {
	return func(p *ScrollPane) { p.containerOpts = append(p.containerOpts, opts...) }
}

// View returns the scroll container the content shows in.
func (p *ScrollPane) View() *ScrollContainer {
	p.init.Do()
	return p.view
}

// Scrollbar returns the scrollbar: its range is the pixels of content below
// the view.
func (p *ScrollPane) Scrollbar() *Slider {
	p.init.Do()
	return p.bar
}

// SetMaxHeight caps the view's height, in pixels, from the next layout on;
// 0 leaves it uncapped.
func (p *ScrollPane) SetMaxHeight(h int) {
	p.init.Do()
	if h == p.maxHeight {
		return
	}
	p.maxHeight = h
	p.view.GetWidget().LayoutData = GridLayoutData{MaxHeight: h}
	p.container.RequestRelayout()
}

// GetWidget implements HasWidget.
func (p *ScrollPane) GetWidget() *Widget {
	p.init.Do()
	return p.container.GetWidget()
}

// PreferredSize implements PreferredSizer.
func (p *ScrollPane) PreferredSize() (int, int) {
	p.init.Do()
	return p.container.PreferredSize()
}

// SetLocation implements Locateable.
func (p *ScrollPane) SetLocation(rect img.Rectangle) {
	p.init.Do()
	p.container.SetLocation(rect)
}

func (p *ScrollPane) Validate() {
	p.init.Do()
	p.container.Validate()
}

// RequestRelayout implements Relayoutable.
func (p *ScrollPane) RequestRelayout() {
	p.init.Do()
	p.container.RequestRelayout()
}

// SetupInputLayer implements InputLayerer.
func (p *ScrollPane) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	p.init.Do()
	p.container.SetupInputLayer(def)
}

// Render implements Renderer.
func (p *ScrollPane) Render(screen *ebiten.Image) {
	p.init.Do()
	p.container.Render(screen)
}

// Update updates the view and the scrollbar, then fits the scrollbar to
// the last layout and keeps it and the view's scroll together, so what is
// drawn next shows both as they are.
func (p *ScrollPane) Update(updObj *UpdateObject) {
	p.init.Do()
	p.container.Update(updObj)
	p.fit()
}

func (p *ScrollPane) GetFocusers() []Focuser {
	p.init.Do()
	return p.container.GetFocusers()
}

func (p *ScrollPane) GetDropTargets() []HasWidget {
	p.init.Do()
	return p.container.GetDropTargets()
}

// fit shows the scrollbar while the content, as last laid out, overflows
// the view, its range the pixels below the view, and keeps it and the
// view's scroll together: a scrollbar moved since the last fit, dragged or
// set by the game, scrolls the view; otherwise it follows the view. Content
// that comes to fit is scrolled back to its top.
func (p *ScrollPane) fit() {
	view, content := p.view.ViewRect(), p.content.GetWidget().Rect
	if span := max(0, content.Dy()-view.Dy()); span != p.span {
		p.span = span
		p.bar.Max = max(span, 1)
		v := VisibilityShow
		if span == 0 {
			v = VisibilityHideBlocking
			p.view.ScrollTop = 0
		}
		p.bar.GetWidget().SetVisibility(v)
	}
	if p.bar.Current != p.barAt && p.span > 0 { // moved by hand or by the game
		p.view.ScrollTop = min(1, max(0, float64(p.bar.Current)/float64(p.span)))
	}
	p.barAt = int(math.Round(p.view.ScrollTop * float64(p.span)))
	p.bar.Current = p.barAt
}

// scroll scrolls the view by dy pixels, down for a positive dy, within the
// content.
func (p *ScrollPane) scroll(dy int) {
	if p.span > 0 {
		p.view.ScrollTop = min(1, max(0, p.view.ScrollTop+float64(dy)/float64(p.span)))
	}
}

func (p *ScrollPane) createWidget() {
	p.view = NewScrollContainer(append(p.viewOpts,
		ScrollContainerOpts.Content(p.content),
		ScrollContainerOpts.StretchContentWidth(),
		ScrollContainerOpts.WidgetOpts(WidgetOpts.LayoutData(GridLayoutData{MaxHeight: p.maxHeight})),
	)...)
	p.bar = NewSlider(append(p.sliderOpts,
		SliderOpts.Direction(DirectionVertical),
		SliderOpts.MinMax(0, 1),
		// A page is the view's height: a click on the track scrolls a view,
		// and a handle sized by the page is to the track what the view is
		// to the content.
		SliderOpts.PageSizeFunc(func() int { return max(1, p.view.ViewRect().Dy()) }),
	)...)
	p.bar.GetWidget().SetVisibility(VisibilityHideBlocking)

	// The wheel scrolls the pixels the platform's settings give, or, from a
	// device that gives only the wheel's units, a third of the view a
	// notch. Less than a pixel waits for more.
	p.view.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		if a, ok := args.(*ScrolledEventArgs); ok {
			dy := a.DY
			if dy == 0 {
				dy = a.Y * float64(p.view.ViewRect().Dy()) / 3
			}
			p.scroll(-p.rest.take(dy))
		}
	})
	p.view.GetWidget().FocusEvent.AddHandler(func(args any) {
		if a, ok := args.(*FocusEventArgs); ok && a.Focused && a.Widget != nil {
			p.view.ScrollVisible(a.Widget, p.focusAt)
		}
	})

	p.container = NewContainer(append(p.containerOpts, ContainerOpts.Layout(NewGridLayout(
		GridLayoutOpts.Columns(2),
		GridLayoutOpts.Spacing(p.spacing, 0),
		GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
	)))...)
	p.container.AddChild(p.view, p.bar)
	p.viewOpts, p.sliderOpts, p.containerOpts = nil, nil, nil
}
