package image

import "github.com/hajimehoshi/ebiten/v2"

// What the external tests reach inside the package.

var RoundedPixels = roundedPixels

func (x *MaskedRenderBuffer) MaskedIn() *ebiten.Image { return x.maskedIn }
func (x *MaskedRenderBuffer) RenderIn() *ebiten.Image { return x.renderIn }
func (x *NineSlice) Heights() [3]int                  { return x.heights }
func (x *NineSlice) Transparent() bool                { return x.transparent }
func (x *NineSlice) Widths() [3]int                   { return x.widths }
