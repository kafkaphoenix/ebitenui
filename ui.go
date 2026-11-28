package ebitenui

import (
	"image"
	"slices"
	"sort"
	"time"

	"github.com/ebitenui/ebitenui/internal/changes"
	internalinput "github.com/ebitenui/ebitenui/internal/input"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/utilities/sliceutil"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
)

// UI encapsulates a complete user interface that can be rendered onto the screen.
// There should only be exactly one UI per application.
type UI struct {
	// Container is the root container of the UI hierarchy.
	Container widget.Containerer
	// If true the default tab/shift-tab to focus will be disabled
	DisableDefaultFocus bool

	// GamepadFocus, when true, lets the gamepads with the standard layout
	// navigate: their D-pad and left stick give the direction actions, as
	// the arrow keys do, and their bottom face button, A on an Xbox pad or
	// cross on a PlayStation one, the Activate action (widget.Action), and
	// their Menu button, the right one in their middle, the Menu action, as
	// the context-menu key does. The focused widget takes those it acts on:
	// a list moves through its entries, a slider its handle. A gamepad's
	// direction it leaves moves the focus (ChangeFocus), an Activate it
	// leaves activates it (ActivateFocused), a Menu opens its context menu;
	// the arrow keys never move the focus. A button pressed on two gamepads
	// at once acts once.
	//
	// A context menu the keys or a gamepad open has the focus on its first
	// widget, and gives it back as it closes. One that closes on a click
	// (WindowCloseClick) closes too as a key or a gamepad chooses in it.
	GamepadFocus bool

	// A direction held, from a key or a gamepad, acts as it goes down, then
	// after NavigationRepeatDelay once every NavigationRepeatInterval;
	// unset, 0.4 s and 0.08 s. They count in updates at Ebitengine's TPS,
	// 60 a second when the TPS follows the display.
	NavigationRepeatDelay    time.Duration
	NavigationRepeatInterval time.Duration

	// ScreenTransform is the transform a game draws the interface with onto
	// the screen, when it draws the interface on an image of its own and
	// draws that image enlarged or moved: a pixel-art interface drawn small
	// and scaled up, or one drawn into a part of the screen. The pointer and
	// the fingers, read in the screen's pixels, are mapped back through it
	// to the interface's, so the widgets take them where they show. The
	// zero GeoM, the identity, for an interface drawn straight on the
	// screen; one that can't be undone, a scale of 0, maps nothing.
	//
	//	small := ebiten.NewImage(w/3, h/3)
	//	ui.Draw(small)
	//	var op ebiten.DrawImageOptions
	//	op.GeoM.Scale(3, 3)
	//	screen.DrawImage(small, &op)
	//	ui.ScreenTransform = op.GeoM
	//
	// For ebitenui/ebitenui#14, #285 and #66.
	ScreenTransform ebiten.GeoM

	// BackHandler, when set, is called for a Back action no widget takes:
	// Escape, or with GamepadFocus on a gamepad's right face button, B on an
	// Xbox pad or circle on a PlayStation one. Closing a menu or a window is
	// the usual answer. An open list combo button takes Back and closes; an
	// open context menu closes, and BackHandler is not called.
	BackHandler func()

	// OnScreenKeyboard, when set, opens for a gamepad's Activate on a focused
	// text input instead of submitting it: in a modal window across the
	// bottom of the screen, focused on its first key, typing into the input.
	// Done or Back closes it and gives the focus back to the input.
	OnScreenKeyboard *widget.OnScreenKeyboard

	// If true the default relayering of Windows will be disabled
	DisableWindowRelayering bool

	// This exposes a Render call before the Container is drawn,
	// but after the Windows with DrawLayer < 0 are drawn.
	PreRenderHook widget.RenderFunc

	// This exposes a Render call after the Container is drawn,
	// but before the Windows with DrawLayer >= 0 (all by default) are drawn.
	PostRenderHook widget.RenderFunc

	// Theme settings
	PrimaryTheme  *widget.Theme
	previousTheme *widget.Theme

	// DarkTheme and LightTheme, when either is set, make the UI follow the
	// system's color mode: PrimaryTheme becomes DarkTheme while the system
	// is dark, LightTheme while it is light or doesn't say, or the one
	// given when only one is. The UI reads the mode as it starts and then
	// every half second. The themes are the caller's own, so changing one
	// changes no other UI.
	DarkTheme, LightTheme *widget.Theme
	colorMode             func() ebiten.ColorMode // the system's, or a test's
	modeIn                int                     // updates until the mode is read again

	focusedWidget      widget.Focuser
	focusedWindow      *widget.Window
	focusedWindowIndex int
	inputLayerers      []input.Layerer
	windows            []*widget.Window

	previousContainer          widget.Containerer
	previousRemoveHandlerFuncs []event.RemoveHandlerFunc
	tabWasPressed              bool
	navigation                 navigation
	keyboardWindow             *widget.Window // the window OnScreenKeyboard opens in
	contextMenu                contextMenu    // the context menu open
	updObj                     *widget.UpdateObject

	debugMode bool

	changes uint64 // the changes widgets made by themselves, as Changed last saw them
}

// Changed reports whether widgets changed the interface's picture by
// themselves since the last call: a text input's caret blinking, a
// tooltip showing or going, a GIF's next frame, a window's transition, a
// list gliding on after a drag, a loading indicator's band moving, spinner
// turning or next tip, a carousel's slide, a canvas marked changed
// (Canvas.MarkChanged). A game that draws only frames that show
// something new draws then; what its own code changes and the player's
// input it knows itself.
func (u *UI) Changed() bool {
	c := changes.Count()
	changed := c != u.changes
	u.changes = c
	return changed
}

// Update updates u. This method should be called in the Ebiten Update function.
func (u *UI) Update() {
	internalinput.InputHandler.SetScreenTransform(u.ScreenTransform)
	input.Update()
	defer input.AfterUpdate()
	u.followColorMode()

	if u.previousContainer == nil || u.previousContainer != u.Container {
		for _, removeHandler := range u.previousRemoveHandlerFuncs {
			removeHandler()
		}
		u.previousRemoveHandlerFuncs = []event.RemoveHandlerFunc{
			u.Container.GetWidget().ContextMenuEvent.AddHandler(u.handleContextMenu),
			u.Container.GetWidget().FocusEvent.AddHandler(u.handleFocusEvent),
			u.Container.GetWidget().ToolTipEvent.AddHandler(u.handleToolTipEvent),
			u.Container.GetWidget().DragAndDropEvent.AddHandler(u.handleDragAndDropEvent),
		}
		u.previousContainer = u.Container
		// Close all Ephemeral Windows (tooltip/dnd/etc).
		u.closeEphemeralWindows(0)
	}
	u.setTheme()
	u.handleFocusChangeRequest()
	u.giveActions()

	// If widget is not visible or disabled, change focus to next widget.
	if u.focusedWidget != nil && (u.focusedWidget.GetWidget().Disabled || !u.focusedWidget.GetWidget().IsVisible()) {
		u.ChangeFocus(widget.FocusNext)
	}
	resortFocusedWindow := false
	index := 0
	for ; index < len(u.windows); index++ {
		if u.windows[index].DrawLayer < 0 {
			u.resetUpdateObject()
			u.windows[index].Update(u.updObj)
			if u.updObj.RelayoutRequested {
				u.windows[index].RequestRelayout()
			}
			if u.windows[index].FocusedWindow {
				u.windows[index].FocusedWindow = false
				u.focusedWindow = u.windows[index]
				u.focusedWindowIndex = index
				resortFocusedWindow = true
			}
		} else {
			break
		}
	}

	u.resetUpdateObject()
	u.Container.Update(u.updObj)
	if u.updObj.RelayoutRequested {
		u.Container.RequestRelayout()
	}
	if u.updObj.CloseEphemeralWindows {
		u.closeEphemeralWindows(0)
	}
	for ; index < len(u.windows); index++ {
		u.resetUpdateObject()
		u.windows[index].Update(u.updObj)
		if u.updObj.RelayoutRequested {
			u.windows[index].RequestRelayout()
		}
		if u.windows[index].FocusedWindow {
			u.windows[index].FocusedWindow = false
			u.focusedWindow = u.windows[index]
			u.focusedWindowIndex = index
			resortFocusedWindow = true
		}
	}

	if !u.DisableWindowRelayering && resortFocusedWindow {
		u.windows = sliceutil.ShiftEnd(u.windows, u.focusedWindowIndex)
	}

	u.actOnActionsLeft()
	event.ExecuteDeferred()
}

func (u *UI) resetUpdateObject() {
	// Reset update object
	if u.updObj == nil {
		u.updObj = &widget.UpdateObject{}
	}

	u.updObj.RelayoutRequested = false
	u.updObj.CloseEphemeralWindows = false
	u.updObj.DebugMode = u.debugMode
	u.updObj.Actions = &u.navigation.actions
}

// Draw renders u onto screen. This function should be called in the Ebiten Draw function.
func (u *UI) Draw(screen *ebiten.Image) {
	u.setTheme()
	input.Draw(screen)
	defer input.AfterDraw(screen)
	x, y := screen.Bounds().Dx(), screen.Bounds().Dy()
	rect := image.Rect(0, 0, x, y)
	u.setupInputLayers()
	u.Container.SetLocation(rect)
	u.render(screen)
	// Render elements that pop up (like combobox) on top of everything else
	widget.RenderDeferred(screen)
}

// Layout lays the interface out on a screen width by height pixels, as
// Draw does before it draws, without drawing: after it, every widget in
// its containers, scroll containers, lists, tab books, flip books, text
// areas and windows, and in widgets of the game's that place children
// (widget.ChildPlacer), is where it will show. A game places things by
// them before the first frame, or right after a change rather than a frame
// later. A widget's own parts, a button's text, are placed as it draws.
// Draw lays out again only what changed since. For ebitenui/ebitenui#233.
func (u *UI) Layout(width, height int) {
	u.setTheme()
	u.Container.SetLocation(image.Rect(0, 0, width, height))
	widget.PlaceChildren(u.Container)
	for _, w := range u.windows {
		w.PlaceChildren()
	}
}

func (u *UI) setupInputLayers() {
	num := 1
	if len(u.windows) > 0 {
		num += len(u.windows)
	}

	if cap(u.inputLayerers) < num {
		u.inputLayerers = make([]input.Layerer, num)
	}

	u.inputLayerers = u.inputLayerers[:0]
	u.inputLayerers = append(u.inputLayerers, u.Container)
	for _, w := range u.windows {
		u.inputLayerers = append(u.inputLayerers, w)
	}

	input.SetupInputLayersWithDeferred(u.inputLayerers)
}

func (u *UI) render(screen *ebiten.Image) {
	index := 0
	for ; index < len(u.windows); index++ {
		if u.windows[index].DrawLayer < 0 {
			u.windows[index].Render(screen)
		} else {
			break
		}
	}

	if u.PreRenderHook != nil {
		u.PreRenderHook(screen)
	}

	u.Container.Render(screen)

	if u.PostRenderHook != nil {
		u.PostRenderHook(screen)
	}

	for ; index < len(u.windows); index++ {
		u.windows[index].Render(screen)
	}
}

// followColorMode sets PrimaryTheme to DarkTheme or LightTheme by the
// system's color mode, read every half second, when either is set.
func (u *UI) followColorMode() {
	if u.DarkTheme == nil && u.LightTheme == nil {
		return
	}
	if u.modeIn > 0 {
		u.modeIn--
		return
	}
	u.modeIn = max(1, ebiten.TPS()/2) - 1
	mode := ebiten.SystemColorMode
	if u.colorMode != nil {
		mode = u.colorMode
	}
	th := u.LightTheme
	if th == nil || u.DarkTheme != nil && mode() == ebiten.ColorModeDark {
		th = u.DarkTheme
	}
	u.PrimaryTheme = th
}

func (u *UI) setTheme() {
	// Handle the user setting a new theme.
	if u.Container != nil {
		if (u.PrimaryTheme != nil && u.Container.GetWidget().GetTheme() == nil) || u.PrimaryTheme != u.previousTheme {
			widget.SetUITheme(u.PrimaryTheme)
			u.Container.GetWidget().SetTheme(u.PrimaryTheme)
			u.previousTheme = u.PrimaryTheme

			// Validate the main container and the open windows with the new theme.
			u.Container.Validate()
			for _, w := range u.windows {
				w.GetContainer().Validate()
			}
		} else if !u.Container.IsValidated() {
			u.Container.Validate()
		}
	}
}
func (u *UI) handleFocusEvent(args any) {
	if a, ok := args.(*widget.FocusEventArgs); ok {
		switch {
		case a.Focused: // New widget focused
			u.focusedWidget = a.Widget
		case a.Widget == u.focusedWidget: // Current widget focus removed
			u.focusedWidget = nil
		case a.Widget == nil: // Clicked out of focusable widgets
			// If we didnt just click on the same widget
			if !a.Location.In(u.focusedWidget.GetWidget().Rect) {
				u.focusedWidget.Focus(false)
				u.focusedWidget = nil
			}
		}
	}
}

func (u *UI) handleToolTipEvent(args any) {
	if a, ok := args.(*widget.ToolTipEventArgs); ok {
		a.Window.Ephemeral = true
		if a.Show {
			u.addWindow(a.Window)
		} else {
			u.removeWindow(a.Window)
		}
	}
}

func (u *UI) handleDragAndDropEvent(args any) {
	if a, ok := args.(*widget.DragAndDropEventArgs); ok {
		if a.Show {
			a.Window.Ephemeral = true
			a.DnD.AvailableDropTargets = u.getDropTargets()
			u.addWindow(a.Window)
		} else {
			a.DnD.AvailableDropTargets = nil
			u.removeWindow(a.Window)
		}
	}
}

func (u *UI) getDropTargets() []widget.HasWidget {
	dropTargets := u.Container.GetDropTargets()
	// Loop through the windows array in reverse. If we find a modal window, only loop through its droppable widgets
	for _, v := range slices.Backward(u.windows) {
		if !v.Modal {
			dropTargets = append(dropTargets, v.GetContainer().GetDropTargets()...)
		} else {
			dropTargets = v.GetContainer().GetDropTargets()
			break
		}
	}
	return dropTargets
}

func (u *UI) handleFocusChangeRequest() {
	if !u.DisableDefaultFocus {
		if input.KeyPressed(ebiten.KeyTab) {
			if !u.tabWasPressed {
				u.tabWasPressed = true
				if input.KeyPressed(ebiten.KeyShift) {
					u.ChangeFocus(widget.FocusPrevious)
				} else {
					u.ChangeFocus(widget.FocusNext)
				}
			}
		} else {
			u.tabWasPressed = false
		}
	}
}

// ChangeFocus moves the focus. FocusNext and FocusPrevious follow the tab
// order. A direction follows the focused widget's link for it (AddFocus)
// when it has one that can take the focus, and otherwise goes to the
// nearest focusable widget that way, keeping the focus where it is when
// there is none.
func (u *UI) ChangeFocus(direction widget.FocusDirection) {
	focusableWidgets := u.Container.GetFocusers()
	// Loop through the windows array in reverse. If we find a modal window, only loop through its focusable widgets
	for _, v := range slices.Backward(u.windows) {
		if !v.Modal {
			focusableWidgets = append(focusableWidgets, v.GetContainer().GetFocusers()...)
		} else {
			focusableWidgets = v.GetContainer().GetFocusers()
			break
		}
	}
	fwLen := len(focusableWidgets)

	if direction == widget.FocusNext || direction == widget.FocusPrevious {
		if fwLen == 1 {
			if u.focusedWidget != nil && u.focusedWidget != focusableWidgets[0] {
				u.focusedWidget.Focus(false)
			}
			focusableWidgets[0].Focus(true)
			u.focusedWidget = focusableWidgets[0]
		} else if fwLen > 0 {
			sort.SliceStable(focusableWidgets, func(i, j int) bool {
				return focusableWidgets[i].TabOrder() < focusableWidgets[j].TabOrder()
			})
			if u.focusedWidget != nil {
				if direction == widget.FocusPrevious {
					for i := range fwLen {
						if focusableWidgets[i] == u.focusedWidget {
							u.focusedWidget.Focus(false)
							if i == 0 {
								focusableWidgets[fwLen-1].Focus(true)
								u.focusedWidget = focusableWidgets[fwLen-1]
							} else {
								focusableWidgets[i-1].Focus(true)
								u.focusedWidget = focusableWidgets[i-1]
							}
							return
						}
					}
				} else {
					for i := 0; i < fwLen-1; i++ {
						if focusableWidgets[i] == u.focusedWidget {
							u.focusedWidget.Focus(false)
							focusableWidgets[i+1].Focus(true)
							u.focusedWidget = focusableWidgets[i+1]
							return
						}
					}
				}
				u.focusedWidget.Focus(false)
			}
			focusableWidgets[0].Focus(true)
			u.focusedWidget = focusableWidgets[0]
		}
	} else {
		if u.focusedWidget != nil {
			next := u.focusedWidget.GetFocus(direction)
			if next == nil || next.GetWidget().Disabled || !next.GetWidget().IsVisible() {
				next = nearestFocuser(u.focusedWidget, direction, focusableWidgets)
			}
			if next != nil {
				u.focusedWidget.Focus(false)
				next.Focus(true)
				u.focusedWidget = next
			}
		} else if fwLen > 0 {
			focusableWidgets[0].Focus(true)
			u.focusedWidget = focusableWidgets[0]
		}
	}
}

// AddWindow adds window w to ui for rendering. It returns a function to remove w from ui.
func (u *UI) AddWindow(w *widget.Window) widget.RemoveWindowFunc {
	return u.AddWindowQuietly(w, true)
}

// AddWindowQuietly adds window w to ui for rendering. It returns a function to remove w from ui.
// This function allows you to specify if you would like it to close any open ephemeralWindows (tooltip/dnd/etc).
func (u *UI) AddWindowQuietly(w *widget.Window, closeEphemeralWindows bool) widget.RemoveWindowFunc {
	if u.addWindow(w) {
		w.GetContainer().GetWidget().ContextMenuEvent.AddHandler(u.handleContextMenu)
		w.GetContainer().GetWidget().FocusEvent.AddHandler(u.handleFocusEvent)
		w.GetContainer().GetWidget().ToolTipEvent.AddHandler(u.handleToolTipEvent)
		w.GetContainer().GetWidget().DragAndDropEvent.AddHandler(u.handleDragAndDropEvent)

		if w.Modal && u.focusedWidget != nil {
			u.focusedWidget.Focus(false)
		}
		if closeEphemeralWindows {
			// Close all Ephemeral Windows (tooltip/dnd/etc)
			u.closeEphemeralWindows(0)
		}
	}

	return w.GetCloseFunction()
}

func (u *UI) addWindow(w *widget.Window) bool {
	if u.IsWindowOpen(w) {
		w.Reopen() // a window still on its way out comes back
		return false
	}

	// A window without a theme of its own takes the UI's, through
	// widget.SetUITheme, and so follows it when it changes. The UI hands
	// its theme out here too, for windows added before its first update.
	u.setTheme()
	w.GetContainer().Validate()

	closeFunc := func() {
		u.removeWindow(w)
	}
	w.SetCloseFunction(closeFunc)

	u.windows = append(u.windows, w)

	u.SortWindows()

	return true
}

func (u *UI) removeWindow(w *widget.Window) {
	windowIdx := -1
	for i := range u.windows {
		if u.windows[i] == w {
			u.windows = append(u.windows[:i], u.windows[i+1:]...)
			windowIdx = i
			break
		}
	}
	if windowIdx != -1 {
		w.ReleaseBuffers() // gone from the screen: what its widgets draw through goes back
	}
	if windowIdx != -1 && !w.Ephemeral {
		u.closeEphemeralWindows(windowIdx)
	}
}

// Used to close tooltips/dnd etc.
func (u *UI) closeEphemeralWindows(windowIdx int) {
	for i := len(u.windows) - 1; i >= windowIdx; i-- {
		if u.windows[i].Ephemeral {
			u.windows = append(u.windows[:i], u.windows[i+1:]...)
		}
	}
}

// This function returns true if the provided window object is currently active in this UI.
func (u *UI) IsWindowOpen(w *widget.Window) bool {
	return slices.Contains(u.windows, w)
}

// This function will re-sort the current windows attached to this UI based on its DrawLayer value.
func (u *UI) SortWindows() {
	sort.SliceStable(u.windows, func(i, j int) bool {
		return u.windows[i].DrawLayer < u.windows[j].DrawLayer
	})
}

// This function will return true if any widget is currently focused or a Modal window is open.
func (u *UI) HasFocus() bool {
	for _, v := range slices.Backward(u.windows) {
		if v.Modal {
			return true
		}
	}
	return u.focusedWidget != nil
}

// This function will unfocus the currently focused widget.
func (u *UI) ClearFocus() {
	if u.focusedWidget != nil {
		u.focusedWidget.Focus(false)
		u.focusedWidget = nil
	}
}

// This function will return the currently focused widget if available otherwise it returns nil.
func (u *UI) GetFocusedWidget() widget.Focuser {
	return u.focusedWidget
}

// This function will set a specific focusable widget as the currently focused widget.
func (u *UI) SetFocusedWidget(focused widget.Focuser) {
	if u.focusedWidget != nil {
		u.focusedWidget.Focus(false)
	}

	u.focusedWidget = focused

	if u.focusedWidget != nil {
		u.focusedWidget.Focus(true)
	}
}

// SetDebugMode enables or disables DebugMode which shows the
// margins of the Widgets when hovering the cursor on them.
func (u *UI) SetDebugMode(dm bool) {
	u.debugMode = dm
}

// GetDebugMode get's the current value of DebugMode.
func (u *UI) GetDebugMode() bool {
	return u.debugMode
}
