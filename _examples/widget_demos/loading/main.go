package main

import (
	"image/color"
	"log"
	"time"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Loading indicators, in the theme's look: a spinner at the center, a
// small one at the left and a short bar at the right, and along the bottom
// a bar spanning the window with tips under it. They draw only frames that
// change: the demo skips the others, as UI.Changed tells.
func main() {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(30)))))
	at := func(h, v widget.AnchorLayoutPosition, stretch bool) widget.LoadingOpt {
		return widget.LoadingOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: h, VerticalPosition: v, StretchHorizontal: stretch,
		}))
	}

	root.AddChild(
		widget.NewLoading(widget.LoadingOpts.Spinner(), widget.LoadingOpts.SpinnerSize(48),
			at(widget.AnchorLayoutPositionCenter, widget.AnchorLayoutPositionCenter, false)),
		widget.NewLoading(widget.LoadingOpts.Spinner(), widget.LoadingOpts.SpinnerSize(20),
			at(widget.AnchorLayoutPositionStart, widget.AnchorLayoutPositionStart, false)),
		widget.NewLoading(widget.LoadingOpts.BarSize(80, 6), widget.LoadingOpts.Period(1500*time.Millisecond),
			at(widget.AnchorLayoutPositionEnd, widget.AnchorLayoutPositionStart, false)),
		// Spanning the window: a bar of no width of its own, stretched.
		widget.NewLoading(widget.LoadingOpts.Position(widget.TextPositionCenter),
			widget.LoadingOpts.Tips([]string{
				"Tip: a spinner suits a small space, a bar a wide one.",
				"Tip: with a width of its own, a bar sits at its start, center or end.",
				"Tip: a loading indicator draws only the frames that change.",
			}, 3*time.Second),
			at(widget.AnchorLayoutPositionCenter, widget.AnchorLayoutPositionEnd, true)),
	)

	ui := &ebitenui.UI{Container: root, PrimaryTheme: themes.GetBasicDarkTheme()}
	ebiten.SetWindowSize(600, 400)
	ebiten.SetWindowTitle("Ebiten UI - Loading")
	ebiten.SetScreenClearedEveryFrame(false) // the frames not drawn keep the last
	if err := ebiten.RunGame(&game{ui: ui}); err != nil {
		log.Println(err)
	}
}

// game draws a frame only when the interface changed, or the window's
// size did.
type game struct {
	ui    *ebitenui.UI
	dirty bool
	drawn bool
	w, h  int
}

func (g *game) Update() error {
	g.ui.Update()
	g.dirty = g.dirty || g.ui.Changed()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.drawn && !g.dirty {
		return
	}
	screen.Clear()
	g.ui.Draw(screen)
	g.drawn, g.dirty = true, false
}

func (g *game) Layout(w, h int) (int, int) {
	if w != g.w || h != g.h {
		g.w, g.h, g.dirty = w, h, true
	}
	return w, h
}
