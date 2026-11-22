package widget_test

import (
	"image"
	"image/color"
	"image/gif"
	"slices"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
)

// A GIF's frames follow their delays, in hundredths of a second, counted
// in updates: 50 ms is 3 at 60 a second and 100 ms is 6; after the last
// frame the first comes again.
func TestGraphic_GIFFramesInUpdates(t *testing.T) {
	frame := func() *image.Paletted {
		return image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	}
	g := newGraphic(t, widget.GraphicOpts.GIF(&gif.GIF{
		Image: []*image.Paletted{frame(), frame(), frame()},
		Delay: []int{5, 10, 5},
	}))
	shown := make([]int, 0, 16)
	for range 16 {
		g.Update(&widget.UpdateObject{})
		shown = append(shown, g.GifCurrentImage())
	}
	want := []int{0, 0, 1, 1, 1, 1, 1, 1, 2, 2, 2, 0, 0, 0, 1, 1}
	if !slices.Equal(shown, want) {
		t.Errorf("frames shown update by update %v, want %v", shown, want)
	}
}
