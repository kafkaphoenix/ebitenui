package widget

import (
	"image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/event"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ColorPicker picks an opaque color: a wheel of every hue around, red to the
// right, and every saturation outward, white at the middle, and under it a
// slider of the color's brightness, from black to full, and optionally a row
// of preset swatches. With Square, a square of every saturation across and
// every brightness down, of one hue, and under it a slider of hues, take
// the wheel's and the brightness slider's place. The color can be picked by
// pressing the wheel or the square or dragging across it, or by pressing a
// swatch.
type ColorPicker struct {
	ChangedEvent *event.Event

	widgetOpts  []WidgetOpt
	wheelSize   int
	squared     bool
	spacing     int
	handleImage *ButtonImage
	handleSize  int
	swatchSize  int
	presets     []color.NRGBA

	// With rounded swatches, their frames: plain and chosen, outer then
	// inner.
	swatchRadius int
	swatchFrames [2][2]*eimage.NineSlice
	frames       [2]*eimage.NineSlice // the game's swatch frames, plain and chosen
	marker       *ebiten.Image        // the game's marker, in the ring's place

	init      *MultiOnce
	container *Container
	wheel     *colorWheel
	square    *colorSquare  // in the wheel's place, with Square
	slider    *Slider       // of brightness, or of hue with Square
	track     *ebiten.Image // the slider's track: black to the hue at full brightness, or every hue
	ring      *ebiten.Image // the ring that marks the color, drawn once
	ringOp    ebiten.DrawImageOptions

	h, s, v        float64 // the color: hue from 0 to 1 around from red, saturation and value
	sliderDragging bool
}

// ColorPickerChangedEventArgs reports a color picked, and whether the wheel
// or the slider is still being dragged.
type ColorPickerChangedEventArgs struct {
	ColorPicker *ColorPicker
	Color       color.NRGBA
	Dragging    bool
}

type ColorPickerChangedHandlerFunc func(args *ColorPickerChangedEventArgs)

type ColorPickerOpt func(p *ColorPicker)

type ColorPickerOptions struct {
}

var ColorPickerOpts ColorPickerOptions

// trackSteps is how many shades the brightness slider's track holds.
const trackSteps = 64

// hueSteps is how many hues the hue slider of Square holds, a degree each.
const hueSteps = 360

func NewColorPicker(opts ...ColorPickerOpt) *ColorPicker {
	p := &ColorPicker{
		ChangedEvent: &event.Event{},
		init:         &MultiOnce{},
		wheelSize:    160,
		spacing:      8,
		v:            1,
	}
	p.init.Append(p.createWidget)
	for _, o := range opts {
		o(p)
	}
	return p
}

func (o ColorPickerOptions) WidgetOpts(opts ...WidgetOpt) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.widgetOpts = append(p.widgetOpts, opts...)
	}
}

// WheelSize sets the wheel's diameter, or the square's side, in pixels; 160
// by default.
func (o ColorPickerOptions) WheelSize(size int) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.wheelSize = size
	}
}

// Square makes the picker a square of saturations, left to right, and
// brightnesses, top to bottom, of one hue, with a slider of hues under it,
// in place of the wheel and the brightness slider.
func (o ColorPickerOptions) Square() ColorPickerOpt {
	return func(p *ColorPicker) {
		p.squared = true
	}
}

// Spacing sets the space between the wheel and the slider, in pixels; 8 by
// default.
func (o ColorPickerOptions) Spacing(spacing int) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.spacing = spacing
	}
}

// InitialColor sets the color the picker starts with; white by default.
// Its alpha is ignored.
func (o ColorPickerOptions) InitialColor(c color.Color) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.h, p.s, p.v = rgbToHSV(color.NRGBAModel.Convert(c).(color.NRGBA)) //nolint:forcetypeassert // NRGBAModel returns color.NRGBA
	}
}

// Swatches adds a row of square swatches, size pixels wide, of preset
// colors under the slider. Pressing one picks its color; the one of the
// color picked is framed in white over black.
func (o ColorPickerOptions) Swatches(size int, colors ...color.Color) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.swatchSize = size
		for _, c := range colors {
			p.presets = append(p.presets, color.NRGBAModel.Convert(c).(color.NRGBA)) //nolint:forcetypeassert // NRGBAModel returns color.NRGBA
		}
	}
}

// Marker sets the image marking the color picked on the wheel or the
// square, drawn centered on it, in place of the ring the picker draws.
func (o ColorPickerOptions) Marker(marker *ebiten.Image) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.marker = marker
	}
}

// SwatchFrames sets the frames drawn around the swatches, plain and for
// the color picked, in place of the picker's: each swatch's color is drawn
// over its frame, 4 pixels in from its edge.
func (o ColorPickerOptions) SwatchFrames(plain, chosen *eimage.NineSlice) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.frames = [2]*eimage.NineSlice{plain, chosen}
	}
}

// SwatchRadius rounds the swatches' corners by radius pixels; they are
// square by default.
func (o ColorPickerOptions) SwatchRadius(radius int) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.swatchRadius = radius
	}
}

// SliderHandle sets the image and the fixed size of the brightness slider's
// handle; without it, the slider takes them from the theme.
func (o ColorPickerOptions) SliderHandle(handle *ButtonImage, size int) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.handleImage = handle
		p.handleSize = size
	}
}

func (o ColorPickerOptions) ChangedHandler(f ColorPickerChangedHandlerFunc) ColorPickerOpt {
	return func(p *ColorPicker) {
		p.ChangedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*ColorPickerChangedEventArgs); ok {
				f(arg)
			}
		})
	}
}

// Color returns the color picked.
func (p *ColorPicker) Color() color.NRGBA {
	return hsvToRGB(p.h, p.s, p.v)
}

// SetColor sets the color picked, without firing ChangedEvent. Its alpha is
// ignored.
func (p *ColorPicker) SetColor(c color.Color) {
	p.init.Do()
	p.take(color.NRGBAModel.Convert(c).(color.NRGBA)) //nolint:forcetypeassert // NRGBAModel returns color.NRGBA
	p.updateTrack()
	p.slider.Current = p.sliderValue()
}

// take makes c the color picked. A gray, which has no hue, keeps the hue
// picked, so a square stays on it.
func (p *ColorPicker) take(c color.NRGBA) {
	h, s, v := rgbToHSV(c)
	if s == 0 {
		h = p.h
	}
	p.h, p.s, p.v = h, s, v
}

func (p *ColorPicker) GetWidget() *Widget {
	p.init.Do()
	return p.container.GetWidget()
}

func (p *ColorPicker) PreferredSize() (int, int) {
	p.init.Do()
	return p.container.PreferredSize()
}

func (p *ColorPicker) SetLocation(rect image.Rectangle) {
	p.init.Do()
	p.container.SetLocation(rect)
}

func (p *ColorPicker) RequestRelayout() {
	p.init.Do()
	p.container.RequestRelayout()
}

func (p *ColorPicker) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	p.init.Do()
	p.container.SetupInputLayer(def)
}

func (p *ColorPicker) Validate() {
	p.init.Do()
	p.container.Validate()
}

func (p *ColorPicker) Render(screen *ebiten.Image) {
	p.init.Do()
	p.container.Render(screen)
}

func (p *ColorPicker) Update(updObj *UpdateObject) {
	p.init.Do()
	p.container.Update(updObj)
}

func (p *ColorPicker) createWidget() {
	p.container = NewContainer(
		ContainerOpts.WidgetOpts(p.widgetOpts...),
		ContainerOpts.Layout(NewRowLayout(RowLayoutOpts.Direction(DirectionVertical), RowLayoutOpts.Spacing(p.spacing))),
	)
	p.widgetOpts = nil

	if p.ring = p.marker; p.ring == nil {
		p.ring = ringImage(max(4, float64(p.wheelSize)/24))
	}
	steps, sliderMax := trackSteps, 255
	if p.squared {
		p.square = newColorSquare(p)
		p.container.AddChild(p.square)
		steps, sliderMax = hueSteps, hueSteps-1 // drawn once, so a shade a degree
		p.track = ebiten.NewImage(steps, 1)
		p.track.WritePixels(hueTrack())
	} else {
		p.wheel = newColorWheel(p)
		p.container.AddChild(p.wheel)
		p.track = ebiten.NewImage(steps, 1)
		p.updateTrack()
	}
	track := eimage.NewNineSlice(p.track, [3]int{0, steps, 0}, [3]int{0, 1, 0})
	sliderOpts := []SliderOpt{
		SliderOpts.Orientation(DirectionHorizontal),
		SliderOpts.MinMax(0, sliderMax),
		SliderOpts.InitialCurrent(p.sliderValue()),
		SliderOpts.TrackImage(&SliderTrackImage{Idle: track, Hover: track}),
		SliderOpts.WidgetOpts(WidgetOpts.LayoutData(RowLayoutData{Stretch: true})),
		SliderOpts.ChangedHandler(p.sliderChanged),
	}
	if p.handleImage != nil {
		sliderOpts = append(sliderOpts, SliderOpts.HandleImage(p.handleImage), SliderOpts.FixedHandleSize(p.handleSize))
	}
	p.slider = NewSlider(sliderOpts...)
	p.container.AddChild(p.slider)

	if len(p.presets) > 0 {
		if p.frames[0] != nil {
			p.swatchFrames = [2][2]*eimage.NineSlice{{p.frames[0], nil}, {p.frames[1], nil}}
		} else if r := p.swatchRadius; r > 0 {
			for chosen, colors := range [2][2]color.NRGBA{
				{{0x60, 0x60, 0x60, 0xff}, {0x60, 0x60, 0x60, 0xff}},
				{{0xff, 0xff, 0xff, 0xff}, {A: 0xff}},
			} {
				p.swatchFrames[chosen][0] = eimage.NewNineSliceRounded(colors[0], r, nil, 0)
				p.swatchFrames[chosen][1] = eimage.NewNineSliceRounded(colors[1], max(r-2, 0), nil, 0)
			}
		}
		row := NewContainer(ContainerOpts.Layout(NewRowLayout(RowLayoutOpts.Spacing(p.spacing))))
		for _, c := range p.presets {
			row.AddChild(newColorSwatch(p, c))
		}
		p.container.AddChild(row)
	}
}

// sliderValue is the slider's place for the color's brightness, or for its
// hue with Square.
func (p *ColorPicker) sliderValue() int {
	if p.squared {
		return int(math.Round(p.h*hueSteps)) % hueSteps
	}
	return int(math.Round(p.v * 255))
}

// sliderChanged takes a brightness, or a hue with Square, from the slider.
// The slider also reports values set by SetColor, which change nothing.
func (p *ColorPicker) sliderChanged(args *SliderChangedEventArgs) {
	value := &p.v
	x := float64(args.Current) / 255
	if p.squared {
		value, x = &p.h, float64(args.Current)/hueSteps
	}
	ended := p.sliderDragging && !args.Dragging
	p.sliderDragging = args.Dragging
	if x == *value && !ended {
		return
	}
	*value = x
	p.fireChanged(args.Dragging)
}

// pick takes a hue and saturation from the wheel, keeping the brightness,
// or full brightness from black, where no hue shows.
func (p *ColorPicker) pick(h, s float64, dragging bool) {
	p.h, p.s = h, s
	if p.v == 0 {
		p.v = 1
		p.slider.Current = p.sliderValue()
	}
	p.updateTrack()
	p.fireChanged(dragging)
}

// pickSV takes a saturation and a brightness from the square, keeping the
// hue.
func (p *ColorPicker) pickSV(s, v float64, dragging bool) {
	p.s, p.v = s, v
	p.fireChanged(dragging)
}

// choose picks a swatch's color.
func (p *ColorPicker) choose(c color.NRGBA) {
	p.take(c)
	p.updateTrack()
	p.slider.Current = p.sliderValue()
	p.fireChanged(false)
}

func (p *ColorPicker) fireChanged(dragging bool) {
	p.ChangedEvent.Fire(&ColorPickerChangedEventArgs{ColorPicker: p, Color: p.Color(), Dragging: dragging})
}

// updateTrack shades the brightness slider's track from black to the
// color's hue and saturation at full brightness. The hue slider's track of
// Square never changes.
func (p *ColorPicker) updateTrack() {
	if p.track == nil || p.squared {
		return
	}
	pix := make([]byte, 4*trackSteps)
	for x := range trackSteps {
		c := hsvToRGB(p.h, p.s, float64(x)/(trackSteps-1))
		pix[4*x], pix[4*x+1], pix[4*x+2], pix[4*x+3] = c.R, c.G, c.B, 0xff
	}
	p.track.WritePixels(pix)
}

// colorField is the part of a picker picked by pressing it or dragging
// across it, its wheel or its square, as wide and tall as WheelSize. pickAt
// picks the color under a point, from the field's top left corner.
type colorField struct {
	widget   *Widget
	picker   *ColorPicker
	pickAt   func(x, y int, dragging bool)
	dragging bool
	op       ebiten.DrawImageOptions
}

// newColorField returns a field of p that picks with pickAt.
func newColorField(p *ColorPicker, pickAt func(x, y int, dragging bool)) *colorField {
	f := &colorField{widget: NewWidget(), picker: p, pickAt: pickAt}
	f.widget.MouseButtonPressedEvent.AddHandler(func(args any) {
		a, ok := args.(*MouseButtonPressedEventArgs)
		if !ok || a.Button != ebiten.MouseButtonLeft || f.widget.Disabled {
			return
		}
		f.dragging = true
		f.pickAt(a.OffsetX, a.OffsetY, true)
	})
	return f
}

func (f *colorField) GetWidget() *Widget { return f.widget }

func (f *colorField) PreferredSize() (int, int) { return f.picker.wheelSize, f.picker.wheelSize }

func (f *colorField) SetLocation(rect image.Rectangle) { f.widget.Rect = rect }

func (f *colorField) Validate() {}

// Update follows a drag that began on the field, picking under the pointer
// until the button is released.
func (f *colorField) Update(updObj *UpdateObject) {
	f.widget.Update(updObj)
	if !f.dragging {
		return
	}
	x, y := input.CursorPosition()
	at := image.Pt(x, y).Sub(f.widget.Rect.Min)
	if !input.MouseButtonPressed(ebiten.MouseButtonLeft) {
		f.dragging = false
		f.pickAt(at.X, at.Y, false)
		return
	}
	f.pickAt(at.X, at.Y, true)
}

// colorWheel is the picker's disc of hues and saturations, drawn once, with
// a ring where the color lies.
type colorWheel struct {
	*colorField
	disc *ebiten.Image
}

func newColorWheel(p *ColorPicker) *colorWheel {
	w := &colorWheel{disc: wheelDisc(p.wheelSize)}
	w.colorField = newColorField(p, w.pickAt)
	return w
}

func (w *colorWheel) Render(screen *ebiten.Image) {
	w.widget.Render(screen)
	r := w.widget.Rect
	w.op.GeoM.Reset()
	w.op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	w.op.ColorScale.Reset()
	if w.widget.Disabled {
		w.op.ColorScale.ScaleAlpha(0.35)
	}
	screen.DrawImage(w.disc, &w.op)

	radius := float64(w.picker.wheelSize) / 2
	h, s := w.picker.h, w.picker.s
	x := float64(r.Min.X) + radius + s*radius*math.Cos(2*math.Pi*h)
	y := float64(r.Min.Y) + radius - s*radius*math.Sin(2*math.Pi*h)
	w.picker.drawRing(screen, x, y)
}

// ringImage returns the ring that marks the color, of radius r (ringPixels).
// Its pixels are worked out once: stroked with vector each frame, its two
// circles cost a frame about 0.1 ms on Ebitengine 2.10.6, and before 2.10.5
// far more, with a gap where they close.
func ringImage(r float64) *ebiten.Image {
	pix, size := ringPixels(r)
	img := ebiten.NewImage(size, size)
	img.WritePixels(pix)
	return img
}

// ringPixels returns the premultiplied pixels of a ring of radius r, centered
// in a square size pixels wide: a white line 1.5 pixels wide on a black one
// 3 wide, smoothed over a pixel.
func ringPixels(r float64) (pix []byte, size int) {
	size = int(math.Ceil(2*r + 5)) // the black line's half width and a pixel of smoothing on each side
	pix = make([]byte, 4*size*size)
	c := float64(size) / 2
	cover := func(halfWidth, d float64) float64 { return min(1, max(0, halfWidth+0.5-d)) }
	for y := range size {
		for x := range size {
			d := math.Abs(math.Hypot(float64(x)+0.5-c, float64(y)+0.5-c) - r)
			white, black := cover(0.75, d), cover(1.5, d)
			a := white + black*(1-white) // white over black
			i := 4 * (y*size + x)
			w := uint8(math.Round(255 * white))
			pix[i], pix[i+1], pix[i+2], pix[i+3] = w, w, w, uint8(math.Round(255*a))
		}
	}
	return pix, size
}

// drawRing draws the ring centered on x, y.
func (p *ColorPicker) drawRing(screen *ebiten.Image, x, y float64) {
	half := float64(p.ring.Bounds().Dx()) / 2
	p.ringOp.GeoM.Reset()
	p.ringOp.GeoM.Translate(x-half, y-half)
	screen.DrawImage(p.ring, &p.ringOp)
}

// pickAt picks the hue and saturation under a point, from the wheel's top
// left corner; points past the rim take the rim's.
func (w *colorWheel) pickAt(x, y int, dragging bool) {
	r := float64(w.picker.wheelSize) / 2
	dx, dy := float64(x)+0.5-r, r-(float64(y)+0.5)
	h := math.Atan2(dy, dx) / (2 * math.Pi)
	w.picker.pick(h-math.Floor(h), min(1, math.Hypot(dx, dy)/r), dragging)
}

// colorSquare is the picker's square of saturations, left to right, and
// brightnesses, top to bottom, of the color's hue, with a ring where the
// color lies. It is drawn as the hue, under white fading out across it and
// black fading in down it, which is the hue at each saturation and
// brightness: a change of hue draws nothing new.
type colorSquare struct {
	*colorField
	across, down *ebiten.Image // the white and the black, a pixel thick
	fill         *ebiten.Image // a white pixel, tinted the hue
}

func newColorSquare(p *ColorPicker) *colorSquare {
	n := p.wheelSize
	across, down := make([]byte, 4*n), make([]byte, 4*n)
	for i := range n {
		white := uint8(math.Round(255 * (1 - float64(i)/float64(max(1, n-1)))))
		black := uint8(math.Round(255 * float64(i) / float64(max(1, n-1))))
		across[4*i], across[4*i+1], across[4*i+2], across[4*i+3] = white, white, white, white // premultiplied
		down[4*i+3] = black
	}
	s := &colorSquare{across: ebiten.NewImage(n, 1), down: ebiten.NewImage(1, n), fill: eimage.NewImageColor(color.White)}
	s.across.WritePixels(across)
	s.down.WritePixels(down)
	s.colorField = newColorField(p, s.pickAt)
	return s
}

func (s *colorSquare) Render(screen *ebiten.Image) {
	s.widget.Render(screen)
	r := s.widget.Rect
	w, h := float64(r.Dx()), float64(r.Dy())
	alpha := float32(1)
	if s.widget.Disabled {
		alpha = 0.35
	}
	hue := hsvToRGB(s.picker.h, 1, 1)
	s.draw(screen, s.fill, float32(hue.R)/0xff, float32(hue.G)/0xff, float32(hue.B)/0xff, alpha)
	s.draw(screen, s.across, 1, 1, 1, alpha)
	s.draw(screen, s.down, 1, 1, 1, alpha)

	s.picker.drawRing(screen, float64(r.Min.X)+s.picker.s*w, float64(r.Min.Y)+(1-s.picker.v)*h)
}

// draw draws img stretched over the square, tinted r, g, b, at alpha.
func (s *colorSquare) draw(screen, img *ebiten.Image, r, g, b, alpha float32) {
	rect, size := s.widget.Rect, img.Bounds()
	s.op.GeoM.Reset()
	s.op.GeoM.Scale(float64(rect.Dx())/float64(size.Dx()), float64(rect.Dy())/float64(size.Dy()))
	s.op.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
	s.op.ColorScale.Reset()
	s.op.ColorScale.Scale(r, g, b, 1)
	s.op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(img, &s.op)
}

// pickAt picks the saturation and the brightness under a point, from the
// square's top left corner; points past its edges take the edges'.
func (s *colorSquare) pickAt(x, y int, dragging bool) {
	r := s.widget.Rect
	clamp := func(f float64) float64 { return min(1, max(0, f)) }
	s.picker.pickSV(clamp((float64(x)+0.5)/float64(r.Dx())), clamp(1-(float64(y)+0.5)/float64(r.Dy())), dragging)
}

// hueTrack returns the pixels of the hue slider's track, hueSteps wide:
// every hue, a degree a pixel, from red, the slider's 0.
func hueTrack() []byte {
	pix := make([]byte, 4*hueSteps)
	for x := range hueSteps {
		c := hsvToRGB(float64(x)/hueSteps, 1, 1)
		pix[4*x], pix[4*x+1], pix[4*x+2], pix[4*x+3] = c.R, c.G, c.B, 0xff
	}
	return pix
}

// colorSwatch is a square of a preset color, framed in white over black
// while it is the color picked and in gray otherwise.
type colorSwatch struct {
	widget *Widget
	picker *ColorPicker
	color  color.NRGBA
	fill   *eimage.NineSlice // the color, rounded; nil for square swatches
}

func newColorSwatch(p *ColorPicker, c color.NRGBA) *colorSwatch {
	s := &colorSwatch{widget: NewWidget(), picker: p, color: c}
	switch {
	case p.swatchRadius > 0:
		s.fill = eimage.NewNineSliceRounded(c, max(p.swatchRadius-4, 0), nil, 0)
	case p.frames[0] != nil:
		s.fill = eimage.NewNineSliceColor(c)
	}
	s.widget.MouseButtonPressedEvent.AddHandler(func(args any) {
		a, ok := args.(*MouseButtonPressedEventArgs)
		if !ok || a.Button != ebiten.MouseButtonLeft || s.widget.Disabled {
			return
		}
		p.choose(s.color)
	})
	return s
}

func (s *colorSwatch) GetWidget() *Widget { return s.widget }

func (s *colorSwatch) PreferredSize() (int, int) { return s.picker.swatchSize, s.picker.swatchSize }

func (s *colorSwatch) SetLocation(rect image.Rectangle) { s.widget.Rect = rect }

func (s *colorSwatch) Validate() {}

func (s *colorSwatch) Update(updObj *UpdateObject) { s.widget.Update(updObj) }

// chosen reports whether the swatch's color is the color picked.
func (s *colorSwatch) chosen() bool { return s.picker.Color() == s.color }

func (s *colorSwatch) Render(screen *ebiten.Image) {
	s.widget.Render(screen)
	if s.fill != nil {
		chosen := 0
		if s.chosen() {
			chosen = 1
		}
		s.drawRounded(screen, s.picker.swatchFrames[chosen][0], 0)
		s.drawRounded(screen, s.picker.swatchFrames[chosen][1], 2)
		s.drawRounded(screen, s.fill, 4)
		return
	}
	outer, inner := color.NRGBA{0x60, 0x60, 0x60, 0xff}, color.NRGBA{0x60, 0x60, 0x60, 0xff}
	if s.chosen() {
		outer, inner = color.NRGBA{0xff, 0xff, 0xff, 0xff}, color.NRGBA{A: 0xff}
	}
	r := s.widget.Rect
	fill := func(inset int, c color.NRGBA) {
		x, y := float32(r.Min.X+inset), float32(r.Min.Y+inset)
		w, h := float32(r.Dx()-2*inset), float32(r.Dy()-2*inset)
		if w > 0 && h > 0 {
			vector.FillRect(screen, x, y, w, h, c, false)
		}
	}
	fill(0, outer)
	fill(2, inner)
	fill(4, s.color)
}

// drawRounded draws a rounded layer of the swatch, inset pixels inside its
// edge.
func (s *colorSwatch) drawRounded(screen *ebiten.Image, n *eimage.NineSlice, inset int) {
	r := s.widget.Rect.Inset(inset)
	if r.Empty() || n == nil {
		return
	}
	n.Draw(screen, r.Dx(), r.Dy(), func(opts *ebiten.DrawImageOptions) {
		opts.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	})
}

// wheelDisc draws a disc size pixels wide of every hue around and every
// saturation outward at full brightness, its rim smoothed over a pixel.
func wheelDisc(size int) *ebiten.Image {
	pix := make([]byte, 4*size*size)
	r := float64(size) / 2
	for y := range size {
		for x := range size {
			dx, dy := float64(x)+0.5-r, r-(float64(y)+0.5)
			d := math.Hypot(dx, dy) / r
			if d > 1 {
				continue
			}
			c := hsvToRGB(math.Atan2(dy, dx)/(2*math.Pi), d, 1)
			edge := min(1, (1-d)*r)
			i := 4 * (y*size + x)
			pix[i], pix[i+1], pix[i+2], pix[i+3] = uint8(float64(c.R)*edge), uint8(float64(c.G)*edge), uint8(float64(c.B)*edge), uint8(255*edge)
		}
	}
	disc := ebiten.NewImage(size, size)
	disc.WritePixels(pix)
	return disc
}

// hsvToRGB returns the opaque color of a hue, from 0 to 1 around from red,
// a saturation and a value, both from 0 to 1.
func hsvToRGB(h, s, v float64) color.NRGBA {
	h = (h - math.Floor(h)) * 6
	sector := math.Floor(h)
	f := h - sector
	p, q, t := v*(1-s), v*(1-s*f), v*(1-s*(1-f))
	var r, g, b float64
	switch int(sector) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	channel := func(x float64) uint8 { return uint8(math.Round(x * 255)) }
	return color.NRGBA{R: channel(r), G: channel(g), B: channel(b), A: 0xff}
}

// rgbToHSV returns a color's hue, from 0 to 1 around from red, its
// saturation and its value; gray has hue 0.
func rgbToHSV(c color.NRGBA) (h, s, v float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	hi, lo := max(r, g, b), min(r, g, b)
	v = hi
	if hi == 0 {
		return 0, 0, 0
	}
	s = (hi - lo) / hi
	if hi == lo {
		return 0, s, v
	}
	switch hi {
	case r:
		h = (g - b) / (hi - lo)
	case g:
		h = 2 + (b-r)/(hi-lo)
	default:
		h = 4 + (r-g)/(hi-lo)
	}
	h /= 6
	return h - math.Floor(h), s, v
}
