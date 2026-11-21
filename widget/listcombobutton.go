package widget

import (
	"image"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/utilities/constantutil"
	"golang.org/x/exp/slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type ListComboButtonParams struct {
	List               *ListParams
	Button             *ButtonParams
	DisableDefaultKeys *bool
	MaxContentHeight   *int
}

type ListComboButton struct {
	definedParams  ListComboButtonParams
	computedParams ListComboButtonParams

	EntrySelectedEvent *event.Event

	init       *MultiOnce
	widget     *Widget
	widgetOpts []WidgetOpt
	entries    []any

	button          *SelectComboButton
	list            *List
	buttonLabelFunc SelectComboButtonEntryLabelFunc
	listLabelFunc   ListEntryLabelFunc
	listWidgetFunc  ListEntryWidgetFunc

	selectedEntry any

	tabOrder int
	focusMap map[FocusDirection]Focuser
}

type ListComboButtonOpt func(l *ListComboButton)

type ListComboButtonEntrySelectedEventArgs struct {
	Button        *ListComboButton
	Entry         any
	PreviousEntry any
}

type ListComboButtonEntrySelectedHandlerFunc func(args *ListComboButtonEntrySelectedEventArgs)

type ListComboButtonOptions struct {
}

var ListComboButtonOpts ListComboButtonOptions

func NewListComboButton(opts ...ListComboButtonOpt) *ListComboButton {
	l := &ListComboButton{
		EntrySelectedEvent: &event.Event{},

		init:     &MultiOnce{},
		focusMap: make(map[FocusDirection]Focuser),
	}

	l.init.Append(l.createWidget)

	for _, o := range opts {
		o(l)
	}

	return l
}

func (l *ListComboButton) Validate() {
	l.init.Do()
	l.populateComputedParams()

	l.initWidget()
}

func (l *ListComboButton) populateComputedParams() {
	params := ListComboButtonParams{Button: &ButtonParams{}, List: &ListParams{}}

	theme := l.GetWidget().GetTheme()

	// Set theme values
	if theme != nil {
		if theme.ListComboButtonTheme != nil {
			params.DisableDefaultKeys = theme.ListComboButtonTheme.DisableDefaultKeys
			params.MaxContentHeight = theme.ListComboButtonTheme.MaxContentHeight
			if theme.ListComboButtonTheme.Button != nil {
				params.Button.GraphicImage = theme.ListComboButtonTheme.Button.GraphicImage
				params.Button.GraphicPadding = theme.ListComboButtonTheme.Button.GraphicPadding
				params.Button.Image = theme.ListComboButtonTheme.Button.Image
				params.Button.MinSize = theme.ListComboButtonTheme.Button.MinSize
				params.Button.TextColor = theme.ListComboButtonTheme.Button.TextColor
				if theme.ListComboButtonTheme.Button.TextFace != nil {
					params.Button.TextFace = theme.ListComboButtonTheme.Button.TextFace
				} else {
					params.Button.TextFace = theme.DefaultFace
				}
				params.Button.TextPadding = theme.ListComboButtonTheme.Button.TextPadding
				params.Button.TextPosition = theme.ListComboButtonTheme.Button.TextPosition
			}
			if theme.ListComboButtonTheme.List != nil {
				params.List.AllowReselect = theme.ListComboButtonTheme.List.AllowReselect
				params.List.ControlWidgetSpacing = theme.ListComboButtonTheme.List.ControlWidgetSpacing
				params.List.DisableDefaultKeys = theme.ListComboButtonTheme.List.DisableDefaultKeys
				if theme.ListComboButtonTheme.List.EntryFace != nil {
					params.List.EntryFace = theme.ListComboButtonTheme.List.EntryFace
				} else {
					params.List.EntryFace = theme.DefaultFace
				}
				params.List.EntryColor = theme.ListComboButtonTheme.List.EntryColor
				params.List.EntryImage = theme.ListComboButtonTheme.List.EntryImage
				params.List.EntryTextHorizontalPosition = theme.ListComboButtonTheme.List.EntryTextHorizontalPosition
				params.List.EntryTextPadding = theme.ListComboButtonTheme.List.EntryTextPadding
				params.List.EntryTextVerticalPosition = theme.ListComboButtonTheme.List.EntryTextVerticalPosition
				params.List.ScrollContainerImage = theme.ListComboButtonTheme.List.ScrollContainerImage
				params.List.ScrollContainerPadding = theme.ListComboButtonTheme.List.ScrollContainerPadding
				params.List.SelectFocus = theme.ListComboButtonTheme.List.SelectFocus
				params.List.SelectPressed = theme.ListComboButtonTheme.List.SelectPressed
				params.List.Slider = theme.ListComboButtonTheme.List.Slider
				params.List.MinSize = theme.ListComboButtonTheme.List.MinSize
			}
		}
	}

	// Set definedParam values
	if l.definedParams.DisableDefaultKeys != nil {
		params.DisableDefaultKeys = l.definedParams.DisableDefaultKeys
	}
	if l.definedParams.MaxContentHeight != nil {
		params.MaxContentHeight = l.definedParams.MaxContentHeight
	}
	if l.definedParams.Button != nil {
		if l.definedParams.Button.GraphicImage != nil {
			params.Button.GraphicImage = l.definedParams.Button.GraphicImage
		}
		if l.definedParams.Button.GraphicPadding != nil {
			params.Button.GraphicPadding = l.definedParams.Button.GraphicPadding
		}
		if l.definedParams.Button.Image != nil {
			params.Button.Image = l.definedParams.Button.Image
		}
		if l.definedParams.Button.MinSize != nil {
			params.Button.MinSize = l.definedParams.Button.MinSize
		}
		if l.definedParams.Button.TextColor != nil {
			params.Button.TextColor = l.definedParams.Button.TextColor
		}
		if l.definedParams.Button.TextFace != nil {
			params.Button.TextFace = l.definedParams.Button.TextFace
		}
		if l.definedParams.Button.TextPadding != nil {
			params.Button.TextPadding = l.definedParams.Button.TextPadding
		}
		if l.definedParams.Button.TextPosition != nil {
			params.Button.TextPosition = l.definedParams.Button.TextPosition
		}
	}
	if l.definedParams.List != nil {
		if l.definedParams.List.AllowReselect != nil {
			params.List.AllowReselect = l.definedParams.List.AllowReselect
		}
		if l.definedParams.List.ControlWidgetSpacing != nil {
			params.List.ControlWidgetSpacing = l.definedParams.List.ControlWidgetSpacing
		}
		if l.definedParams.List.DisableDefaultKeys != nil {
			params.List.DisableDefaultKeys = l.definedParams.List.DisableDefaultKeys
		}
		if l.definedParams.List.EntryFace != nil {
			params.List.EntryFace = l.definedParams.List.EntryFace
		}
		if l.definedParams.List.EntryColor != nil {
			params.List.EntryColor = l.definedParams.List.EntryColor
		}
		if l.definedParams.List.EntryImage != nil || l.definedParams.List.EntryColor != nil {
			params.List.EntryImage = l.definedParams.List.EntryImage // colors given show, not the theme's images
		}

		if l.definedParams.List.EntryTextHorizontalPosition != nil {
			params.List.EntryTextHorizontalPosition = l.definedParams.List.EntryTextHorizontalPosition
		}
		if l.definedParams.List.EntryTextPadding != nil {
			params.List.EntryTextPadding = l.definedParams.List.EntryTextPadding
		}
		if l.definedParams.List.EntryTextVerticalPosition != nil {
			params.List.EntryTextVerticalPosition = l.definedParams.List.EntryTextVerticalPosition
		}
		if l.definedParams.List.EntryColor != nil {
			params.List.EntryColor = l.definedParams.List.EntryColor
		}
		if l.definedParams.List.ScrollContainerImage != nil {
			params.List.ScrollContainerImage = l.definedParams.List.ScrollContainerImage
		}
		if l.definedParams.List.ScrollContainerPadding != nil {
			params.List.ScrollContainerPadding = l.definedParams.List.ScrollContainerPadding
		}
		if l.definedParams.List.SelectFocus != nil {
			params.List.SelectFocus = l.definedParams.List.SelectFocus
		}
		if l.definedParams.List.SelectPressed != nil {
			params.List.SelectPressed = l.definedParams.List.SelectPressed
		}
		if l.definedParams.List.Slider != nil {
			params.List.Slider = l.definedParams.List.Slider
		}
		if l.definedParams.List.MinSize != nil {
			params.List.MinSize = l.definedParams.List.MinSize
		}
	}
	// Set defaults
	if params.MaxContentHeight == nil {
		params.MaxContentHeight = constantutil.ConstantToPointer(200)
	}
	if params.DisableDefaultKeys == nil {
		params.DisableDefaultKeys = constantutil.ConstantToPointer(false)
	}

	l.computedParams = params
}

func (o ListComboButtonOptions) WidgetOpts(opts ...WidgetOpt) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.widgetOpts = append(l.widgetOpts, opts...)
	}
}

func (o ListComboButtonOptions) ButtonParams(buttonParams *ButtonParams) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.definedParams.Button = buttonParams
	}
}

func (o ListComboButtonOptions) ListParams(listParams *ListParams) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.definedParams.List = listParams
	}
}

func (o ListComboButtonOptions) Text(face *text.Face, image *GraphicImage, color *ButtonTextColor) ListComboButtonOpt {
	return func(l *ListComboButton) {
		if l.definedParams.Button == nil {
			l.definedParams.Button = &ButtonParams{}
		}
		l.definedParams.Button.GraphicImage = image
		l.definedParams.Button.TextColor = color
		l.definedParams.Button.TextFace = face
	}
}

func (o ListComboButtonOptions) Entries(e []any) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.entries = slices.CompactFunc(e, func(a any, b any) bool { return a == b })
	}
}

func (o ListComboButtonOptions) InitialEntry(e any) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.selectedEntry = e
	}
}

func (o ListComboButtonOptions) EntryLabelFunc(button SelectComboButtonEntryLabelFunc, list ListEntryLabelFunc) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.buttonLabelFunc = button
		l.listLabelFunc = list
	}
}

// EntryWidgetFunc shows each entry of the open list as a widget f returns,
// in place of the list's label (see ListOptions.EntryWidgetFunc). The button
// still shows its label for the selected entry.
func (o ListComboButtonOptions) EntryWidgetFunc(f ListEntryWidgetFunc) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.listWidgetFunc = f
	}
}

func (o ListComboButtonOptions) EntrySelectedHandler(f ListComboButtonEntrySelectedHandlerFunc) ListComboButtonOpt {
	return func(l *ListComboButton) {
		l.EntrySelectedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*ListComboButtonEntrySelectedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o ListComboButtonOptions) MaxContentHeight(h int) ListComboButtonOpt {
	return func(c *ListComboButton) {
		c.definedParams.MaxContentHeight = &h
	}
}

func (o ListComboButtonOptions) TabOrder(tabOrder int) ListComboButtonOpt {
	return func(sl *ListComboButton) {
		sl.tabOrder = tabOrder
	}
}

func (o ListComboButtonOptions) DisableDefaultKeys(val bool) ListComboButtonOpt {
	return func(sl *ListComboButton) {
		sl.definedParams.DisableDefaultKeys = &val
	}
}

/** Focuser Interface - Start **/

func (l *ListComboButton) Focus(focused bool) {
	l.init.Do()
	l.GetWidget().FireFocusEvent(l, focused, image.Point{-1, -1})
	l.button.button.button.focused = focused
	if !focused {
		l.SetContentVisible(false)
	}
}

func (l *ListComboButton) IsFocused() bool {
	return l.button.button.button.focused
}

func (l *ListComboButton) TabOrder() int {
	return l.tabOrder
}

func (l *ListComboButton) GetFocus(direction FocusDirection) Focuser {
	return l.focusMap[direction]
}

func (l *ListComboButton) AddFocus(direction FocusDirection, focus Focuser) {
	l.focusMap[direction] = focus
}

/** Focuser Interface - End **/

func (l *ListComboButton) FocusNext() {
	if l.list != nil {
		l.SetContentVisible(true)
		l.list.FocusNext()
	}
}

func (l *ListComboButton) FocusPrevious() {
	if l.list != nil {
		l.SetContentVisible(true)
		l.list.FocusPrevious()
	}
}

func (l *ListComboButton) SelectFocused() {
	if l.list != nil {
		l.SetContentVisible(true)
		l.list.SelectFocused()
	}
}

func (l *ListComboButton) GetWidget() *Widget {
	l.init.Do()
	return l.widget
}

func (l *ListComboButton) PreferredSize() (int, int) {
	l.init.Do()
	return l.button.PreferredSize()
}

func (l *ListComboButton) SetLocation(rect image.Rectangle) {
	l.init.Do()
	l.widget.Rect = rect
	l.button.SetLocation(rect)
}

func (l *ListComboButton) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	l.init.Do()
	l.button.SetupInputLayer(def)
}

func (l *ListComboButton) Render(screen *ebiten.Image) {
	l.init.Do()

	l.button.Render(screen)
}

func (l *ListComboButton) Update(updObj *UpdateObject) {
	l.init.Do()

	l.widget.Update(updObj) // its own widget: tooltips, hover and the events set on it
	l.button.GetWidget().Disabled = l.widget.Disabled
	l.button.Update(updObj) // the open list takes its actions here
	l.handleActions(updObj.Actions)
}

// handleActions opens the focused combo with Activate, or the Down or Up
// key. Open, it keeps the directions its list leaves, so the focus stays in
// it, and closes with Back.
func (l *ListComboButton) handleActions(actions *Actions) {
	if !l.button.button.button.focused || *l.computedParams.DisableDefaultKeys {
		return
	}
	if !l.ContentVisible() {
		if actions.Take(ActionActivate) || input.KeyJustPressed(ebiten.KeyArrowDown) || input.KeyJustPressed(ebiten.KeyArrowUp) {
			l.SetContentVisible(true)
		}
		return
	}
	for _, d := range [...]Action{ActionUp, ActionDown, ActionLeft, ActionRight} {
		actions.Take(d)
	}
	if actions.Take(ActionBack) {
		l.SetContentVisible(false)
	}
}

func (l *ListComboButton) createWidget() {
	l.widget = NewWidget(l.widgetOpts...)
}

func (l *ListComboButton) initWidget() {
	l.list = NewList(ListOpts.HideHorizontalSlider(), ListOpts.Entries(l.entries), ListOpts.EntryLabelFunc(l.listLabelFunc), ListOpts.EntryWidgetFunc(l.listWidgetFunc))
	l.list.definedParams = *l.computedParams.List
	l.list.definedParams.DisableDefaultKeys = l.computedParams.DisableDefaultKeys
	l.list.definedParams.AllowReselect = constantutil.ConstantToPointer(true)
	l.list.GetWidget().parent = l.widget
	l.list.Validate()
	btnOpts := []ButtonOpt{
		ButtonOpts.Graphic(l.computedParams.Button.GraphicImage),
		ButtonOpts.Image(l.computedParams.Button.Image),
		ButtonOpts.Text("", l.computedParams.Button.TextFace, l.computedParams.Button.TextColor),
	}
	if l.computedParams.Button.MinSize != nil {
		btnOpts = append(btnOpts, ButtonOpts.WidgetOpts(WidgetOpts.MinSize(l.computedParams.Button.MinSize.X, l.computedParams.Button.MinSize.Y)))
	}
	if l.computedParams.Button.TextPadding != nil {
		btnOpts = append(btnOpts, ButtonOpts.TextPadding(l.computedParams.Button.TextPadding))
	}

	l.button = NewSelectComboButton(
		SelectComboButtonOpts.ComboButtonOpts(
			ComboButtonOpts.Content(l.list),
			ComboButtonOpts.MaxContentHeight(*l.computedParams.MaxContentHeight),
			ComboButtonOpts.ButtonOpts(btnOpts...),
		),
		SelectComboButtonOpts.EntryLabelFunc(l.buttonLabelFunc),
	)
	l.button.GetWidget().parent = l.widget
	l.button.Validate()

	l.button.EntrySelectedEvent.AddHandler(func(args any) {
		if a, ok := args.(*SelectComboButtonEntrySelectedEventArgs); ok {
			l.EntrySelectedEvent.Fire(&ListComboButtonEntrySelectedEventArgs{
				Button:        l,
				Entry:         a.Entry,
				PreviousEntry: a.PreviousEntry,
			})
		}
	})

	l.list.EntrySelectedEvent.AddHandler(func(args any) {
		if a, ok := args.(*ListEntrySelectedEventArgs); ok {
			l.SetContentVisible(false)
			l.SetSelectedEntry(a.Entry)
		}
	})

	if l.selectedEntry != nil {
		l.button.SetSelectedEntry(l.selectedEntry)
		l.list.setSelectedEntry(l.selectedEntry, false)
	} else if len(l.list.entries) > 0 {
		firstEntry := l.list.entries[0]
		l.button.SetSelectedEntry(firstEntry)
		l.list.SetSelectedEntry(firstEntry)
	}
}

func (l *ListComboButton) SetSelectedEntry(e any) {
	l.init.Do()
	l.selectedEntry = e
	if l.button != nil && l.list != nil {
		l.button.SetSelectedEntry(e)
		l.list.setSelectedEntry(e, false)
	}
}

func (l *ListComboButton) SelectedEntry() any {
	l.init.Do()
	if l.button == nil {
		return l.selectedEntry
	}
	return l.button.SelectedEntry()
}

func (l *ListComboButton) SetContentVisible(v bool) {
	if l.list != nil {
		l.init.Do()
		l.list.Focus(v)
		l.button.SetContentVisible(v)
		if !v {
			l.list.resetFocusIndex()
		}
	}
}

func (l *ListComboButton) ContentVisible() bool {
	l.init.Do()
	if l.button == nil {
		return false
	}
	return l.button.ContentVisible()
}

func (l *ListComboButton) Label() string {
	l.init.Do()
	if l.button == nil {
		return ""
	}
	return l.button.Label()
}
