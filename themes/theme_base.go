package themes

import (
	"bytes"
	"fmt"
	"image/color"
	"sync"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// goRegular is Go Regular, which the library bundles, parsed once for
// every theme: a parsed font is most of what a theme keeps alive, and its
// glyph caches serve them all. It panics if the font doesn't parse, as a
// bundled font always does.
var goRegular = sync.OnceValue(func() *text.GoTextFaceSource {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(fmt.Sprintf("themes: the bundled Go Regular doesn't parse: %v", err))
	}
	return s
})

// loadFont returns Go Regular at a size.
func loadFont(size float64) text.Face {
	return &text.GoTextFace{
		Source: goRegular(),
		Size:   size,
	}
}

// ThemeOpt shapes the boxes a basic theme draws: buttons, inputs, panels,
// tabs, lists and their entries, sliders, check boxes and the rest.
type ThemeOpt func(s *shape)

// Rounded rounds the boxes' corners to radius pixels, and makes list
// entries' backgrounds and text selections rounded boxes too. Small parts,
// sliders' handles and tracks, progress bars and text selections, round by
// smallRadius at most, check boxes by checkboxRadius, so the corners fit.
func Rounded(radius int) ThemeOpt {
	return func(s *shape) { s.radius = max(0, radius) }
}

// Borders draws every border width pixels wide in place of the theme's own
// widths; 0 draws none. Boxes the theme fills without a border stay so.
func Borders(width int) ThemeOpt {
	return func(s *shape) { s.border, s.bordered = max(0, width), true }
}

// The largest radius of small parts, and of check boxes, which are 32 px.
const (
	smallRadius    = 3
	checkboxRadius = 10
)

// shape is what the options set; its zero value keeps a theme as it was.
type shape struct {
	radius   int
	border   int
	bordered bool // border replaces the theme's widths
}

func newShape(opts []ThemeOpt) shape {
	var s shape
	for _, o := range opts {
		o(&s)
	}
	return s
}

// upTo returns s with its radius n at most.
func (s shape) upTo(n int) shape {
	s.radius = min(s.radius, n)
	return s
}

// box returns a box filled with fill and bordered with edge, width pixels
// wide unless Borders set another width.
func (s shape) box(fill, edge color.Color, width int) *image.NineSlice {
	if s.bordered {
		width = s.border
	}
	switch {
	case s.radius > 0:
		return image.NewNineSliceRounded(fill, s.radius, edge, width)
	case width > 0:
		return image.NewBorderedNineSliceColor(fill, edge, width)
	}
	return image.NewNineSliceColor(fill)
}

// fill returns a box filled with c, without a border.
func (s shape) fill(c color.Color) *image.NineSlice {
	if s.radius > 0 {
		return image.NewNineSliceRounded(c, s.radius, nil, 0)
	}
	return image.NewNineSliceColor(c)
}

// entries returns rounded backgrounds of c's for list entries, or nil when
// s is square, so that the colors draw them.
func (s shape) entries(c *widget.ListEntryColor) *widget.ListEntryImage {
	if s.radius == 0 {
		return nil
	}
	return entryImages(c, s.fill)
}

// entryImages returns list entries' backgrounds, each fill of its color in c.
func entryImages(c *widget.ListEntryColor, fill func(color.Color) *image.NineSlice) *widget.ListEntryImage {
	return &widget.ListEntryImage{
		Focused:          fill(c.FocusedBackground),
		Selecting:        fill(c.SelectingBackground),
		Selected:         fill(c.SelectedBackground),
		DisabledSelected: fill(c.DisabledSelectedBackground),
		SelectedFocused:  fill(c.SelectedFocusedBackground),
		SelectingFocused: fill(c.SelectingFocusedBackground),
	}
}

// highlight returns a rounded text selection in the inputs' own color, or
// nil when s is square, so that the inputs draw their own.
func (s shape) highlight() *image.NineSlice {
	if s.radius == 0 {
		return nil
	}
	return s.upTo(smallRadius).fill(color.NRGBA{6, 67, 161, 100})
}
