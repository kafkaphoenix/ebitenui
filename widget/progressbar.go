package widget

import (
	img "image"

	"github.com/ebitenui/ebitenui/image"

	"github.com/hajimehoshi/ebiten/v2"
)

type ProgressBarParams struct {
	TrackImage   *ProgressBarImage
	FillImage    *ProgressBarImage
	TrackPadding *Insets
}

type ProgressBar struct {
	definedParams  ProgressBarParams
	computedParams ProgressBarParams

	Min     int
	Max     int
	current int

	widgetOpts []WidgetOpt
	direction  Direction
	inverted   bool

	init   *MultiOnce
	widget *Widget
}

type ProgressBarImage struct {
	Idle     *image.NineSlice
	Hover    *image.NineSlice
	Disabled *image.NineSlice
}

type ProgressBarOpt func(s *ProgressBar)

type ProgressBarOptions struct {
}

var ProgressBarOpts ProgressBarOptions

func NewProgressBar(opts ...ProgressBarOpt) *ProgressBar {
	pb := &ProgressBar{
		Min:     1,
		Max:     100,
		current: 1,

		init: &MultiOnce{},
	}

	pb.init.Append(pb.createWidget)

	for _, o := range opts {
		o(pb)
	}

	return pb
}

func (p *ProgressBar) Validate() {
	p.init.Do()
	p.populateComputedParams()

	if p.computedParams.TrackImage == nil {
		panic("ProgressBar: TrackImage is required.")
	}
	if p.computedParams.TrackImage.Idle == nil {
		panic("ProgressBar: TrackImage.Idle is required")
	}
}

func (p *ProgressBar) populateComputedParams() {
	params := ProgressBarParams{}
	theme := p.widget.GetTheme()
	if theme != nil {
		if theme.ProgressBarTheme != nil {
			params.FillImage = theme.ProgressBarTheme.FillImage
			params.TrackImage = theme.ProgressBarTheme.TrackImage
			params.TrackPadding = theme.ProgressBarTheme.TrackPadding
		}
	}
	if p.definedParams.FillImage != nil {
		if params.FillImage == nil {
			params.FillImage = p.definedParams.FillImage
		} else {
			if p.definedParams.FillImage.Idle != nil {
				params.FillImage.Idle = p.definedParams.FillImage.Idle
			}
			if p.definedParams.FillImage.Hover != nil {
				params.FillImage.Hover = p.definedParams.FillImage.Hover
			}
			if p.definedParams.FillImage.Disabled != nil {
				params.FillImage.Disabled = p.definedParams.FillImage.Disabled
			}
		}
	}
	if p.definedParams.TrackImage != nil {
		if params.TrackImage == nil {
			params.TrackImage = p.definedParams.TrackImage
		} else {
			if p.definedParams.TrackImage.Idle != nil {
				params.TrackImage.Idle = p.definedParams.TrackImage.Idle
			}
			if p.definedParams.TrackImage.Hover != nil {
				params.TrackImage.Hover = p.definedParams.TrackImage.Hover
			}
			if p.definedParams.TrackImage.Disabled != nil {
				params.TrackImage.Disabled = p.definedParams.TrackImage.Disabled
			}
		}
	}

	if p.definedParams.TrackPadding != nil {
		params.TrackPadding = p.definedParams.TrackPadding
	}

	if params.TrackPadding == nil {
		params.TrackPadding = &Insets{}
	}

	p.computedParams = params
}

func (o ProgressBarOptions) WidgetOpts(opts ...WidgetOpt) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.widgetOpts = append(s.widgetOpts, opts...)
	}
}

// Direction sets the direction of the progress bar.
// The default is horizontal.
func (o ProgressBarOptions) Direction(d Direction) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.direction = d
	}
}

// Inverted sets whether the progress bar is inverted.
// The default is false, which means from left to right or top to bottom.
func (o ProgressBarOptions) Inverted(inverted bool) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.inverted = inverted
	}
}

func (o ProgressBarOptions) Images(track *ProgressBarImage, fill *ProgressBarImage) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.definedParams.TrackImage = track
		s.definedParams.FillImage = fill
	}
}

func (o ProgressBarOptions) TrackImage(track *ProgressBarImage) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.definedParams.TrackImage = track
	}
}

func (o ProgressBarOptions) FillImage(fill *ProgressBarImage) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.definedParams.FillImage = fill
	}
}

func (o ProgressBarOptions) TrackPadding(i *Insets) ProgressBarOpt {
	return func(s *ProgressBar) {
		s.definedParams.TrackPadding = i
	}
}

func (o ProgressBarOptions) Values(lowest, highest, current int) ProgressBarOpt {
	return func(p *ProgressBar) {
		p.Min = lowest
		p.Max = highest
		p.current = current
	}
}

func (p *ProgressBar) GetWidget() *Widget {
	p.init.Do()
	return p.widget
}

func (p *ProgressBar) PreferredSize() (int, int) {
	w, h := p.computedParams.TrackImage.Idle.MinSize()
	if p.widget != nil {
		if w < p.widget.MinWidth {
			w = p.widget.MinWidth
		}
		if h < p.widget.MinHeight {
			h = p.widget.MinHeight
		}
	}
	return w, h
}

func (p *ProgressBar) SetLocation(rect img.Rectangle) {
	p.init.Do()
	p.widget.Rect = rect
}

func (p *ProgressBar) Render(screen *ebiten.Image) {
	p.init.Do()
	p.widget.Render(screen)
	p.draw(screen)
}

func (p *ProgressBar) Update(updObj *UpdateObject) {
	p.init.Do()

	p.widget.Update(updObj)
}

func (p *ProgressBar) draw(screen *ebiten.Image) {
	i := p.computedParams.TrackImage.Idle
	var fill *image.NineSlice
	if p.computedParams.FillImage != nil {
		fill = p.computedParams.FillImage.Idle
	}
	if p.widget.Disabled {
		if p.computedParams.TrackImage.Disabled != nil {
			i = p.computedParams.TrackImage.Disabled
		}
		if p.computedParams.FillImage.Disabled != nil {
			fill = p.computedParams.FillImage.Disabled
		}
	}

	if i != nil {
		i.Draw(screen, p.widget.Rect.Dx(), p.widget.Rect.Dy(), p.widget.drawImageOptions)
	}
	if fill != nil && p.currentPercentage() > 0 {
		fillX := p.widget.Rect.Dx() - p.computedParams.TrackPadding.Left - p.computedParams.TrackPadding.Right
		fillY := p.widget.Rect.Dy() - p.computedParams.TrackPadding.Top - p.computedParams.TrackPadding.Bottom
		if p.direction == DirectionHorizontal {
			fillX = int(float64(fillX) * p.currentPercentage())
		} else {
			fillY = int(float64(fillY) * p.currentPercentage())
		}
		fill.Draw(screen, fillX, fillY, func(opts *ebiten.DrawImageOptions) {
			tx := p.widget.Rect.Min.X + p.computedParams.TrackPadding.Left
			ty := p.widget.Rect.Min.Y + p.computedParams.TrackPadding.Top
			if p.inverted {
				if p.direction == DirectionHorizontal {
					tx = p.widget.Rect.Max.X - p.computedParams.TrackPadding.Right - fillX
				} else {
					ty = p.widget.Rect.Max.Y - p.computedParams.TrackPadding.Bottom - fillY
				}
			}
			opts.GeoM.Translate(float64(tx), float64(ty))
		})
	}
}

func (p *ProgressBar) currentPercentage() float64 {
	if p.Max <= p.Min {
		return 0
	}

	return float64(p.current-p.Min) / float64(p.Max-p.Min)
}

func (p *ProgressBar) SetCurrent(value int) bool {
	oldValue := p.current
	switch {
	case value < p.Min:
		p.current = p.Min
	case value > p.Max:
		p.current = p.Max
	default:
		p.current = value
	}
	return oldValue != p.current
}

func (p *ProgressBar) GetCurrent() int {
	return p.current
}

func (p *ProgressBar) createWidget() {
	p.widget = NewWidget(append([]WidgetOpt{WidgetOpts.TrackHover(true)}, p.widgetOpts...)...)
}
