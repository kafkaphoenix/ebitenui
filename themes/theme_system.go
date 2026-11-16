package themes

import (
	"sync"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// The basic themes, each made once, the first time it is asked for.
var (
	basicDarkTheme  = sync.OnceValue(GetBasicDarkTheme)
	basicLightTheme = sync.OnceValue(GetBasicLightTheme)
)

// The shaped basic themes made so far, by mode and shape.
var (
	shapedMu     sync.Mutex
	shapedThemes = map[shapedKey]*widget.Theme{}
)

type shapedKey struct {
	mode ebiten.ColorMode
	s    shape
}

// GetBasicSystemTheme returns the basic theme matching the system's color
// mode: the dark one when the system is dark, the light one when it is
// light or doesn't say. It returns the same theme for the same mode, so
// setting a UI's PrimaryTheme to it every frame makes the UI follow the
// system and re-validate only when the mode changes.
func GetBasicSystemTheme() *widget.Theme {
	return basicThemeFor(ebiten.SystemColorMode())
}

// GetBasicSystemThemeWith returns GetBasicSystemTheme's theme shaped by
// opts (GetBasicDarkThemeWith), the same theme for the same mode and
// options. Called every frame, it allocates nothing when given the same
// slice of options each time.
func GetBasicSystemThemeWith(opts ...ThemeOpt) *widget.Theme {
	return shapedThemeFor(ebiten.SystemColorMode(), newShape(opts))
}

// shapedThemeFor returns the basic theme for a color mode shaped by s,
// made once.
func shapedThemeFor(mode ebiten.ColorMode, s shape) *widget.Theme {
	if s == (shape{}) {
		return basicThemeFor(mode)
	}
	shapedMu.Lock()
	defer shapedMu.Unlock()
	k := shapedKey{mode, s}
	if th, ok := shapedThemes[k]; ok {
		return th
	}
	th := newBasicLightTheme(s)
	if mode == ebiten.ColorModeDark {
		th = newBasicDarkTheme(s)
	}
	shapedThemes[k] = th
	return th
}

// basicThemeFor returns the basic theme for a color mode.
func basicThemeFor(mode ebiten.ColorMode) *widget.Theme {
	if mode == ebiten.ColorModeDark {
		return basicDarkTheme()
	}
	return basicLightTheme()
}
