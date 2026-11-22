package widget_test

import (
	"cmp"
	img "image"
	"image/color"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui/event"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// body is a row of the tables tested: a body's name, mass and the year it
// was found.
type body struct {
	name  string
	mass  float64
	found int
}

func bodies() []any {
	return []any{
		&body{"Mars", 6.42e23, 1659},
		&body{"Earth", 5.97e24, 0},
		&body{"Io", 8.93e22, 1610},
		&body{"Europa", 4.80e22, 1610},
	}
}

func bodyName(r any) string { return r.(*body).name }

var bodyColumns = []widget.TableColumn{
	{Header: "Name", Text: bodyName, Compare: func(a, b any) int { return cmp.Compare(bodyName(a), bodyName(b)) }},
	{Header: "Mass (kg)", Text: func(r any) string { return strconv.FormatFloat(r.(*body).mass, 'e', 2, 64) },
		Compare: func(a, b any) int { return cmp.Compare(a.(*body).mass, b.(*body).mass) }, Align: widget.TextPositionEnd},
	{Header: "Found", Text: func(r any) string { return strconv.Itoa(r.(*body).found) },
		Compare: func(a, b any) int { return cmp.Compare(a.(*body).found, b.(*body).found) }, Width: 70},
	{Header: "Notes", Text: func(any) string { return "" }}, // not sortable
}

// newTable returns a table of bodies at 0,0 to w,200, drawn twice.
func newTable(t *testing.T, w int, rows []any, opts ...widget.TableOpt) *widget.Table {
	t.Helper()
	tb := widget.NewTable(append([]widget.TableOpt{
		widget.TableOpts.Columns(bodyColumns...),
		widget.TableOpts.Rows(rows),
		widget.TableOpts.Face(loadFont(t)),
		widget.TableOpts.HeaderImage(&widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}),
		widget.TableOpts.ListParams(&widget.ListParams{
			ScrollContainerImage: &widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Disabled: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)},
			Slider:               &widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}},
			EntryColor: &widget.ListEntryColor{
				Unselected: color.Transparent, Selected: color.Transparent, DisabledUnselected: color.Transparent,
				DisabledSelected: color.Transparent, SelectedBackground: color.Transparent, DisabledSelectedBackground: color.Transparent,
			},
		}),
	}, opts...)...)
	tb.Validate()
	tb.SetLocation(img.Rect(0, 0, w, 200))
	drawTable(tb)
	return tb
}

func drawTable(tb *widget.Table) {
	screen := ebiten.NewImage(tb.GetWidget().Rect.Dx(), 200)
	for range 2 {
		tb.Render(screen)
	}
}

func names(rows []any) []string {
	s := make([]string, len(rows))
	for i, r := range rows {
		s[i] = bodyName(r)
	}
	return s
}

// A table shows each row's text in each column, laid out under the
// header's cells, the rows in the order given until sorted.
func TestTable_ShowsColumns(t *testing.T) {
	is := is.New(t)

	tb := newTable(t, 500, bodies())
	is.Equal(tb.CellText(0, 0), "Mars")
	is.Equal(tb.CellText(0, 1), "6.42e+23")
	is.Equal(tb.CellText(1, 2), "0")
	is.Equal(tb.ColumnWidth(2), 70) // as given
	h := tb.Header()
	is.Equal(h.CellRect(1).Min.X, h.CellRect(0).Max.X)
	is.Equal(h.CellRect(0).Min.X, 0)
	is.True(tb.ColumnWidth(0) >= 60) // fitted to its header, no narrower than 60
}

// clickHeader presses and lets go in column i's header cell, with Shift
// held or not.
func clickHeader(tb *widget.Table, d *uitest.Devices, i int, shift bool) {
	r := tb.Header().CellRect(i)
	d.Cursor = img.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
	d.Keys[ebiten.KeyShift] = shift
	for _, down := range []bool{true, false} {
		d.Mouse[ebiten.MouseButtonLeft] = down
		input.Update()
		tb.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
		drawTable(tb)
	}
	d.Keys[ebiten.KeyShift] = false
}

// A click on a sortable column's header sorts by it, again the other way;
// Shift and a click sorts by one more column among rows equal so far; an
// unsortable column's header sorts nothing.
func TestTable_Sorts(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	tb := newTable(t, 500, bodies(), widget.TableOpts.SortBy(widget.TableSort{Column: 1}))
	is.Equal(names(tb.ShownRows()), []string{"Europa", "Io", "Mars", "Earth"}) // by mass

	clickHeader(tb, d, 0, false)
	is.Equal(names(tb.ShownRows()), []string{"Earth", "Europa", "Io", "Mars"})
	is.Equal(tb.CellText(0, 0), "Earth") // the rows show the order
	clickHeader(tb, d, 0, false)
	is.Equal(names(tb.ShownRows()), []string{"Mars", "Io", "Europa", "Earth"})

	clickHeader(tb, d, 2, false) // by year found, Europa and Io tied
	clickHeader(tb, d, 1, true)  // then by mass
	is.Equal(tb.SortedBy(), []widget.TableSort{{Column: 2}, {Column: 1}})
	is.Equal(names(tb.ShownRows()), []string{"Earth", "Europa", "Io", "Mars"})
	clickHeader(tb, d, 1, true) // the second key the other way
	is.Equal(names(tb.ShownRows()), []string{"Earth", "Io", "Europa", "Mars"})

	clickHeader(tb, d, 3, false)
	is.Equal(tb.SortedBy(), []widget.TableSort{{Column: 2}, {Column: 1, Descending: true}}) // unsortable: unchanged
}

// A filter shows only the rows it keeps, in their order; Rows stays every
// row, and rows added later are sorted into place and filtered.
func TestTable_Filter(t *testing.T) {
	is := is.New(t)

	tb := newTable(t, 500, bodies(), widget.TableOpts.SortBy(widget.TableSort{Column: 0}))
	tb.SetFilter(func(r any) bool { return strings.Contains(bodyName(r), "a") })
	is.Equal(names(tb.ShownRows()), []string{"Earth", "Europa", "Mars"})
	is.Equal(len(tb.Rows()), 4)
	tb.AddRow(&body{"Ganymede", 1.48e23, 1610})
	tb.AddRow(&body{"Io II", 1, 1})
	is.Equal(names(tb.ShownRows()), []string{"Earth", "Europa", "Ganymede", "Mars"})
	tb.SetFilter(nil)
	is.Equal(len(tb.ShownRows()), 6)
}

// Text too long for its cell ends with "…", within the cell.
func TestTable_FitsText(t *testing.T) {
	is := is.New(t)

	rows := []any{&body{"Mars and its moons, Phobos and Deimos", 6.42e23, 1659}}
	tb := newTable(t, 500, rows)
	tb.SetColumnWidth(0, 50)
	drawTable(tb)
	s := tb.CellText(0, 0)
	is.True(strings.HasSuffix(s, "…"))
	is.True(utf8.RuneCountInString(s) < 10)
	is.Equal(widget.FontAdvanceAt(s, utf8.RuneCountInString(s), loadFont(t)) <= 50-12, true) // within the padding
}

// Dragging a column's edge in the header resizes it, no narrower than its
// minimum; the columns after it move along.
func TestTable_ResizesColumns(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	tb := newTable(t, 500, bodies())
	w0 := tb.ColumnWidth(0)
	edge := tb.Header().CellRect(0).Max.X
	drag := func(from, to int) {
		d.Cursor = img.Pt(from, 5)
		for _, s := range []struct {
			x    int
			down bool
		}{{from, true}, {to, true}, {to, false}} {
			d.Cursor = img.Pt(s.x, 5)
			d.Mouse[ebiten.MouseButtonLeft] = s.down
			input.Update()
			tb.Update(&widget.UpdateObject{})
			input.AfterUpdate()
			event.ExecuteDeferred()
			drawTable(tb)
		}
	}
	drag(edge, edge+40)
	is.Equal(tb.ColumnWidth(0), w0+40)
	is.Equal(tb.Header().CellRect(1).Min.X, edge+40)
	is.Equal(tb.SortedBy(), []widget.TableSort(nil)) // a resize sorts nothing
	drag(edge+40, -500)
	is.Equal(tb.ColumnWidth(0), 24) // the minimum
}

// A rank column numbers the rows in the table's order, 1 the first, and a
// row keeps its rank while a filter hides the rows before it: a
// leaderboard. Highlight marks rows, such as the player's.
func TestTable_RankAndHighlight(t *testing.T) {
	is := is.New(t)

	tb := newTable(t, 600, bodies(), widget.TableOpts.RankColumn("#"),
		widget.TableOpts.SortBy(widget.TableSort{Column: 1, Descending: true}),
		widget.TableOpts.Highlight(func(r any) bool { return bodyName(r) == "Mars" }))
	is.Equal(tb.CellText(0, 1), "Earth") // the user's columns after the rank
	is.Equal([]string{tb.CellText(0, 0), tb.CellText(1, 0), tb.CellText(2, 0)}, []string{"1", "2", "3"})
	is.Equal(tb.SortedBy(), []widget.TableSort{{Column: 2, Descending: true}}) // past the rank column
	is.True(tb.RowHighlighted(1) && !tb.RowHighlighted(0))

	tb.SetFilter(func(r any) bool { return bodyName(r) == "Io" })
	drawTable(tb)
	is.Equal(tb.CellText(0, 0), "3") // its place, whatever is hidden
}

// A click on a row chooses it; a double click activates it.
func TestTable_ChoosesAndActivates(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	var chosen, activated any
	tb := newTable(t, 500, bodies(),
		widget.TableOpts.RowSelectedHandler(func(a *widget.TableRowEventArgs) { chosen = a.Row }),
		widget.TableOpts.RowActivatedHandler(func(a *widget.TableRowEventArgs) { activated = a.Row }))
	rows := tb.List().Rows()
	top, bottom := rows.RowSpan(1)
	y := rows.Widget().Rect.Min.Y + (top+bottom)/2
	for range 2 {
		for _, down := range []bool{true, false} {
			d.Cursor, d.Mouse[ebiten.MouseButtonLeft] = img.Pt(100, y), down
			input.Update()
			tb.Update(&widget.UpdateObject{})
			input.AfterUpdate()
			event.ExecuteDeferred()
			drawTable(tb)
		}
	}
	is.Equal(bodyName(chosen), "Earth")
	is.Equal(bodyName(activated), "Earth")
	is.Equal(bodyName(tb.SelectedRow()), "Earth")
}

// The header scrolls sideways with the rows, on a table narrower than its
// columns.
func TestTable_HeaderFollowsTheRows(t *testing.T) {
	is := is.New(t)

	tb := newTable(t, 150, bodies())
	sc := tb.List().ScrollContainer()
	sc.ScrollLeft = 1
	drawTable(tb)
	is.True(tb.Header().CellRect(0).Min.X < 0)
	is.Equal(tb.Header().CellRect(0).Min.X, sc.ContentRect().Min.X)
}

// A table of 10 000 rows keeps a screenful of row widgets.
func TestTable_ManyRows(t *testing.T) {
	is := is.New(t)

	rows := make([]any, 10_000)
	for i := range rows {
		rows[i] = &body{name: "Body " + strconv.Itoa(i), mass: float64(i)}
	}
	tb := newTable(t, 500, rows, widget.TableOpts.SortBy(widget.TableSort{Column: 1, Descending: true}))
	is.True(len(tb.List().Rows().Labels()) <= 200/10+2)
	tb.SetColumnWidth(0, 150)
	drawTable(tb)
	is.Equal(tb.CellText(0, 0), "Body 9999")
}

// A theme gives a table its look; a game's options win over it.
func TestTable_Theme(t *testing.T) {
	is := is.New(t)

	glow := eimage.NewNineSliceColor(color.White)
	theme := themes.GetBasicDarkTheme()
	theme.TableTheme = &widget.TableParams{HighlightImage: glow}
	tb := widget.NewTable(widget.TableOpts.Columns(bodyColumns...))
	tb.GetWidget().SetTheme(theme)
	tb.Validate()
	is.True(tb.ComputedParams().HighlightImage == glow)
	is.True(tb.ComputedParams().Face == theme.DefaultFace)
	is.True(tb.ComputedParams().HeaderImage == theme.ButtonTheme.Image)
}

// An update of a still table allocates nothing.
func TestTable_IdleUpdateAllocatesNothing(t *testing.T) {
	uitest.Use(t)
	tb := newTable(t, 500, bodies())
	u := &widget.UpdateObject{}
	if n := testing.AllocsPerRun(100, func() {
		input.Update()
		tb.Update(u)
		input.AfterUpdate()
	}); n != 0 {
		t.Errorf("an update allocated %v times", n)
	}
}

// A press on a header moved to another before letting go sorts nothing.
func TestTable_APressMovedOffSortsNothing(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	tb := newTable(t, 500, bodies())
	center := func(i int) img.Point {
		r := tb.Header().CellRect(i)
		return img.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
	}
	for _, s := range []struct {
		at   img.Point
		down bool
	}{{center(0), true}, {center(2), true}, {center(2), false}} {
		d.Cursor, d.Mouse[ebiten.MouseButtonLeft] = s.at, s.down
		input.Update()
		tb.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
	}
	is.Equal(tb.SortedBy(), []widget.TableSort(nil))
}
