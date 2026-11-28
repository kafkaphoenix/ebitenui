package widget_test

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/matryer/is"
	"golang.org/x/image/font/gofont/goregular"
)

// glyphFace returns Go Regular at 13 pixels.
func glyphFace(tb testing.TB) *text.GoTextFace {
	tb.Helper()
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		tb.Fatal(err)
	}
	return &text.GoTextFace{Source: src, Size: 13}
}

// A text's size is text.Measure's, the first time and again from the
// cache: one line, lines spaced apart, and none, and the same line in a
// larger face.
func TestGlyphCache_Measure(t *testing.T) {
	is := is.New(t)
	face := glyphFace(t)
	large := &text.GoTextFace{Source: face.Source, Size: 20}
	var c widget.GlyphCache
	for _, tc := range []struct {
		s       string
		face    text.Face
		spacing float64
	}{{"Score 12 · 1 496 m", face, 0}, {"two\nlines", face, 18}, {"", face, 0}, {"Score 12 · 1 496 m", large, 0}} {
		w, h := text.Measure(tc.s, tc.face, tc.spacing)
		for range 2 {
			gw, gh := c.Measure(tc.s, tc.face, tc.spacing)
			is.Equal(gw, w) // width
			is.Equal(gh, h) // height
		}
	}
}

// A text is laid out as text.AppendGlyphs lays it out, without its glyphs
// that have no image, in each layout: aligned, spaced, or neither.
func TestGlyphCache_LaysOutAsAppendGlyphs(t *testing.T) {
	is := is.New(t)
	face := glyphFace(t)
	var c widget.GlyphCache
	const s = "Level 3\nThe Old Mill"
	for _, layout := range []text.LayoutOptions{
		{},
		{LineSpacing: 20},
		{LineSpacing: 20, PrimaryAlign: text.AlignCenter, SecondaryAlign: text.AlignEnd},
	} {
		want := slices.DeleteFunc(text.AppendGlyphs(nil, s, face, &layout), func(g text.Glyph) bool { return g.Image == nil })
		is.True(len(want) > 0)
		for range 2 {
			is.True(slices.Equal(c.Glyphs(s, face, layout), want)) // the glyphs
		}
	}
	is.Equal(c.Len(), 3) // one text a layout
}

// The cache keeps up to its limit of texts, by default 256, then starts
// over; a text asked for again is not laid out again.
func TestGlyphCache_Limit(t *testing.T) {
	is := is.New(t)
	face := glyphFace(t)
	for _, tc := range []struct{ limit, keeps int }{{0, 256}, {3, 3}} {
		c := widget.GlyphCache{Limit: tc.limit}
		for i := range tc.keeps {
			c.Measure(fmt.Sprint(i), face, 0)
			c.Measure("0", face, 0)
		}
		is.Equal(c.Len(), tc.keeps) // full
		c.Measure("one more", face, 0)
		is.Equal(c.Len(), 1) // started over
	}
}

// Drawing a text the cache holds lays out nothing: it allocates what
// drawing each of its glyphs does, no more.
func TestGlyphCache_DrawFromTheCache(t *testing.T) {
	is := is.New(t)
	if raceEnabled {
		t.Skip("Ebitengine's allocations per draw vary under the race detector")
	}
	face := glyphFace(t)
	const s = "Score 12 · 1 496 m"
	var c widget.GlyphCache
	dst := ebiten.NewImage(200, 32)
	op := &text.DrawOptions{}
	op.GeoM.Translate(4.3, 6.7)
	c.Draw(dst, s, face, op)
	glyphs := c.Glyphs(s, face, text.LayoutOptions{})
	one := &ebiten.DrawImageOptions{}
	oneDraw := testing.AllocsPerRun(10, func() { dst.DrawImage(glyphs[0].Image, one) })
	is.Equal(testing.AllocsPerRun(10, func() { c.Draw(dst, s, face, op) }), float64(len(glyphs))*oneDraw)
	is.Equal(c.Len(), 1) // laid out once
}
