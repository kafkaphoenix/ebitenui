package widget_test

import (
	"image"
	"math"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/matryer/is"
)

func TestGridLayout_PositionsInCell(t *testing.T) {
	is := is.New(t)

	l := widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Spacing(10, 10), widget.GridLayoutOpts.DefaultStretch(true, true))
	at := func(h, v widget.GridLayoutPosition) *simpleWidget {
		return newSimpleWidget(10, 10, widget.GridLayoutData{MaxWidth: 10, MaxHeight: 10, HorizontalPosition: h, VerticalPosition: v})
	}
	// Cells are 50x30; the second column starts at 60, the second row at 40.
	cells := []widget.PreferredSizeLocateableWidget{
		newSimpleWidget(50, 30, nil),
		at(widget.GridLayoutPositionCenter, widget.GridLayoutPositionCenter),
		at(widget.GridLayoutPositionEnd, widget.GridLayoutPositionEnd),
		at(widget.GridLayoutPositionStart, widget.GridLayoutPositionStart),
	}

	l.Layout(cells, image.Rect(0, 0, 110, 70))

	is.Equal(cells[0].GetWidget().Rect, image.Rect(0, 0, 50, 30))
	is.Equal(cells[1].GetWidget().Rect, image.Rect(80, 10, 90, 20)) // centered
	is.Equal(cells[2].GetWidget().Rect, image.Rect(40, 60, 50, 70)) // at the end
	is.Equal(cells[3].GetWidget().Rect, image.Rect(60, 40, 70, 50)) // at the start
}

// Stretched columns and rows fill the layout; cells in a row share its
// height, and cells in a column its width.
func TestGridLayout_StretchedCellsFill(t *testing.T) {
	is := is.New(t)

	l := widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.DefaultStretch(true, true))
	cells := make([]widget.PreferredSizeLocateableWidget, 9)
	for i := range cells {
		cells[i] = newSimpleWidget(10, 10, nil)
	}

	l.Layout(cells, image.Rect(0, 0, 100, 100))

	// 100 does not divide by 3: the first column and row take the extra pixel.
	edges := []int{0, 34, 67, 100}
	for i, w := range cells {
		c, r := i%3, i/3
		is.Equal(w.GetWidget().Rect, image.Rect(edges[c], edges[r], edges[c+1], edges[r+1]))
	}
}

// A row ends where its cells do, though its last widget is hidden.
func TestGridLayout_RowAfterHiddenWidget(t *testing.T) {
	is := is.New(t)

	l := widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.DefaultStretch(false, true))
	cells := make([]widget.PreferredSizeLocateableWidget, 4)
	for i := range cells {
		cells[i] = newSimpleWidget(10, 10, nil)
	}
	cells[1].GetWidget().SetVisibility(widget.VisibilityHide)

	l.Layout(cells, image.Rect(0, 0, 100, 100))

	is.Equal(cells[0].GetWidget().Rect, image.Rect(0, 0, 10, 50))
	is.Equal(cells[2].GetWidget().Rect, image.Rect(0, 50, 10, 100))
}

// Stretched columns share what the others leave, the first taking the
// extra pixel.
func TestGridLayout_StretchedColumnsShareTheRest(t *testing.T) {
	is := is.New(t)

	l := widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Spacing(5, 0), widget.GridLayoutOpts.Stretch([]bool{false, true, true}, nil))
	cells := []widget.PreferredSizeLocateableWidget{
		newSimpleWidget(20, 10, nil),
		newSimpleWidget(10, 10, nil),
		newSimpleWidget(10, 10, nil),
	}

	l.Layout(cells, image.Rect(0, 0, 101, 50))

	is.Equal(cells[0].GetWidget().Rect, image.Rect(0, 0, 20, 10))
	is.Equal(cells[1].GetWidget().Rect, image.Rect(25, 0, 61, 10))
	is.Equal(cells[2].GetWidget().Rect, image.Rect(66, 0, 101, 10))
}

// A grid layout measures and lays out without allocating once it has
// measured, and a layout shared by nested containers measures each right.
func TestGridLayout_MeasuresWithoutAllocating(t *testing.T) {
	is := is.New(t)

	l := widget.NewGridLayout(widget.GridLayoutOpts.Columns(2))
	cells := []widget.PreferredSizeLocateableWidget{
		newSimpleWidget(20, 10, nil),
		newSimpleWidget(10, 10, nil),
		newSimpleWidget(10, 30, nil),
	}
	w, h := l.PreferredSize(cells)
	is.Equal(w, 30)
	is.Equal(h, 40)
	is.Equal(testing.AllocsPerRun(10, func() { l.PreferredSize(cells) }), 0.0)
	is.Equal(testing.AllocsPerRun(10, func() { l.Layout(cells, image.Rect(0, 0, 30, 40)) }), 0.0)

	inner := widget.NewContainer(widget.ContainerOpts.Layout(l))
	inner.AddChild(newSimpleWidget(10, 10, nil), newSimpleWidget(10, 10, nil))
	outer := widget.NewContainer(widget.ContainerOpts.Layout(l))
	outer.AddChild(newSimpleWidget(40, 10, nil), inner)
	w, h = outer.PreferredSize()
	is.Equal(w, 60) // 40, and the inner container's 20 measured with the same layout
	is.Equal(h, 10)
}

// Weighted columns and rows share the space left after padding and spacing
// in proportion to their weights, also after a resize, wherever the
// container is.
func TestGridLayout_WeightsResize(t *testing.T) {
	is := is.New(t)
	g := widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
		widget.GridLayoutOpts.Weights([]float64{20, 50, 30}, []float64{1, 3}),
		widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		widget.GridLayoutOpts.Spacing(5, 10))
	widgets := make([]widget.PreferredSizeLocateableWidget, 6)
	for i := range widgets {
		widgets[i] = newSimpleWidget(10, 10, nil)
	}
	// Both rows share column widths, even after resizing. Padding, spacing,
	// and a non-zero container origin must all be respected.
	for _, size := range []int{100, 200} {
		rect := image.Rect(30, 40, 30+size+30, 40+size+30)
		g.Layout(widgets, rect)
		widths := []int{size * 20 / 100, size * 50 / 100, size * 30 / 100}
		heights := []int{size / 4, size * 3 / 4}
		y := 50
		for r, height := range heights {
			x := 40
			for c, width := range widths {
				is.Equal(widgets[r*3+c].GetWidget().Rect, image.Rect(x, y, x+width, y+height))
				x += width + 5
			}
			y += height + 10
		}
	}
	w, h := g.PreferredSize(widgets)
	is.Equal(w, 60)
	is.Equal(h, 50)
}

// How three columns of 10 pixels, 5 apart, share a width by their weights
// and their stretch: rounding left to the first weighted one, a zero
// weight keeping the preferred size, the stretch for columns without a
// weight, and nothing below 0.
func TestGridLayout_WeightedCellSizes(t *testing.T) {
	tests := []struct {
		name           string
		weights        []float64
		stretch        []bool
		defaultStretch bool
		available      int
		want           []int
	}{
		{"relative fractions", []float64{.2, .5, .3}, nil, false, 110, []int{20, 50, 30}},
		{"rounding", []float64{2, 5, 3}, nil, false, 111, []int{21, 50, 30}},
		{"fixed cell", []float64{0, 1, 3}, nil, true, 110, []int{10, 23, 67}},
		{"partial weights", []float64{3}, nil, true, 110, []int{60, 20, 20}},
		{"zero overrides stretch", []float64{0, 0, 0}, nil, true, 110, []int{10, 10, 10}},
		{"legacy stretch", nil, []bool{true, false, true}, false, 111, []int{46, 10, 45}},
		{"default stretch", nil, nil, true, 110, []int{34, 33, 33}},
		{"undersized container", []float64{0, 1, 3}, nil, false, 15, []int{10, 0, 0}},
		{"large weights", []float64{math.MaxFloat64, math.MaxFloat64, 0}, nil, false, 110, []int{45, 45, 10}},
		{"a weight too small to count", []float64{math.SmallestNonzeroFloat64, math.MaxFloat64, 0}, nil, false, 110, []int{0, 90, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			g := widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
				widget.GridLayoutOpts.Weights(tt.weights, nil),
				widget.GridLayoutOpts.Stretch(tt.stretch, nil),
				widget.GridLayoutOpts.Spacing(5, 0),
				widget.GridLayoutOpts.DefaultStretch(tt.defaultStretch, false))
			widgets := []widget.PreferredSizeLocateableWidget{newSimpleWidget(10, 10, nil), newSimpleWidget(10, 10, nil), newSimpleWidget(10, 10, nil)}
			g.Layout(widgets, image.Rect(0, 0, tt.available, 10))
			got := make([]int, len(widgets))
			for i, w := range widgets {
				got[i] = w.GetWidget().Rect.Dx()
			}
			is.Equal(got, tt.want)
		})
	}
}

// A weighted cell's widget keeps to its maximum size and its position in
// the cell.
func TestGridLayout_WeightsMaxSizeAndPosition(t *testing.T) {
	is := is.New(t)
	g := widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Weights([]float64{1}, []float64{1}))
	w := newSimpleWidget(10, 10, widget.GridLayoutData{
		MaxWidth: 20, MaxHeight: 30,
		HorizontalPosition: widget.GridLayoutPositionCenter,
		VerticalPosition:   widget.GridLayoutPositionEnd,
	})
	g.Layout([]widget.PreferredSizeLocateableWidget{w}, image.Rect(10, 20, 110, 120))
	is.Equal(w.GetWidget().Rect, image.Rect(50, 90, 70, 120))
}

// A negative, not-a-number or infinite weight is refused as the layout is
// made.
func TestGridLayout_InvalidWeights(t *testing.T) {
	for _, weight := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, row := range []bool{false, true} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("expected panic for weight %v (row=%v)", weight, row)
					}
				}()
				c, r := []float64{weight}, []float64(nil)
				if row {
					c, r = r, c
				}
				widget.NewGridLayout(widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Weights(c, r))
			}()
		}
	}
}

// A weighted layout of no widgets lays out nothing, and doesn't panic.
func TestGridLayout_EmptyWeightedLayout(_ *testing.T) {
	g := widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Weights([]float64{20, 50, 30}, nil))
	g.Layout(nil, image.Rect(0, 0, 100, 100))
}

// Weighted layouts allocate nothing either.
func TestGridLayout_WeightsWithoutAllocating(t *testing.T) {
	is := is.New(t)
	g := widget.NewGridLayout(widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Weights([]float64{20, 50, 30}, []float64{1}))
	cells := []widget.PreferredSizeLocateableWidget{newSimpleWidget(10, 10, nil), newSimpleWidget(10, 10, nil), newSimpleWidget(10, 10, nil)}
	g.Layout(cells, image.Rect(0, 0, 100, 40))
	is.Equal(testing.AllocsPerRun(10, func() { g.Layout(cells, image.Rect(0, 0, 100, 40)) }), 0.0)
}

// A grid of no columns, or fewer, is refused as it is made.
func TestGridLayout_ColumnsMustBePositive(t *testing.T) {
	for _, columns := range []int{0, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for %d columns", columns)
				}
			}()
			widget.NewGridLayout(widget.GridLayoutOpts.Columns(columns))
		}()
	}
}

// The layout keeps its own copy of the weights: changing the slices given
// changes nothing.
func TestGridLayout_WeightsAreCopied(t *testing.T) {
	is := is.New(t)
	c, r := []float64{1, 3}, []float64{1}
	g := widget.NewGridLayout(widget.GridLayoutOpts.Columns(2), widget.GridLayoutOpts.Weights(c, r))
	c[0], c[1], r[0] = 3, 1, 0
	cells := []widget.PreferredSizeLocateableWidget{newSimpleWidget(10, 10, nil), newSimpleWidget(10, 10, nil)}
	g.Layout(cells, image.Rect(0, 0, 100, 50))
	is.Equal(cells[0].GetWidget().Rect, image.Rect(0, 0, 25, 50))
}
