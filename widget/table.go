package widget

import (
	img "image"
	"image/color"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TableColumn is a column of a table: its header, each row's text in it,
// how rows sort by it, its width and the cells' alignment.
type TableColumn struct {
	Header string
	// Text returns a row's text in the column, as the game formats it:
	// a number in scientific notation, a date, a name.
	Text func(row any) string
	// Compare orders two rows by the column, as slices.SortFunc takes:
	// negative when a comes first. Nil leaves the column unsortable.
	Compare func(a, b any) int
	// Width is the column's width in pixels; 0 fits its header.
	// MinWidth is the narrowest a drag makes it; 24 by default.
	Width, MinWidth int
	// Align places the cells' text: at the start by default, at the end
	// for numbers.
	Align TextPosition
}

// TableSort is a sort key: a column, and whether it sorts descending.
type TableSort struct {
	Column     int
	Descending bool
}

// TableParams is a table's look: the body's, as a list's (its rows'
// backgrounds, the selection, the sliders), the header's cells, the
// cells' face, colors and padding, the sort marks, and the background of
// highlighted rows.
type TableParams struct {
	List            *ListParams
	HeaderImage     *ButtonImage
	HeaderTextColor color.Color
	TextColor       color.Color
	Face            *text.Face
	CellPadding     *Insets
	SortAscending   *ebiten.Image
	SortDescending  *ebiten.Image
	HighlightImage  *image.NineSlice
}

// Table shows rows of data in columns under a header, as a spreadsheet
// does: thousands of rows, as it keeps widgets only for those in view.
// Each column gives a row's text in it, formatted as the game wants. A
// click on a sortable column's header sorts the rows by it, again
// reverses them; Shift and a click sorts by a further column among rows
// equal so far. Dragging a column's edge in the header resizes it; text
// too long for its cell ends with "…". A filter shows only the rows it
// keeps, as a search does. Its sliders show only when the rows overflow
// it. A rank column shows each row's place in the
// order, filters aside, and highlighted rows stand out: a leaderboard
// with the player's own row. The rows scroll, and are chosen, as a
// list's entries are: rows are pointers or other values that compare
// equal only to themselves.
type Table struct {
	definedParams  TableParams
	computedParams TableParams

	RowSelectedEvent  *event.Event
	RowActivatedEvent *event.Event

	columns   []TableColumn
	widths    []int
	rows      []any       // every row, as given
	order     []any       // the rows in the sort's order
	rank      map[any]int // each row's place in order
	sorts     []TableSort
	highlight func(row any) bool
	rankName  string
	hasRank   bool

	listOpts   []ListOpt
	widgetOpts []WidgetOpt
	container  *Container
	header     *tableHeader
	list       *List
	hooked     *listRows // the body's rows the double click is hooked to
	cellH      int

	init *MultiOnce
}

// TableRowEventArgs tell which row was chosen or activated.
type TableRowEventArgs struct {
	Table *Table
	Row   any
}

type TableRowHandlerFunc func(args *TableRowEventArgs)

type TableOpt func(t *Table)

type TableOptions struct{}

// TableOpts holds the options of NewTable.
var TableOpts TableOptions

// NewTable returns a table of the columns and rows its options give.
func NewTable(opts ...TableOpt) *Table {
	t := &Table{
		RowSelectedEvent:  &event.Event{},
		RowActivatedEvent: &event.Event{},
		init:              &MultiOnce{},
	}
	for _, o := range opts {
		o(t)
	}
	t.init.Append(t.createWidget)
	return t
}

func (o TableOptions) WidgetOpts(opts ...WidgetOpt) TableOpt {
	return func(t *Table) { t.widgetOpts = append(t.widgetOpts, opts...) }
}

// Columns sets the columns, from the left.
func (o TableOptions) Columns(columns ...TableColumn) TableOpt {
	return func(t *Table) { t.columns = append(t.columns, columns...) }
}

// Rows sets the rows.
func (o TableOptions) Rows(rows []any) TableOpt {
	return func(t *Table) { t.rows = slices.Clone(rows) }
}

// SortBy sorts the rows by these keys from the start, the first first.
func (o TableOptions) SortBy(keys ...TableSort) TableOpt {
	return func(t *Table) { t.sorts = append(t.sorts[:0], keys...) }
}

// RankColumn adds a first column, headed header, of each row's place in
// the table's order, 1 for the first, whatever a filter hides: a
// leaderboard's ranks.
func (o TableOptions) RankColumn(header string) TableOpt {
	return func(t *Table) { t.hasRank, t.rankName = true, header }
}

// Highlight draws HighlightImage behind the rows f keeps, such as the
// player's own on a leaderboard.
func (o TableOptions) Highlight(f func(row any) bool) TableOpt {
	return func(t *Table) { t.highlight = f }
}

// Filter shows only the rows f keeps, from the start (SetFilter).
func (o TableOptions) Filter(f func(row any) bool) TableOpt {
	return func(t *Table) { t.listOpts = append(t.listOpts, ListOpts.EntryFilter(f)) }
}

func (o TableOptions) RowSelectedHandler(f TableRowHandlerFunc) TableOpt {
	return func(t *Table) { t.RowSelectedEvent.AddHandler(handle(f)) }
}

// RowActivatedHandler is called as a row is double clicked.
func (o TableOptions) RowActivatedHandler(f TableRowHandlerFunc) TableOpt {
	return func(t *Table) { t.RowActivatedEvent.AddHandler(handle(f)) }
}

// ListParams sets the body's look, as a list's; its entries' padding is
// the cells'.
func (o TableOptions) ListParams(p *ListParams) TableOpt {
	return func(t *Table) { t.definedParams.List = p }
}

func (o TableOptions) HeaderImage(i *ButtonImage) TableOpt {
	return func(t *Table) { t.definedParams.HeaderImage = i }
}

func (o TableOptions) Face(f *text.Face) TableOpt {
	return func(t *Table) { t.definedParams.Face = f }
}

// TextColors sets the cells' and the header's text colors.
func (o TableOptions) TextColors(cells, header color.Color) TableOpt {
	return func(t *Table) { t.definedParams.TextColor, t.definedParams.HeaderTextColor = cells, header }
}

func (o TableOptions) CellPadding(p *Insets) TableOpt {
	return func(t *Table) { t.definedParams.CellPadding = p }
}

// SortMarks sets the marks shown in a sorted column's header, in place of
// the triangles the table draws.
func (o TableOptions) SortMarks(ascending, descending *ebiten.Image) TableOpt {
	return func(t *Table) { t.definedParams.SortAscending, t.definedParams.SortDescending = ascending, descending }
}

func (o TableOptions) HighlightImage(i *image.NineSlice) TableOpt {
	return func(t *Table) { t.definedParams.HighlightImage = i }
}

// Rows returns every row, as given.
func (t *Table) Rows() []any { return t.rows }

// ShownRows returns the rows shown, in their order: sorted, and those the
// filter keeps.
func (t *Table) ShownRows() []any {
	t.init.Do()
	return t.list.ShownEntries()
}

// SetRows sets the rows, sorted as the table is.
func (t *Table) SetRows(rows []any) {
	t.init.Do()
	t.rows = slices.Clone(rows)
	t.reorder()
}

// AddRow adds a row, in its place in the order.
func (t *Table) AddRow(row any) {
	t.init.Do()
	t.rows = append(t.rows, row)
	t.reorder()
}

// RemoveRow removes a row.
func (t *Table) RemoveRow(row any) {
	t.init.Do()
	if i := slices.Index(t.rows, row); i >= 0 {
		t.rows = slices.Delete(t.rows, i, i+1)
		t.reorder()
	}
}

// UpdateRow shows a row again, as it changed, sorted again.
func (t *Table) UpdateRow(row any) {
	t.init.Do()
	t.reorder()
	t.list.UpdateEntry(row)
}

// SetFilter shows only the rows f keeps, as a search does; nil shows all.
func (t *Table) SetFilter(f func(row any) bool) {
	t.init.Do()
	t.list.SetFilter(f)
}

// SortBy sorts the rows by these keys, the first first; none keeps them
// as given.
func (t *Table) SortBy(keys ...TableSort) {
	t.init.Do()
	t.sorts = append(t.sorts[:0], keys...)
	t.reorder()
}

// SortedBy returns the sort's keys, the first first.
func (t *Table) SortedBy() []TableSort { return t.sorts }

// SelectedRow returns the row chosen, or nil.
func (t *Table) SelectedRow() any {
	t.init.Do()
	return t.list.SelectedEntry()
}

// ColumnWidth returns column i's width, in pixels.
func (t *Table) ColumnWidth(i int) int {
	t.init.Do()
	return t.widths[i]
}

// SetColumnWidth sets column i's width, no narrower than its MinWidth.
func (t *Table) SetColumnWidth(i, w int) {
	t.init.Do()
	w = max(w, t.columns[i].MinWidth)
	if w == t.widths[i] {
		return
	}
	t.widths[i] = w
	if t.list.rows != nil {
		t.list.rows.changed() // the rows bound again, their texts cut to the new width
	}
}

// sortClick sorts by column i as its header is clicked: by it alone, or
// with more, after the keys there are; again, the other way.
func (t *Table) sortClick(i int, more bool) {
	if t.columns[i].Compare == nil {
		return
	}
	at := slices.IndexFunc(t.sorts, func(k TableSort) bool { return k.Column == i })
	switch {
	case at >= 0 && (more || at == 0):
		t.sorts[at].Descending = !t.sorts[at].Descending
	case more:
		t.sorts = append(t.sorts, TableSort{Column: i})
	default:
		t.sorts = append(t.sorts[:0], TableSort{Column: i})
	}
	t.reorder()
}

// reorder sorts the rows by the keys, stably, and gives them to the body,
// keeping its selection; ranks follow the order.
func (t *Table) reorder() {
	t.order = append(t.order[:0], t.rows...)
	if len(t.sorts) > 0 {
		slices.SortStableFunc(t.order, func(a, b any) int {
			for _, k := range t.sorts {
				c := t.columns[k.Column].Compare(a, b)
				if k.Descending {
					c = -c
				}
				if c != 0 {
					return c
				}
			}
			return 0
		})
	}
	clear(t.rank)
	if t.rank == nil {
		t.rank = make(map[any]int, len(t.order))
	}
	for i, r := range t.order {
		t.rank[r] = i
	}
	t.list.setOrder(t.order)
}

func (t *Table) GetWidget() *Widget {
	t.init.Do()
	return t.container.GetWidget()
}

func (t *Table) SetLocation(rect img.Rectangle) {
	t.init.Do()
	t.container.SetLocation(rect)
}

func (t *Table) PreferredSize() (int, int) {
	t.init.Do()
	return t.container.PreferredSize()
}

func (t *Table) RequestRelayout() {
	t.init.Do()
	t.container.RequestRelayout()
}

func (t *Table) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	t.init.Do()
	t.container.SetupInputLayer(def)
}

// GetFocusers returns the body, which takes the keys as a list does.
func (t *Table) GetFocusers() []Focuser {
	t.init.Do()
	return []Focuser{t.list}
}

// PlaceChildren places the header and the body's rows in view.
func (t *Table) PlaceChildren() {
	t.init.Do()
	t.container.PlaceChildren()
}

// Validate takes the look, from the options, the theme or defaults, and
// fits each column without a width to its header.
func (t *Table) Validate() {
	t.init.Do()
	p := t.computeParams()
	t.computedParams = p
	_, lineH := text.Measure(" ", *p.Face, 0)
	t.cellH = int(lineH) + p.CellPadding.Top + p.CellPadding.Bottom
	t.fitWidths()
	if p.List != nil {
		t.list.definedParams = *p.List
	}
	t.list.definedParams.EntryTextPadding = &Insets{}
	t.list.definedParams.EntryFace = p.Face
	t.header.validate()
	t.container.Validate()
	t.hookActivation()
}

// computeParams returns the look: each of the options, else the theme's
// table, else the theme's buttons, face and text color, else defaults.
func (t *Table) computeParams() TableParams {
	theme := t.container.GetWidget().GetTheme()
	p := TableParams{}
	if theme != nil && theme.TableTheme != nil {
		p = *theme.TableTheme
	}
	var face *text.Face
	var ink color.Color = color.White
	var header *ButtonImage
	if theme != nil {
		face = theme.DefaultFace
		if theme.DefaultTextColor != nil {
			ink = theme.DefaultTextColor
		}
		if theme.ButtonTheme != nil {
			header = theme.ButtonTheme.Image
		}
	}
	d := t.definedParams
	p.List = first(d.List, p.List)
	p.HeaderImage = first(d.HeaderImage, p.HeaderImage, header)
	p.Face = first(d.Face, p.Face, face)
	p.TextColor = first(d.TextColor, p.TextColor, ink)
	p.HeaderTextColor = first(d.HeaderTextColor, p.HeaderTextColor, p.TextColor)
	p.CellPadding = first(d.CellPadding, p.CellPadding, &Insets{Left: 6, Right: 6, Top: 3, Bottom: 3})
	p.SortAscending = first(d.SortAscending, p.SortAscending)
	p.SortDescending = first(d.SortDescending, p.SortDescending)
	p.HighlightImage = first(d.HighlightImage, p.HighlightImage, image.NewNineSliceColor(color.NRGBA{255, 255, 255, 40}))
	if p.Face == nil {
		panic("Table: Face is required.")
	}
	if p.SortAscending == nil {
		p.SortAscending, p.SortDescending = sortMark(true, p.HeaderTextColor), sortMark(false, p.HeaderTextColor)
	}
	return p
}

// fitWidths gives each column its width, the first time: its own, or its
// header's with a sort mark's room.
func (t *Table) fitWidths() {
	if t.widths != nil {
		return
	}
	p := t.computedParams
	markW := p.SortAscending.Bounds().Dx() + 4
	for _, c := range t.columns {
		w := c.Width
		if w <= 0 {
			w = max(60, fontAdvanceAt(c.Header, utf8.RuneCountInString(c.Header), p.Face)+p.CellPadding.Left+p.CellPadding.Right+markW)
		}
		t.widths = append(t.widths, w)
	}
}

// hookActivation listens for double clicks on the body's rows, made anew
// as the body is validated.
func (t *Table) hookActivation() {
	r := t.list.rows
	if r == t.hooked {
		return
	}
	t.hooked = r
	r.widget.MouseButtonDoubleClickedEvent.AddHandler(func(args any) {
		if a, ok := args.(*MouseButtonDoubleClickedEventArgs); ok && a.Button == ebiten.MouseButtonLeft {
			if i := r.entryAt(r.widget.Rect.Min.Y + a.OffsetY); i >= 0 {
				t.RowActivatedEvent.Fire(&TableRowEventArgs{Table: t, Row: t.list.entries[i]})
			}
		}
	})
}

// sortMark draws a small triangle pointing up or down, once.
func sortMark(up bool, c color.Color) *ebiten.Image {
	i := ebiten.NewImage(8, 5)
	var path vector.Path
	if up {
		path.MoveTo(0, 5)
		path.LineTo(4, 0)
		path.LineTo(8, 5)
	} else {
		path.MoveTo(0, 0)
		path.LineTo(4, 5)
		path.LineTo(8, 0)
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(i, &path, &vector.FillOptions{}, op)
	return i
}

func (t *Table) Update(updObj *UpdateObject) {
	t.init.Do()
	t.container.Update(updObj)
}

// Render draws the body, and then the header, so the header follows the
// body's sideways scroll of this frame.
func (t *Table) Render(screen *ebiten.Image) {
	t.init.Do()
	t.container.GetWidget().Render(screen)
	t.container.draw(screen)
	t.container.doLayout()
	t.list.Render(screen)
	t.header.Render(screen)
}

func (t *Table) createWidget() {
	if t.hasRank {
		t.columns = append([]TableColumn{{
			Header: t.rankName,
			Text:   func(row any) string { return strconv.Itoa(t.rank[row] + 1) },
			Align:  TextPositionEnd,
		}}, t.columns...)
		for i := range t.sorts {
			t.sorts[i].Column++
		}
	}
	for i := range t.columns {
		if t.columns[i].MinWidth <= 0 {
			t.columns[i].MinWidth = 24
		}
	}
	t.container = NewContainer(
		ContainerOpts.WidgetOpts(t.widgetOpts...),
		ContainerOpts.Layout(NewGridLayout(
			GridLayoutOpts.Columns(1),
			GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
		)),
	)
	t.widgetOpts = nil
	t.header = &tableHeader{table: t, widget: NewWidget(), pressed: -1, resizing: -1, hover: -1}
	t.list = NewList(append([]ListOpt{
		ListOpts.SlidersWhenNeeded(),
		ListOpts.EntryRows(t.newRow, t.bindRow),
		ListOpts.EntrySelectedHandler(func(a *ListEntrySelectedEventArgs) {
			t.RowSelectedEvent.Fire(&TableRowEventArgs{Table: t, Row: a.Entry})
		}),
	}, t.listOpts...)...)
	t.list.GetWidget().LayoutData = GridLayoutData{}
	t.container.AddChild(t.header, t.list)
	t.reorder()
}

// newRow makes a row of the body, a text a column.
func (t *Table) newRow() PreferredSizeLocateableWidget {
	r := &tableRow{table: t, widget: NewWidget()}
	r.place = func(opts *ebiten.DrawImageOptions) { opts.GeoM.Translate(float64(r.at.X), float64(r.at.Y)) }
	for _, c := range t.columns {
		tx := NewText(TextOpts.Text("", t.computedParams.Face, t.computedParams.TextColor),
			TextOpts.Position(c.Align, TextPositionCenter))
		tx.GetWidget().parent = r.widget
		tx.Validate()
		r.cells = append(r.cells, tx)
	}
	return r
}

// bindRow has a row of the body show row: each column's text, cut to the
// column's width.
func (t *Table) bindRow(w PreferredSizeLocateableWidget, row any) {
	r, ok := w.(*tableRow) // always, made by newRow
	if !ok {
		return
	}
	pad := t.computedParams.CellPadding
	for i, c := range t.columns {
		r.cells[i].Label = fitText(c.Text(row), t.computedParams.Face, t.widths[i]-pad.Left-pad.Right)
	}
	r.highlighted = t.highlight != nil && t.highlight(row)
}

// fitText returns s, or as much of it as fits in width pixels with "…"
// after it.
func fitText(s string, face *text.Face, width int) string {
	n := utf8.RuneCountInString(s)
	if fontAdvanceAt(s, n, face) <= width {
		return s
	}
	room := width - fontAdvanceAt("…", 1, face)
	lo, hi := 0, n
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if fontAdvanceAt(s, mid, face) <= room {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return s[:runeOffset(s, lo)] + "…"
}

// tableRow is a row of a table's body: a text a column, and a background
// when highlighted.
type tableRow struct {
	table       *Table
	widget      *Widget
	cells       []*Text
	highlighted bool
	at          img.Point
	place       func(*ebiten.DrawImageOptions) // places the highlight at at, made once
}

func (r *tableRow) GetWidget() *Widget { return r.widget }

func (r *tableRow) Validate() {}

// PreferredSize is every column's width, and a line's height.
func (r *tableRow) PreferredSize() (int, int) {
	w := 0
	for _, cw := range r.table.widths {
		w += cw
	}
	return w, r.table.cellH
}

// SetLocation places the row and each cell in its column.
func (r *tableRow) SetLocation(rect img.Rectangle) {
	r.widget.Rect = rect
	pad := r.table.computedParams.CellPadding
	x := rect.Min.X
	for i, c := range r.cells {
		w := r.table.widths[i]
		c.SetLocation(img.Rect(x+pad.Left, rect.Min.Y, x+w-pad.Right, rect.Max.Y))
		x += w
	}
}

func (r *tableRow) Render(screen *ebiten.Image) {
	if r.highlighted {
		rect := r.widget.Rect
		r.at = rect.Min
		r.table.computedParams.HighlightImage.Draw(screen, rect.Dx(), rect.Dy(), r.place)
	}
	for _, c := range r.cells {
		c.Render(screen)
	}
}
