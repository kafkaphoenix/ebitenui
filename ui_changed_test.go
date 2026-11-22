package ebitenui_test

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// A UI reports a change of its picture that no input made, here a
// focused text input's caret blinking every 27 updates, in the update it
// happens, and none in the updates between.
func TestUIChangedAsTheCaretBlinks(t *testing.T) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	var face text.Face = &text.GoTextFace{Source: src, Size: 12}
	in := widget.NewTextInput(
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White, Disabled: color.White, Caret: color.White}),
	)
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	root.AddChild(in)
	ui := &ebitenui.UI{Container: root}
	d := uitest.New(t, 200, 50, ui.Update, ui.Draw)
	d.Frames(2)
	in.Focus(true)
	d.Frames(1)
	ui.Changed() // what the frames so far did

	var changedAt []int
	for i := range 60 {
		d.Frames(1)
		if ui.Changed() {
			changedAt = append(changedAt, i)
		}
	}
	if len(changedAt) != 2 || changedAt[1]-changedAt[0] != 27 {
		t.Errorf("over 60 updates the UI changed by itself at %v, want twice, 27 updates apart", changedAt)
	}
}

// changedOver runs n frames and returns those, counted from 0, in which
// the UI changed by itself.
func changedOver(d *uitest.Driver, ui *ebitenui.UI, n int) []int {
	ui.Changed() // what came before
	var at []int
	for i := range n {
		d.Frames(1)
		if ui.Changed() {
			at = append(at, i)
		}
	}
	return at
}

// A window coming in changes the picture each update of its transition,
// and no more once it is in.
func TestUIChangedAsAWindowComesIn(t *testing.T) {
	ui := &ebitenui.UI{Container: widget.NewContainer()}
	d := uitest.New(t, 200, 100, ui.Update, ui.Draw)
	d.Frames(1)
	w := widget.NewWindow(
		widget.WindowOpts.Contents(widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(50, 50)))),
		widget.WindowOpts.Location(image.Rect(10, 10, 60, 60)),
		widget.WindowOpts.Transition(widget.WindowTransition{Duration: 8}),
	)
	ui.AddWindow(w)
	if at := changedOver(d, ui, 20); len(at) != 8 || at[0] != 0 || at[7] != 7 {
		t.Errorf("as the window came in over 8 updates, the UI changed by itself at %v", at)
	}
}

// A GIF changes the picture as it moves to its next frame: every 3
// updates for frames of 50 ms.
func TestUIChangedAsAGIFPlays(t *testing.T) {
	frame := func() *image.Paletted {
		return image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	}
	g := widget.NewGraphic(widget.GraphicOpts.GIF(&gif.GIF{Image: []*image.Paletted{frame(), frame()}, Delay: []int{5, 5}}))
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	root.AddChild(g)
	ui := &ebitenui.UI{Container: root}
	d := uitest.New(t, 100, 100, ui.Update, ui.Draw)
	d.Frames(1)
	if at := changedOver(d, ui, 12); len(at) != 4 || at[1]-at[0] != 3 {
		t.Errorf("over 12 updates of frames of 3 updates, the UI changed by itself at %v", at)
	}
}

// A tooltip showing, the pointer resting on its widget, changes the
// picture once, as it shows.
func TestUIChangedAsATooltipShows(t *testing.T) {
	tip := widget.NewToolTip(widget.ToolTipOpts.Content(widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(20, 10)))))
	target := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 40), widget.WidgetOpts.ToolTip(tip)))
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	root.AddChild(target)
	ui := &ebitenui.UI{Container: root}
	d := uitest.New(t, 200, 100, ui.Update, ui.Draw)
	d.Frames(1)
	d.Cursor = image.Pt(20, 20)
	if at := changedOver(d, ui, 10); len(at) != 1 {
		t.Errorf("as the tooltip showed, the UI changed by itself at %v, want once", at)
	}
}

// A list let go of while dragged glides on, changing the picture each
// update, until it stops.
func TestUIChangedAsAListGlides(t *testing.T) {
	content := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	for range 40 {
		content.AddChild(widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 20))))
	}
	empty := eimage.NewNineSliceColor(color.Transparent)
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{Idle: empty, Mask: empty}),
		widget.ScrollContainerOpts.DragToScroll(),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 100), widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxHeight: 100})),
	)
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	root.AddChild(scroll)
	ui := &ebitenui.UI{Container: root}
	d := uitest.New(t, 200, 200, ui.Update, ui.Draw)
	d.Frames(1)
	d.Cursor = image.Pt(50, 90)
	d.Mouse[ebiten.MouseButtonLeft] = true
	for y := 90; y >= 30; y -= 20 { // a fast drag up
		d.Cursor = image.Pt(50, y)
		d.Frames(1)
	}
	d.Mouse[ebiten.MouseButtonLeft] = false
	d.Frames(1)
	at := changedOver(d, ui, 200) // a glide slows by 5% an update: some 70 updates
	if len(at) < 3 || at[0] != 0 || at[len(at)-1] != len(at)-1 {
		t.Fatalf("let go of while dragged, the list changed by itself at %v, want every update from the first until it stops", at)
	}
	if len(at) == 200 {
		t.Errorf("the list still glided 200 updates after it was let go of")
	}
}
