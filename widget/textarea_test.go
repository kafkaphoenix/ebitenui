package widget_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
)

// A text area starts at the top of its text, and the left of it: its
// sliders, which count thousandths, start at 0.
func TestTextAreaStartsAtTheTop(t *testing.T) {
	a := widget.NewTextArea(
		widget.TextAreaOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
		widget.TextAreaOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
		widget.TextAreaOpts.FontFace(loadFont(t)),
		widget.TextAreaOpts.FontColor(color.White),
		widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.ShowHorizontalScrollbar(),
		widget.TextAreaOpts.Text(strings.Repeat("a long line, longer than the area is wide\n", 40)),
	)
	a.GetWidget().Rect = image.Rect(0, 0, 100, 60)
	render(a, t)
	if v, h := a.VSlider().Current, a.HSlider().Current; v != 0 || h != 0 {
		t.Errorf("the sliders start at %d down and %d across, want 0 and 0", v, h)
	}
	if sc := a.ScrollContainer(); sc.ScrollTop != 0 || sc.ScrollLeft != 0 {
		t.Errorf("the text starts scrolled to %v, %v, want 0, 0", sc.ScrollTop, sc.ScrollLeft)
	}
}

// As its text changes, a text area set to scroll to its end or its start
// goes there; one set to neither, or only to start at its end, stays where
// the reader left it.
func TestTextAreaScrollsAsItsTextChanges(t *testing.T) {
	text := strings.Repeat("a line\n", 40)
	const left = 7 // where the reader left the slider
	for name, tc := range map[string]struct {
		mode widget.ScrollMode
		want func(s *widget.Slider) int
	}{
		"to the end":         {widget.ScrollEnd, func(s *widget.Slider) int { return s.Max }},
		"to the start":       {widget.ScrollBeginning, func(*widget.Slider) int { return 0 }},
		"not at all":         {widget.None, func(*widget.Slider) int { return left }},
		"only as it is made": {widget.PositionAtEnd, func(*widget.Slider) int { return left }},
	} {
		t.Run(name, func(t *testing.T) {
			a := widget.NewTextArea(
				widget.TextAreaOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
				widget.TextAreaOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
				widget.TextAreaOpts.FontFace(loadFont(t)),
				widget.TextAreaOpts.FontColor(color.White),
				widget.TextAreaOpts.ShowVerticalScrollbar(),
				widget.TextAreaOpts.VerticalScrollMode(tc.mode),
				widget.TextAreaOpts.Text(text),
			)
			a.GetWidget().Rect = image.Rect(0, 0, 200, 60)
			render(a, t)
			if m := a.VSlider().Max; m == 0 || m == left {
				t.Fatalf("the slider ends at %d, which doesn't tell the modes apart", m)
			}
			a.VSlider().Current = left
			a.SetText(text + "one more\n")
			if got, want := a.VSlider().Current, tc.want(a.VSlider()); got != want {
				t.Errorf("the slider is at %d after the text changed, want %d", got, want)
			}
		})
	}
}

// A text area places its text and its slider without drawing
// (PlaceChildren), as UI.Layout has it.
func TestTextAreaPlacesItsParts(t *testing.T) {
	a := widget.NewTextArea(
		widget.TextAreaOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
		widget.TextAreaOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
		widget.TextAreaOpts.FontFace(loadFont(t)),
		widget.TextAreaOpts.FontColor(color.White),
		widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.Text(strings.Repeat("a line\n", 40)),
	)
	a.Validate()
	a.SetLocation(image.Rect(0, 0, 200, 60))
	a.PlaceChildren()
	if r := a.VSlider().GetWidget().Rect; r.Empty() {
		t.Errorf("the slider is at %v: not placed", r)
	}
}
