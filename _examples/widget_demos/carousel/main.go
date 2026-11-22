package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Two carousels in the theme's look. At the left, a loading screen: its
// screens go round by themselves every three seconds, waiting while the
// pointer is over them, with a loading bar under them. At the right, a
// tutorial in steps: each step's page asks for its box to be ticked, as a
// game would ask for something to be done, and the carousel moves on as it
// is. Arrows, dots and a swipe across a page move them too.
func main() {
	theme := themes.GetBasicDarkTheme()
	face := theme.DefaultFace

	page := func(c color.Color, label string, extra ...widget.PreferredSizeLocateableWidget) *widget.Container {
		p := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(c)),
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(40)),
				widget.RowLayoutOpts.Spacing(20))),
		)
		p.AddChild(widget.NewText(widget.TextOpts.Text(label, face, color.White)))
		p.AddChild(extra...)
		return p
	}

	screens := make([]widget.PreferredSizeLocateableWidget, 3)
	for i, c := range []color.Color{color.NRGBA{40, 70, 120, 255}, color.NRGBA{120, 60, 40, 255}, color.NRGBA{50, 110, 60, 255}} {
		screens[i] = page(c, fmt.Sprintf("Loading screen %d", i+1))
	}
	loading := widget.NewCarousel(widget.CarouselOpts.Pages(screens...), widget.CarouselOpts.Loop(),
		widget.CarouselOpts.AutoAdvance(3*time.Second),
		widget.CarouselOpts.WidgetOpts(widget.WidgetOpts.MinSize(320, 240)))

	steps := []string{"Step 1: tick the box below.", "Step 2: and this one.", "Step 3: one more."}
	var tutorial *widget.Carousel
	pages := make([]widget.PreferredSizeLocateableWidget, 0, len(steps)+1)
	for i, s := range steps {
		box := widget.NewCheckbox(widget.CheckboxOpts.Text("Done", face, &widget.LabelColor{Idle: color.White}),
			widget.CheckboxOpts.StateChangedHandler(func(a *widget.CheckboxChangedEventArgs) {
				tutorial.SetStepDone(i, a.State == widget.Checked)
			}))
		pages = append(pages, page(color.NRGBA{60, 60, 80, 255}, s, box))
	}
	pages = append(pages, page(color.NRGBA{60, 90, 60, 255}, "All done."))
	tutorial = widget.NewCarousel(widget.CarouselOpts.Pages(pages...), widget.CarouselOpts.Steps(),
		widget.CarouselOpts.AdvanceWhenDone(),
		widget.CarouselOpts.WidgetOpts(widget.WidgetOpts.MinSize(320, 240)))

	left := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(10))))
	left.AddChild(loading, widget.NewLoading(widget.LoadingOpts.BarSize(320, 6)))
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(30)), widget.RowLayoutOpts.Spacing(30))))
	root.AddChild(left, tutorial)

	ui := &ebitenui.UI{Container: root, PrimaryTheme: theme}
	ebiten.SetWindowSize(760, 360)
	ebiten.SetWindowTitle("Ebiten UI - Carousel")
	if err := ebiten.RunGame(&game{ui: ui}); err != nil {
		log.Println(err)
	}
}

type game struct{ ui *ebitenui.UI }

func (g *game) Update() error {
	g.ui.Update()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) { g.ui.Draw(screen) }

func (g *game) Layout(w, h int) (int, int) { return w, h }
