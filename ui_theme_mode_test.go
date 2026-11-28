package ebitenui_test

import (
	"testing"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// Given a dark and a light theme, the UI follows the system's color mode:
// the dark theme while the system is dark, the light one while it is light
// or doesn't say. It reads the mode at once and then every half second, and
// the themes are the caller's own.
func TestUI_FollowsTheSystemMode(t *testing.T) {
	is := is.New(t)
	t.Cleanup(func() { widget.SetUITheme(nil) })

	dark, light := themes.GetBasicDarkTheme(), themes.GetBasicLightTheme()
	mode := ebiten.ColorModeDark
	ui := &ebitenui.UI{Container: widget.NewContainer(), DarkTheme: dark, LightTheme: light}
	ui.SetColorMode(func() ebiten.ColorMode { return mode })
	ui.Update()
	is.True(ui.PrimaryTheme == dark) // at once

	mode = ebiten.ColorModeLight
	for range ebiten.TPS()/2 - 1 {
		ui.Update()
	}
	is.True(ui.PrimaryTheme == dark) // not read again yet
	ui.Update()
	is.True(ui.PrimaryTheme == light) // half a second on

	mode = ebiten.ColorModeUnknown
	for range ebiten.TPS() / 2 {
		ui.Update()
	}
	is.True(ui.PrimaryTheme == light) // a system that doesn't say

	onlyDark := &ebitenui.UI{Container: widget.NewContainer(), DarkTheme: dark}
	onlyDark.SetColorMode(func() ebiten.ColorMode { return ebiten.ColorModeLight })
	onlyDark.Update()
	is.True(onlyDark.PrimaryTheme == dark) // the one given, whatever the mode

	own := themes.GetBasicDarkTheme()
	plain := &ebitenui.UI{Container: widget.NewContainer(), PrimaryTheme: own}
	plain.SetColorMode(func() ebiten.ColorMode { return ebiten.ColorModeLight })
	plain.Update()
	is.True(plain.PrimaryTheme == own) // neither given: the UI keeps its own
}
