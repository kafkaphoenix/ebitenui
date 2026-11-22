package main

import (
	"image"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// A minimap in a window: a Canvas draws a game's world, made once, scaled
// to its place, and a frame for the part in view; a click or a drag on it
// moves the view there. The window drags and resizes as any.
func main() {
	world := makeWorld()
	view := image.Rect(0, 0, 64, 48) // the part of the world in view, in world pixels

	var canvas *widget.Canvas
	moveView := func(x, y int) { // x, y from the canvas's corner
		r := canvas.GetWidget().Rect
		wx := x * world.Bounds().Dx() / max(1, r.Dx())
		wy := y * world.Bounds().Dy() / max(1, r.Dy())
		view = view.Sub(view.Min).Add(image.Pt(wx-view.Dx()/2, wy-view.Dy()/2))
		canvas.MarkChanged()
	}
	canvas = widget.NewCanvas(
		widget.CanvasOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(200, 200),
			widget.WidgetOpts.MouseButtonPressedHandler(func(a *widget.MouseButtonPressedEventArgs) { moveView(a.OffsetX, a.OffsetY) }),
			widget.WidgetOpts.CursorMoveHandler(func(a *widget.CursorMoveEventArgs) {
				if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
					moveView(a.OffsetX, a.OffsetY)
				}
			}),
		),
		widget.CanvasOpts.DrawFunc(func(screen *ebiten.Image, rect image.Rectangle) {
			sx := float64(rect.Dx()) / float64(world.Bounds().Dx())
			sy := float64(rect.Dy()) / float64(world.Bounds().Dy())
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(sx, sy)
			op.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
			screen.DrawImage(world, op)
			vector.StrokeRect(screen, float32(rect.Min.X)+float32(float64(view.Min.X)*sx), float32(rect.Min.Y)+float32(float64(view.Min.Y)*sy),
				float32(float64(view.Dx())*sx), float32(float64(view.Dy())*sy), 2, color.White, false)
		}),
	)

	theme := themes.GetBasicDarkTheme()
	contents := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{45, 50, 60, 255})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(10)))))
	canvas.GetWidget().LayoutData = widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}
	contents.AddChild(canvas)
	window := widget.NewWindow(
		widget.WindowOpts.Contents(contents),
		widget.WindowOpts.Draggable(), widget.WindowOpts.Resizeable(),
		widget.WindowOpts.MinSize(120, 120),
		widget.WindowOpts.Location(image.Rect(40, 40, 260, 260)),
	)
	root := widget.NewContainer(widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})))
	ui := &ebitenui.UI{Container: root, PrimaryTheme: theme}
	ui.AddWindow(window)

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Ebiten UI - Canvas")
	if err := ebiten.RunGame(&game{ui: ui}); err != nil {
		log.Println(err)
	}
}

// makeWorld draws a world of land and sea once: the canvas scales it.
func makeWorld() *ebiten.Image {
	const size = 256
	w := ebiten.NewImage(size, size)
	w.Fill(color.NRGBA{30, 60, 120, 255})
	for i := range 40 {
		x, y := float32((i*97)%size), float32((i*57)%size)
		vector.FillCircle(w, x, y, float32(8+(i*13)%24), color.NRGBA{60, 120, 60, 255}, true)
	}
	return w
}

type game struct{ ui *ebitenui.UI }

func (g *game) Update() error {
	g.ui.Update()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) { g.ui.Draw(screen) }

func (g *game) Layout(w, h int) (int, int) { return w, h }
