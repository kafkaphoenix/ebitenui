package widget

import (
	img "image"
	"image/color"
	"slices"
	"strings"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ChecklistState is the state of a checklist's item.
type ChecklistState int

const (
	// ChecklistPending is an item still to do.
	ChecklistPending ChecklistState = iota
	// ChecklistDone is an item done.
	ChecklistDone
	// ChecklistFailed is an item that can't be done any more.
	ChecklistFailed
)

// ChecklistItem is one item of a checklist: a task, a mission's
// objective, a tutorial's step.
type ChecklistItem struct {
	Label string
	// Detail is shown at the row's end, such as "3/5" for a count so far.
	Detail string
	State  ChecklistState
	// Optional is an item the list is complete without.
	Optional bool
	// Tickable is an item the player ticks and unticks: with a click, or
	// with Space or Enter on the checklist focused.
	Tickable bool
}

// ChecklistMarks are the marks shown before the items, one a state: the
// game's icons, drawn as they are. A mark not given is drawn: an empty box,
// a box with a tick, a box with a cross.
type ChecklistMarks struct {
	Pending, Done, Failed *ebiten.Image
}

// ChecklistColors are a checklist's colors: each state's text, that of
// the items after the current one in a sequential list, the title's, and
// those of the tick and the cross it draws when given no marks. A color
// not given comes from the theme's text color: the pending items' and the
// title's are it, done and failed items' and upcoming ones' it faded.
type ChecklistColors struct {
	Pending, Done, Failed, Upcoming, Title color.Color
	DoneMark, FailedMark                   color.Color
}

// ChecklistParams is a checklist's look: its marks and colors, its faces,
// the images behind a hovered or focused tickable item and behind the
// current item of a sequential list, the padding around the list and
// inside each row, and the spacing between a row's mark, label and detail.
type ChecklistParams struct {
	Marks        *ChecklistMarks
	Colors       *ChecklistColors
	Face         *text.Face
	TitleFace    *text.Face
	HoverImage   *image.NineSlice
	CurrentImage *image.NineSlice
	Padding      *Insets
	RowPadding   *Insets
	Spacing      *int
}

// Checklist shows items to do, each with a mark for its state: a game's
// pending tasks, a mission's objectives, a tutorial's steps. The game sets
// each item's state as it happens, and the player ticks the items that are
// Tickable, with a click or, with the checklist focused, Space or Enter,
// Up and Down moving between them. A done item takes a faded color, and a
// line through it with StrikeDone; a failed one a cross. Each item may
// show a detail at its end, such as a count so far. Above the items, a
// title may show, and the progress as the game words it. In a sequential
// list, a tutorial's, the first pending item that isn't optional is the
// current one, shown on its own background, and the items after it are
// faded or hidden until it is reached.
//
// It tells of each item's change of state, and when the last item that
// isn't optional is done; and UI.Changed of each change it shows.
type Checklist struct {
	definedParams  ChecklistParams
	computedParams ChecklistParams

	ItemChangedEvent *event.Event
	CompletedEvent   *event.Event

	items    []ChecklistItem
	rows     []checklistRow
	title    string
	progress func(done, total int) string
	maxWidth int

	sequential, hideUpcoming, strike bool
	current                          int // the current item, -1 for none
	complete                         bool
	tickable                         int // the tickable items

	titleText, progressText *Text
	marks                   ChecklistMarks
	markW, markH            int
	lineH, strikeY          int // a line's height, and where in it a line through goes
	strikeW                 int
	spaceW                  float64
	pixel                   *ebiten.Image

	dirty, laidOut bool
	prefW, prefH   int
	titleH         int

	hovering                bool
	hover, pressed, focused int // rows, -1 for none; focused is the focus's row
	hasFocus                bool
	tabOrder                int
	focusMap                map[FocusDirection]Focuser

	at       img.Point
	place    func(*ebiten.DrawImageOptions)
	drawOpts ebiten.DrawImageOptions

	widgetOpts []WidgetOpt
	init       *MultiOnce
	widget     *Widget
}

// checklistRow is an item's row: its texts, and its top and height in the
// checklist, 0 high while hidden.
type checklistRow struct {
	label, detail *Text
	y, h          int
}

// ChecklistItemChangedEventArgs tell which item of a checklist changed
// state, and from which.
type ChecklistItemChangedEventArgs struct {
	Checklist *Checklist
	Index     int
	Item      ChecklistItem
	Previous  ChecklistState
}

// ChecklistCompletedEventArgs tell which checklist is complete.
type ChecklistCompletedEventArgs struct {
	Checklist *Checklist
}

type (
	ChecklistItemChangedHandlerFunc func(args *ChecklistItemChangedEventArgs)
	ChecklistCompletedHandlerFunc   func(args *ChecklistCompletedEventArgs)
)

type ChecklistOpt func(c *Checklist)

type ChecklistOptions struct{}

// ChecklistOpts holds the options of NewChecklist.
var ChecklistOpts ChecklistOptions

// NewChecklist returns a checklist.
func NewChecklist(opts ...ChecklistOpt) *Checklist {
	c := &Checklist{
		ItemChangedEvent: &event.Event{},
		CompletedEvent:   &event.Event{},
		current:          -1,
		hover:            -1,
		pressed:          -1,
		focused:          -1,
		focusMap:         map[FocusDirection]Focuser{},
		init:             &MultiOnce{},
	}
	c.init.Append(c.createWidget)
	for _, o := range opts {
		o(c)
	}
	c.current, c.complete = c.findCurrent(), c.findComplete()
	c.tickable = c.countTickable()
	return c
}

func (o ChecklistOptions) WidgetOpts(opts ...WidgetOpt) ChecklistOpt {
	return func(c *Checklist) { c.widgetOpts = append(c.widgetOpts, opts...) }
}

// Items sets the items, in order.
func (o ChecklistOptions) Items(items ...ChecklistItem) ChecklistOpt {
	return func(c *Checklist) { c.items = append(c.items, items...) }
}

// Title shows title above the items.
func (o ChecklistOptions) Title(title string) ChecklistOpt {
	return func(c *Checklist) { c.title = title }
}

// Progress shows the progress at the end of the title's line, worded by
// f from the items done and those that count, the optional ones left
// out: "2/5", "2 of 5 done". f is called as an item changes.
func (o ChecklistOptions) Progress(f func(done, total int) string) ChecklistOpt {
	return func(c *Checklist) { c.progress = f }
}

// Sequential makes the items steps done in order, as a tutorial's: the
// first pending item that isn't optional is the current one, on the
// CurrentImage, and those after it take the Upcoming color.
func (o ChecklistOptions) Sequential() ChecklistOpt {
	return func(c *Checklist) { c.sequential = true }
}

// HideUpcoming makes the list sequential, and hides the items after the
// current one until it is reached.
func (o ChecklistOptions) HideUpcoming() ChecklistOpt {
	return func(c *Checklist) { c.sequential, c.hideUpcoming = true, true }
}

// StrikeDone draws a line through each done item's label.
func (o ChecklistOptions) StrikeDone() ChecklistOpt {
	return func(c *Checklist) { c.strike = true }
}

// MaxWidth makes the checklist at most width pixels wide, its labels
// wrapping to fit; without it, labels don't wrap.
func (o ChecklistOptions) MaxWidth(width int) ChecklistOpt {
	return func(c *Checklist) { c.maxWidth = width }
}

func (o ChecklistOptions) TabOrder(order int) ChecklistOpt {
	return func(c *Checklist) { c.tabOrder = order }
}

func (o ChecklistOptions) ItemChangedHandler(f ChecklistItemChangedHandlerFunc) ChecklistOpt {
	return func(c *Checklist) { c.ItemChangedEvent.AddHandler(handle(f)) }
}

func (o ChecklistOptions) CompletedHandler(f ChecklistCompletedHandlerFunc) ChecklistOpt {
	return func(c *Checklist) { c.CompletedEvent.AddHandler(handle(f)) }
}

func (o ChecklistOptions) Marks(m *ChecklistMarks) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.Marks = m }
}

func (o ChecklistOptions) Colors(col *ChecklistColors) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.Colors = col }
}

func (o ChecklistOptions) Face(f *text.Face) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.Face = f }
}

func (o ChecklistOptions) TitleFace(f *text.Face) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.TitleFace = f }
}

func (o ChecklistOptions) HoverImage(i *image.NineSlice) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.HoverImage = i }
}

func (o ChecklistOptions) CurrentImage(i *image.NineSlice) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.CurrentImage = i }
}

func (o ChecklistOptions) Padding(p *Insets) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.Padding = p }
}

func (o ChecklistOptions) RowPadding(p *Insets) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.RowPadding = p }
}

func (o ChecklistOptions) Spacing(s int) ChecklistOpt {
	return func(c *Checklist) { c.definedParams.Spacing = &s }
}

// Len returns the number of items.
func (c *Checklist) Len() int { return len(c.items) }

// Item returns item i.
func (c *Checklist) Item(i int) ChecklistItem { return c.items[i] }

// Items returns a copy of the items.
func (c *Checklist) Items() []ChecklistItem { return slices.Clone(c.items) }

// SetItems replaces the items.
func (c *Checklist) SetItems(items []ChecklistItem) {
	c.items = append(c.items[:0], items...)
	c.hover, c.pressed = -1, -1
	c.refresh()
}

// AddItem adds item at the end.
func (c *Checklist) AddItem(item ChecklistItem) {
	c.items = append(c.items, item)
	c.refresh()
}

// RemoveItem removes item i.
func (c *Checklist) RemoveItem(i int) {
	if i < 0 || i >= len(c.items) {
		return
	}
	c.items = slices.Delete(c.items, i, i+1)
	c.hover, c.pressed = -1, -1
	if c.focused >= i {
		c.focused--
	}
	c.refresh()
}

// SetItem replaces item i, telling of a change of its state.
func (c *Checklist) SetItem(i int, item ChecklistItem) {
	if i < 0 || i >= len(c.items) || c.items[i] == item {
		return
	}
	prev := c.items[i].State
	c.items[i] = item
	c.refresh()
	if item.State != prev {
		c.ItemChangedEvent.Fire(&ChecklistItemChangedEventArgs{Checklist: c, Index: i, Item: item, Previous: prev})
	}
	c.checkComplete()
}

// SetState sets item i's state, telling of the change.
func (c *Checklist) SetState(i int, s ChecklistState) {
	if i >= 0 && i < len(c.items) {
		item := c.items[i]
		item.State = s
		c.SetItem(i, item)
	}
}

// SetDetail sets item i's detail, such as a count so far.
func (c *Checklist) SetDetail(i int, detail string) {
	if i >= 0 && i < len(c.items) {
		item := c.items[i]
		item.Detail = detail
		c.SetItem(i, item)
	}
}

// Toggle ticks item i if it isn't done, else unticks it, as the player
// does a tickable item.
func (c *Checklist) Toggle(i int) {
	if i < 0 || i >= len(c.items) {
		return
	}
	if c.items[i].State == ChecklistDone {
		c.SetState(i, ChecklistPending)
	} else {
		c.SetState(i, ChecklistDone)
	}
}

// SetTitle sets the title shown above the items; "" shows none, unless
// the progress shows.
func (c *Checklist) SetTitle(title string) {
	if title != c.title {
		c.title = title
		c.refresh()
	}
}

// Progress returns the items done and those that count, the optional ones
// left out.
func (c *Checklist) Progress() (done, total int) {
	for _, it := range c.items {
		if !it.Optional {
			total++
			if it.State == ChecklistDone {
				done++
			}
		}
	}
	return done, total
}

// Complete reports whether every item that isn't optional is done.
func (c *Checklist) Complete() bool { return c.complete }

// Current returns the first pending item that isn't optional, the current
// step of a sequential list, or -1 for none.
func (c *Checklist) Current() int { return c.current }

func (c *Checklist) findCurrent() int {
	for i, it := range c.items {
		if !it.Optional && it.State == ChecklistPending {
			return i
		}
	}
	return -1
}

func (c *Checklist) findComplete() bool {
	done, total := c.Progress()
	return done == total
}

func (c *Checklist) countTickable() int {
	n := 0
	for _, it := range c.items {
		if it.Tickable {
			n++
		}
	}
	return n
}

// checkComplete tells as the list becomes complete.
func (c *Checklist) checkComplete() {
	was := c.complete
	c.complete = c.findComplete()
	if c.complete && !was {
		c.CompletedEvent.Fire(&ChecklistCompletedEventArgs{Checklist: c})
	}
}

// refresh shows a change of the items: the current item, the rows'
// texts and colors, the progress, laid out again before it draws.
func (c *Checklist) refresh() {
	c.current = c.findCurrent()
	c.tickable = c.countTickable()
	if c.focused >= len(c.items) || c.focused >= 0 && !c.focusable(c.focused) {
		c.focused = -1
	}
	if c.hasFocus && c.focused < 0 {
		c.focused = c.nextFocusable(-1, 1)
	}
	c.dirty = true
	changes.Mark()
	if c.computedParams.Face != nil {
		c.syncRows()
	}
}

// syncRows makes a row for each item, and sets each its texts and colors,
// the title's and the progress's too.
func (c *Checklist) syncRows() {
	p := c.computedParams
	for len(c.rows) < len(c.items) {
		c.rows = append(c.rows, checklistRow{label: c.newText(p.Face, TextPositionStart), detail: c.newText(p.Face, TextPositionEnd)})
	}
	c.rows = c.rows[:len(c.items)]
	for i, it := range c.items {
		r := &c.rows[i]
		r.label.Label, r.detail.Label = it.Label, it.Detail
		col := c.textColor(i)
		r.label.SetColor(col)
		r.detail.SetColor(col)
	}
	switch {
	case c.title == "" && c.progress == nil:
		c.titleText, c.progressText = nil, nil
	case c.titleText == nil:
		c.titleText = c.newText(p.TitleFace, TextPositionStart)
		c.titleText.SetColor(p.Colors.Title)
		c.progressText = c.newText(p.Face, TextPositionEnd)
		c.progressText.SetColor(p.Colors.Title)
	}
	if c.titleText != nil {
		c.titleText.Label = c.title
		if c.progress != nil {
			if s := c.progress(c.Progress()); s != c.progressText.Label {
				c.progressText.Label = s
			}
		}
	}
}

func (c *Checklist) newText(face *text.Face, h TextPosition) *Text {
	t := NewText(TextOpts.Text("", face, c.computedParams.Colors.Pending), TextOpts.Position(h, TextPositionStart),
		TextOpts.Padding(&Insets{}))
	t.GetWidget().parent = c.widget
	t.Validate()
	return t
}

// textColor returns item i's text color, for its state and, in a
// sequential list, whether it comes after the current item.
func (c *Checklist) textColor(i int) color.Color {
	col := c.computedParams.Colors
	switch c.items[i].State {
	case ChecklistDone:
		return col.Done
	case ChecklistFailed:
		return col.Failed
	case ChecklistPending:
	}
	if c.sequential && c.current >= 0 && i > c.current {
		return col.Upcoming
	}
	return col.Pending
}

// shown reports whether item i shows: all do but, with HideUpcoming,
// those after the current item.
func (c *Checklist) shown(i int) bool {
	return !c.hideUpcoming || c.current < 0 || i <= c.current
}

func (c *Checklist) GetWidget() *Widget {
	c.init.Do()
	return c.widget
}

func (c *Checklist) SetLocation(rect img.Rectangle) {
	c.init.Do()
	c.widget.Rect = rect
}

// Validate takes the look from the options, then the theme, then the
// theme's face and text color, and draws the marks not given.
func (c *Checklist) Validate() {
	c.init.Do()
	c.computedParams = c.computeParams()
	p := c.computedParams
	m := (*p.Face).Metrics()
	c.lineH = int(m.HAscent + m.HDescent + 0.5)
	x := m.XHeight
	if x <= 0 {
		x = m.HAscent * 0.7
	}
	c.strikeY = int(m.HAscent - x/2 + 0.5)
	c.strikeW = max(1, c.lineH/14)
	c.spaceW, _ = text.Measure(" ", *p.Face, 0)
	c.makeMarks()
	if c.pixel == nil {
		c.pixel = ebiten.NewImage(1, 1)
		c.pixel.Fill(color.White)
	}
	c.titleText, c.progressText = nil, nil
	c.rows = c.rows[:0]
	c.syncRows()
	c.dirty = true
	c.layout()
}

// computeParams returns the look: each of the options, else the theme's
// checklist, else the theme's face and text color, else defaults.
func (c *Checklist) computeParams() ChecklistParams {
	theme := c.widget.GetTheme()
	p := ChecklistParams{}
	var face *text.Face
	var ink color.Color = color.White
	if theme != nil {
		if theme.ChecklistTheme != nil {
			p = *theme.ChecklistTheme
		}
		face = theme.DefaultFace
		if theme.DefaultTextColor != nil {
			ink = theme.DefaultTextColor
		}
	}
	d := c.definedParams
	p.Face = first(d.Face, p.Face, face)
	if p.Face == nil {
		panic("Checklist: Face is required.")
	}
	p.TitleFace = first(d.TitleFace, p.TitleFace, p.Face)
	p.HoverImage = first(d.HoverImage, p.HoverImage, image.NewNineSliceColor(color.NRGBA{255, 255, 255, 24}))
	p.CurrentImage = first(d.CurrentImage, p.CurrentImage, image.NewNineSliceColor(color.NRGBA{255, 255, 255, 40}))
	p.Padding = first(d.Padding, p.Padding, &Insets{})
	p.RowPadding = first(d.RowPadding, p.RowPadding, &Insets{Left: 4, Right: 4, Top: 2, Bottom: 2})
	p.Spacing = first(d.Spacing, p.Spacing, new(6))

	var mine, theirs ChecklistMarks
	if d.Marks != nil {
		mine = *d.Marks
	}
	if p.Marks != nil {
		theirs = *p.Marks
	}
	p.Marks = &ChecklistMarks{
		Pending: first(mine.Pending, theirs.Pending),
		Done:    first(mine.Done, theirs.Done),
		Failed:  first(mine.Failed, theirs.Failed),
	}

	var given, themed ChecklistColors
	if d.Colors != nil {
		given = *d.Colors
	}
	if p.Colors != nil {
		themed = *p.Colors
	}
	pending := first(given.Pending, themed.Pending, ink)
	p.Colors = &ChecklistColors{
		Pending:    pending,
		Done:       first(given.Done, themed.Done, faded(pending, 0.5)),
		Failed:     first(given.Failed, themed.Failed, faded(pending, 0.5)),
		Upcoming:   first(given.Upcoming, themed.Upcoming, faded(pending, 0.4)),
		Title:      first(given.Title, themed.Title, pending),
		DoneMark:   first(given.DoneMark, themed.DoneMark, color.Color(color.NRGBA{110, 200, 120, 255})),
		FailedMark: first(given.FailedMark, themed.FailedMark, color.Color(color.NRGBA{230, 90, 80, 255})),
	}
	return p
}

// faded returns col at alpha times its opacity.
func faded(col color.Color, alpha float64) color.Color {
	n, _ := color.NRGBAModel.Convert(col).(color.NRGBA)
	n.A = uint8(float64(n.A)*alpha + 0.5)
	return n
}

// makeMarks takes the marks given, and draws those not, once: boxes as
// high as most of a line, a tick in DoneMark and a cross in FailedMark.
func (c *Checklist) makeMarks() {
	p := c.computedParams
	s := max(9, c.lineH*3/4)
	st := max(1.5, float32(s)/9)
	box := func(col color.Color) *ebiten.Image {
		i := ebiten.NewImage(s, s)
		vector.StrokeRect(i, st/2, st/2, float32(s)-st, float32(s)-st, st, col, true)
		return i
	}
	c.marks = *p.Marks
	if c.marks.Pending == nil {
		c.marks.Pending = box(p.Colors.Pending)
	}
	f := float32(s)
	if c.marks.Done == nil {
		c.marks.Done = box(p.Colors.DoneMark)
		var path vector.Path
		path.MoveTo(f*0.22, f*0.52)
		path.LineTo(f*0.42, f*0.72)
		path.LineTo(f*0.78, f*0.3)
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(p.Colors.DoneMark)
		vector.StrokePath(c.marks.Done, &path, &vector.StrokeOptions{Width: st * 1.3, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}, op)
	}
	if c.marks.Failed == nil {
		c.marks.Failed = box(p.Colors.FailedMark)
		vector.StrokeLine(c.marks.Failed, f*0.3, f*0.3, f*0.7, f*0.7, st*1.3, p.Colors.FailedMark, true)
		vector.StrokeLine(c.marks.Failed, f*0.7, f*0.3, f*0.3, f*0.7, st*1.3, p.Colors.FailedMark, true)
	}
	c.markW, c.markH = 0, 0
	for _, m := range [...]*ebiten.Image{c.marks.Pending, c.marks.Done, c.marks.Failed} {
		b := m.Bounds()
		c.markW, c.markH = max(c.markW, b.Dx()), max(c.markH, b.Dy())
	}
}

// mark returns the mark of a state.
func (c *Checklist) mark(s ChecklistState) *ebiten.Image {
	switch s {
	case ChecklistDone:
		return c.marks.Done
	case ChecklistFailed:
		return c.marks.Failed
	case ChecklistPending:
	}
	return c.marks.Pending
}

// layout places the rows, after a change: each below the one before,
// wrapped to MaxWidth, the hidden ones 0 high but as wide as they will be;
// the parent lays out again as the checklist's size changes.
func (c *Checklist) layout() {
	if !c.dirty || c.computedParams.Face == nil {
		return
	}
	c.dirty = false
	p := c.computedParams
	pad, rp, gap := p.Padding, p.RowPadding, *p.Spacing
	y, w := pad.Top, 0
	c.titleH = 0
	if c.titleText != nil {
		tw, th := c.titleText.PreferredSize()
		pw, ph := c.progressText.PreferredSize()
		c.titleH = rp.Top + max(th, ph) + rp.Bottom
		w = rp.Left + tw + gap + pw + rp.Right
		y += c.titleH
	}
	for i := range c.rows {
		r := &c.rows[i]
		r.y, r.h = y, 0
		dw, dh := 0, 0
		if c.items[i].Detail != "" {
			dw, dh = r.detail.PreferredSize()
			dw += gap
		}
		if c.maxWidth > 0 {
			r.label.MaxWidth = float64(max(1, c.maxWidth-pad.Left-pad.Right-rp.Left-rp.Right-c.markW-gap-dw))
		}
		lw, lh := r.label.PreferredSize()
		w = max(w, rp.Left+c.markW+gap+lw+dw+rp.Right) // hidden rows too: the list keeps its width as they show
		if !c.shown(i) {
			continue
		}
		r.h = rp.Top + max(c.markH, lh, dh) + rp.Bottom
		y += r.h
	}
	w += pad.Left + pad.Right
	if c.maxWidth > 0 {
		w = min(w, c.maxWidth)
	}
	h := y + pad.Bottom
	if w != c.prefW || h != c.prefH {
		c.prefW, c.prefH = w, h
		c.widget.relayoutParent = c.laidOut
	}
	c.laidOut = true
}

// PreferredSize is the title's line and the rows shown, as wide as the
// widest, or MaxWidth.
func (c *Checklist) PreferredSize() (int, int) {
	c.init.Do()
	c.layout()
	return max(c.prefW, c.widget.MinWidth), max(c.prefH, c.widget.MinHeight)
}

// rowAt returns the row shown at y pixels below the checklist's top, or -1.
func (c *Checklist) rowAt(y int) int {
	for i, r := range c.rows {
		if r.h > 0 && y >= r.y && y < r.y+r.h {
			return i
		}
	}
	return -1
}

// rowRect returns row i's place on the screen.
func (c *Checklist) rowRect(i int) img.Rectangle {
	r, p := c.widget.Rect, c.computedParams.Padding
	top := r.Min.Y + c.rows[i].y
	return img.Rect(r.Min.X+p.Left, top, r.Max.X-p.Right, top+c.rows[i].h)
}

// focusable reports whether the focus may rest on row i: a tickable item
// shown.
func (c *Checklist) focusable(i int) bool {
	return i >= 0 && i < len(c.items) && c.items[i].Tickable && c.shown(i)
}

// nextFocusable returns the next row from row from, by dir, the focus may
// rest on, or -1.
func (c *Checklist) nextFocusable(from, dir int) int {
	for i := from + dir; i >= 0 && i < len(c.items); i += dir {
		if c.focusable(i) {
			return i
		}
	}
	return -1
}

func (c *Checklist) Update(updObj *UpdateObject) {
	c.init.Do()
	c.layout()
	c.widget.Update(updObj)
	hover := -1
	if c.hovering && !c.widget.Disabled {
		_, y := input.CursorPosition()
		if i := c.rowAt(y - c.widget.Rect.Min.Y); i >= 0 && c.items[i].Tickable {
			hover = i
		}
	}
	if hover != c.hover {
		c.hover = hover
		changes.Mark()
	}
	if c.hasFocus && !c.widget.Disabled {
		c.handleKeys(updObj.Actions)
	}
}

// handleKeys moves the focus through the tickable items with Up and Down,
// leaving them at its ends so a gamepad moves the focus on from there, and
// ticks the item it is on with Activate, Space or Enter.
func (c *Checklist) handleKeys(actions *Actions) {
	if n := c.nextFocusable(c.focused, 1); n >= 0 && actions.Take(ActionDown) {
		c.focused = n
		changes.Mark()
	}
	if n := c.nextFocusable(c.focused, -1); n >= 0 && actions.Take(ActionUp) {
		c.focused = n
		changes.Mark()
	}
	if c.focused >= 0 && (actions.Take(ActionActivate) || input.KeyJustPressed(ebiten.KeyEnter) || input.KeyJustPressed(ebiten.KeySpace)) {
		c.Toggle(c.focused)
	}
}

// Render draws each row shown: its background, hovered, focused or
// current, its mark on its first line, its label, struck through when
// done with StrikeDone, and its detail at its end.
func (c *Checklist) Render(screen *ebiten.Image) {
	c.init.Do()
	c.layout()
	c.widget.Render(screen)
	p := c.computedParams
	rp := p.RowPadding
	if c.titleText != nil {
		r := c.widget.Rect
		top := r.Min.Y + p.Padding.Top + rp.Top
		x0, x1 := r.Min.X+p.Padding.Left+rp.Left, r.Max.X-p.Padding.Right-rp.Right
		c.titleText.SetLocation(img.Rect(x0, top, x1, top+c.titleH))
		c.titleText.Render(screen)
		c.progressText.SetLocation(img.Rect(x0, top, x1, top+c.titleH))
		c.progressText.Render(screen)
	}
	for i := range c.rows {
		if c.rows[i].h > 0 {
			c.renderRow(screen, i, rp)
		}
	}
}

func (c *Checklist) renderRow(screen *ebiten.Image, i int, rp *Insets) {
	row := c.rowRect(i)
	r := &c.rows[i]
	if bg := c.background(i); bg != nil {
		c.at = row.Min
		bg.Draw(screen, row.Dx(), row.Dy(), c.place)
	}
	top := row.Min.Y + rp.Top
	mark := c.mark(c.items[i].State)
	o := &c.drawOpts
	o.GeoM.Reset()
	o.GeoM.Translate(float64(row.Min.X+rp.Left), float64(top+(c.lineH-mark.Bounds().Dy())/2))
	o.ColorScale.Reset()
	screen.DrawImage(mark, o)

	right := row.Max.X - rp.Right
	if c.items[i].Detail != "" {
		dw, dh := r.detail.PreferredSize()
		r.detail.SetLocation(img.Rect(right-dw, top, right, top+dh))
		r.detail.Render(screen)
	}
	at := c.labelAt(i)
	lw, lh := r.label.PreferredSize()
	r.label.SetLocation(img.Rect(at.X, at.Y, at.X+lw, at.Y+lh))
	r.label.Render(screen)
	if c.strike && c.items[i].State == ChecklistDone {
		for k := range r.label.measurements.processedLineWidths {
			s := c.strikeRect(i, k)
			o.GeoM.Reset()
			o.GeoM.Scale(float64(s.Dx()), float64(s.Dy()))
			o.GeoM.Translate(float64(s.Min.X), float64(s.Min.Y))
			o.ColorScale.Reset()
			o.ColorScale.ScaleWithColor(c.computedParams.Colors.Done)
			screen.DrawImage(c.pixel, o)
		}
	}
}

// labelAt returns where item i's label starts: after its mark, at its
// row's top inside the padding.
func (c *Checklist) labelAt(i int) img.Point {
	row := c.rowRect(i)
	p := c.computedParams
	return img.Pt(row.Min.X+p.RowPadding.Left+c.markW+*p.Spacing, row.Min.Y+p.RowPadding.Top)
}

// strikeRect returns the line through line k of item i's label, as it was
// last measured, up to its last letter.
func (c *Checklist) strikeRect(i, k int) img.Rectangle {
	at := c.labelAt(i)
	w := int(c.rows[i].label.measurements.processedLineWidths[k] - c.trailingSpace(i, k) + 0.5)
	y := at.Y + k*c.lineH + c.strikeY - c.strikeW/2
	return img.Rect(at.X, y, at.X+w, y+c.strikeW)
}

// trailingSpace returns the width of the spaces line k of item i's label
// ends with, as a line wrapped after a space does: its last piece, a word
// and the spaces after it.
func (c *Checklist) trailingSpace(i, k int) float64 {
	pieces := c.rows[i].label.measurements.processedLines[k] // never none
	last := pieces[len(pieces)-1].text
	return float64(len(last)-len(strings.TrimRight(last, " "))) * c.spaceW
}

// background returns what row i is drawn on: HoverImage for a tickable
// item hovered or focused, CurrentImage for the current item of a
// sequential list, or nil.
func (c *Checklist) background(i int) *image.NineSlice {
	switch {
	case !c.widget.Disabled && (i == c.hover || c.hasFocus && i == c.focused):
		return c.computedParams.HoverImage
	case c.sequential && i == c.current:
		return c.computedParams.CurrentImage
	}
	return nil
}

/** Focuser Interface - Start **/

// Focus focuses the checklist, on its first tickable item shown unless
// the focus was on one.
func (c *Checklist) Focus(focused bool) {
	c.init.Do()
	c.widget.FireFocusEvent(c, focused, img.Point{-1, -1})
	c.hasFocus = focused
	if focused && !c.focusable(c.focused) {
		c.focused = c.nextFocusable(-1, 1)
	}
	changes.Mark()
}

func (c *Checklist) IsFocused() bool { return c.hasFocus }

// TabOrder is the checklist's place in the focus order, or -1, not
// focusable, while no item is tickable.
func (c *Checklist) TabOrder() int {
	if c.tickable == 0 {
		return -1
	}
	return c.tabOrder
}

func (c *Checklist) GetFocus(direction FocusDirection) Focuser { return c.focusMap[direction] }

func (c *Checklist) AddFocus(direction FocusDirection, focus Focuser) {
	c.focusMap[direction] = focus
}

/** Focuser Interface - End **/

func (c *Checklist) createWidget() {
	c.widget = NewWidget(append([]WidgetOpt{
		WidgetOpts.TrackHover(true),
		WidgetOpts.CursorEnterHandler(func(*CursorEnterEventArgs) { c.hovering = true }),
		WidgetOpts.CursorExitHandler(func(*CursorExitEventArgs) { c.hovering = false }),
		WidgetOpts.MouseButtonPressedHandler(func(a *MouseButtonPressedEventArgs) {
			if a.Button == ebiten.MouseButtonLeft {
				c.pressed = c.rowAt(a.OffsetY)
			}
		}),
		WidgetOpts.MouseButtonClickedHandler(func(a *MouseButtonClickedEventArgs) {
			i := c.rowAt(a.OffsetY)
			if a.Button == ebiten.MouseButtonLeft && !c.widget.Disabled && i >= 0 && i == c.pressed && c.items[i].Tickable {
				c.focused = i
				c.Toggle(i)
			}
			c.pressed = -1
		}),
	}, c.widgetOpts...)...)
	c.widgetOpts = nil
	c.place = func(o *ebiten.DrawImageOptions) { o.GeoM.Translate(float64(c.at.X), float64(c.at.Y)) }
}
