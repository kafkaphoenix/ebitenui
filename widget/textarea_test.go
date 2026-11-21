package widget

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

// As its text changes, a text area set to scroll to its end or its start
// goes there; one set to neither, or only to start at its end, stays where
// the reader left it.
func TestTextAreaScrollsAsItsTextChanges(t *testing.T) {
	text := strings.Repeat("a line\n", 40)
	const left = 7 // where the reader left the slider
	for name, tc := range map[string]struct {
		mode ScrollMode
		want func(s *Slider) int
	}{
		"to the end":         {ScrollEnd, func(s *Slider) int { return s.Max }},
		"to the start":       {ScrollBeginning, func(*Slider) int { return 0 }},
		"not at all":         {None, func(*Slider) int { return left }},
		"only as it is made": {PositionAtEnd, func(*Slider) int { return left }},
	} {
		t.Run(name, func(t *testing.T) {
			a := NewTextArea(
				TextAreaOpts.ScrollContainerImage(&ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
				TextAreaOpts.SliderParams(&SliderParams{TrackImage: &SliderTrackImage{}, HandleImage: &ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
				TextAreaOpts.FontFace(loadFont(t)),
				TextAreaOpts.FontColor(color.White),
				TextAreaOpts.ShowVerticalScrollbar(),
				TextAreaOpts.VerticalScrollMode(tc.mode),
				TextAreaOpts.Text(text),
			)
			a.GetWidget().Rect = image.Rect(0, 0, 200, 60)
			render(a, t)
			if m := a.vSlider.Max; m == 0 || m == left {
				t.Fatalf("the slider ends at %d, which doesn't tell the modes apart", m)
			}
			a.vSlider.Current = left
			a.SetText(text + "one more\n")
			if got, want := a.vSlider.Current, tc.want(a.vSlider); got != want {
				t.Errorf("the slider is at %d after the text changed, want %d", got, want)
			}
		})
	}
}
