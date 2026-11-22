package widget

import (
	img "image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"

	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type ListParams struct {
	EntryFace                   *text.Face
	EntryColor                  *ListEntryColor
	EntryImage                  *ListEntryImage
	EntryTextPadding            *Insets
	EntryTextHorizontalPosition *TextPosition
	EntryTextVerticalPosition   *TextPosition
	ControlWidgetSpacing        *int
	MinSize                     *img.Point

	AllowReselect      *bool
	SelectFocus        *bool
	SelectPressed      *bool
	DisableDefaultKeys *bool

	Slider                 *SliderParams
	ScrollContainerImage   *ScrollContainerImage
	ScrollContainerPadding *Insets

	entryUnselectedColor     *ButtonImage
	entrySelectedColor       *ButtonImage
	entryUnselectedTextColor *ButtonTextColor
	entryTextColor           *ButtonTextColor
}

type List struct {
	definedParams  ListParams
	computedParams ListParams

	EntrySelectedEvent *event.Event

	containerOpts        []ContainerOpt
	hideHorizontalSlider bool
	hideVerticalSlider   bool
	slidersWhenNeeded    bool
	dragToScroll         bool

	entries         []any
	entryLabelFunc  ListEntryLabelFunc
	entryWidgetFunc ListEntryWidgetFunc
	entrySortFunc   ListEntrySortFunc

	init            *MultiOnce
	container       *Container
	rows            *listRows
	scrollContainer *ScrollContainer
	layout          *GridLayout
	vSlider         *Slider
	hSlider         *Slider
	selectedEntry   any
	validated       bool

	focused        bool
	tabOrder       int
	focusIndex     int
	prevFocusIndex int

	focusMap map[FocusDirection]Focuser
}

type ListOpt func(l *List)

type ListEntryLabelFunc func(e any) string

// ListEntryWidgetFunc returns the widget an entry shows, in place of a label.
type ListEntryWidgetFunc func(e any) PreferredSizeLocateableWidget
type ListEntrySortFunc func(a, b any) int

type ListEntryColor struct {
	Unselected                 color.Color
	Selected                   color.Color
	DisabledUnselected         color.Color
	DisabledSelected           color.Color
	SelectingBackground        color.Color
	SelectedBackground         color.Color
	FocusedBackground          color.Color
	SelectingFocusedBackground color.Color
	SelectedFocusedBackground  color.Color
	DisabledSelectedBackground color.Color
}

// ListEntryImage draws the entries' backgrounds, such as rounded or
// bordered ones, each in place of its color in ListEntryColor; a state
// without an image keeps its color. Unselected entries have no color: without
// an image they stay clear. A theme's images apply to a list given neither
// colors nor images.
type ListEntryImage struct {
	Unselected         *image.NineSlice
	DisabledUnselected *image.NineSlice
	Focused            *image.NineSlice
	Selecting          *image.NineSlice
	Selected           *image.NineSlice
	DisabledSelected   *image.NineSlice
	SelectedFocused    *image.NineSlice
	SelectingFocused   *image.NineSlice
}

type ListEntrySelectedEventArgs struct {
	List          *List
	Entry         any
	PreviousEntry any
}

type ListEntrySelectedHandlerFunc func(args *ListEntrySelectedEventArgs)

type ListOptions struct {
}

var ListOpts ListOptions

func NewList(opts ...ListOpt) *List {
	l := &List{
		EntrySelectedEvent: &event.Event{},

		init:           &MultiOnce{},
		focusIndex:     0,
		prevFocusIndex: -1,
		focusMap:       make(map[FocusDirection]Focuser),
	}

	l.init.Append(l.createWidget)

	for _, o := range opts {
		o(l)
	}

	return l
}

func (l *List) Validate() {
	l.init.Do()
	l.populateComputedParams()

	if l.entryWidgetFunc == nil && l.computedParams.EntryFace == nil {
		panic("List: EntryFontFace is required.")
	}
	if l.entryWidgetFunc == nil && l.entryLabelFunc == nil {
		panic("List: EntryLabelFunc is required.")
	}
	l.initWidget()
	l.resetFocusIndex()
	l.validated = true
}

func (l *List) populateComputedParams() {
	params := ListParams{}

	theme := l.GetWidget().GetTheme()

	// Set theme values
	if theme != nil {
		if theme.ListTheme != nil {
			params.AllowReselect = theme.ListTheme.AllowReselect
			params.ControlWidgetSpacing = theme.ListTheme.ControlWidgetSpacing
			params.DisableDefaultKeys = theme.ListTheme.DisableDefaultKeys
			if theme.ListTheme.EntryFace != nil {
				params.EntryFace = theme.ListTheme.EntryFace
			} else {
				params.EntryFace = theme.DefaultFace
			}
			params.EntryColor = theme.ListTheme.EntryColor
			params.EntryImage = theme.ListTheme.EntryImage
			params.EntryTextHorizontalPosition = theme.ListTheme.EntryTextHorizontalPosition
			params.EntryTextPadding = theme.ListTheme.EntryTextPadding
			params.EntryTextVerticalPosition = theme.ListTheme.EntryTextVerticalPosition
			params.ScrollContainerImage = theme.ListTheme.ScrollContainerImage
			params.ScrollContainerPadding = theme.ListTheme.ScrollContainerPadding
			params.SelectFocus = theme.ListTheme.SelectFocus
			params.SelectPressed = theme.ListTheme.SelectPressed
			params.Slider = theme.ListTheme.Slider
			params.MinSize = theme.ListTheme.MinSize
		}
	}

	// Set definedParam values
	if l.definedParams.AllowReselect != nil {
		params.AllowReselect = l.definedParams.AllowReselect
	}
	if l.definedParams.ControlWidgetSpacing != nil {
		params.ControlWidgetSpacing = l.definedParams.ControlWidgetSpacing
	}
	if l.definedParams.DisableDefaultKeys != nil {
		params.DisableDefaultKeys = l.definedParams.DisableDefaultKeys
	}
	if l.definedParams.EntryFace != nil {
		params.EntryFace = l.definedParams.EntryFace
	}
	if l.definedParams.EntryColor != nil {
		params.EntryColor = l.definedParams.EntryColor
	}
	if l.definedParams.EntryImage != nil || l.definedParams.EntryColor != nil {
		params.EntryImage = l.definedParams.EntryImage // colors given show, not the theme's images
	}
	if l.definedParams.EntryTextHorizontalPosition != nil {
		params.EntryTextHorizontalPosition = l.definedParams.EntryTextHorizontalPosition
	}
	if l.definedParams.EntryTextPadding != nil {
		params.EntryTextPadding = l.definedParams.EntryTextPadding
	}
	if l.definedParams.EntryTextVerticalPosition != nil {
		params.EntryTextVerticalPosition = l.definedParams.EntryTextVerticalPosition
	}
	if l.definedParams.ScrollContainerImage != nil {
		params.ScrollContainerImage = l.definedParams.ScrollContainerImage
	}
	if l.definedParams.ScrollContainerPadding != nil {
		params.ScrollContainerPadding = l.definedParams.ScrollContainerPadding
	}
	if l.definedParams.SelectFocus != nil {
		params.SelectFocus = l.definedParams.SelectFocus
	}
	if l.definedParams.SelectPressed != nil {
		params.SelectPressed = l.definedParams.SelectPressed
	}
	if l.definedParams.MinSize != nil {
		params.MinSize = l.definedParams.MinSize
	}

	if l.definedParams.Slider != nil {
		if params.Slider == nil {
			params.Slider = &SliderParams{}
		}
		if l.definedParams.Slider.FixedHandleSize != nil {
			params.Slider.FixedHandleSize = l.definedParams.Slider.FixedHandleSize
		}
		if l.definedParams.Slider.HandleImage != nil {
			params.Slider.HandleImage = l.definedParams.Slider.HandleImage
		}
		if l.definedParams.Slider.MinHandleSize != nil {
			params.Slider.MinHandleSize = l.definedParams.Slider.MinHandleSize
		}
		if l.definedParams.Slider.TrackImage != nil {
			params.Slider.TrackImage = l.definedParams.Slider.TrackImage
		}
		if l.definedParams.Slider.TrackOffset != nil {
			params.Slider.TrackOffset = l.definedParams.Slider.TrackOffset
		}
		if l.definedParams.Slider.TrackPadding != nil {
			params.Slider.TrackPadding = l.definedParams.Slider.TrackPadding
		}
	}

	// Set defaults
	if params.EntryTextPadding == nil {
		params.EntryTextPadding = &Insets{}
	}
	if params.EntryTextHorizontalPosition == nil {
		pos := TextPositionCenter
		params.EntryTextHorizontalPosition = &pos
	}
	if params.EntryTextVerticalPosition == nil {
		pos := TextPositionCenter
		params.EntryTextVerticalPosition = &pos
	}
	if params.ControlWidgetSpacing == nil {
		spacing := 0
		params.ControlWidgetSpacing = &spacing
	}
	FALSE := false
	if params.AllowReselect == nil {
		params.AllowReselect = &FALSE
	}
	if params.SelectFocus == nil {
		params.SelectFocus = &FALSE
	}
	if params.SelectPressed == nil {
		params.SelectPressed = &FALSE
	}
	if params.DisableDefaultKeys == nil {
		params.DisableDefaultKeys = &FALSE
	}
	if params.ScrollContainerPadding == nil {
		params.ScrollContainerPadding = &Insets{}
	}

	params.entryUnselectedColor, params.entrySelectedColor = entryBackgrounds(params.EntryColor, params.EntryImage)

	params.entryUnselectedTextColor = &ButtonTextColor{
		Idle:     params.EntryColor.Unselected,
		Disabled: params.EntryColor.DisabledUnselected,
	}

	params.entryTextColor = &ButtonTextColor{
		Idle:     params.EntryColor.Selected,
		Disabled: params.EntryColor.DisabledSelected,
	}

	l.computedParams = params
}

// entryBackgrounds returns the backgrounds of unselected and of selected
// entries, idle, disabled, focused (Hover) and being selected (Pressed):
// each state's image, or else its color. Either may be nil.
func entryBackgrounds(colors *ListEntryColor, images *ListEntryImage) (unselected, selected *ButtonImage) {
	if colors == nil {
		colors = &ListEntryColor{}
	}
	if images == nil {
		images = &ListEntryImage{}
	}
	unselected = &ButtonImage{
		Idle:     entryBackground(images.Unselected, color.Transparent),
		Disabled: entryBackground(images.DisabledUnselected, color.Transparent),
		Hover:    entryBackground(images.Focused, colors.FocusedBackground),
		Pressed:  entryBackground(images.Selecting, colors.SelectingBackground),
	}
	selected = &ButtonImage{
		Idle:     entryBackground(images.Selected, colors.SelectedBackground),
		Disabled: entryBackground(images.DisabledSelected, colors.DisabledSelectedBackground),
		Hover:    entryBackground(images.SelectedFocused, colors.SelectedFocusedBackground),
		Pressed:  entryBackground(images.SelectingFocused, colors.SelectingFocusedBackground),
	}
	return unselected, selected
}

// entryBackground returns an entry's background in one state: its image, or
// else its color.
func entryBackground(i *image.NineSlice, c color.Color) *image.NineSlice {
	if i != nil {
		return i
	}
	return image.NewNineSliceColor(c)
}

func (o ListOptions) ContainerOpts(opts ...ContainerOpt) ListOpt {
	return func(l *List) {
		l.containerOpts = append(l.containerOpts, opts...)
	}
}

// Specify the images for the scroll container.
func (o ListOptions) ScrollContainerImage(image *ScrollContainerImage) ListOpt {
	return func(l *List) {
		l.definedParams.ScrollContainerImage = image
	}
}

// Specify the padding for the scroll container.
func (o ListOptions) ScrollContainerPadding(padding *Insets) ListOpt {
	return func(l *List) {
		l.definedParams.ScrollContainerPadding = padding
	}
}

// Specify the options for the scroll bars.
func (o ListOptions) SliderParams(sliderParams *SliderParams) ListOpt {
	return func(l *List) {
		l.definedParams.Slider = sliderParams
	}
}

func (o ListOptions) ControlWidgetSpacing(s int) ListOpt {
	return func(l *List) {
		l.definedParams.ControlWidgetSpacing = &s
	}
}

func (o ListOptions) HideHorizontalSlider() ListOpt {
	return func(l *List) {
		l.hideHorizontalSlider = true
	}
}

func (o ListOptions) HideVerticalSlider() ListOpt {
	return func(l *List) {
		l.hideVerticalSlider = true
	}
}

// DragToScroll lets the pointer scroll the entries by dragging them, as
// ScrollContainerOpts.DragToScroll does; the sliders follow.
func (o ListOptions) DragToScroll() ListOpt {
	return func(l *List) {
		l.dragToScroll = true
	}
}

// SlidersWhenNeeded shows each slider only while the entries overflow the
// list its way, so a list whose entries fit shows none; a hidden slider
// keeps its room, so the entries do not move when it shows.
func (o ListOptions) SlidersWhenNeeded() ListOpt {
	return func(l *List) {
		l.slidersWhenNeeded = true
	}
}

func (o ListOptions) Entries(e []any) ListOpt {
	return func(l *List) {
		l.entries = l.normalizeEntries(e)
	}
}

func (o ListOptions) EntryLabelFunc(f ListEntryLabelFunc) ListOpt {
	return func(l *List) {
		l.entryLabelFunc = f
	}
}

// EntryWidgetFunc shows each entry as a widget f returns, such as a
// container of an icon and texts, in place of a label: the list draws its
// selected and unselected backgrounds behind it and selects the entry when
// it is pressed. UpdateEntry asks f for the entry's widget again. With it,
// neither EntryLabelFunc nor EntryFontFace is required.
func (o ListOptions) EntryWidgetFunc(f ListEntryWidgetFunc) ListOpt {
	return func(l *List) {
		l.entryWidgetFunc = f
	}
}

func (o ListOptions) EntrySortFunc(f ListEntrySortFunc) ListOpt {
	return func(l *List) {
		l.entrySortFunc = f
		l.entries = l.normalizeEntries(l.entries)
	}
}

func (o ListOptions) EntryFontFace(f *text.Face) ListOpt {
	return func(l *List) {
		l.definedParams.EntryFace = f
	}
}

func (o ListOptions) DisableDefaultKeys(val bool) ListOpt {
	return func(l *List) {
		l.definedParams.DisableDefaultKeys = &val
	}
}

func (o ListOptions) EntryColor(c *ListEntryColor) ListOpt {
	return func(l *List) {
		l.definedParams.EntryColor = c
	}
}

// EntryImage draws the entries' backgrounds with images, such as rounded or
// bordered ones, in place of their colors; see ListEntryImage.
func (o ListOptions) EntryImage(i *ListEntryImage) ListOpt {
	return func(l *List) {
		l.definedParams.EntryImage = i
	}
}

func (o ListOptions) EntryTextPadding(i *Insets) ListOpt {
	return func(l *List) {
		l.definedParams.EntryTextPadding = i
	}
}

// EntryTextPosition sets the position of the text for entries.
// Defaults to both TextPositionCenter.
func (o ListOptions) EntryTextPosition(h TextPosition, v TextPosition) ListOpt {
	return func(l *List) {
		l.definedParams.EntryTextHorizontalPosition = &h
		l.definedParams.EntryTextVerticalPosition = &v
	}
}

func (o ListOptions) EntrySelectedHandler(f ListEntrySelectedHandlerFunc) ListOpt {
	return func(l *List) {
		l.EntrySelectedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*ListEntrySelectedEventArgs); ok {
				f(arg)
			}
		})
	}
}

func (o ListOptions) AllowReselect() ListOpt {
	return func(l *List) {
		l.definedParams.AllowReselect = new(true)
	}
}

// SelectFocus automatically selects each focused entry.
func (o ListOptions) SelectFocus() ListOpt {
	return func(l *List) {
		l.definedParams.SelectFocus = new(true)
	}
}

// SelectPressed selects entries when pressing instead of releasing (the default).
func (o ListOptions) SelectPressed() ListOpt {
	return func(l *List) {
		l.definedParams.SelectPressed = new(true)
	}
}

func (o ListOptions) TabOrder(tabOrder int) ListOpt {
	return func(l *List) {
		l.tabOrder = tabOrder
	}
}

func (l *List) GetWidget() *Widget {
	l.init.Do()
	return l.container.GetWidget()
}

func (l *List) PreferredSize() (int, int) {
	l.init.Do()
	return l.container.PreferredSize()
}

func (l *List) SetLocation(rect img.Rectangle) {
	l.init.Do()
	l.container.GetWidget().Rect = rect
}

func (l *List) RequestRelayout() {
	l.init.Do()
	l.container.RequestRelayout()
}

func (l *List) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	l.init.Do()
	l.container.SetupInputLayer(def)
}

func (l *List) Render(screen *ebiten.Image) {
	l.init.Do()

	d := l.container.GetWidget().Disabled
	if l.vSlider != nil {
		l.vSlider.DrawTrackDisabled = d
	}
	if l.hSlider != nil {
		l.hSlider.DrawTrackDisabled = d
	}
	l.scrollContainer.GetWidget().Disabled = d

	if l.focusIndex != l.prevFocusIndex && l.focusIndex >= 0 && l.focusIndex < len(l.entries) {
		l.scrollVisible(l.rows.rowRect(l.focusIndex))
	}

	if *l.computedParams.SelectFocus {
		l.SelectFocused()
	}

	if l.slidersWhenNeeded {
		l.showSlidersWhenNeeded()
	}
	l.syncVerticalSlider()

	l.container.Render(screen)
}

// showSlidersWhenNeeded shows each slider while the entries overflow the
// list its way, and hides it, keeping its room, while they fit. Before the
// list is first laid out its view has no size, and the sliders stay as
// they start, hidden.
func (l *List) showSlidersWhenNeeded() {
	view := l.scrollContainer.ViewRect()
	if view.Empty() {
		return
	}
	w, h := l.rows.PreferredSize()
	show := func(s *Slider, needed bool) {
		v := VisibilityHideBlocking
		if needed {
			v = VisibilityShow
		}
		s.GetWidget().SetVisibility(v)
	}
	if l.vSlider != nil {
		show(l.vSlider, h > view.Dy())
	}
	if l.hSlider != nil {
		show(l.hSlider, w > view.Dx())
	}
}

func (l *List) Update(updObj *UpdateObject) {
	l.init.Do()

	l.handleInput(updObj.Actions)
	l.container.Update(updObj)
}

/** Focuser Interface - Start **/

func (l *List) Focus(focused bool) {
	l.init.Do()
	l.GetWidget().FireFocusEvent(l, focused, img.Point{-1, -1})
	l.focused = focused
}

func (l *List) IsFocused() bool {
	return l.focused
}

func (l *List) TabOrder() int {
	return l.tabOrder
}

func (l *List) GetFocus(direction FocusDirection) Focuser {
	return l.focusMap[direction]
}

func (l *List) AddFocus(direction FocusDirection, focus Focuser) {
	l.focusMap[direction] = focus
}

/** Focuser Interface - End **/

// handleInput moves the focused list through its entries with Up and Down,
// leaving them at its ends so a gamepad moves the focus on from there, and
// selects the entry it is on with Activate.
func (l *List) handleInput(actions *Actions) {
	if l.focused && !l.GetWidget().Disabled && len(l.entries) > 0 && !*l.computedParams.DisableDefaultKeys {
		if l.focusIndex < len(l.entries)-1 && actions.Take(ActionDown) {
			l.FocusNext()
		}
		if l.focusIndex > 0 && actions.Take(ActionUp) {
			l.FocusPrevious()
		}
		// The player chooses the focused entry, as a click on it would.
		if actions.Take(ActionActivate) || input.KeyJustPressed(ebiten.KeyEnter) || input.KeyJustPressed(ebiten.KeySpace) {
			l.setSelectedEntry(l.entries[l.focusIndex], true)
		}
	}
}

// FocusNext moves the focus to the next entry, from the last to the first.
func (l *List) FocusNext() { l.moveFocus(1) }

// FocusPrevious moves the focus to the entry before, from the first to the
// last.
func (l *List) FocusPrevious() { l.moveFocus(-1) }

func (l *List) moveFocus(by int) {
	if n := len(l.entries); n > 0 {
		l.prevFocusIndex = l.focusIndex
		l.focusIndex = ((l.focusIndex+by)%n + n) % n
	}
}

// SelectFocused selects the focused entry. It scrolls nothing: the list
// scrolls to an entry as the focus moves to it, so with SelectFocus, which
// calls this as the list draws, the wheel and the sliders still scroll away
// from the selected entry.
func (l *List) SelectFocused() {
	if l.focusIndex >= 0 && l.focusIndex < len(l.entries) {
		l.setSelectedEntry(l.entries[l.focusIndex], false)
	}
}

func (l *List) resetFocusIndex() {
	if len(l.entries) > 0 {
		for i := range len(l.entries) {
			if l.entries[i] == l.selectedEntry {
				if i != l.focusIndex {
					l.prevFocusIndex = l.focusIndex
					l.focusIndex = i
				}
				return
			}
		}
		l.focusIndex = 0
	}
}

func (l *List) createWidget() {
	var cols int
	if l.hideVerticalSlider {
		cols = 1
	} else {
		cols = 2
	}
	l.layout = NewGridLayout(
		GridLayoutOpts.Columns(cols),
		GridLayoutOpts.Stretch([]bool{true, false}, []bool{true, false}))

	l.container = NewContainer(
		append([]ContainerOpt{
			ContainerOpts.WidgetOpts(WidgetOpts.TrackHover(true)),
			ContainerOpts.Layout(l.layout)}, l.containerOpts...)...,
	)
}

func (l *List) initWidget() {
	l.container.RemoveChildren()

	l.layout.columnSpacing = *l.computedParams.ControlWidgetSpacing
	l.layout.rowSpacing = *l.computedParams.ControlWidgetSpacing

	if l.computedParams.MinSize != nil {
		l.container.GetWidget().MinWidth = l.computedParams.MinSize.X
		l.container.GetWidget().MinHeight = l.computedParams.MinSize.Y
	}

	l.rows = newListRows(l)
	l.rows.changed()

	scrollOpts := []ScrollContainerOpt{
		ScrollContainerOpts.Content(l.rows),
		ScrollContainerOpts.StretchContentWidth(),
		ScrollContainerOpts.Image(l.computedParams.ScrollContainerImage),
		ScrollContainerOpts.Padding(l.computedParams.ScrollContainerPadding),
	}
	if l.dragToScroll {
		scrollOpts = append(scrollOpts, ScrollContainerOpts.DragToScroll())
	}
	l.scrollContainer = NewScrollContainer(scrollOpts...)

	if l.dragToScroll {
		l.scrollContainer.ScrollMovedEvent.AddHandler(func(args any) {
			if a, ok := args.(*ScrollContainerScrollMovedEventArgs); ok {
				if l.vSlider != nil {
					l.vSlider.Current = int(math.Round(a.Top * float64(l.verticalSpan())))
				}
				if l.hSlider != nil {
					l.hSlider.Current = int(math.Round(a.Left * 1000))
				}
			}
		})
	}

	l.container.AddChild(l.scrollContainer)

	var sliderOpts []SliderOpt
	if l.computedParams.Slider != nil {
		if l.computedParams.Slider.FixedHandleSize != nil {
			sliderOpts = append(sliderOpts, SliderOpts.FixedHandleSize(*l.computedParams.Slider.FixedHandleSize))
		} else {
			sliderOpts = append(sliderOpts, SliderOpts.FixedHandleSize(0))
		}
		if l.computedParams.Slider.HandleImage != nil {
			sliderOpts = append(sliderOpts, SliderOpts.HandleImage(l.computedParams.Slider.HandleImage))
		}
		if l.computedParams.Slider.TrackImage != nil {
			sliderOpts = append(sliderOpts, SliderOpts.TrackImage(l.computedParams.Slider.TrackImage))
		}
		if l.computedParams.Slider.MinHandleSize != nil {
			sliderOpts = append(sliderOpts, SliderOpts.MinHandleSize(*l.computedParams.Slider.MinHandleSize))
		}
		if l.computedParams.Slider.TrackOffset != nil {
			sliderOpts = append(sliderOpts, SliderOpts.TrackOffset(*l.computedParams.Slider.TrackOffset))
		}
		if l.computedParams.Slider.TrackPadding != nil {
			sliderOpts = append(sliderOpts, SliderOpts.TrackPadding(l.computedParams.Slider.TrackPadding))
		}
	}
	if l.computedParams.DisableDefaultKeys != nil {
		sliderOpts = append(sliderOpts, SliderOpts.DisableDefaultKeys(*l.computedParams.DisableDefaultKeys))
	}

	if !l.hideVerticalSlider {
		// The vertical slider counts pixels from the top, so it reaches every
		// row of a long list; a page is the view's height in its steps,
		// scaled so the handle is to the track what the view is to the list.
		pageSizeFunc := func() int {
			_, total := l.rows.PreferredSize()
			if total <= 0 {
				return 1
			}
			view := float64(l.scrollContainer.ViewRect().Dy())
			return max(1, int(math.Round(view*float64(l.vSlider.Max+1)/float64(total))))
		}

		l.vSlider = NewSlider(append(sliderOpts, []SliderOpt{
			SliderOpts.Direction(DirectionVertical),
			SliderOpts.MinMax(0, 0), // the span, set as the list draws
			SliderOpts.PageSizeFunc(pageSizeFunc),
			SliderOpts.ChangedHandler(func(args *SliderChangedEventArgs) {
				l.scrollContainer.ScrollTop = 0
				if span := l.verticalSpan(); span > 0 {
					l.scrollContainer.ScrollTop = float64(args.Slider.Current) / float64(span)
				}
			}),
		}...)...)
		l.container.AddChild(l.vSlider)
		if l.slidersWhenNeeded {
			l.vSlider.GetWidget().SetVisibility(VisibilityHideBlocking)
		}

		// The wheel scrolls the pixels the platform's settings give, or,
		// from a device that gives only the wheel's units, three rows a
		// notch. Less than a pixel waits for more.
		var rest scrollRest
		l.scrollContainer.widget.ScrolledEvent.AddHandler(func(args any) {
			if a, ok := args.(*ScrolledEventArgs); ok {
				dy := a.DY
				if dy == 0 {
					dy = a.Y * float64(3*l.rows.rowStep())
				}
				l.vSlider.Current -= rest.take(dy)
			}
		})
	}

	if !l.hideHorizontalSlider {
		l.hSlider = NewSlider(append(sliderOpts, []SliderOpt{
			SliderOpts.Direction(DirectionHorizontal),
			SliderOpts.MinMax(0, 1000),
			SliderOpts.PageSizeFunc(func() int {
				return int(math.Round(float64(l.scrollContainer.ViewRect().Dx()) / float64(max(1, l.rows.GetWidget().Rect.Dx())) * 1000))
			}),
			SliderOpts.ChangedHandler(func(args *SliderChangedEventArgs) {
				l.scrollContainer.ScrollLeft = float64(args.Slider.Current) / 1000
			}),
		}...)...)
		l.container.AddChild(l.hSlider)
		if l.slidersWhenNeeded {
			l.hSlider.GetWidget().SetVisibility(VisibilityHideBlocking)
		}
	}

	if l.selectedEntry != nil {
		se := l.selectedEntry
		l.selectedEntry = nil
		l.SetSelectedEntry(se)
	}
}

// Updates the entries in the list.
// Note: Duplicates will be removed.
func (l *List) SetEntries(newEntries []any) {
	l.entries = l.normalizeEntries(newEntries)
	if l.validated {
		l.rows.changed()
	}

	l.selectedEntry = nil
	l.resetFocusIndex()
}

// UpdateEntry shows the entry again, as it changed: its label, or a widget
// made for it again.
func (l *List) UpdateEntry(entry any) {
	l.init.Do()
	if l.validated {
		l.rows.remake(entry)
	}
}

// Remove the passed in entry from the list if it exists.
func (l *List) RemoveEntry(entry any) {
	l.init.Do()

	if len(l.entries) > 0 && entry != nil {
		for i, e := range l.entries {
			if e == entry {
				l.entries = append(l.entries[:i], l.entries[i+1:]...)
				if l.validated {
					l.rows.changed()
				}

				entryLen := len(l.entries)
				if l.focusIndex >= entryLen {
					l.focusIndex = i - 1
				}

				if l.focusIndex >= 0 && l.focusIndex < entryLen {
					l.setSelectedEntry(l.entries[l.focusIndex], false)
				}
				break
			}
		}
		l.resetFocusIndex()
	}
}

// Add a new entry to the end of the list
// Note: Duplicates will not be added.
func (l *List) AddEntry(entry any) {
	l.init.Do()
	if !l.checkForDuplicates(l.entries, entry) {
		if l.entrySortFunc != nil {
			index, _ := slices.BinarySearchFunc(l.entries, entry, l.entrySortFunc)
			l.entries = slices.Insert(l.entries, index, entry)
		} else {
			l.entries = append(l.entries, entry)
		}
		if l.validated {
			l.rows.changed()
		}
	}
	l.resetFocusIndex()
}

// Return the current entries in the list.
func (l *List) Entries() []any {
	l.init.Do()
	return l.entries
}

// Return the currently selected entry in the list.
func (l *List) SelectedEntry() any {
	l.init.Do()
	return l.selectedEntry
}

// Set the Selected Entry to e if it is found.
func (l *List) SetSelectedEntry(entry any) {
	l.setSelectedEntry(entry, false)
}

func (l *List) setSelectedEntry(e any, user bool) {
	if e != l.selectedEntry || (user && *l.computedParams.AllowReselect) {
		l.init.Do()

		prev := l.selectedEntry
		l.selectedEntry = e
		l.resetFocusIndex()

		l.EntrySelectedEvent.Fire(&ListEntrySelectedEventArgs{
			List:          l,
			Entry:         e,
			PreviousEntry: prev,
		})
	}
}

func (l *List) checkForDuplicates(entries []any, entry any) bool {
	return slices.ContainsFunc(entries, func(cmp any) bool {
		return cmp == entry
	})
}

// normalizeEntries returns the entries without duplicates, in their
// order, sorted if the list sorts them.
func (l *List) normalizeEntries(entries []any) []any {
	normalized := make([]any, 0, len(entries))
	seen := make(map[any]bool, len(entries))
	for _, entry := range entries {
		if !seen[entry] {
			seen[entry] = true
			normalized = append(normalized, entry)
		}
	}
	if l.entrySortFunc != nil {
		slices.SortStableFunc(normalized, l.entrySortFunc)
	}
	return normalized
}

func (l *List) setScrollTop(t float64) {
	l.init.Do()
	if l.vSlider != nil {
		l.vSlider.Current = int(math.Round(t * float64(l.verticalSpan())))
	}
	l.scrollContainer.ScrollTop = t
}

// verticalSpan returns how many pixels the rows scroll: their height less
// the view's.
func (l *List) verticalSpan() int {
	_, h := l.rows.PreferredSize()
	return max(0, h-l.scrollContainer.ViewRect().Dy())
}

// syncVerticalSlider gives the vertical slider the rows' span, as entries
// come and go or the view changes size, keeping the pixel it is at, so the
// entry at the top stays in place; the slider clamps as it draws.
func (l *List) syncVerticalSlider() {
	if l.vSlider == nil {
		return
	}
	span := l.verticalSpan()
	l.vSlider.Max = span
	l.scrollContainer.ScrollTop = 0
	if span > 0 {
		l.scrollContainer.ScrollTop = float64(min(l.vSlider.Current, span)) / float64(span)
	}
}

func (l *List) setScrollLeft(left float64) {
	l.init.Do()
	if l.hSlider != nil {
		l.hSlider.Current = int(math.Round(left * 1000))
	}
	l.scrollContainer.ScrollLeft = left
}

// scrollVisible scrolls toward showing w, a step a call: along each axis as
// little as shows it whole, or to its start if it is larger than the view
// that way. Once there, within a pixel, it stops scrolling for the focused
// entry, so the wheel and the sliders scroll freely again.
// scrollVisible scrolls, a step at most, toward showing wrect, a row laid
// out in the content.
func (l *List) scrollVisible(wrect img.Rectangle) {
	vrect, crect := l.scrollContainer.ViewRect(), l.scrollContainer.ContentRect()
	top, left := l.scrollContainer.ScrollTop, l.scrollContainer.ScrollLeft
	spanY, spanX := crect.Dy()-vrect.Dy(), crect.Dx()-vrect.Dx()
	toTop := scrollTarget(wrect.Min.Y, wrect.Max.Y, vrect.Min.Y, vrect.Max.Y, crect.Min.Y, spanY, top)
	toLeft := scrollTarget(wrect.Min.X, wrect.Max.X, vrect.Min.X, vrect.Max.X, crect.Min.X, spanX, left)
	if math.Abs(toTop-top)*float64(spanY) < 1 && math.Abs(toLeft-left)*float64(spanX) < 1 {
		l.prevFocusIndex = l.focusIndex // there
		return
	}
	l.setScrollTop(scrollClamp(toTop, top))
	l.setScrollLeft(scrollClamp(toLeft, left))
}

// scrollTarget returns the scroll along one axis, from 0 to 1, that shows
// the span lo to hi of the content: as little a move from current as shows
// it whole within the view's span vlo to vhi, or its start at the view's if
// it is larger than the view. The content starts at clo and scrolls span
// pixels.
func scrollTarget(lo, hi, vlo, vhi, clo, span int, current float64) float64 {
	if span <= 0 {
		return 0
	}
	target := current
	switch {
	case hi-lo > vhi-vlo || lo < vlo: // larger than the view, or before it
		target = float64(lo-clo) / float64(span)
	case hi > vhi: // after it
		target = float64(hi-(vhi-vlo)-clo) / float64(span)
	}
	return math.Max(0, math.Min(1, target))
}

func scrollClamp(targetScroll, currentScroll float64) float64 {
	const maxScrollStep = 0.1
	minScroll := currentScroll - maxScrollStep
	maxScroll := currentScroll + maxScrollStep
	return math.Max(minScroll, math.Min(targetScroll, maxScroll))
}
