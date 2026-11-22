package widget

import (
	"image"
	"math"
)

// GridLayout layouts widgets in a grid fashion, with columns or rows optionally being stretched or weighted.
//
// Widget.LayoutData of widgets being layouted by GridLayout need to be of type GridLayoutData.
type GridLayout struct {
	columns       int
	padding       *Insets
	columnSpacing int
	rowSpacing    int
	columnStretch []bool
	rowStretch    []bool
	columnWeights []float64
	rowWeights    []float64

	defaultColumnStretch bool
	defaultRowStretch    bool

	// The column widths and row heights of the last measure, for the next
	// to reuse; nil while a measure holds them.
	spareWidths, spareHeights []int
}

// GridLayoutOpt is a function that configures g.
type GridLayoutOpt func(g *GridLayout)

type GridLayoutOptions struct {
}

// GridLayoutData specifies layout settings for a widget.
type GridLayoutData struct {
	// MaxWidth specifies the maximum width.
	MaxWidth int

	// MaxHeight specifies the maximum height..
	MaxHeight int

	// HorizontalPosition specifies the horizontal anchoring position inside the grid cell.
	HorizontalPosition GridLayoutPosition

	// VerticalPosition specifies the vertical anchoring position inside the grid cell.
	VerticalPosition GridLayoutPosition
}

// GridLayoutPosition is the type used to specify an anchoring position.
type GridLayoutPosition int

const (
	// GridLayoutPositionStart is the anchoring position for "left" (in the horizontal direction) or "top" (in the vertical direction.)
	GridLayoutPositionStart = GridLayoutPosition(iota)

	// GridLayoutPositionStart is the center anchoring position.
	GridLayoutPositionCenter

	// GridLayoutPositionStart is the anchoring position for "right" (in the horizontal direction) or "bottom" (in the vertical direction.)
	GridLayoutPositionEnd
)

// GridLayoutOpts contains functions that configure a GridLayout.
var GridLayoutOpts GridLayoutOptions

// NewGridLayout constructs a new GridLayout, configured by opts.
func NewGridLayout(opts ...GridLayoutOpt) *GridLayout {
	g := &GridLayout{}

	for _, o := range opts {
		o(g)
	}

	if g.padding == nil {
		g.padding = &Insets{}
	}

	g.validate()

	return g
}

func (g *GridLayout) validate() {
	if g.columns <= 0 {
		panic("GridLayout: columns must be positive.")
	}
	for _, weights := range [][]float64{g.columnWeights, g.rowWeights} {
		for _, weight := range weights {
			if weight < 0 || math.IsNaN(weight) || math.IsInf(weight, 0) {
				panic("GridLayout: weights must be finite and non-negative.")
			}
		}
	}
}

// Columns configures a grid layout to use c columns.
func (o GridLayoutOptions) Columns(c int) GridLayoutOpt {
	return func(g *GridLayout) {
		g.columns = c
	}
}

// Padding configures a grid layout to use padding i.
func (o GridLayoutOptions) Padding(i *Insets) GridLayoutOpt {
	return func(g *GridLayout) {
		g.padding = i
	}
}

// Spacing configures a grid layout to separate columns by spacing c and rows by spacing r.
func (o GridLayoutOptions) Spacing(c int, r int) GridLayoutOpt {
	return func(g *GridLayout) {
		g.columnSpacing = c
		g.rowSpacing = r
	}
}

// Stretch configures a grid layout to stretch columns according to c and rows according to r.
// The number of elements of c and r must correspond with the number of columns and rows in the
// layout.
func (o GridLayoutOptions) Stretch(c []bool, r []bool) GridLayoutOpt {
	return func(g *GridLayout) {
		g.columnStretch = c
		g.rowStretch = r
	}
}

// Weights configures relative column and row sizes. Positive weights divide the
// available space proportionally, after padding, spacing, and fixed cells are
// deducted. For example, []float64{20, 50, 30} gives three columns 20%, 50%, and
// 30% of that space. Weights need not sum to 100.
//
// A zero weight keeps the preferred size. Supplied weights override Stretch and
// DefaultStretch for those indices. Unspecified columns and rows use their
// stretch settings (stretched cells have weight 1). Pass nil to leave an axis
// governed entirely by its stretch settings. PreferredSize is unaffected.
func (o GridLayoutOptions) Weights(c []float64, r []float64) GridLayoutOpt {
	return func(g *GridLayout) {
		g.columnWeights = append([]float64(nil), c...)
		g.rowWeights = append([]float64(nil), r...)
	}
}

// DefaultStretch will set the stretch value to the columns/rows that are
// extra not defined on the main Stretch.
func (o GridLayoutOptions) DefaultStretch(c bool, r bool) GridLayoutOpt {
	return func(g *GridLayout) {
		g.defaultColumnStretch = c
		g.defaultRowStretch = r
	}
}

// PreferredSize implements Layouter.
func (g *GridLayout) PreferredSize(widgets []PreferredSizeLocateableWidget) (int, int) {
	colWidths, rowHeights := g.preferredColumnWidthsAndRowHeights(widgets)
	defer g.keep(colWidths, rowHeights)
	return g.padding.Dx() + g.columnSpacing*(len(colWidths)-1) + sumInts(colWidths),
		g.padding.Dy() + g.rowSpacing*(len(rowHeights)-1) + sumInts(rowHeights)
}

// Layout implements Layouter.
func (g *GridLayout) Layout(widgets []PreferredSizeLocateableWidget, rect image.Rectangle) {
	rect = g.padding.Apply(rect)

	colWidths, rowHeights := g.preferredColumnWidthsAndRowHeights(widgets)
	defer g.keep(colWidths, rowHeights)
	stretch(colWidths, rect.Dx()-g.columnSpacing*(len(colWidths)-1), g.columnWeights, g.columnStretched)
	stretch(rowHeights, rect.Dy()-g.rowSpacing*(len(rowHeights)-1), g.rowWeights, g.rowStretched)

	c, r := 0, 0
	x, y := 0, 0
	for _, w := range widgets {
		cw := colWidths[c]
		ch := rowHeights[r]
		if w.GetWidget().GetVisibility() != VisibilityHide {
			ww, wh := cw, ch
			wx, wy := x, y

			ld := w.GetWidget().LayoutData
			if gld, ok := ld.(GridLayoutData); ok {
				wx, wy, ww, wh = g.applyLayoutData(gld, wx, wy, ww, wh, x, y, cw, ch)
			}

			w.SetLocation(image.Rect(rect.Min.X+wx, rect.Min.Y+wy, rect.Min.X+wx+ww, rect.Min.Y+wy+wh))

			x += cw + g.columnSpacing
		}

		c++
		if c >= g.columns {
			c = 0
			r++
			x = 0
			y += ch + g.rowSpacing
		}
	}
}

// stretch sets the weighted sizes to share what is left of space after the
// sizes that are not, none below 0, in proportion to their weights, the
// first weighted size taking what does not divide evenly.
func stretch(sizes []int, space int, weights []float64, stretched func(int) bool) {
	first, top := -1, 0.0
	for i, size := range sizes {
		if w := cellWeight(i, weights, stretched); w > 0 {
			if first < 0 {
				first = i
			}
			top = math.Max(top, w)
		} else {
			space -= size
		}
	}
	if first < 0 {
		return
	}
	space = max(space, 0)

	// The weights over the largest, so that large finite weights don't sum
	// to infinity.
	total := 0.0
	for i := range sizes {
		total += cellWeight(i, weights, stretched) / top
	}
	left := space
	for i := range sizes {
		w := cellWeight(i, weights, stretched)
		switch {
		case w/top > 0:
			sizes[i] = int(math.Floor(float64(space) * (w / top / total)))
			left -= sizes[i]
		case w > 0: // so small beside the largest that it comes to nothing
			sizes[i] = 0
		}
	}
	sizes[first] += left
}

// cellWeight returns column or row i's weight: the one given, or 1 when it
// is stretched without one, and 0 when it keeps its preferred size.
func cellWeight(i int, weights []float64, stretched func(int) bool) float64 {
	switch {
	case i < len(weights):
		return weights[i]
	case stretched(i):
		return 1
	}
	return 0
}

func (g *GridLayout) columnStretched(c int) bool {
	if c >= len(g.columnStretch) {
		return g.defaultColumnStretch
	}
	return g.columnStretch[c]
}

func (g *GridLayout) rowStretched(r int) bool {
	if r >= len(g.rowStretch) {
		return g.defaultRowStretch
	}
	return g.rowStretch[r]
}

func (g *GridLayout) preferredColumnWidthsAndRowHeights(widgets []PreferredSizeLocateableWidget) ([]int, []int) {
	// With fewer widgets than columns, the widgets set the column count.
	colLen := min(len(widgets), g.columns)
	colWidths := take(&g.spareWidths, colLen)
	rowHeights := take(&g.spareHeights, int(math.Ceil(float64(len(widgets))/float64(g.columns))))

	c := 0
	r := 0
	for _, w := range widgets {
		ww, wh := w.PreferredSize()

		ld := w.GetWidget().LayoutData
		if gld, ok := ld.(GridLayoutData); ok {
			ww, wh = g.applyMaxSize(gld, ww, wh)
		}

		if ww > colWidths[c] {
			colWidths[c] = ww
		}

		if wh > rowHeights[r] {
			rowHeights[r] = wh
		}

		c++
		if c >= g.columns {
			c = 0
			r++
		}
	}

	return colWidths, rowHeights
}

// take returns n zeroed ints from *spare, emptying it so that a layout
// shared by nested containers doesn't measure into them twice, or new ints
// when *spare is short or held.
func take(spare *[]int, n int) []int {
	s := *spare
	*spare = nil
	if cap(s) < n {
		return make([]int, n)
	}
	s = s[:n]
	clear(s)
	return s
}

// keep keeps a measure's column widths and row heights for the next.
func (g *GridLayout) keep(colWidths, rowHeights []int) {
	g.spareWidths, g.spareHeights = colWidths, rowHeights
}

func (g *GridLayout) applyLayoutData(ld GridLayoutData, wx int, wy int, ww int, wh int, x int, y int, cw int, ch int) (int, int, int, int) {
	if ld.MaxWidth > 0 && ww > ld.MaxWidth {
		ww = ld.MaxWidth
	}

	if ld.MaxHeight > 0 && wh > ld.MaxHeight {
		wh = ld.MaxHeight
	}

	switch ld.HorizontalPosition {
	case GridLayoutPositionStart:
		// Do Nothing
	case GridLayoutPositionCenter:
		wx = x + (cw-ww)/2
	case GridLayoutPositionEnd:
		wx = x + cw - ww
	}

	switch ld.VerticalPosition {
	case GridLayoutPositionStart:
		// Do Nothing
	case GridLayoutPositionCenter:
		wy = y + (ch-wh)/2
	case GridLayoutPositionEnd:
		wy = y + ch - wh
	}

	return wx, wy, ww, wh
}

func (g *GridLayout) applyMaxSize(ld GridLayoutData, ww int, wh int) (int, int) {
	if ld.MaxWidth > 0 && ww > ld.MaxWidth {
		ww = ld.MaxWidth
	}

	if ld.MaxHeight > 0 && wh > ld.MaxHeight {
		wh = ld.MaxHeight
	}

	return ww, wh
}

func sumInts(ints []int) int {
	s := 0
	for _, i := range ints {
		s += i
	}
	return s
}
