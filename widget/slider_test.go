package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestSlider_Current_Initial(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.SliderChangedEventArgs
	s := newSlider(t,
		widget.SliderOpts.MinMax(10, 20),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			eventArgs = args
		}))

	is.Equal(s.Current, 10)
	is.Equal(eventArgs.Current, 10)
}

func newSlider(t *testing.T, opts ...widget.SliderOpt) *widget.Slider {
	s := widget.NewSlider(append(opts, widget.SliderOpts.Images(&widget.SliderTrackImage{
		Idle: newNineSliceEmpty(t),
	}, &widget.ButtonImage{
		Idle:    newNineSliceEmpty(t),
		Pressed: newNineSliceEmpty(t),
	}))...)
	event.ExecuteDeferred()
	render(s, t)
	return s
}

// An inverted slider has its Max at the start of its track: a vertical
// one, a volume control, its largest values at the top. Up raises it, and
// a click on the track above the handle does too.
func TestSlider_Inverted(t *testing.T) {
	s := newSlider(t, widget.SliderOpts.Direction(widget.DirectionVertical), widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.InitialCurrent(100), widget.SliderOpts.Inverted(), widget.SliderOpts.PageSizeFunc(func() int { return 10 }))
	s.SetLocation(img.Rect(0, 0, 20, 200))
	screen := ebiten.NewImage(20, 200)
	handleAt := func() img.Rectangle {
		s.Render(screen)
		return s.Handle().GetWidget().Rect
	}
	if top := handleAt(); top.Min.Y != 0 {
		t.Errorf("at its Max the handle is at %v, want at the top", top)
	}
	s.Current = 0
	if bottom := handleAt(); bottom.Max.Y != 200 {
		t.Errorf("at its Min the handle is at %v, want at the bottom", bottom)
	}

	s.Focus(true)
	s.Update(&widget.UpdateObject{Actions: given(widget.ActionUp)})
	if s.Current != 10 {
		t.Errorf("Up took it to %d, want 10: up raises an inverted slider", s.Current)
	}
	d := uitest.Use(t)
	s.Current = 50
	h := handleAt()
	d.Cursor = img.Pt(10, h.Min.Y-5) // on the track, above the handle
	input.Update()
	s.GetWidget().MouseButtonPressedEvent.Fire(&widget.MouseButtonPressedEventArgs{Widget: s.GetWidget(), Button: ebiten.MouseButtonLeft})
	event.ExecuteDeferred()
	input.AfterUpdate()
	if s.Current != 60 {
		t.Errorf("a click above the handle took it to %d, want 60", s.Current)
	}
}
