package themes_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/themes"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// The basic themes share one parsed Go Regular: a parsed font is most of
// what a theme keeps alive (0.38 of 0.48 MB), and its glyph caches serve
// every theme.
func TestThemesShareTheirFont(t *testing.T) {
	source := func(th *widget.Theme) *text.GoTextFaceSource {
		t.Helper()
		f, ok := (*th.DefaultFace).(*text.GoTextFace)
		if !ok {
			t.Fatalf("the theme's face is a %T, want a *text.GoTextFace", *th.DefaultFace)
		}
		return f.Source
	}
	dark, light := source(themes.GetBasicDarkTheme()), source(themes.GetBasicLightTheme())
	if dark != light || source(themes.GetBasicDarkThemeWith(themes.Rounded(4))) != dark {
		t.Error("the basic themes parse Go Regular each, want one shared source")
	}
}
