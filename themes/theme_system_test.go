package themes_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/themes"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// Each color mode gets one theme, the same every time: the dark one for a
// dark system, the light one for a light system or one that doesn't say.
func TestBasicThemeFor(t *testing.T) {
	is := is.New(t)

	dark, light := themes.BasicThemeFor(ebiten.ColorModeDark), themes.BasicThemeFor(ebiten.ColorModeLight)
	is.True(dark != light)
	is.Equal(themes.BasicThemeFor(ebiten.ColorModeDark), dark)     // made once
	is.Equal(themes.BasicThemeFor(ebiten.ColorModeLight), light)   // made once
	is.Equal(themes.BasicThemeFor(ebiten.ColorModeUnknown), light) // light when the system doesn't say
	is.Equal(themes.GetBasicSystemTheme(), themes.BasicThemeFor(ebiten.SystemColorMode()))
}
