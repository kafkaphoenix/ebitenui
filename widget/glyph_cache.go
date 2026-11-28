package widget

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// defaultGlyphCacheLimit is how many texts a GlyphCache keeps with no
// Limit set.
const defaultGlyphCacheLimit = 256

// GlyphCache draws and measures text as text.Draw and text.Measure do,
// from glyphs laid out the first time each text is asked for in its face
// and layout: drawing it again is a DrawImage for each of its glyphs, with
// no shaping and nothing allocated. It is for text drawn outside widgets,
// as on a Canvas, where the same labels are drawn frame after frame. The
// glyphs' images stay valid while the cache holds them, even after
// text/v2's own cache lets them go.
//
// The zero value is ready to use. A GlyphCache is not safe for concurrent
// use.
type GlyphCache struct {
	// Limit is how many texts the cache keeps laid out before it starts
	// over; 0 keeps 256. A text that changes as the game runs, such as a
	// readout, adds one for each new value.
	Limit int

	texts map[glyphKey]laidOutText
	op    ebiten.DrawImageOptions // reused for every glyph
}

// glyphKey is what a text's glyphs depend on.
type glyphKey struct {
	face   text.Face
	s      string
	layout text.LayoutOptions
}

// laidOutText is a text's glyphs that have an image, placed from its
// origin, and its size.
type laidOutText struct {
	glyphs        []text.Glyph
	width, height float64
}

// text returns s laid out in face, laying it out if it is new.
func (c *GlyphCache) text(s string, face text.Face, layout text.LayoutOptions) laidOutText {
	key := glyphKey{face: face, s: s, layout: layout}
	if t, ok := c.texts[key]; ok {
		return t
	}
	limit := c.Limit
	if limit <= 0 {
		limit = defaultGlyphCacheLimit
	}
	if c.texts == nil {
		c.texts = make(map[glyphKey]laidOutText)
	}
	if len(c.texts) >= limit {
		clear(c.texts)
	}
	glyphs := slices.DeleteFunc(text.AppendGlyphs(nil, s, face, &layout), func(g text.Glyph) bool { return g.Image == nil })
	w, h := text.Measure(s, face, layout.LineSpacing)
	t := laidOutText{glyphs: slices.Clip(glyphs), width: w, height: h}
	c.texts[key] = t
	return t
}

// Draw draws s in face as text.Draw does, with options's layout, GeoM,
// color scale and the rest; nil options draw at the origin, untinted.
// Color glyphs, such as emoji, keep their own colors and take only the
// color scale's alpha.
func (c *GlyphCache) Draw(dst *ebiten.Image, s string, face text.Face, options *text.DrawOptions) {
	var layout text.LayoutOptions
	op := &c.op
	*op = ebiten.DrawImageOptions{}
	if options != nil {
		layout = options.LayoutOptions
		*op = options.DrawImageOptions
	}
	geoM, tint := op.GeoM, op.ColorScale
	var colored ebiten.ColorScale
	colored.ScaleAlpha(tint.A())
	for _, g := range c.text(s, face, layout).glyphs {
		op.GeoM.Reset()
		op.GeoM.Translate(g.X, g.Y)
		op.GeoM.Concat(geoM)
		op.ColorScale = tint
		if g.Colored {
			op.ColorScale = colored
		}
		dst.DrawImage(g.Image, op)
	}
}

// Measure returns the width and height of s in face, its lines lineSpacing
// pixels apart, as text.Measure does.
func (c *GlyphCache) Measure(s string, face text.Face, lineSpacing float64) (width, height float64) {
	t := c.text(s, face, text.LayoutOptions{LineSpacing: lineSpacing})
	return t.width, t.height
}
