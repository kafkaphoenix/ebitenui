package widget_test

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

func TestButton_PressedEvent_User(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.ButtonPressedEventArgs

	b := newButton(t,
		widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
			eventArgs = args
		}))

	leftMouseButtonPress(b, t)

	is.True(eventArgs != nil)
}

func TestButton_ReleasedEvent_User(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.ButtonReleasedEventArgs

	b := newButton(t,
		widget.ButtonOpts.ReleasedHandler(func(args *widget.ButtonReleasedEventArgs) {
			eventArgs = args
		}))

	leftMouseButtonPress(b, t)
	leftMouseButtonRelease(b, t)

	is.True(eventArgs != nil)
}

func TestButton_ClickedEvent_User(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.ButtonClickedEventArgs

	b := newButton(t,
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			eventArgs = args
		}))

	leftMouseButtonClick(b, t)

	is.True(eventArgs != nil)
}

func TestButton_ManualClick(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.ButtonClickedEventArgs

	b := newButton(t,
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			eventArgs = args
		}))

	b.Click()
	event.ExecuteDeferred()

	is.True(eventArgs != nil)
}

func TestButton_PreferredSizeBeforeValidation(t *testing.T) {
	is := is.New(t)

	buttonImage := &widget.ButtonImage{
		Idle:    newNineSliceEmpty(t),
		Pressed: newNineSliceEmpty(t),
	}
	b := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(60, 70)),
	)
	c := widget.NewContainer()
	c.AddChild(b)

	w, h := b.PreferredSize()
	is.Equal(w, 60)
	is.Equal(h, 70)
}

// A button that takes its look from a theme, sized before it joins a
// themed container, as a layout made ahead does, is sized without its image
// instead of panicking; in the container, its size takes the theme's image.
// Validated without an image anywhere, it still panics, saying what it
// lacks.
func TestButton_PreferredSizeBeforeItsTheme(t *testing.T) {
	is := is.New(t)

	b := widget.NewButton(widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(20, 10)))
	w, h := b.PreferredSize()
	is.Equal([]int{w, h}, []int{50, 50}) // no image yet: the size of a button with no content

	big := image.NewBorderedNineSliceColor(color.White, color.Black, 40)
	c := widget.NewContainer()
	c.GetWidget().SetTheme(&widget.Theme{ButtonTheme: &widget.ButtonParams{Image: &widget.ButtonImage{Idle: big, Pressed: big}}})
	c.AddChild(b)
	w, h = b.PreferredSize()
	is.Equal([]int{w, h}, []int{80, 80}) // the theme's image's corners

	labeled := widget.NewButton(widget.ButtonOpts.Image(&widget.ButtonImage{Idle: big, Pressed: big}),
		widget.ButtonOpts.Text("A label much wider than eighty pixels", loadFont(t), &widget.ButtonTextColor{Idle: color.White}))
	w, _ = labeled.PreferredSize()
	is.True(w > 200) // all it needs given: validated, so sized by its label

	ns := newNineSliceEmpty(t)
	for _, c := range []struct {
		opts []widget.ButtonOpt
		want string
	}{
		{nil, "Image is required."},
		{[]widget.ButtonOpt{widget.ButtonOpts.Image(&widget.ButtonImage{Pressed: ns})}, "Image.Idle is required."},
		{[]widget.ButtonOpt{widget.ButtonOpts.Image(&widget.ButtonImage{Idle: ns})}, "Image.Pressed is required."},
		{[]widget.ButtonOpt{widget.ButtonOpts.Image(&widget.ButtonImage{Idle: ns, Pressed: ns}), widget.ButtonOpts.TextLabel("x")}, "TextFace is required if TextLabel is set."},
		{[]widget.ButtonOpt{widget.ButtonOpts.Image(&widget.ButtonImage{Idle: ns, Pressed: ns}), widget.ButtonOpts.TextLabel("x"), widget.ButtonOpts.TextFace(loadFont(t))}, "TextColor is required if TextLabel is set."},
		{[]widget.ButtonOpt{widget.ButtonOpts.Image(&widget.ButtonImage{Idle: ns, Pressed: ns}), widget.ButtonOpts.TextLabel("x"), widget.ButtonOpts.TextFace(loadFont(t)), widget.ButtonOpts.TextColor(&widget.ButtonTextColor{})}, "TextColor.Idle is required if TextLabel is set."},
	} {
		func() {
			defer func() { is.Equal(recover(), "Button: "+c.want) }()
			widget.NewButton(c.opts...).Validate()
		}()
	}
}

func newButton(t *testing.T, opts ...widget.ButtonOpt) *widget.Button {
	t.Helper()

	b := widget.NewButton(append(opts, widget.ButtonOpts.Image(&widget.ButtonImage{
		Idle:    newNineSliceEmpty(t),
		Pressed: newNineSliceEmpty(t),
	}))...)
	event.ExecuteDeferred()
	render(b, t)
	return b
}

// A widget in a button's content is under the button: it takes the
// button's input layer, and the theme of what holds the button.
func TestButton_ContentUnderTheButton(t *testing.T) {
	is := is.New(t)

	inner := widget.NewButton(widget.ButtonOpts.Image(&widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}))
	b := widget.NewButton(widget.ButtonOpts.Image(&widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}), widget.ButtonOpts.Content(inner))
	theme := &widget.Theme{}
	holder := widget.NewContainer()
	holder.GetWidget().SetTheme(theme)
	holder.AddChild(b)
	holder.Validate()
	layer := &input.Layer{DebugLabel: "button", EventTypes: input.LayerEventTypeAll, RectFunc: func() img.Rectangle { return b.GetWidget().Rect }}
	t.Cleanup(func() { input.SetupInputLayersWithDeferred(nil) }) // the layers are the input package's
	b.GetWidget().ElevateToNewInputLayer(layer)

	is.Equal(inner.GetWidget().EffectiveInputLayer(), layer) // the button's layer
	is.Equal(inner.GetWidget().GetTheme(), theme)            // its holder's theme
}

// placedWidget counts the times a layout places it.
type placedWidget struct {
	widget *widget.Widget
	placed int
}

func (p *placedWidget) GetWidget() *widget.Widget { return p.widget }
func (p *placedWidget) PreferredSize() (int, int) { return 10, 10 }
func (p *placedWidget) Validate()                 {}

func (p *placedWidget) SetLocation(rect img.Rectangle) {
	p.widget.Rect = rect
	p.placed++
}

// A button lays its content out when it moves or resizes, or when asked to
// (a window's relayout asks every widget in it), not on every frame.
func TestButton_ContentLaidOutWhenNeeded(t *testing.T) {
	is := is.New(t)

	content := &placedWidget{widget: widget.NewWidget()}
	b := newButton(t, widget.ButtonOpts.Content(content))
	screen := ebiten.NewImage(100, 100)
	b.SetLocation(img.Rect(0, 0, 40, 20))
	b.Render(screen)
	placed, at := content.placed, content.widget.Rect

	b.Render(screen)
	b.Render(screen)
	is.Equal(content.placed, placed) // the button stayed put

	b.SetLocation(img.Rect(10, 5, 50, 25))
	b.Render(screen)
	is.Equal(content.placed, placed+1)                      // laid out again
	is.Equal(content.widget.Rect, at.Add(img.Point{10, 5})) // following the button

	b.RequestRelayout()
	b.Render(screen)
	is.Equal(content.placed, placed+2) // asked to
}
