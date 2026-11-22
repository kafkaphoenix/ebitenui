package widget_test

import (
	"image"
	"strconv"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/matryer/is"
)

func TestAnchorLayout_PreferredSize(t *testing.T) {
	is := is.New(t)

	padding := &widget.Insets{
		Top:    10,
		Left:   20,
		Right:  30,
		Bottom: 40,
	}

	l := newAnchorLayout(t, widget.AnchorLayoutOpts.Padding(padding))

	wi := newSimpleWidget(35, 45, nil)

	w, h := l.PreferredSize([]widget.PreferredSizeLocateableWidget{wi})

	is.Equal(w, wi.preferredWidth+padding.Dx())
	is.Equal(h, wi.preferredHeight+padding.Dy())
}

func TestAnchorLayout_Layout(t *testing.T) {
	ww, wh := 25, 35
	wrect := image.Rect(0, 0, ww, wh)
	rect := image.Rect(45, 55, 200, 200)
	padding := &widget.Insets{
		Top:    10,
		Left:   20,
		Right:  30,
		Bottom: 40,
	}

	prect := padding.Apply(rect)

	tests := []struct {
		ld       widget.AnchorLayoutData
		expected image.Rectangle
	}{
		{
			widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			},
			prect,
		},
		{
			widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			},
			wrect.Add(prect.Min).Add(image.Point{prect.Dx(), 0}).Sub(image.Point{ww, 0}),
		},
		{
			widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			},
			wrect.Add(prect.Min).Add(image.Point{(prect.Dx() - ww) / 2, (prect.Dy() - wh) / 2}),
		},
		{
			widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			},
			wrect.Add(prect.Min).Add(image.Point{0, prect.Dy()}).Sub(image.Point{0, wh}),
		},
		{
			widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				StretchVertical:    true,
			},
			image.Rect(prect.Min.X+(prect.Dx()-ww)/2, prect.Min.Y, prect.Min.X+(prect.Dx()-ww)/2+ww, prect.Max.Y),
		},
		{
			widget.AnchorLayoutData{
				StretchHorizontal: true,
				VerticalPosition:  widget.AnchorLayoutPositionCenter,
			},
			image.Rect(prect.Min.X, prect.Min.Y+(prect.Dy()-wh)/2, prect.Max.X, prect.Min.Y+(prect.Dy()-wh)/2+wh),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			is := is.New(t)

			l := newAnchorLayout(t, widget.AnchorLayoutOpts.Padding(padding))

			w := newSimpleWidget(ww, wh, test.ld)
			l.Layout([]widget.PreferredSizeLocateableWidget{w}, rect)

			is.Equal(w.GetWidget().Rect, test.expected)
		})
	}
}

func newAnchorLayout(t *testing.T, opts ...widget.AnchorLayoutOpt) widget.Layouter {
	t.Helper()
	l := widget.NewAnchorLayout(opts...)
	return l
}
