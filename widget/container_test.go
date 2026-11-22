package widget_test

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/input"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// control is a widget 50 pixels square that counts how often it is drawn
// and sets its input layers up, and runs onRender as it is drawn.
type control struct {
	widget   *widget.Widget
	renders  int
	setups   int
	onRender func()
}

func newControl() *control { return &control{widget: widget.NewWidget()} }

func (c *control) GetWidget() *widget.Widget     { return c.widget }
func (c *control) PreferredSize() (int, int)     { return 50, 50 }
func (c *control) SetLocation(_ image.Rectangle) {}
func (c *control) Validate()                     {}

func (c *control) Render(_ *ebiten.Image) {
	c.renders++
	if c.onRender != nil {
		c.onRender()
	}
}

func (c *control) SetupInputLayer(_ input.DeferredSetupInputLayerFunc) { c.setups++ }

func TestContainer_Render(t *testing.T) {
	m := newControl()
	c := newContainer(t,
		widget.ContainerOpts.Layout(newRowLayout(t)))
	c.AddChild(m)

	render(c, t)

	if m.renders != 1 {
		t.Errorf("the child was drawn %d times, want once", m.renders)
	}
}

func TestContainer_Render_AutoDisableChildren(t *testing.T) {
	is := is.New(t)

	m := newControl()
	c := newContainer(t,
		widget.ContainerOpts.AutoDisableChildren(),
		widget.ContainerOpts.Layout(newRowLayout(t)))
	c.AddChild(m)

	c.Widget().Disabled = true
	render(c, t)

	is.True(m.widget.Disabled)
}

func TestContainer_SetupInputLayer(t *testing.T) {
	def := func(_ input.SetupInputLayerFunc) {
		// nothing to do
	}

	m := newControl()
	c := newContainer(t,
		widget.ContainerOpts.Layout(newRowLayout(t)))
	c.AddChild(m)

	c.SetupInputLayer(def)

	if m.setups != 1 {
		t.Errorf("the child set its input layers up %d times, want once", m.setups)
	}
}

func TestContainer_Replace(t *testing.T) {
	c := newContainer(t)
	m1, m2, m3 := newControl(), newControl(), newControl()

	c.AddChild(m1, m2, m3)

	if len(c.ChildrenForTest()) != 3 {
		t.Fatalf("expected 3 children, got %d", len(c.ChildrenForTest()))
	}

	replaced := newControl()

	c.ReplaceChild(m2, replaced)

	if len(c.ChildrenForTest()) != 3 {
		t.Fatalf("expected 3 children, got %d", len(c.ChildrenForTest()))
	}

	if c.ChildrenForTest()[0] != m1 {
		t.Fatalf("expected original child at index 0")
	}

	if c.ChildrenForTest()[1] != replaced {
		t.Fatalf("expected replaced child at index 1")
	}

	if c.ChildrenForTest()[2] != m3 {
		t.Fatalf("expected original child at index 2")
	}
}

func newContainer(t *testing.T, opts ...widget.ContainerOpt) *widget.Container {
	t.Helper()
	return widget.NewContainer(opts...)
}
