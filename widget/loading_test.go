package widget_test

import (
	img "image"
	"image/color"
	"testing"
	"time"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// newLoading returns a validated loading indicator at rect.
func newLoading(t *testing.T, rect img.Rectangle, opts ...widget.LoadingOpt) *widget.Loading {
	t.Helper()
	l := widget.NewLoading(append([]widget.LoadingOpt{widget.LoadingOpts.TipFace(loadFont(t))}, opts...)...)
	l.Validate()
	l.SetLocation(rect)
	return l
}

// marks returns how many changes n updates of l mark, those UI.Changed
// tells of, which keep a game that skips unchanged frames drawing them.
func marks(l *widget.Loading, n int) int {
	c := changes.Count()
	for range n {
		l.Update(&widget.UpdateObject{})
	}
	return int(changes.Count() - c)
}

// A spinner turns a dot a step, Period over its dots, and marks a change
// only as it steps: 8 a second by default.
func TestLoading_SpinnerSteps(t *testing.T) {
	is := is.New(t)

	l := newLoading(t, img.Rect(0, 0, 32, 32), widget.LoadingOpts.Spinner())
	is.Equal(l.Step(), 0)
	is.Equal(marks(l, ebiten.TPS()), 8)
	is.Equal(l.Step(), 0)      // a turn
	marks(l, ebiten.TPS()/8+1) // past the first step's 7.5 updates
	is.Equal(l.Step(), 1)
}

// A bar's band sweeps along its track and marks a change only as it moves
// a pixel: every update on a wide bar, seldom on a narrow, slow one.
func TestLoading_BarSweeps(t *testing.T) {
	is := is.New(t)

	wide := newLoading(t, img.Rect(0, 0, 200, 8), widget.LoadingOpts.Period(time.Second))
	from := wide.BandX()
	is.Equal(from, -60) // a band of 30 % of the track, starting off its left
	is.Equal(marks(wide, ebiten.TPS()), ebiten.TPS())
	is.Equal(wide.BandX(), from) // a sweep

	marks(wide, ebiten.TPS()/2)
	is.Equal(wide.BandX(), 70) // halfway along 260 pixels

	narrow := newLoading(t, img.Rect(0, 0, 10, 8), widget.LoadingOpts.Period(10*time.Second))
	is.True(marks(narrow, 10*ebiten.TPS()) <= 13) // 13 pixels a sweep
}

// Tips show under the indicator, the next every TipEvery; a tip alone
// changes nothing.
func TestLoading_Tips(t *testing.T) {
	is := is.New(t)

	slow := widget.LoadingOpts.Period(1000 * time.Second) // so that the spinner keeps still
	one := newLoading(t, img.Rect(0, 0, 200, 60), widget.LoadingOpts.Spinner(), slow, widget.LoadingOpts.Tips([]string{"only"}, time.Second))
	is.Equal(one.Tip(), "only")
	is.Equal(marks(one, 5*ebiten.TPS()), 0)

	l := newLoading(t, img.Rect(0, 0, 200, 60), widget.LoadingOpts.Spinner(), slow, widget.LoadingOpts.Tips([]string{"a", "b", "c"}, time.Second))
	is.Equal(l.Tip(), "a")
	is.Equal(marks(l, ebiten.TPS()), 1)
	is.Equal(l.Tip(), "b")
	is.Equal(marks(l, 2*ebiten.TPS()), 2)
	is.Equal(l.Tip(), "a") // round again
}

// A bar spans its widget unless given a width; a narrower bar, and a
// spinner, sit at its start, center or end, the tips under them.
func TestLoading_Place(t *testing.T) {
	is := is.New(t)

	rect := img.Rect(10, 20, 210, 80)
	is.Equal(newLoading(t, rect).IndicatorRect(), img.Rect(10, 20, 210, 28)) // 8 high, spanning
	is.Equal(newLoading(t, rect, widget.LoadingOpts.BarSize(50, 6), widget.LoadingOpts.Position(widget.TextPositionEnd)).IndicatorRect(),
		img.Rect(160, 20, 210, 26))
	is.Equal(newLoading(t, rect, widget.LoadingOpts.BarSize(50, 6), widget.LoadingOpts.Position(widget.TextPositionCenter)).IndicatorRect(),
		img.Rect(85, 20, 135, 26))
	is.Equal(newLoading(t, rect, widget.LoadingOpts.Spinner(), widget.LoadingOpts.SpinnerSize(24)).IndicatorRect(),
		img.Rect(10, 20, 34, 44)) // at the start by default
	is.Equal(newLoading(t, rect, widget.LoadingOpts.Spinner(), widget.LoadingOpts.Position(widget.TextPositionCenter)).IndicatorRect(),
		img.Rect(94, 20, 126, 52)) // 32 across by default
}

// Its preferred size is its indicator's, and its tips' under it.
func TestLoading_PreferredSize(t *testing.T) {
	is := is.New(t)

	w, h := newLoading(t, img.Rectangle{}, widget.LoadingOpts.BarSize(120, 6)).PreferredSize()
	is.Equal([]int{w, h}, []int{120, 6})
	w, h = newLoading(t, img.Rectangle{}, widget.LoadingOpts.Spinner()).PreferredSize()
	is.Equal([]int{w, h}, []int{32, 32})
	l := newLoading(t, img.Rectangle{}, widget.LoadingOpts.Spinner(), widget.LoadingOpts.Tips([]string{"a tip far wider than the spinner"}, time.Second))
	tw, th := l.TipText().PreferredSize()
	w, h = l.PreferredSize()
	is.Equal([]int{w, h}, []int{tw, 32 + 4 + th})
}

// An update of a loading indicator allocates nothing, bar, spinner or tips.
// (Its draws are measured in a running game: before one, Ebitengine
// records draws, which allocates.)
func TestLoading_UpdatesWithoutAllocating(t *testing.T) {
	screen := ebiten.NewImage(200, 80)
	for name, opts := range map[string][]widget.LoadingOpt{
		"bar":     nil,
		"spinner": {widget.LoadingOpts.Spinner()},
		"tips":    {widget.LoadingOpts.Tips([]string{"a", "b"}, time.Second)},
	} {
		l := newLoading(t, img.Rect(0, 0, 200, 80), opts...)
		l.Render(screen)
		if n := testing.AllocsPerRun(100, func() { l.Update(&widget.UpdateObject{}) }); n != 0 {
			t.Errorf("%s: an update allocated %v times", name, n)
		}
	}
}

// A theme gives a loading indicator its track, band, dots and tips' look.
func TestLoading_Theme(t *testing.T) {
	is := is.New(t)

	track, fill := eimage.NewNineSliceColor(color.White), eimage.NewNineSliceColor(color.Black)
	l := widget.NewLoading()
	l.GetWidget().SetTheme(&widget.Theme{LoadingTheme: &widget.LoadingParams{
		TrackImage: track, FillImage: fill, Color: color.Black, TipFace: loadFont(t), TipColor: color.Black,
	}})
	l.Validate()
	p := l.ComputedParams()
	is.Equal([]any{p.TrackImage, p.FillImage, p.Color, p.TipColor}, []any{track, fill, color.Black, color.Black})
}
