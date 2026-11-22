package main

import (
	"image/color"
	"log"
	"strconv"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Three checklists in the theme's look. A mission's objectives, which the
// game sets as they happen: one counts its progress, one may fail, one is
// optional, and the done ones are struck through. A tutorial whose steps
// show one at a time as each is done. A to-do list the player ticks with a
// click, or with Tab to focus it, Up and Down, and Space.
func main() {
	theme := themes.GetBasicDarkTheme()

	moons := 0
	mission := widget.NewChecklist(
		widget.ChecklistOpts.Title("Survey Jupiter"),
		widget.ChecklistOpts.Progress(func(done, total int) string { return strconv.Itoa(done) + " of " + strconv.Itoa(total) }),
		widget.ChecklistOpts.Items(
			widget.ChecklistItem{Label: "Reach orbit", State: widget.ChecklistDone},
			widget.ChecklistItem{Label: "Survey the Galilean moons", Detail: "0/4"},
			widget.ChecklistItem{Label: "Bring the probe back"},
			widget.ChecklistItem{Label: "Photograph the rings (optional)", Optional: true},
		),
		widget.ChecklistOpts.StrikeDone(),
		widget.ChecklistOpts.MaxWidth(300),
	)
	missionState := label(theme, "")
	mission.CompletedEvent.AddHandler(func(any) { missionState.Label = "Mission complete" })
	survey := button("Survey a moon", func() {
		if moons < 4 {
			moons++
			mission.SetDetail(1, strconv.Itoa(moons)+"/4")
		}
		if moons == 4 {
			mission.SetState(1, widget.ChecklistDone)
		}
	})
	ret := button("Return", func() { mission.SetState(2, widget.ChecklistDone) })
	lose := button("Lose the probe", func() { mission.SetState(2, widget.ChecklistFailed) })

	tutorial := widget.NewChecklist(
		widget.ChecklistOpts.Title("Tutorial"),
		widget.ChecklistOpts.HideUpcoming(),
		widget.ChecklistOpts.Items(
			widget.ChecklistItem{Label: "Open the star map"},
			widget.ChecklistItem{Label: "Zoom in on a planet"},
			widget.ChecklistItem{Label: "Pick a moon to visit"},
			widget.ChecklistItem{Label: "Plot a course"},
			widget.ChecklistItem{Label: "Launch"},
		),
	)
	step := button("Do the step", func() { tutorial.SetState(tutorial.Current(), widget.ChecklistDone) })

	todo := widget.NewChecklist(
		widget.ChecklistOpts.Title("To do"),
		widget.ChecklistOpts.Progress(func(done, total int) string { return strconv.Itoa(total-done) + " left" }),
		widget.ChecklistOpts.Items(
			widget.ChecklistItem{Label: "Refuel the tug", Tickable: true},
			widget.ChecklistItem{Label: "Repair the antenna", Tickable: true, State: widget.ChecklistDone},
			widget.ChecklistItem{Label: "Name the new moon", Tickable: true},
			widget.ChecklistItem{Label: "Write the log", Tickable: true},
		),
	)

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)), widget.RowLayoutOpts.Spacing(36))))
	root.AddChild(
		column(mission, missionState, row(survey, ret, lose)),
		column(tutorial, step),
		column(todo),
	)

	ui := &ebitenui.UI{Container: root, PrimaryTheme: theme}
	ebiten.SetWindowSize(1000, 360)
	ebiten.SetWindowTitle("Ebiten UI - Checklist")
	if err := ebiten.RunGame(&game{ui: ui}); err != nil {
		log.Println(err)
	}
}

func button(text string, clicked func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextLabel(text),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 10, Right: 10, Top: 4, Bottom: 4}),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { clicked() }),
	)
}

func label(theme *widget.Theme, text string) *widget.Text {
	return widget.NewText(widget.TextOpts.Text(text, theme.DefaultFace, theme.DefaultTextColor))
}

func column(children ...widget.PreferredSizeLocateableWidget) *widget.Container {
	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(12))))
	c.AddChild(children...)
	return c
}

func row(children ...widget.PreferredSizeLocateableWidget) *widget.Container {
	c := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Spacing(8))))
	c.AddChild(children...)
	return c
}

type game struct{ ui *ebitenui.UI }

func (g *game) Update() error {
	g.ui.Update()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) { g.ui.Draw(screen) }

func (g *game) Layout(w, h int) (int, int) { return w, h }
