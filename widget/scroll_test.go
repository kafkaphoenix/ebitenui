package widget_test

import (
	img "image"
	"image/color"
	"strconv"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// longList is a list of n entries, laid out on a 200 by 300 screen.
func longList(t *testing.T, n int) (*widget.List, *ebiten.Image) {
	t.Helper()
	entries := make([]any, n)
	for i := range entries {
		entries[i] = strconv.Itoa(i)
	}
	return newList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryLabelFunc(label)), ebiten.NewImage(200, 300)
}

// scrollList fires the list's scrolled event with a wheel's notches and
// the pixels the platform makes of them, and lays the list out again.
func scrollList(l *widget.List, screen *ebiten.Image, notches, pixels float64) {
	sc := l.ScrollContainer().GetWidget()
	sc.ScrolledEvent.Fire(&widget.ScrolledEventArgs{Widget: sc, Y: notches, DY: pixels})
	event.ExecuteDeferred()
	for range 2 {
		l.Render(screen)
	}
}

// A wheel turned over a list scrolls it the pixels the platform's settings
// make of the turn, which the widget under the pointer is told alongside
// the notches. For ebitenui/ebitenui#200.
func TestList_TheWheelScrollsThePlatformsPixels(t *testing.T) {
	l, screen := longList(t, 10_000)
	d := uitest.Use(t)
	d.Cursor = img.Pt(100, 150)
	d.WheelY, d.ScrollY = -1, -7
	input.Update()
	l.Update(&widget.UpdateObject{})
	input.AfterUpdate()
	event.ExecuteDeferred()
	for range 2 {
		l.Render(screen)
	}
	if got := l.VSlider().Current; got != 7 {
		t.Errorf("a notch the platform makes 7 pixels scrolled %d", got)
	}
}

// A precision touchpad scrolls a fraction of a pixel at a time; the
// fractions add up rather than round away.
func TestList_FractionsOfAPixelAddUp(t *testing.T) {
	l, screen := longList(t, 10_000)
	scrollList(l, screen, -0.1, -0.4)
	if got := l.VSlider().Current; got != 0 {
		t.Errorf("0.4 pixels scrolled %d", got)
	}
	scrollList(l, screen, -0.1, -0.4)
	scrollList(l, screen, -0.1, -0.4)
	if got := l.VSlider().Current; got != 1 {
		t.Errorf("1.2 pixels scrolled %d, want 1", got)
	}
}

// A slider moves a page a notch of the wheel; half a notch, as a touchpad
// gives, waits for the other half.
func TestSlider_HalfNotchesAddUp(t *testing.T) {
	s := newSlider(t, widget.SliderOpts.Direction(widget.DirectionVertical), widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.InitialCurrent(50), widget.SliderOpts.PageSizeFunc(func() int { return 10 }))
	scroll := func(notches float64) {
		w := s.GetWidget()
		w.ScrolledEvent.Fire(&widget.ScrolledEventArgs{Widget: w, Y: notches})
		event.ExecuteDeferred()
	}
	scroll(-0.5)
	if s.Current != 50 {
		t.Errorf("half a notch down moved it to %d", s.Current)
	}
	scroll(-0.5)
	if s.Current != 60 {
		t.Errorf("two halves down moved it to %d, want a page, 60", s.Current)
	}
}

// A text area's slider counts thousandths of how far its text scrolls; the
// platform's pixels scroll it as many pixels.
func TestTextArea_TheWheelScrollsThePlatformsPixels(t *testing.T) {
	a := widget.NewTextArea(
		widget.TextAreaOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
		widget.TextAreaOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
		widget.TextAreaOpts.FontFace(loadFont(t)),
		widget.TextAreaOpts.FontColor(color.White),
		widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.Text(strings.Repeat("a line\n", 40)),
	)
	a.GetWidget().Rect = img.Rect(0, 0, 200, 60)
	render(a, t)
	view := a.ScrollContainer().ViewRect().Dy()
	span := a.ScrollContainer().ContentRect().Dy() - view
	if span <= 0 {
		t.Fatalf("the text scrolls %d pixels: it fits in view", span)
	}
	w := a.ScrollContainer().GetWidget()
	from := a.VSlider().Current
	w.ScrolledEvent.Fire(&widget.ScrolledEventArgs{Widget: w, Y: -1, DY: -float64(span) / 4})
	event.ExecuteDeferred()
	if got := a.VSlider().Current - from; got != 250 {
		t.Errorf("a quarter of the span moved the slider %d thousandths, want 250", got)
	}
}
