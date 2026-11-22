package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

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
