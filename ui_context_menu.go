package ebitenui

import (
	"image"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// contextMenu is the context menu open, from a right click, a long press or
// the keys.
type contextMenu struct {
	window *widget.Window
	// closeOnChoice is set for a menu the keys opened that closes on a
	// click: it closes as a key chooses in it, as it would on a click.
	// choosing is whether Enter or Space is held in it.
	closeOnChoice, choosing bool
}

func (u *UI) handleContextMenu(args any) {
	if a, ok := args.(*widget.ContextMenuEventArgs); ok {
		u.openContextMenu(a.Widget, a.Location)
	}
}

// openContextMenu opens owner's context menu in a modal window at location.
func (u *UI) openContextMenu(owner *widget.Widget, location image.Point) *widget.Window {
	x, y := owner.ContextMenu.PreferredSize()
	owner.ContextMenuWindow = widget.NewWindow(
		widget.WindowOpts.Contents(owner.ContextMenu),
		widget.WindowOpts.CloseMode(owner.ContextMenuCloseMode),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Location(image.Rect(0, 0, x, y).Add(location)),
	)
	u.contextMenu = contextMenu{window: owner.ContextMenuWindow}
	u.AddWindow(owner.ContextMenuWindow)
	return owner.ContextMenuWindow
}

// openFocusedContextMenu opens, for ActionMenu, the context menu of the
// focused widget, or of its nearest parent with one, below the focused
// widget, and focuses the menu's first widget that takes the focus; as the
// menu closes, the focus goes back. For ebitenui/ebitenui#44.
func (u *UI) openFocusedContextMenu() {
	focused := u.focusedWidget
	if focused == nil {
		return
	}
	owner := focused.GetWidget()
	for owner != nil && owner.ContextMenu == nil {
		owner = owner.Parent()
	}
	if owner == nil {
		return
	}
	r := focused.GetWidget().Rect
	window := u.openContextMenu(owner, image.Pt(r.Min.X, r.Max.Y))
	u.contextMenu.closeOnChoice = owner.ContextMenuCloseMode == widget.WindowCloseClick
	window.ClosedEvent.AddHandler(func(any) {
		u.SetFocusedWidget(focused)
	})
	if f := owner.ContextMenu.GetFocusers(); len(f) > 0 {
		u.SetFocusedWidget(f[0])
	}
}

// closeContextMenu closes the context menu open, and reports whether one
// was.
func (u *UI) closeContextMenu() bool {
	w := u.contextMenu.window
	if w == nil || !u.IsWindowOpen(w) {
		return false
	}
	w.Close()
	return true
}

// closeContextMenuOnChoice closes a menu the keys opened that closes on a
// click, once a key chose in it: Enter or Space let go, which click a
// button as they come up, or Activate, by which the UI activated the
// focused widget this update.
func (u *UI) closeContextMenuOnChoice(activated bool) {
	c := &u.contextMenu
	if !c.closeOnChoice {
		return
	}
	held := input.KeyPressed(ebiten.KeyEnter) || input.KeyPressed(ebiten.KeySpace)
	chose := activated || c.choosing && !held
	c.choosing = held
	if chose {
		u.closeContextMenu()
	}
}
