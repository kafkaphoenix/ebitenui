package widget_test

import (
	"bytes"
	"image"
	"image/color"
	"strconv"
	"testing"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// showList makes a list 200 by 100 pixels with its required looks and
// draws it a few frames, so its rows in view are laid out.
func showList(t *testing.T, opts ...widget.ListOpt) {
	t.Helper()
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	var face text.Face = &text.GoTextFace{Source: src, Size: 12}
	none := eimage.NewNineSliceColor(color.Transparent)
	l := widget.NewList(append([]widget.ListOpt{
		widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: none, Mask: none}),
		widget.ListOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: none, Pressed: none}}),
		widget.ListOpts.EntryFontFace(&face),
		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Unselected: color.White, Selected: color.White, DisabledUnselected: color.White, DisabledSelected: color.White,
			SelectedBackground: color.Transparent, DisabledSelectedBackground: color.Transparent,
		}),
	}, opts...)...)
	l.SetLocation(image.Rect(0, 0, 200, 100))
	l.Validate()
	screen := ebiten.NewImage(200, 100)
	for range 3 {
		l.Update(&widget.UpdateObject{})
		l.Render(screen)
	}
}

// A list of ten thousand labeled entries asks only for the labels of the
// rows in view, a few, not every entry's: it keeps a screenful of rows.
func TestList_LabelsOfTheRowsInViewOnly(t *testing.T) {
	entries := make([]any, 10_000)
	for i := range entries {
		entries[i] = i
	}
	asked := map[any]bool{}
	showList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryLabelFunc(func(e any) string {
		asked[e] = true
		return "Entry " + strconv.Itoa(e.(int))
	}))
	if len(asked) == 0 || len(asked) > 20 {
		t.Errorf("a list 100 pixels tall asked for %d labels, want those of the few rows in view", len(asked))
	}
}

// counted is an entry widget h pixels tall that counts its updates and
// draws.
type counted struct {
	widget           *widget.Widget
	h                int
	updates, renders int
}

func (c *counted) GetWidget() *widget.Widget        { return c.widget }
func (c *counted) PreferredSize() (int, int)        { return 50, c.h }
func (c *counted) SetLocation(rect image.Rectangle) { c.widget.Rect = rect }
func (c *counted) Validate()                        {}
func (c *counted) Update(_ *widget.UpdateObject)    { c.updates++ }
func (c *counted) Render(_ *ebiten.Image)           { c.renders++ }

// Entries with widgets of their own take rows as tall as their widgets,
// one under the other, and only the widgets in view are updated and drawn.
func TestList_EntryWidgetsOfTheirHeightsInViewOnly(t *testing.T) {
	entries := make([]any, 50)
	widgets := map[any]*counted{}
	for i := range entries {
		entries[i] = i
		widgets[i] = &counted{widget: widget.NewWidget(), h: 20 + 20*(i%2)} // 20 and 40 pixels, in turn
	}
	showList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryTextPadding(&widget.Insets{}),
		widget.ListOpts.EntryWidgetFunc(func(e any) widget.PreferredSizeLocateableWidget { return widgets[e] }))

	for i := 1; i < 4; i++ {
		above, row := widgets[i-1].widget.Rect, widgets[i].widget.Rect
		if row.Min.Y != above.Max.Y || row.Dy() != widgets[i].h {
			t.Errorf("entry %d's row is %v under %v, want right under it and %d tall", i, row, above, widgets[i].h)
		}
	}
	if w := widgets[0]; w.updates == 0 || w.renders == 0 {
		t.Errorf("the first entry's widget, in view, was updated %d and drawn %d times", w.updates, w.renders)
	}
	if w := widgets[49]; w.updates != 0 || w.renders != 0 {
		t.Errorf("the last entry's widget, out of view, was updated %d and drawn %d times, want none", w.updates, w.renders)
	}
}
