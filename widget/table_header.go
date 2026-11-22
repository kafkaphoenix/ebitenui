package widget

import (
	img "image"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
)

// tableEdge is how near a column's right edge, in pixels, a press in the
// header resizes the column rather than sorting by it.
const tableEdge = 4

// tableHeader is a table's header: a cell a column, sorting the rows on a
// click and resizing a column on a drag at its edge, scrolled sideways
// with the body.
type tableHeader struct {
	table  *Table
	widget *Widget
	texts  []*Text
	fitW   []int // the width each header text was cut to

	hover, pressed, resizing int // the columns hovered, pressed, resized; -1 for none
	dragFrom, dragWidth      int

	at       img.Point
	place    func(*ebiten.DrawImageOptions)
	markOpts ebiten.DrawImageOptions
}

func (h *tableHeader) validate() {
	t := h.table
	p := t.computedParams
	h.texts, h.fitW = h.texts[:0], h.fitW[:0]
	for _, c := range t.columns {
		tx := NewText(TextOpts.Text(c.Header, p.Face, p.HeaderTextColor), TextOpts.Position(c.Align, TextPositionCenter))
		tx.GetWidget().parent = h.widget
		tx.Validate()
		h.texts = append(h.texts, tx)
		h.fitW = append(h.fitW, -1)
	}
	h.place = func(o *ebiten.DrawImageOptions) { o.GeoM.Translate(float64(h.at.X), float64(h.at.Y)) }
}

func (h *tableHeader) GetWidget() *Widget { return h.widget }

func (h *tableHeader) Validate() {}

func (h *tableHeader) SetLocation(rect img.Rectangle) { h.widget.Rect = rect }

// PreferredSize is every column's width, and a line's height.
func (h *tableHeader) PreferredSize() (int, int) {
	w := 0
	for _, cw := range h.table.widths {
		w += cw
	}
	return w, h.table.cellH
}

// x0 returns where the first column starts: where the body's rows do,
// scrolled sideways with them.
func (h *tableHeader) x0() int {
	if sc := h.table.list.scrollContainer; sc != nil && sc.content != nil {
		return sc.ContentRect().Min.X
	}
	return h.widget.Rect.Min.X
}

// clip returns the part of the header the cells show in: across the
// body's view.
func (h *tableHeader) clip() img.Rectangle {
	r := h.widget.Rect
	if sc := h.table.list.scrollContainer; sc != nil {
		v := sc.ViewRect()
		r.Min.X, r.Max.X = max(r.Min.X, v.Min.X), min(r.Max.X, v.Max.X)
	}
	return r
}

// cellRect returns column i's cell in the header.
func (h *tableHeader) cellRect(i int) img.Rectangle {
	x := h.x0()
	for _, w := range h.table.widths[:i] {
		x += w
	}
	r := h.widget.Rect
	return img.Rect(x, r.Min.Y, x+h.table.widths[i], r.Max.Y)
}

// columnAt returns the column under x, -1 for none, and whether x is at
// its right edge.
func (h *tableHeader) columnAt(x int) (int, bool) {
	cx := h.x0()
	for i, w := range h.table.widths {
		if abs(x-(cx+w)) <= tableEdge {
			return i, true
		}
		if x >= cx && x < cx+w {
			return i, false
		}
		cx += w
	}
	return -1, false
}

// Update follows the pointer: a press on a cell and its release there
// sorts by the column, with Shift by one more; a press at a column's edge
// drags its width.
func (h *tableHeader) Update(updObj *UpdateObject) {
	h.widget.Update(updObj)
	t := h.table
	x, y := input.CursorPosition()
	in := img.Pt(x, y).In(h.clip())
	col, edge := -1, false
	if in {
		col, edge = h.columnAt(x)
	}
	h.hover = -1
	if !edge {
		h.hover = col
	}
	if edge || h.resizing >= 0 {
		input.SetCursorShape(input.CursorEWResize)
	}
	switch {
	case input.MouseButtonJustPressedLayer(ebiten.MouseButtonLeft, h.widget.EffectiveInputLayer()):
		switch {
		case col >= 0 && edge:
			h.resizing, h.dragFrom, h.dragWidth = col, x, t.widths[col]
		case col >= 0:
			h.pressed = col
		}
	case h.resizing >= 0 && input.MouseButtonPressed(ebiten.MouseButtonLeft):
		t.SetColumnWidth(h.resizing, h.dragWidth+x-h.dragFrom)
	case h.resizing >= 0:
		h.resizing = -1
	case h.pressed >= 0 && !input.MouseButtonPressed(ebiten.MouseButtonLeft):
		if col == h.pressed && !edge {
			t.sortClick(col, input.KeyPressed(ebiten.KeyShift))
		}
		h.pressed = -1
	}
}

// Render draws each cell within the body's view across: its face for its
// state, its text, cut to fit, and a sort mark in a sorted column's, faded
// for a key after the first.
func (h *tableHeader) Render(screen *ebiten.Image) {
	t := h.table
	p := t.computedParams
	area, ok := screen.SubImage(h.clip()).(*ebiten.Image)
	if !ok || area.Bounds().Empty() {
		return
	}
	for i := range t.columns {
		cell := h.cellRect(i)
		if face := h.face(i); face != nil {
			h.at = cell.Min
			face.Draw(area, cell.Dx(), cell.Dy(), h.place)
		}
		mark, faded := h.mark(i)
		room := cell.Dx() - p.CellPadding.Left - p.CellPadding.Right
		if mark != nil {
			room -= mark.Bounds().Dx() + 4
		}
		if room != h.fitW[i] {
			h.fitW[i] = room
			h.texts[i].Label = fitText(t.columns[i].Header, p.Face, room)
		}
		h.texts[i].SetLocation(img.Rect(cell.Min.X+p.CellPadding.Left, cell.Min.Y, cell.Min.X+p.CellPadding.Left+max(room, 0), cell.Max.Y))
		h.texts[i].Render(area)
		if mark != nil {
			b := mark.Bounds()
			o := &h.markOpts
			o.GeoM.Reset()
			o.GeoM.Translate(float64(cell.Max.X-p.CellPadding.Right-b.Dx()), float64(cell.Min.Y+(cell.Dy()-b.Dy())/2))
			o.ColorScale.Reset()
			if faded {
				o.ColorScale.ScaleAlpha(0.5)
			}
			area.DrawImage(mark, o)
		}
	}
}

// face returns column i's cell's face for its state: pressed, hovered or
// idle.
func (h *tableHeader) face(i int) *image.NineSlice {
	faces := h.table.computedParams.HeaderImage
	if faces == nil {
		return nil
	}
	switch {
	case i == h.pressed && faces.Pressed != nil:
		return faces.Pressed
	case i == h.hover && faces.Hover != nil:
		return faces.Hover
	}
	return faces.Idle
}

// mark returns the sort mark column i shows, if the rows sort by it, and
// whether it is faded, for a key after the first.
func (h *tableHeader) mark(i int) (*ebiten.Image, bool) {
	p := h.table.computedParams
	for k, s := range h.table.sorts {
		if s.Column == i {
			if s.Descending {
				return p.SortDescending, k > 0
			}
			return p.SortAscending, k > 0
		}
	}
	return nil, false
}
