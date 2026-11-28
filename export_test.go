package ebitenui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// What the external tests reach inside the package.

func (x *UI) KeyboardWindow() *widget.Window { return x.keyboardWindow }

// SetColorMode has the UI read the system's color mode from f.
func (x *UI) SetColorMode(f func() ebiten.ColorMode) { x.colorMode = f }
