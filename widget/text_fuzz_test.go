package widget_test

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// Any label lays out and draws, never panics, whether its BBCode is
// processed, stripped or wrapped at a width, and measures no smaller than
// nothing.
func FuzzTextBBCode(f *testing.F) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		f.Fatal(err)
	}
	var face text.Face = &text.GoTextFace{Source: src, Size: 12}
	screen := ebiten.NewImage(120, 80)
	for _, s := range []string{
		"[color=#ff0000]red[/color] plain",
		"[link=a]there[/link]",
		"[color=#ff0000][link=x args]both[/link][/color]",
		"[color=#zz]bad color[/color]",
		"[/color][/link]",
		"[link]", "[color]", "[[[]]]", "a\nb\r\nc", "", "\n",
		"a [color=#00ff00]long line of text that wraps at the width given[/color]",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, label string) {
		for _, opts := range [][]widget.TextOpt{
			{widget.TextOpts.ProcessBBCode(true)},
			{widget.TextOpts.StripBBCode(true)},
			{widget.TextOpts.ProcessBBCode(true), widget.TextOpts.MaxWidth(60)},
		} {
			tw := widget.NewText(append(opts, widget.TextOpts.Text(label, &face, color.White))...)
			tw.Validate()
			if w, h := tw.PreferredSize(); w < 0 || h < 0 {
				t.Errorf("%q measures %d×%d", label, w, h)
			}
			tw.SetLocation(image.Rect(0, 0, 120, 80))
			tw.Render(screen)
		}
	})
}
