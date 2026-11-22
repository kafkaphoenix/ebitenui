package widget

import (
	img "image"
	"image/color"
	"slices"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
)

// listRows is a List's scrolled content: a row for each entry, as tall as
// all of them, laying out, updating and drawing only the rows in view.
// Labeled entries share a few rows of text of one height, each showing the
// entry in view it is bound to (entry i in row i modulo their number), so a
// list of thousands of labels keeps a screenful of widgets. Entries with
// widgets of their own (EntryWidgetFunc) keep one each, its own height:
// the tops of their rows are measured once, again only as the entries or
// their sizes change.
type listRows struct {
	list   *List
	widget *Widget

	// Labeled entries' rows, their height, and the widest label bound yet:
	// the content is as wide as that, so a wide entry scrolls the view.
	labels []listLabelRow
	rowH   int
	widest int

	// Entries with widgets of their own: each entry's widget, the top of
	// each row with the content's height last, and the widest widget.
	widgets  map[any]PreferredSizeLocateableWidget
	tops     []int
	width    int
	measured bool // whether tops and width match the entries

	first, last int // the entries in view, last excluded

	hovering bool
	hoverY   int // the pointer's y while it is over the rows
	pressed  int // the entry the left button is down on, -1 for none
	downOn   int // the entry the left button last went down on: a click selects it only

	// Where the next background goes, and the placement that puts it
	// there, made once.
	at    img.Point
	place image.DrawImageOptionsFunc
}

type listLabelRow struct {
	text  *Text
	entry int // the entry it shows, -1 for none
}

func newListRows(l *List) *listRows {
	r := &listRows{list: l, widget: NewWidget(), pressed: -1, downOn: -1}
	r.place = func(opts *ebiten.DrawImageOptions) { opts.GeoM.Translate(float64(r.at.X), float64(r.at.Y)) }
	r.watchPointer()
	return r
}

func (r *listRows) GetWidget() *Widget { return r.widget }

func (r *listRows) Validate() {}

func (r *listRows) SetLocation(rect img.Rectangle) { r.widget.Rect = rect }

// PreferredSize is every row's height and the widest row: labels as wide
// as the widest bound so far, entry widgets as the widest of them.
func (r *listRows) PreferredSize() (int, int) {
	if r.list.entryWidgetFunc != nil {
		r.measure()
		return r.width, r.tops[len(r.tops)-1]
	}
	r.styleLabels()
	return r.widest, len(r.list.entries) * r.rowH
}

// changed has the rows show the entries again, after they or their order
// changed: labels are bound again, widgets made for new entries and dropped
// for those gone, and the rows measured again.
func (r *listRows) changed() {
	for i := range r.labels {
		r.labels[i].entry = -1
	}
	if r.list.entryWidgetFunc == nil {
		r.relayout()
		return
	}
	if r.widgets == nil {
		r.widgets = map[any]PreferredSizeLocateableWidget{}
	}
	keep := make(map[any]bool, len(r.list.entries))
	for _, e := range r.list.entries {
		keep[e] = true
		if r.widgets[e] == nil {
			r.widgets[e] = r.makeWidget(e)
		}
	}
	for e := range r.widgets {
		if !keep[e] {
			delete(r.widgets, e)
		}
	}
	r.measured = false
	r.relayout()
}

// relayout has the list lay its rows out again, as their size changed.
func (r *listRows) relayout() {
	r.widget.relayoutParent = true
	if r.list.container != nil {
		r.list.container.RequestRelayout()
	}
}

// remake makes entry e's widget again, as UpdateEntry asks.
func (r *listRows) remake(e any) {
	if r.list.entryWidgetFunc == nil {
		for i := range r.labels {
			r.labels[i].entry = -1
		}
		return
	}
	if _, ok := r.widgets[e]; ok {
		r.widgets[e] = r.makeWidget(e)
		r.measured = false
		r.relayout()
	}
}

func (r *listRows) makeWidget(e any) PreferredSizeLocateableWidget {
	w := r.list.entryWidgetFunc(e)
	w.GetWidget().parent = r.widget
	w.Validate()
	return w
}

// measure finds each entry widget's row: its height, the padding around it
// included, and the widest row.
func (r *listRows) measure() {
	if r.measured && len(r.tops) == len(r.list.entries)+1 {
		return
	}
	pad := r.list.computedParams.EntryTextPadding
	r.tops = append(r.tops[:0], 0)
	r.width = 0
	for _, e := range r.list.entries {
		w, h := r.widgets[e].PreferredSize()
		r.tops = append(r.tops, r.tops[len(r.tops)-1]+h+pad.Top+pad.Bottom)
		r.width = max(r.width, w+pad.Left+pad.Right)
	}
	r.measured = true
}

// styleLabels makes the first labeled row, to learn a row's height.
func (r *listRows) styleLabels() {
	if r.rowH > 0 {
		return
	}
	t := r.labelRow()
	t.Label = " " // a line, as an empty label measures nothing
	_, r.rowH = t.PreferredSize()
	r.rowH = max(r.rowH, 1)
	t.Label = ""
	r.labels = append(r.labels, listLabelRow{text: t, entry: -1})
}

func (r *listRows) labelRow() *Text {
	p := r.list.computedParams
	t := NewText(
		TextOpts.Text("", p.EntryFace, color.Transparent),
		TextOpts.Padding(p.EntryTextPadding),
		TextOpts.Position(*p.EntryTextHorizontalPosition, *p.EntryTextVerticalPosition),
	)
	t.GetWidget().parent = r.widget
	t.Validate()
	return t
}

// rowSpan returns the top and bottom of entry i's row, from the content's
// top.
func (r *listRows) rowSpan(i int) (top, bottom int) {
	if r.list.entryWidgetFunc != nil {
		r.measure()
		return r.tops[i], r.tops[i+1]
	}
	r.styleLabels()
	return i * r.rowH, (i + 1) * r.rowH
}

// rowStep returns a row's height, or for rows of their own heights their
// mean: what the wheel moves the rows by, three a notch.
func (r *listRows) rowStep() int {
	if n := len(r.list.entries); r.list.entryWidgetFunc != nil && n > 0 {
		r.measure()
		return max(1, r.tops[n]/n)
	}
	r.styleLabels()
	return r.rowH
}

// rowRect returns entry i's row where it is laid out now.
func (r *listRows) rowRect(i int) img.Rectangle {
	top, bottom := r.rowSpan(i)
	rect := r.widget.Rect
	return img.Rect(rect.Min.X, rect.Min.Y+top, rect.Max.X, rect.Min.Y+bottom)
}

// entryAt returns the entry whose row holds the content's y, -1 for none.
func (r *listRows) entryAt(y int) int {
	y -= r.widget.Rect.Min.Y
	n := len(r.list.entries)
	if y < 0 || n == 0 {
		return -1
	}
	if r.list.entryWidgetFunc != nil {
		r.measure()
		i, _ := slices.BinarySearch(r.tops, y+1)
		if i == 0 || i > n {
			return -1
		}
		return i - 1
	}
	r.styleLabels()
	if i := y / r.rowH; i < n {
		return i
	}
	return -1
}

// layOut finds the entries in view and places their rows: labeled rows
// made when the view needs more and bound to the entries they come to
// show, entry widgets where their rows are.
func (r *listRows) layOut() {
	l, rect, view := r.list, r.widget.Rect, r.list.scrollContainer.ViewRect()
	r.first = max(0, r.entryAt(max(view.Min.Y, rect.Min.Y)))
	r.last = r.entryAt(view.Max.Y - 1)
	if r.last < 0 {
		r.last = len(l.entries) - 1
	}
	r.last++
	if len(l.entries) == 0 || view.Empty() || r.last <= r.first {
		r.first, r.last = 0, 0
		return
	}
	pad := l.computedParams.EntryTextPadding
	if l.entryWidgetFunc == nil {
		for len(r.labels) < r.last-r.first {
			r.labels = append(r.labels, listLabelRow{text: r.labelRow(), entry: -1})
		}
	}
	for i := r.first; i < r.last; i++ {
		row := r.rowRect(i)
		if l.entryWidgetFunc != nil {
			r.widgets[l.entries[i]].SetLocation(pad.Apply(row))
			continue
		}
		lr := &r.labels[i%len(r.labels)]
		if lr.entry != i {
			lr.text.Label = l.entryLabelFunc(l.entries[i])
			lr.entry = i
			if w, _ := lr.text.PreferredSize(); w > r.widest {
				r.widest = w
				r.widget.relayoutParent = true // the content grew wider
			}
		}
		lr.text.SetLocation(row)
	}
}

// background returns what entry i's row is drawn on: the selected or
// unselected background in the row's state.
func (r *listRows) background(i int) *image.NineSlice {
	l := r.list
	set := l.computedParams.entryUnselectedColor
	if l.entries[i] == l.selectedEntry {
		set = l.computedParams.entrySelectedColor
	}
	switch {
	case l.container.GetWidget().Disabled:
		return set.Disabled
	case i == r.pressed && set.Pressed != nil:
		return set.Pressed
	case (r.hovering && i == r.entryAt(r.hoverY) || l.focused && i == l.focusIndex) && set.Hover != nil:
		return set.Hover
	}
	return set.Idle
}

// labelColor returns the color of entry i's label: selected or not, enabled
// or not, or the idle one where that state has none.
func (r *listRows) labelColor(i int) color.Color {
	l := r.list
	c := l.computedParams.entryUnselectedTextColor
	if l.entries[i] == l.selectedEntry {
		c = l.computedParams.entryTextColor
	}
	if l.container.GetWidget().Disabled && c.Disabled != nil {
		return c.Disabled
	}
	return c.Idle
}

func (r *listRows) Render(screen *ebiten.Image) {
	r.widget.Render(screen)
	r.layOut()
	for i := r.first; i < r.last; i++ {
		row := r.rowRect(i)
		if bg := r.background(i); bg != nil {
			r.at = row.Min
			bg.Draw(screen, row.Dx(), row.Dy(), r.place)
		}
		if r.list.entryWidgetFunc != nil {
			if w, ok := r.widgets[r.list.entries[i]].(Renderer); ok {
				w.Render(screen)
			}
			continue
		}
		t := r.labels[i%len(r.labels)].text
		t.SetColor(r.labelColor(i))
		t.Render(screen)
	}
}

func (r *listRows) Update(updObj *UpdateObject) {
	r.widget.Update(updObj)
	if r.list.entryWidgetFunc == nil {
		return
	}
	for i := r.first; i < min(r.last, len(r.list.entries)); i++ {
		if u, ok := r.widgets[r.list.entries[i]].(Updater); ok {
			u.Update(updObj)
		}
	}
}

func (r *listRows) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	if r.list.entryWidgetFunc == nil {
		return
	}
	for i := r.first; i < min(r.last, len(r.list.entries)); i++ {
		if l, ok := r.widgets[r.list.entries[i]].(input.Layerer); ok {
			l.SetupInputLayer(def)
		}
	}
}

// GetFocusers returns what the entry widgets in view hold that takes the
// focus, so a gamepad reaches into a list's entries as into a container.
func (r *listRows) GetFocusers() []Focuser {
	if r.list.entryWidgetFunc == nil {
		return nil
	}
	var fs []Focuser
	for i := r.first; i < min(r.last, len(r.list.entries)); i++ {
		switch w := r.widgets[r.list.entries[i]].(type) {
		case interface{ GetFocusers() []Focuser }:
			fs = append(fs, w.GetFocusers()...)
		case Focuser:
			fs = append(fs, w)
		}
	}
	return fs
}

// watchPointer follows the pointer over the rows: the entry it is over, the
// one it goes down on, and a click (or with SelectPressed, a press), which
// selects the entry under it unless the list is disabled.
func (r *listRows) watchPointer() {
	w := r.widget
	w.CursorEnterEvent.AddHandler(func(args any) {
		if a, ok := args.(*CursorEnterEventArgs); ok {
			r.hovering, r.hoverY = true, w.Rect.Min.Y+a.OffsetY
		}
	})
	w.CursorMoveEvent.AddHandler(func(args any) {
		if a, ok := args.(*CursorMoveEventArgs); ok {
			r.hoverY = w.Rect.Min.Y + a.OffsetY
		}
	})
	w.CursorExitEvent.AddHandler(func(any) { r.hovering = false })
	w.MouseButtonPressedEvent.AddHandler(func(args any) {
		if a, ok := args.(*MouseButtonPressedEventArgs); ok && a.Button == ebiten.MouseButtonLeft {
			r.pressed = r.entryAt(w.Rect.Min.Y + a.OffsetY)
			r.downOn = r.pressed
			if *r.list.computedParams.SelectPressed {
				r.selectAt(w.Rect.Min.Y + a.OffsetY)
			}
		}
	})
	w.MouseButtonReleasedEvent.AddHandler(func(args any) {
		if a, ok := args.(*MouseButtonReleasedEventArgs); ok && a.Button == ebiten.MouseButtonLeft {
			r.pressed = -1
		}
	})
	w.MouseButtonClickedEvent.AddHandler(func(args any) {
		if a, ok := args.(*MouseButtonClickedEventArgs); ok && a.Button == ebiten.MouseButtonLeft && !*r.list.computedParams.SelectPressed &&
			r.entryAt(w.Rect.Min.Y+a.OffsetY) == r.downOn {
			r.selectAt(w.Rect.Min.Y + a.OffsetY)
		}
	})
}

// selectAt selects the entry whose row holds y, as the player chose it.
func (r *listRows) selectAt(y int) {
	if r.list.GetWidget().Disabled {
		return
	}
	if i := r.entryAt(y); i >= 0 {
		r.list.setSelectedEntry(r.list.entries[i], true)
	}
}
