package widget

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/input"
	"github.com/matryer/is"
)

// Setting the input layers up, which the UI does every frame, allocates
// nothing: a list's scroll container and slider, a slider, a focusable
// child raised to a layer of its own, a window and an open combo button
// keep their layers from frame to frame.
func TestInputLayers_SetUpWithoutAllocating(t *testing.T) {
	is := is.New(t)

	list := newList(t, ListOpts.Entries([]any{"a", "b"}), ListOpts.EntryLabelFunc(func(e any) string { return e.(string) }))
	raised := NewContainer()
	raised.AddChild(newButton(t, ButtonOpts.WidgetOpts(WidgetOpts.ElevateLayer(true))))
	window := NewWindow(WindowOpts.Contents(NewContainer()))
	window.SetLocation(img.Rect(0, 0, 100, 100))
	combo := newComboButton(t)
	combo.ContentVisible = true
	layerers := []input.Layerer{list, newSlider(t), raised, window, combo}
	t.Cleanup(func() { input.SetupInputLayersWithDeferred(nil) })

	input.SetupInputLayersWithDeferred(layerers)
	is.Equal(testing.AllocsPerRun(100, func() { input.SetupInputLayersWithDeferred(layerers) }), 0.0)
}

// A window's layer follows the window from frame to frame: it covers the
// screen while the window is modal, and blocks the widgets under it unless
// the window is ephemeral.
func TestWindow_LayerFollowsTheWindow(t *testing.T) {
	is := is.New(t)
	t.Cleanup(func() { input.SetupInputLayersWithDeferred(nil) })

	w := NewWindow(WindowOpts.Contents(NewContainer()))
	w.SetLocation(img.Rect(0, 0, 100, 100))
	setUp := func() { input.SetupInputLayersWithDeferred([]input.Layerer{w}) }
	setUp()
	is.True(!w.layer.FullScreen && w.layer.BlockLower)
	w.Modal = true
	setUp()
	is.True(w.layer.FullScreen)
	w.Ephemeral = true
	setUp()
	is.True(!w.layer.BlockLower)
}
