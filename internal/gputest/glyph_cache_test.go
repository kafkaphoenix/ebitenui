//go:build gpu

package gputest_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/internal/gputest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// TestMain runs the tests in Ebitengine's game loop, where pixels can be
// read back.
func TestMain(m *testing.M) {
	gputest.Main(m)
}

// colorFont is ChromaCheck's COLR test font: U+E900 is a color glyph.
//
//go:embed testdata/chromacheck-colr.ttf
var colorFont []byte

// glyphFace returns Go Regular at 13 pixels.
func glyphFace(tb testing.TB) text.Face {
	tb.Helper()
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		tb.Fatal(err)
	}
	return &text.GoTextFace{Source: src, Size: 13}
}

// A text drawn from its cached glyphs has the pixels text.Draw gives it,
// at a fraction of a pixel, tinted and faded: on one line, centered on two,
// scaled, a color glyph, which keeps its colors, and one between plain
// glyphs: the first time, when it is laid out, and again from the cache;
// then without options.
func TestGlyphCache_DrawsAsTextDraw(t *testing.T) {
	face := glyphFace(t)
	src, err := text.NewGoTextFaceSource(bytes.NewReader(colorFont))
	if err != nil {
		t.Fatal(err)
	}
	colored := &text.GoTextFace{Source: src, Size: 24}
	mixed, err := text.NewMultiFace(face, colored)
	if err != nil {
		t.Fatal(err)
	}
	clr := color.NRGBA{R: 140, G: 150, B: 170, A: 255}
	var c widget.GlyphCache
	for _, tc := range []struct {
		s      string
		face   text.Face
		layout text.LayoutOptions
		scale  float64
	}{
		{"Score 12 · 1 496 m", face, text.LayoutOptions{}, 1},
		{"Level 3\nThe Old Mill", face, text.LayoutOptions{LineSpacing: 18, PrimaryAlign: text.AlignCenter}, 1},
		{"Bridge", face, text.LayoutOptions{}, 1.5},
		{"\uE900", colored, text.LayoutOptions{}, 1},
		{"Gold \uE900 Silver", mixed, text.LayoutOptions{}, 1},
	} {
		op := &text.DrawOptions{LayoutOptions: tc.layout}
		op.GeoM.Scale(tc.scale, tc.scale)
		op.GeoM.Translate(100.3, 6.7)
		op.ColorScale.ScaleWithColor(clr)
		op.ColorScale.ScaleAlpha(0.6)
		for _, pass := range []string{"laid out", "cached"} {
			drawsAsTextDraw(t, fmt.Sprintf("%q, %s", tc.s, pass), &c, tc.s, tc.face, op)
		}
	}
	drawsAsTextDraw(t, "without options", &c, "Harbor", face, nil)
}

// drawsAsTextDraw fails the test unless c draws s as text.Draw does.
func drawsAsTextDraw(t *testing.T, name string, c *widget.GlyphCache, s string, face text.Face, op *text.DrawOptions) {
	t.Helper()
	const w, h = 200, 48
	want, got := make([]byte, 4*w*h), make([]byte, 4*w*h)
	gputest.Frame(func(*ebiten.Image) {
		byText, byCache := ebiten.NewImage(w, h), ebiten.NewImage(w, h)
		defer byText.Deallocate()
		defer byCache.Deallocate()
		text.Draw(byText, s, face, op)
		c.Draw(byCache, s, face, op)
		byText.ReadPixels(want)
		byCache.ReadPixels(got)
	})
	if bytes.Equal(want, make([]byte, len(want))) {
		t.Fatalf("%s: text.Draw drew nothing", name)
	}
	for i := range want {
		if got[i] != want[i] {
			p := i / 4
			t.Fatalf("%s: pixel (%d, %d) is %v, text.Draw's %v", name, p%w, p/w, got[4*p:4*p+4], want[4*p:4*p+4])
		}
	}
}

// Drawing 50 labels a frame from the cache, against text.Draw: the CPU's
// work and the GPU's.
func BenchmarkGlyphCache_Draw(b *testing.B) {
	face := glyphFace(b)
	labels := make([]string, 50)
	for i := range labels {
		labels[i] = fmt.Sprintf("Label %d · %d m", i, 1_000+37*i)
	}
	var c widget.GlyphCache
	for _, draw := range []struct {
		name string
		draw func(dst *ebiten.Image, s string, face text.Face, op *text.DrawOptions)
	}{{"text.Draw", text.Draw}, {"GlyphCache", c.Draw}} {
		b.Run(draw.name, func(b *testing.B) {
			dst := ebiten.NewImage(640, 480)
			defer dst.Deallocate()
			op := &text.DrawOptions{}
			frame := func(*ebiten.Image) {
				dst.Clear()
				for i, s := range labels {
					op.GeoM.Reset()
					op.GeoM.Translate(float64(i%5)*120+0.3, float64(i/5)*40+0.7)
					draw.draw(dst, s, face, op)
				}
			}
			for range 30 {
				gputest.Frame(frame)
			}
			b.ReportAllocs()
			for b.Loop() {
				gputest.Frame(frame)
			}
		})
	}
}
