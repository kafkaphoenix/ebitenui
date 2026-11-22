package widget

import (
	img "image"
	"image/color"
	"math"
	"time"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// LoadingParams is a loading indicator's look: a bar's track and the band
// sweeping along it, a spinner's dots, and the tips' face and color.
type LoadingParams struct {
	TrackImage *image.NineSlice
	FillImage  *image.NineSlice
	Color      color.Color
	TipFace    *text.Face
	TipColor   color.Color
}

// Loading shows that something is loading, for as long as it is in the
// interface: a bar, a band sweeping along a track again and again, or a
// spinner, a ring of dots turning, the dots behind the brightest fading. A
// bar spans its widget unless given a width; a narrower bar, and a spinner,
// sit at the widget's start, center or end (Position). Tips, if given,
// show under it, the next every while. For progress that is known, a
// ProgressBar.
//
// It counts its animation in updates, at Ebitengine's TPS, and tells
// UI.Changed only as what it draws changes: the band moving a pixel, the
// spinner turning a dot, the next tip.
type Loading struct {
	definedParams  LoadingParams
	computedParams LoadingParams

	spinner  bool
	barSize  img.Point // a bar's width, 0 to span the widget, and height
	diameter int       // a spinner's
	position TextPosition
	period   time.Duration // a bar's sweep, a spinner's turn
	tips     []string
	tipEvery time.Duration

	updates int // since it was made

	tip  *Text
	dot  *ebiten.Image // a spinner's dot, drawn once
	dots [loadingDots]img.Point
	// What draws are placed with, made once.
	at      img.Point
	place   image.DrawImageOptionsFunc
	dotOpts ebiten.DrawImageOptions

	widgetOpts []WidgetOpt
	init       *MultiOnce
	widget     *Widget
}

// loadingDots is the number of a spinner's dots.
const loadingDots = 8

// loadingTipGap is the space between the indicator and its tips.
const loadingTipGap = 4

type LoadingOpt func(l *Loading)

type LoadingOptions struct{}

// LoadingOpts holds the options of NewLoading.
var LoadingOpts LoadingOptions

// NewLoading returns a loading indicator: by default a bar 8 pixels high
// spanning its widget, sweeping once a second.
func NewLoading(opts ...LoadingOpt) *Loading {
	l := &Loading{
		barSize:  img.Pt(0, 8),
		diameter: 32,
		position: TextPositionStart,
		period:   time.Second,
		tipEvery: 5 * time.Second,
		init:     &MultiOnce{},
	}
	l.place = func(opts *ebiten.DrawImageOptions) { opts.GeoM.Translate(float64(l.at.X), float64(l.at.Y)) }
	l.init.Append(l.createWidget)
	for _, o := range opts {
		o(l)
	}
	return l
}

func (o LoadingOptions) WidgetOpts(opts ...WidgetOpt) LoadingOpt {
	return func(l *Loading) { l.widgetOpts = append(l.widgetOpts, opts...) }
}

// Spinner makes the indicator a spinner, in place of a bar.
func (o LoadingOptions) Spinner() LoadingOpt {
	return func(l *Loading) { l.spinner = true }
}

// BarSize sets a bar's width, 0 to span the widget, and height.
func (o LoadingOptions) BarSize(width, height int) LoadingOpt {
	return func(l *Loading) { l.barSize = img.Pt(width, height) }
}

// SpinnerSize sets a spinner's diameter.
func (o LoadingOptions) SpinnerSize(diameter int) LoadingOpt {
	return func(l *Loading) { l.diameter = diameter }
}

// Position sets where a bar narrower than the widget, or a spinner, and
// the tips sit across the widget: at its start, the left, by default, its
// center or its end.
func (o LoadingOptions) Position(p TextPosition) LoadingOpt {
	return func(l *Loading) { l.position = p }
}

// Period sets how long a bar's band takes to sweep along it, or a spinner
// to turn: a second by default.
func (o LoadingOptions) Period(d time.Duration) LoadingOpt {
	return func(l *Loading) { l.period = d }
}

// Tips shows tips under the indicator, the next every every, in turn.
func (o LoadingOptions) Tips(tips []string, every time.Duration) LoadingOpt {
	return func(l *Loading) { l.tips, l.tipEvery = tips, every }
}

func (o LoadingOptions) TrackImage(i *image.NineSlice) LoadingOpt {
	return func(l *Loading) { l.definedParams.TrackImage = i }
}

func (o LoadingOptions) FillImage(i *image.NineSlice) LoadingOpt {
	return func(l *Loading) { l.definedParams.FillImage = i }
}

// Color sets a spinner's dots' color.
func (o LoadingOptions) Color(c color.Color) LoadingOpt {
	return func(l *Loading) { l.definedParams.Color = c }
}

func (o LoadingOptions) TipFace(f *text.Face) LoadingOpt {
	return func(l *Loading) { l.definedParams.TipFace = f }
}

func (o LoadingOptions) TipColor(c color.Color) LoadingOpt {
	return func(l *Loading) { l.definedParams.TipColor = c }
}

func (l *Loading) GetWidget() *Widget {
	l.init.Do()
	return l.widget
}

func (l *Loading) SetLocation(rect img.Rectangle) {
	l.init.Do()
	l.widget.Rect = rect
}

// Validate takes the look from the options, then the theme, and else
// defaults: a translucent white track, a white band and dots.
func (l *Loading) Validate() {
	l.init.Do()
	p := LoadingParams{}
	if theme := l.widget.GetTheme(); theme != nil {
		if t := theme.LoadingTheme; t != nil {
			p = *t
		}
		if p.TipFace == nil {
			p.TipFace = theme.DefaultFace
		}
		if p.TipColor == nil {
			p.TipColor = theme.DefaultTextColor
		}
	}
	d := l.definedParams
	p.TrackImage = first(d.TrackImage, p.TrackImage, image.NewNineSliceColor(color.NRGBA{255, 255, 255, 40}))
	p.FillImage = first(d.FillImage, p.FillImage, image.NewNineSliceColor(color.White))
	p.Color = first(d.Color, p.Color, color.Color(color.White))
	p.TipFace = first(d.TipFace, p.TipFace)
	p.TipColor = first(d.TipColor, p.TipColor, color.Color(color.White))
	l.computedParams = p

	if len(l.tips) > 0 {
		if p.TipFace == nil {
			panic("Loading: TipFace is required for tips.")
		}
		l.tip = NewText(TextOpts.Text(l.tips[l.tipIndex()], p.TipFace, p.TipColor), TextOpts.Position(l.position, TextPositionStart))
		l.tip.Validate()
	}
	if l.spinner {
		l.makeDot()
	}
}

// first returns the first of vs that is not nil, or nil.
func first[T comparable](vs ...T) T {
	var zero T
	for _, v := range vs {
		if v != zero {
			return v
		}
	}
	return zero
}

// makeDot draws a spinner's dot once, and finds where each dot goes around
// its center, the first at the top and the rest clockwise.
func (l *Loading) makeDot() {
	r := max(1, l.diameter/10)
	l.dot = ebiten.NewImage(2*r, 2*r)
	vector.FillCircle(l.dot, float32(r), float32(r), float32(r), color.White, true)
	ring := float64(l.diameter)/2 - float64(r)
	for i := range l.dots {
		a := 2*math.Pi*float64(i)/loadingDots - math.Pi/2
		l.dots[i] = img.Pt(int(math.Round(ring*math.Cos(a)))-r, int(math.Round(ring*math.Sin(a)))-r)
	}
}

// PreferredSize is the indicator's size, and its tips' under it.
func (l *Loading) PreferredSize() (int, int) {
	l.init.Do()
	w, h := l.indicatorSize(0)
	if l.tip != nil {
		tw, th := l.tip.PreferredSize()
		w, h = max(w, tw), h+loadingTipGap+th
	}
	if l.widget != nil {
		w, h = max(w, l.widget.MinWidth), max(h, l.widget.MinHeight)
	}
	return w, h
}

// indicatorSize returns the bar's or the spinner's size in a widget span
// pixels wide.
func (l *Loading) indicatorSize(span int) (int, int) {
	if l.spinner {
		return l.diameter, l.diameter
	}
	if l.barSize.X > 0 {
		return l.barSize.X, l.barSize.Y
	}
	return span, l.barSize.Y
}

// indicatorRect returns where the bar or the spinner is drawn: at the
// widget's top, across it where Position puts it.
func (l *Loading) indicatorRect() img.Rectangle {
	r := l.widget.Rect
	w, h := l.indicatorSize(r.Dx())
	x := r.Min.X
	switch l.position {
	case TextPositionStart:
	case TextPositionCenter:
		x += (r.Dx() - w) / 2
	case TextPositionEnd:
		x = r.Max.X - w
	}
	return img.Rect(x, r.Min.Y, x+w, r.Min.Y+h)
}

func (l *Loading) periodUpdates() int { return updatesIn(l.period) }

// step returns the spinner's brightest dot.
func (l *Loading) step() int {
	return l.updates * loadingDots / l.periodUpdates() % loadingDots
}

// bandX returns where the bar's band starts along its track: a band of 30 %
// of the track, sweeping from off its left to off its right.
func (l *Loading) bandX() int {
	w := l.indicatorRect().Dx()
	band := w * 3 / 10
	p := l.periodUpdates()
	return int(math.Floor(float64(-band) + float64(l.updates%p)/float64(p)*float64(w+band)))
}

func (l *Loading) tipIndex() int {
	return l.updates / updatesIn(l.tipEvery) % len(l.tips)
}

// Update moves the animation an update on, telling UI.Changed as what it
// draws changes.
func (l *Loading) Update(updObj *UpdateObject) {
	l.init.Do()
	l.widget.Update(updObj)
	step, band, tip := l.step(), l.bandX(), 0
	if l.tip != nil {
		tip = l.tipIndex()
	}
	l.updates++
	switch {
	case l.spinner && l.step() != step, !l.spinner && l.bandX() != band:
		changes.Mark()
	}
	if l.tip != nil {
		if t := l.tipIndex(); t != tip {
			l.tip.Label = l.tips[t]
			changes.Mark()
		}
	}
}

func (l *Loading) Render(screen *ebiten.Image) {
	l.init.Do()
	l.widget.Render(screen)
	ir := l.indicatorRect()
	if l.spinner {
		l.drawSpinner(screen, ir)
	} else {
		l.drawBar(screen, ir)
	}
	if l.tip != nil {
		_, th := l.tip.PreferredSize()
		r := l.widget.Rect
		l.tip.SetLocation(img.Rect(r.Min.X, ir.Max.Y+loadingTipGap, r.Max.X, ir.Max.Y+loadingTipGap+th))
		l.tip.Render(screen)
	}
}

func (l *Loading) drawBar(screen *ebiten.Image, ir img.Rectangle) {
	l.at = ir.Min
	l.computedParams.TrackImage.Draw(screen, ir.Dx(), ir.Dy(), l.place)
	x := l.bandX()
	x0, x1 := max(x, 0), min(x+ir.Dx()*3/10, ir.Dx())
	if x1 > x0 {
		l.at = ir.Min.Add(img.Pt(x0, 0))
		l.computedParams.FillImage.Draw(screen, x1-x0, ir.Dy(), l.place)
	}
}

func (l *Loading) drawSpinner(screen *ebiten.Image, ir img.Rectangle) {
	c := ir.Min.Add(img.Pt(ir.Dx()/2, ir.Dy()/2))
	head := l.step()
	for i, at := range l.dots {
		behind := (head - i + loadingDots) % loadingDots
		o := &l.dotOpts
		o.GeoM.Reset()
		o.GeoM.Translate(float64(c.X+at.X), float64(c.Y+at.Y))
		o.ColorScale.Reset()
		o.ColorScale.ScaleWithColor(l.computedParams.Color)
		o.ColorScale.ScaleAlpha(1 - 0.85*float32(behind)/loadingDots)
		screen.DrawImage(l.dot, o)
	}
}

func (l *Loading) createWidget() {
	l.widget = NewWidget(l.widgetOpts...)
	l.widgetOpts = nil
}
