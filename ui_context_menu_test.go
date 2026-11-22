package ebitenui_test

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// newMenu returns a context menu of two buttons, one above the other, the
// first one and the count of its clicks.
func newMenu() (menu *widget.Container, first *widget.Button, clicks *int) {
	clicks = new(int)
	size := widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(60, 20))
	first = focusButton(image.Rectangle{}, size, widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { *clicks++ }))
	second := focusButton(image.Rectangle{}, size)
	menu = widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	menu.AddChild(first, second)
	return menu, first, clicks
}

// The context-menu key, and Shift+F10, open the focused widget's context
// menu below it, the focus on the menu's first widget; Back closes it,
// rather than calling BackHandler, and gives the focus back. With
// GamepadFocus on, a gamepad's Menu button opens it too. For
// ebitenui/ebitenui#44.
func TestUI_TheContextMenuKeyOpensTheFocusedWidgetsMenu(t *testing.T) {
	is := is.New(t)

	menu, first, _ := newMenu()
	b := focusButton(image.Rect(10, 10, 60, 30), widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.ContextMenu(menu)))
	ui := focusUI(b)
	backs := 0
	ui.BackHandler = func() { backs++ }
	d, press := useDevices(t, ui)
	ui.SetFocusedWidget(b)
	open := func() bool {
		w := b.GetWidget().ContextMenuWindow
		return w != nil && ui.IsWindowOpen(w)
	}

	press(ebiten.KeyContextMenu)
	is.True(open())
	is.True(ui.GetFocusedWidget() == first)
	ui.Draw(ebiten.NewImage(200, 200))
	is.Equal(menu.GetWidget().Rect.Min, image.Pt(10, 30)) // below the widget

	press(ebiten.KeyEscape)
	is.True(!open())
	is.True(ui.GetFocusedWidget() == b)
	is.Equal(backs, 0)

	press(ebiten.KeyF10)
	is.True(!open()) // F10 alone
	d.Keys[ebiten.KeyShift] = true
	press(ebiten.KeyF10)
	d.Keys[ebiten.KeyShift] = false
	is.True(open())
	press(ebiten.KeyEscape)

	press(ebiten.StandardGamepadButtonCenterRight)
	is.True(!open()) // GamepadFocus off
	ui.GamepadFocus = true
	press(ebiten.StandardGamepadButtonCenterRight)
	is.True(open())
	press(ebiten.StandardGamepadButtonRightRight)
	is.True(!open())
	is.True(ui.GetFocusedWidget() == b)
	is.Equal(backs, 0)

	press(ebiten.KeyEscape)
	is.Equal(backs, 1) // no menu open: BackHandler
}

// A focused widget without a context menu opens its nearest parent's;
// with none up the tree, the key opens nothing.
func TestUI_TheContextMenuKeyOpensTheNearestParentsMenu(t *testing.T) {
	is := is.New(t)

	menu, first, _ := newMenu()
	b := focusButton(image.Rect(10, 10, 60, 30))
	group := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.ContextMenu(menu)))
	group.AddChild(b)
	lone := focusButton(image.Rect(100, 10, 150, 30))
	ui := focusUI(group, lone)
	_, press := useDevices(t, ui)

	ui.SetFocusedWidget(lone)
	press(ebiten.KeyContextMenu)
	is.True(ui.GetFocusedWidget() == lone) // nothing opened

	ui.SetFocusedWidget(b)
	press(ebiten.KeyContextMenu)
	is.True(ui.IsWindowOpen(group.GetWidget().ContextMenuWindow))
	is.True(ui.GetFocusedWidget() == first)
}

// A menu keys opened that closes on a click closes as a key chooses in it,
// after the choice: Enter let go, or a gamepad's A. One set not to close
// stays open.
func TestUI_AMenuKeysOpenedClosesOnAChoice(t *testing.T) {
	for name, tc := range map[string]struct {
		mode   widget.WindowCloseMode
		choose any
		closes bool
	}{
		"Enter":                      {widget.WindowCloseClick, ebiten.KeyEnter, true},
		"Space":                      {widget.WindowCloseClick, ebiten.KeySpace, true},
		"A":                          {widget.WindowCloseClick, ebiten.StandardGamepadButtonRightBottom, true},
		"Enter, closing on none":     {widget.WindowCloseNone, ebiten.KeyEnter, false},
		"Enter, closing on clickout": {widget.WindowCloseClickOut, ebiten.KeyEnter, false},
	} {
		t.Run(name, func(t *testing.T) {
			is := is.New(t)

			menu, _, clicks := newMenu()
			b := focusButton(image.Rect(10, 10, 60, 30), widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.ContextMenu(menu), widget.WidgetOpts.ContextMenuCloseMode(tc.mode)))
			ui := focusUI(b)
			ui.GamepadFocus = true
			_, press := useDevices(t, ui)
			ui.SetFocusedWidget(b)
			press(ebiten.KeyContextMenu)

			press(tc.choose)
			is.Equal(*clicks, 1) // chosen
			is.Equal(ui.IsWindowOpen(b.GetWidget().ContextMenuWindow), !tc.closes)
			if tc.closes {
				is.True(ui.GetFocusedWidget() == b)
			}
		})
	}
}
