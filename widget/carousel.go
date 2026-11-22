package widget

import (
	img "image"
	"image/color"
	"time"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// CarouselParams is a carousel's look: its arrows' buttons, the arrows
// drawn on them and their size, or icons of the game's in their place
// (PrevGraphic, NextGraphic), and its dots' color, a done step's color,
// size and spacing, or an image of the game's for a dot (DotImage), tinted
// by those colors.
type CarouselParams struct {
	ArrowImage   *ButtonImage
	ArrowColor   color.Color
	ArrowSize    *img.Point
	PrevGraphic  *GraphicImage
	NextGraphic  *GraphicImage
	DotColor     color.Color
	DotDoneColor color.Color
	DotSize      *int
	DotSpacing   *int
	DotImage     *ebiten.Image
}

// Carousel shows pages one at a time, any widgets: the screens of a
// loading screen, a tutorial's steps. Arrows at the sides of the page area
// move it back and on, a row of dots under it, one a page, shows where it
// is and jumps to the page clicked, and a swipe across the page moves it
// as well. A move slides the new page in. It may loop round, move on by
// itself every while (pausing while the pointer is over it), and in Steps
// mode reach a page only once the game marks the steps before it done.
//
// It counts its slides and its waits in updates, at Ebitengine's TPS, and
// tells UI.Changed of each move and each update of a slide.
type Carousel struct {
	definedParams  CarouselParams
	computedParams CarouselParams

	PageChangedEvent *event.Event

	pages   []PreferredSizeLocateableWidget
	current int
	from    int // the page sliding out, -1 for none
	dir     int // 1 as the new page comes from the right, -1 from the left
	slid    int // the updates into the slide

	slideTime time.Duration
	autoEvery time.Duration
	idle      int // the updates since it last moved, for AutoAdvance

	loop, noArrows, noDots, noSwipe bool
	steps, advanceWhenDone          bool
	done                            []bool

	hovering bool
	swipe    struct {
		down, moving bool
		x, offset    int // where the press began, how far the page follows it
	}

	prev, next *Button
	dot        *ebiten.Image // a dot, drawn once
	dotOpts    ebiten.DrawImageOptions

	widgetOpts []WidgetOpt
	init       *MultiOnce
	widget     *Widget
}

// carouselArrowMargin is the space between an arrow and the page area's
// side; carouselSwipeStart how far a press moves before it is a swipe.
const (
	carouselArrowMargin = 4
	carouselSwipeStart  = 8
)

// CarouselPageChangedEventArgs tell which page a carousel moved to, and
// from.
type CarouselPageChangedEventArgs struct {
	Carousel       *Carousel
	Page, Previous int
}

type CarouselPageChangedHandlerFunc func(args *CarouselPageChangedEventArgs)

type CarouselOpt func(c *Carousel)

type CarouselOptions struct{}

// CarouselOpts holds the options of NewCarousel.
var CarouselOpts CarouselOptions

// NewCarousel returns a carousel, on its first page, sliding in 0.3 s.
func NewCarousel(opts ...CarouselOpt) *Carousel {
	c := &Carousel{
		PageChangedEvent: &event.Event{},
		from:             -1,
		slideTime:        300 * time.Millisecond,
		init:             &MultiOnce{},
	}
	c.init.Append(c.createWidget)
	for _, o := range opts {
		o(c)
	}
	c.done = make([]bool, len(c.pages))
	return c
}

func (o CarouselOptions) WidgetOpts(opts ...WidgetOpt) CarouselOpt {
	return func(c *Carousel) { c.widgetOpts = append(c.widgetOpts, opts...) }
}

// Pages sets the pages, in order.
func (o CarouselOptions) Pages(pages ...PreferredSizeLocateableWidget) CarouselOpt {
	return func(c *Carousel) { c.pages = append(c.pages, pages...) }
}

// Loop has it go round: on from the last page to the first, back from the
// first to the last. Not in Steps mode.
func (o CarouselOptions) Loop() CarouselOpt {
	return func(c *Carousel) { c.loop = true }
}

// AutoAdvance has it move on by itself every every, waiting while the
// pointer is over it, and a whole while after it moved.
func (o CarouselOptions) AutoAdvance(every time.Duration) CarouselOpt {
	return func(c *Carousel) { c.autoEvery = every }
}

// SlideTime sets how long a new page takes to slide in; 0 shows it at
// once.
func (o CarouselOptions) SlideTime(d time.Duration) CarouselOpt {
	return func(c *Carousel) { c.slideTime = d }
}

// NoArrows leaves the arrows out.
func (o CarouselOptions) NoArrows() CarouselOpt {
	return func(c *Carousel) { c.noArrows = true }
}

// NoDots leaves the dots out; the page area takes their row.
func (o CarouselOptions) NoDots() CarouselOpt {
	return func(c *Carousel) { c.noDots = true }
}

// NoSwipe has a swipe across the page move nothing.
func (o CarouselOptions) NoSwipe() CarouselOpt {
	return func(c *Carousel) { c.noSwipe = true }
}

// Steps makes the pages steps, such as a tutorial's: a page is reached
// only once the steps before it are done, as the game marks them
// (SetStepDone) when what each asks for has happened. Going back is
// always allowed; a done step's dot takes DotDoneColor.
func (o CarouselOptions) Steps() CarouselOpt {
	return func(c *Carousel) { c.steps = true }
}

// AdvanceWhenDone has a carousel in Steps mode move on as the step it
// shows is marked done.
func (o CarouselOptions) AdvanceWhenDone() CarouselOpt {
	return func(c *Carousel) { c.advanceWhenDone = true }
}

func (o CarouselOptions) PageChangedHandler(f CarouselPageChangedHandlerFunc) CarouselOpt {
	return func(c *Carousel) { c.PageChangedEvent.AddHandler(handle(f)) }
}

func (o CarouselOptions) ArrowImage(i *ButtonImage) CarouselOpt {
	return func(c *Carousel) { c.definedParams.ArrowImage = i }
}

func (o CarouselOptions) ArrowColor(col color.Color) CarouselOpt {
	return func(c *Carousel) { c.definedParams.ArrowColor = col }
}

func (o CarouselOptions) ArrowSize(w, h int) CarouselOpt {
	return func(c *Carousel) { c.definedParams.ArrowSize = &img.Point{w, h} }
}

// ArrowGraphics sets the arrows' icons, in place of the arrows the
// carousel draws: each with its Idle image, and its Disabled one for an
// arrow that can't move the carousel.
func (o CarouselOptions) ArrowGraphics(prev, next *GraphicImage) CarouselOpt {
	return func(c *Carousel) { c.definedParams.PrevGraphic, c.definedParams.NextGraphic = prev, next }
}

// DotImage sets the image drawn for each dot, in place of a circle, tinted
// by DotColor, or DotDoneColor for a done step; white keeps its colors. A
// dot is as wide as the image unless DotSize sets it.
func (o CarouselOptions) DotImage(i *ebiten.Image) CarouselOpt {
	return func(c *Carousel) { c.definedParams.DotImage = i }
}

func (o CarouselOptions) DotColor(col color.Color) CarouselOpt {
	return func(c *Carousel) { c.definedParams.DotColor = col }
}

func (o CarouselOptions) DotDoneColor(col color.Color) CarouselOpt {
	return func(c *Carousel) { c.definedParams.DotDoneColor = col }
}

func (o CarouselOptions) DotSize(size, spacing int) CarouselOpt {
	return func(c *Carousel) { c.definedParams.DotSize, c.definedParams.DotSpacing = &size, &spacing }
}

// Page returns the page shown.
func (c *Carousel) Page() int { return c.current }

// Next moves to the next page, round to the first with Loop, if it may.
func (c *Carousel) Next() {
	to := c.current + 1
	if to == len(c.pages) {
		if !c.loop || c.steps {
			return
		}
		to = 0
	}
	if c.canGo(to) {
		c.move(to, 1)
	}
}

// Previous moves to the page before, round to the last with Loop.
func (c *Carousel) Previous() {
	to := c.current - 1
	if to < 0 {
		if !c.loop || c.steps {
			return
		}
		to = len(c.pages) - 1
	}
	c.move(to, -1)
}

// SetPage moves to page i, if it may.
func (c *Carousel) SetPage(i int) {
	if i == c.current || !c.canGo(i) {
		return
	}
	dir := 1
	if i < c.current {
		dir = -1
	}
	c.move(i, dir)
}

// SetStepDone marks step i done or not, in Steps mode; with
// AdvanceWhenDone, the step shown marked done moves on.
func (c *Carousel) SetStepDone(i int, done bool) {
	if i < 0 || i >= len(c.done) || c.done[i] == done {
		return
	}
	c.done[i] = done
	changes.Mark()
	c.updateArrows()
	if done && c.advanceWhenDone && i == c.current {
		c.Next()
	}
}

// StepDone reports whether step i is marked done.
func (c *Carousel) StepDone(i int) bool {
	return i >= 0 && i < len(c.done) && c.done[i]
}

// canGo reports whether page to may be shown: one of the pages, and in
// Steps mode one whose steps before are done.
func (c *Carousel) canGo(to int) bool {
	if to < 0 || to >= len(c.pages) {
		return false
	}
	if c.steps {
		for k := c.current; k < to; k++ {
			if !c.done[k] {
				return false
			}
		}
	}
	return true
}

// move shows page to, sliding it in from dir's side.
func (c *Carousel) move(to, dir int) {
	if to == c.current {
		return
	}
	prev := c.current
	c.current, c.dir, c.slid, c.idle = to, dir, 0, 0
	c.from = -1
	if c.slideTime > 0 {
		c.from = prev
	}
	changes.Mark()
	c.updateArrows()
	c.PageChangedEvent.Fire(&CarouselPageChangedEventArgs{Carousel: c, Page: to, Previous: prev})
}

// updateArrows disables an arrow that can't move the carousel.
func (c *Carousel) updateArrows() {
	if c.prev == nil {
		return
	}
	few := len(c.pages) < 2
	c.prev.GetWidget().Disabled = few || c.neighbor(-1) < 0
	c.next.GetWidget().Disabled = few || !c.canGo(c.neighbor(1))
}

// Sliding reports whether a page is sliding in.
func (c *Carousel) Sliding() bool { return c.from >= 0 }

func (c *Carousel) GetWidget() *Widget {
	c.init.Do()
	return c.widget
}

func (c *Carousel) SetLocation(rect img.Rectangle) {
	c.init.Do()
	c.widget.Rect = rect
}

// Validate takes the look from the options, then the theme, then defaults:
// translucent arrows' buttons, white arrows and dots.
func (c *Carousel) Validate() {
	c.init.Do()
	p := CarouselParams{}
	if theme := c.widget.GetTheme(); theme != nil && theme.CarouselTheme != nil {
		p = *theme.CarouselTheme
	}
	d := c.definedParams
	p.ArrowImage = first(d.ArrowImage, p.ArrowImage, &ButtonImage{
		Idle:     image.NewNineSliceColor(color.NRGBA{0, 0, 0, 96}),
		Hover:    image.NewNineSliceColor(color.NRGBA{0, 0, 0, 128}),
		Pressed:  image.NewNineSliceColor(color.NRGBA{0, 0, 0, 160}),
		Disabled: image.NewNineSliceColor(color.NRGBA{0, 0, 0, 32}),
	})
	p.ArrowColor = first(d.ArrowColor, p.ArrowColor, color.Color(color.White))
	p.ArrowSize = first(d.ArrowSize, p.ArrowSize, &img.Point{24, 32})
	p.DotColor = first(d.DotColor, p.DotColor, color.Color(color.White))
	p.DotDoneColor = first(d.DotDoneColor, p.DotDoneColor, color.Color(color.NRGBA{120, 200, 120, 255}))
	p.PrevGraphic = first(d.PrevGraphic, p.PrevGraphic)
	p.NextGraphic = first(d.NextGraphic, p.NextGraphic)
	p.DotImage = first(d.DotImage, p.DotImage)
	dotSize := 8
	if p.DotImage != nil {
		dotSize = p.DotImage.Bounds().Dx()
	}
	p.DotSize = first(d.DotSize, p.DotSize, &dotSize)
	p.DotSpacing = first(d.DotSpacing, p.DotSpacing, new(6))
	c.computedParams = p

	if c.dot = p.DotImage; c.dot == nil {
		r := float32(*p.DotSize) / 2
		c.dot = ebiten.NewImage(*p.DotSize, *p.DotSize)
		vector.FillCircle(c.dot, r, r, r, color.White, true)
	}
	if !c.noArrows && c.prev == nil {
		c.prev = c.arrowButton(true, p.PrevGraphic, c.Previous)
		c.next = c.arrowButton(false, p.NextGraphic, c.Next)
		c.updateArrows()
	}
	for _, pg := range c.pages {
		pg.GetWidget().parent = c.widget
		pg.Validate()
	}
}

// arrowButton returns an arrow's button, showing icon or else an arrow
// pointing left or right, calling move as it is clicked.
func (c *Carousel) arrowButton(left bool, icon *GraphicImage, move func()) *Button {
	s := *c.computedParams.ArrowSize
	if icon == nil {
		icon = c.arrow(left)
	}
	b := NewButton(
		ButtonOpts.Image(c.computedParams.ArrowImage),
		ButtonOpts.Graphic(icon),
		ButtonOpts.WidgetOpts(WidgetOpts.MinSize(s.X, s.Y)),
		ButtonOpts.ClickedHandler(func(*ButtonClickedEventArgs) { move() }),
	)
	b.GetWidget().parent = c.widget
	b.Validate()
	return b
}

// arrow draws an arrow pointing left or right once, in ArrowColor, and
// faded for an arrow that can't move the carousel.
func (c *Carousel) arrow(left bool) *GraphicImage {
	s := *c.computedParams.ArrowSize
	w, h := s.X/2, s.Y/2
	arrow := func(alpha float32) *ebiten.Image {
		i := ebiten.NewImage(w, h)
		var path vector.Path
		x0, x1 := float32(0), float32(w)
		if left {
			x0, x1 = x1, x0
		}
		path.MoveTo(x0, 0)
		path.LineTo(x1, float32(h)/2)
		path.LineTo(x0, float32(h))
		path.Close()
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(c.computedParams.ArrowColor)
		op.ColorScale.ScaleAlpha(alpha)
		vector.FillPath(i, &path, &vector.FillOptions{}, op)
		return i
	}
	return &GraphicImage{Idle: arrow(1), Disabled: arrow(0.35)}
}

// dotRow returns the height of the dots' row, 0 without dots.
func (c *Carousel) dotRow() int {
	if c.noDots {
		return 0
	}
	return *c.computedParams.DotSize + 2**c.computedParams.DotSpacing
}

// PreferredSize is its largest page's size, and the dots' row under it.
func (c *Carousel) PreferredSize() (int, int) {
	c.init.Do()
	w, h := 0, 0
	for _, p := range c.pages {
		pw, ph := p.PreferredSize()
		w, h = max(w, pw), max(h, ph)
	}
	h += c.dotRow()
	return max(w, c.widget.MinWidth), max(h, c.widget.MinHeight)
}

// pageRect returns the page area: the widget above the dots' row.
func (c *Carousel) pageRect() img.Rectangle {
	r := c.widget.Rect
	r.Max.Y -= c.dotRow()
	return r
}

// dotRect returns dot i's place, the row centered under the page area.
func (c *Carousel) dotRect(i int) img.Rectangle {
	size, gap := *c.computedParams.DotSize, *c.computedParams.DotSpacing
	pr := c.pageRect()
	n := len(c.pages)
	x := (pr.Min.X+pr.Max.X)/2 - (n*size+(n-1)*gap)/2 + i*(size+gap)
	y := pr.Max.Y + gap
	return img.Rect(x, y, x+size, y+size)
}

// place places the page shown, still, and the arrows at the page area's
// sides.
func (c *Carousel) place() {
	pr := c.pageRect()
	if c.from < 0 && c.swipe.offset == 0 && len(c.pages) > 0 {
		c.pages[c.current].SetLocation(pr)
	}
	if c.prev == nil {
		return
	}
	s := *c.computedParams.ArrowSize
	cy := (pr.Min.Y + pr.Max.Y) / 2
	left := pr.Min.X + carouselArrowMargin
	right := pr.Max.X - carouselArrowMargin - s.X
	c.prev.SetLocation(img.Rect(left, cy-s.Y/2, left+s.X, cy-s.Y/2+s.Y))
	c.next.SetLocation(img.Rect(right, cy-s.Y/2, right+s.X, cy-s.Y/2+s.Y))
	c.updateArrows()
}

// PlaceChildren places the page shown, its children and the arrows.
func (c *Carousel) PlaceChildren() {
	c.init.Do()
	c.place()
	if len(c.pages) > 0 {
		PlaceChildren(c.pages[c.current])
	}
}

func (c *Carousel) Update(updObj *UpdateObject) {
	c.init.Do()
	c.widget.Update(updObj)
	if len(c.pages) == 0 {
		return
	}
	if c.prev != nil {
		c.prev.Update(updObj)
		c.next.Update(updObj)
	}
	if u, ok := c.pages[c.current].(Updater); ok {
		u.Update(updObj)
	}
	if c.from >= 0 {
		c.slid++
		changes.Mark()
		if c.slid >= updatesIn(c.slideTime) {
			c.from = -1
		}
	}
	c.updateSwipe()
	c.autoAdvance()
}

// autoAdvance moves on every AutoAdvance while, counting only while the
// pointer is off it and nothing moves.
func (c *Carousel) autoAdvance() {
	if c.autoEvery <= 0 {
		return
	}
	if c.hovering || c.swipe.down || c.from >= 0 {
		c.idle = 0
		return
	}
	c.idle++
	if c.idle >= updatesIn(c.autoEvery) {
		c.idle = 0
		c.Next()
	}
}

// updateSwipe has the page follow a press dragged across it, and on
// letting go move to the page that way when it went over a fifth of the
// page, else spring back.
func (c *Carousel) updateSwipe() {
	if c.noSwipe || len(c.pages) < 2 {
		return
	}
	s := &c.swipe
	x, y := input.CursorPosition()
	pr := c.pageRect()
	switch {
	case input.MouseButtonJustPressedLayer(ebiten.MouseButtonLeft, c.widget.EffectiveInputLayer()):
		if img.Pt(x, y).In(pr) && !c.onArrow(x, y) && c.from < 0 {
			s.down, s.moving, s.x, s.offset = true, false, x, 0
		}
	case s.down && input.MouseButtonPressed(ebiten.MouseButtonLeft):
		c.followSwipe(c.clampSwipe(x-s.x, pr.Dx()))
	case s.down:
		c.endSwipe(pr.Dx())
	}
}

// followSwipe has the page follow the press off pixels across, the press
// clicking nothing on the page once it is a swipe.
func (c *Carousel) followSwipe(off int) {
	s := &c.swipe
	if off != s.offset {
		s.offset = off
		changes.Mark()
	}
	if !s.moving && abs(off) > carouselSwipeStart {
		s.moving = true
		cancelPresses()
	}
}

// endSwipe lets go of the swipe on a page width pixels wide: past a fifth
// of it, the carousel moves that way, its slide going on from where the
// swipe left it.
func (c *Carousel) endSwipe(width int) {
	s := &c.swipe
	off := s.offset
	s.down, s.moving, s.offset = false, false, 0
	changes.Mark()
	if abs(off) <= width/5 {
		return
	}
	if off < 0 {
		c.Next()
	} else {
		c.Previous()
	}
	if c.from >= 0 && width > 0 {
		c.slid = abs(off) * updatesIn(c.slideTime) / width
	}
}

// clampSwipe returns how far the page follows a press moved by off: not
// toward a page it can't move to, and not past a page's width.
func (c *Carousel) clampSwipe(off, width int) int {
	if off < 0 && !c.canGo(c.neighbor(1)) || off > 0 && c.neighbor(-1) < 0 {
		return 0
	}
	return max(-width, min(width, off))
}

// neighbor returns the page on dir's side of the one shown, round with
// Loop, or -1.
func (c *Carousel) neighbor(dir int) int {
	n := len(c.pages)
	i := c.current + dir
	if i < 0 || i >= n {
		if !c.loop || c.steps {
			return -1
		}
		i = (i + n) % n
	}
	return i
}

// onArrow reports whether x, y is on an arrow.
func (c *Carousel) onArrow(x, y int) bool {
	p := img.Pt(x, y)
	return c.prev != nil && (p.In(c.prev.GetWidget().Rect) || p.In(c.next.GetWidget().Rect))
}

func (c *Carousel) Render(screen *ebiten.Image) {
	c.init.Do()
	c.widget.Render(screen)
	if len(c.pages) == 0 {
		return
	}
	c.place()
	if c.from >= 0 || c.swipe.offset != 0 {
		c.renderMoving(screen)
	} else if r, ok := c.pages[c.current].(Renderer); ok {
		r.Render(screen)
	}
	if c.prev != nil {
		c.prev.Render(screen)
		c.next.Render(screen)
	}
	if !c.noDots {
		c.renderDots(screen)
	}
}

// renderMoving draws the pages a slide or a swipe shows, each where it has
// got to, within the page area.
func (c *Carousel) renderMoving(screen *ebiten.Image) {
	pr := c.pageRect()
	area, ok := screen.SubImage(pr).(*ebiten.Image)
	if !ok || pr.Empty() {
		return
	}
	w := pr.Dx()
	draw := func(i, off int) {
		if i < 0 {
			return
		}
		p := c.pages[i]
		p.SetLocation(pr.Add(img.Pt(off, 0)))
		if r, ok := p.(Renderer); ok {
			r.Render(area)
		}
	}
	if c.from >= 0 {
		t := float64(c.slid) / float64(max(1, updatesIn(c.slideTime)))
		t = t * t * (3 - 2*t) // eased in and out
		off := int(float64(c.dir*w) * (1 - t))
		draw(c.from, off-c.dir*w)
		draw(c.current, off)
		return
	}
	off := c.swipe.offset
	if off < 0 {
		draw(c.neighbor(1), off+w)
	} else {
		draw(c.neighbor(-1), off-w)
	}
	draw(c.current, off)
}

// renderDots draws a dot a page: the current's opaque, a done step's in
// DotDoneColor, the rest faded.
func (c *Carousel) renderDots(screen *ebiten.Image) {
	for i := range c.pages {
		r := c.dotRect(i)
		o := &c.dotOpts
		o.GeoM.Reset()
		o.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
		o.ColorScale.Reset()
		col := c.computedParams.DotColor
		if c.steps && c.done[i] {
			col = c.computedParams.DotDoneColor
		}
		o.ColorScale.ScaleWithColor(col)
		if i != c.current {
			o.ColorScale.ScaleAlpha(0.4)
		}
		screen.DrawImage(c.dot, o)
	}
}

// dotAt returns the dot at x, y, -1 for none.
func (c *Carousel) dotAt(x, y int) int {
	if c.noDots {
		return -1
	}
	for i := range c.pages {
		if img.Pt(x, y).In(c.dotRect(i)) {
			return i
		}
	}
	return -1
}

// SetupInputLayer sets up the arrows' and the page shown's input layers.
func (c *Carousel) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	c.init.Do()
	if c.prev != nil {
		c.prev.SetupInputLayer(def)
		c.next.SetupInputLayer(def)
	}
	if len(c.pages) > 0 {
		if l, ok := c.pages[c.current].(input.Layerer); ok {
			l.SetupInputLayer(def)
		}
	}
}

// GetFocusers returns the arrows and what the page shown holds that takes
// the focus, so keys and gamepads reach them.
func (c *Carousel) GetFocusers() []Focuser {
	c.init.Do()
	var fs []Focuser
	if c.prev != nil {
		fs = append(fs, c.prev, c.next)
	}
	if len(c.pages) == 0 {
		return fs
	}
	switch p := c.pages[c.current].(type) {
	case interface{ GetFocusers() []Focuser }:
		fs = append(fs, p.GetFocusers()...)
	case Focuser:
		fs = append(fs, p)
	}
	return fs
}

func (c *Carousel) createWidget() {
	c.widget = NewWidget(append([]WidgetOpt{
		WidgetOpts.CursorEnterHandler(func(*CursorEnterEventArgs) { c.hovering = true }),
		WidgetOpts.CursorExitHandler(func(*CursorExitEventArgs) { c.hovering = false }),
		WidgetOpts.MouseButtonClickedHandler(func(a *MouseButtonClickedEventArgs) {
			if a.Button != ebiten.MouseButtonLeft {
				return
			}
			r := c.widget.Rect
			if i := c.dotAt(r.Min.X+a.OffsetX, r.Min.Y+a.OffsetY); i >= 0 {
				c.SetPage(i)
			}
		}),
	}, c.widgetOpts...)...)
	c.widgetOpts = nil
}
