package widget_test

import (
	"bytes"
	img "image"
	"image/color"
	"runtime"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/matryer/is"
)

// cachedText returns a text drawn from an image of itself, 200 by 100 px,
// drawn once.
func cachedText(t *testing.T, label string, opts ...widget.TextOpt) *widget.Text {
	t.Helper()
	widget.SetTextCache(true)
	t.Cleanup(func() { widget.SetTextCache(false) })
	txt := widget.NewText(append([]widget.TextOpt{widget.TextOpts.Text(label, loadFont(t), color.White)}, opts...)...)
	txt.SetLocation(img.Rect(0, 0, 200, 100))
	txt.Validate()
	txt.Render(ebiten.NewImage(200, 100))
	return txt
}

// A cached text draws its glyphs into its image once, and again only when
// what it shows changes; moving it moves the image.
func TestText_CacheDrawsOnceUntilChanged(t *testing.T) {
	is := is.New(t)

	txt := cachedText(t, "Hello\nworld", widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart))
	screen := ebiten.NewImage(200, 100)
	is.Equal(txt.Cache().Renders(), 1)
	txt.Render(screen)
	txt.Render(screen)
	is.Equal(txt.Cache().Renders(), 1) // drawn from the image

	for i, change := range []func(){
		func() { txt.Label = "Hello" },
		func() { txt.SetColor(color.Black) },
		func() { txt.SetShadow(&widget.TextShadow{Color: color.Black, OffsetX: 1, OffsetY: 1}) },
		func() { txt.SetLocation(img.Rect(0, 0, 150, 100)) }, // centered lines move
		func() { txt.SetPadding(&widget.Insets{Left: 4}) },
		func() { txt.SetFace(loadFont(t)) },
		func() { txt.SetPosition(&widget.TextPositioning{HTextPosition: widget.TextPositionEnd}) },
	} {
		change()
		txt.Render(screen)
		is.Equal(txt.Cache().Renders(), i+2)
	}

	txt.SetLocation(img.Rect(10, 10, 160, 110)) // moved, same width
	txt.Render(screen)
	is.Equal(txt.Cache().Renders(), 8) // the image moves with it
}

// A text aligned at its start doesn't depend on its width; one showing a
// hovered link draws again in the link's hover color.
func TestText_CacheWidthAndLinks(t *testing.T) {
	is := is.New(t)

	txt := cachedText(t, "A [link=x]link[/link]", widget.TextOpts.ProcessBBCode(true),
		widget.TextOpts.LinkColor(&widget.TextLinkColor{Idle: color.White, Hover: color.Black}))
	screen := ebiten.NewImage(200, 100)
	txt.SetLocation(img.Rect(0, 0, 150, 100))
	txt.Render(screen)
	is.Equal(txt.Cache().Renders(), 1) // same lines at any width

	txt.SetPreviousLink(txt.Measurements().ProcessedLines()[0][1]) // hovered, as handleLinkEvents leaves it
	txt.Render(screen)
	is.Equal(txt.Cache().Renders(), 2)

	txt.ComputedParams().LinkColor = &widget.TextLinkColor{Idle: color.Black, Hover: color.White} // as a new theme gives
	txt.Render(screen)
	is.Equal(txt.Cache().Renders(), 3)
}

// A text keeps its slot while its box fits and fills more than half of it;
// otherwise it moves to another slot, freeing the old one, and an empty text
// gives its slot back.
func TestText_CacheKeepsItsSlot(t *testing.T) {
	is := is.New(t)
	widget.EmptyTextPages(t)

	txt := cachedText(t, "Hello")
	screen := ebiten.NewImage(200, 100)
	slot := txt.Cache().Slot()
	is.True(slot != nil)

	txt.Label = "Hell"
	txt.Render(screen)
	is.True(txt.Cache().Slot() == slot) // still fits

	txt.Label = "Hello, world"
	txt.Render(screen)
	is.True(txt.Cache().Slot() != slot) // outgrown
	is.Equal(slot.Page().Used(), 1)     // the old slot freed

	txt.Label = "H"
	txt.Render(screen)
	is.True(txt.Cache().Slot().Rect().Dx() < 40) // a slot its size again

	txt.Label = ""
	txt.Render(screen)
	is.True(txt.Cache().Slot() == nil && txt.Cache().Image() == nil) // nothing to keep
	is.Equal(widget.TextPageUsed(0), 0)
}

// Cached texts share a page; a text too tall for a page's slots gets a
// texture of its own, and back on a page when it fits again.
func TestText_CacheSharesPages(t *testing.T) {
	is := is.New(t)
	widget.EmptyTextPages(t)

	a := cachedText(t, "Hello")
	b := cachedText(t, "World")
	is.True(a.Cache().Slot().Page() == b.Cache().Slot().Page())
	is.Equal(widget.TextPageCount(), 1)

	tall := cachedText(t, strings.Repeat("line\n", 15)) // over 256 px, under 2048
	is.True(tall.Cache().Slot() == nil)
	is.True(tall.Cache().Own() != nil && tall.Cache().Image() == tall.Cache().Own())

	tall.Label = "short"
	tall.Render(ebiten.NewImage(200, 100))
	is.True(tall.Cache().Slot() != nil && tall.Cache().Own() == nil)
}

// A text the collector drops gives its slot back, freed by the first draw of
// a cached text after the collection, and not before it; a slot it gave up
// before is not freed again.
func TestText_CacheSlotFreedWhenCollected(t *testing.T) {
	is := is.New(t)
	widget.EmptyTextPages(t)

	kept := cachedText(t, "Kept")
	screen := ebiten.NewImage(200, 100)
	func() {
		txt := cachedText(t, "Hello")
		txt.Label = "Hello, world" // outgrows its first slot and frees it now
		txt.Render(screen)
	}()
	page := kept.Cache().Slot().Page()
	is.Equal(page.Used(), 2)
	kept.Render(screen)
	is.Equal(page.Used(), 2) // no collection yet
	runtime.GC()
	kept.Render(screen)
	is.Equal(page.Used(), 1) // the kept text's slot alone; each other slot freed once
}

// A text taller or wider than textCacheMaxSide draws glyph by glyph, without
// trying again each frame; an empty one has nothing to keep.
func TestText_CacheLimits(t *testing.T) {
	is := is.New(t)

	tall := cachedText(t, strings.Repeat("line\n", 200))
	tall.Render(ebiten.NewImage(200, 100))
	is.True(tall.Cache().Image() == nil)
	is.True(tall.Cache().TooLarge())

	empty := cachedText(t, "")
	is.True(empty.Cache().Image() == nil)
}

// Without SetTextCache, texts keep no image.
func TestText_CacheOff(t *testing.T) {
	is := is.New(t)

	txt := widget.NewText(widget.TextOpts.Text("Hello", loadFont(t), color.White))
	txt.SetLocation(img.Rect(0, 0, 200, 100))
	render(txt, t)
	is.True(txt.Cache() == nil)
}

// Drawn from its image, a text allocates no more than one draw does (before
// the game runs, Ebitengine allocates as it queues each draw).
func TestText_CacheDrawsWithoutAllocating(t *testing.T) {
	is := is.New(t)
	if raceEnabled {
		t.Skip("Ebitengine's allocations per draw vary under the race detector")
	}

	txt := cachedText(t, "Hello\nworld", widget.TextOpts.Shadow(color.Black, 1, 1))
	screen := ebiten.NewImage(200, 100)
	op := &ebiten.DrawImageOptions{}
	oneDraw := testing.AllocsPerRun(10, func() { screen.DrawImage(txt.Cache().Image(), op) })
	is.Equal(testing.AllocsPerRun(10, func() { txt.Render(screen) }), oneDraw)
}

func BenchmarkText_Draw(b *testing.B) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		b.Fatal(err)
	}
	var face text.Face = &text.GoTextFace{Source: src, Size: 20}
	label := strings.Repeat(strings.Repeat("Lorem ipsum dolor sit amet. ", 2)+"\n", 20) // 1 100 glyphs
	for _, cache := range []bool{false, true} {
		name := "glyphs"
		if cache {
			name = "image"
		}
		b.Run(name, func(b *testing.B) {
			widget.SetTextCache(cache)
			defer widget.SetTextCache(false)
			txt := widget.NewText(widget.TextOpts.Text(label, &face, color.White))
			txt.SetLocation(img.Rect(0, 0, 800, 600))
			txt.Validate()
			screen := ebiten.NewImage(800, 600)
			txt.Render(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				txt.Render(screen)
			}
		})
	}
}
