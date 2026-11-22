package main

import (
	"fmt"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// A list of a hundred thousand entries: the List keeps widgets only for
// the rows in view, and binds them to the entries as it scrolls. A click
// selects an entry; with the list focused (Tab), Up and Down move through
// the entries. The window's title shows the selection.
func main() {
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout(
		widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(30)))))

	entries := make([]any, 100_000)
	for i := range entries {
		entries[i] = i
	}
	list := widget.NewList(
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true, StretchVertical: true}))),
		widget.ListOpts.Entries(entries),
		// Each row a line of its entry's label, in the theme's list look. For
		// rows of any widget, EntryWidgetFunc makes an entry's.
		widget.ListOpts.EntryLabelFunc(func(e any) string { return fmt.Sprintf("Entry %d", e) }),
		// Dragging the rows, with a mouse or a finger, scrolls them too.
		widget.ListOpts.DragToScroll(),
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			ebiten.SetWindowTitle(fmt.Sprintf("Ebiten UI - Long List: entry %d", args.Entry))
		}),
	)
	root.AddChild(list)

	ui := &ebitenui.UI{Container: root, PrimaryTheme: themes.GetBasicDarkTheme()}
	ebiten.SetWindowSize(400, 500)
	ebiten.SetWindowTitle("Ebiten UI - Long List")
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
