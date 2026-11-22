package widget

import (
	img "image"
	"image/color"
	"time"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SearchBarParams is a search bar's look: the field drawn behind it, the
// magnifier at its start, the clear button's icon and face, and the color
// of the icons it draws when given none.
type SearchBarParams struct {
	Background       *image.NineSlice
	Icon             *GraphicImage
	ClearIcon        *GraphicImage
	ClearButtonImage *ButtonImage
	IconColor        color.Color
}

// SearchBar is a field to search with: a magnifier at its start, the text
// typed, and a button at its end, shown while there is text, that clears
// it. It tells of the search once the player pauses typing, 250 ms by
// default, rather than at every key, so a list or a table filters once a
// search (List.SetFilter). Back, from Escape or a gamepad, clears it while
// its input has the focus and holds text.
//
// The text input inside takes its options from TextInputOpts and its face
// and colors from the theme's text inputs; the search bar draws the field
// around it, and the input none of its own.
type SearchBar struct {
	definedParams  SearchBarParams
	computedParams SearchBarParams

	ChangedEvent *event.Event

	inputOpts  []TextInputOpt
	delay      time.Duration
	typed      string // the text as last seen
	fired      string // the search last told of
	waited     int    // the updates since the text last changed
	pending    bool   // whether the text differs from the search told of
	widgetOpts []WidgetOpt

	container *Container
	input     *TextInput
	icon      *Graphic
	clear     *Button

	init *MultiOnce
}

// SearchBarChangedEventArgs tell the search typed.
type SearchBarChangedEventArgs struct {
	SearchBar *SearchBar
	Query     string
}

type SearchBarChangedHandlerFunc func(args *SearchBarChangedEventArgs)

type SearchBarOpt func(s *SearchBar)

type SearchBarOptions struct{}

// SearchBarOpts holds the options of NewSearchBar.
var SearchBarOpts SearchBarOptions

// NewSearchBar returns a search bar telling of a search 250 ms after the
// player stops typing.
func NewSearchBar(opts ...SearchBarOpt) *SearchBar {
	s := &SearchBar{
		ChangedEvent: &event.Event{},
		delay:        250 * time.Millisecond,
		init:         &MultiOnce{},
	}
	for _, o := range opts {
		o(s)
	}
	s.init.Append(s.createWidget)
	return s
}

func (o SearchBarOptions) WidgetOpts(opts ...WidgetOpt) SearchBarOpt {
	return func(s *SearchBar) { s.widgetOpts = append(s.widgetOpts, opts...) }
}

// TextInputOpts sets options of the text input inside: its face, colors,
// placeholder, and the rest.
func (o SearchBarOptions) TextInputOpts(opts ...TextInputOpt) SearchBarOpt {
	return func(s *SearchBar) { s.inputOpts = append(s.inputOpts, opts...) }
}

// Delay sets how long after the last key it tells of the search; 0 tells
// at every key.
func (o SearchBarOptions) Delay(d time.Duration) SearchBarOpt {
	return func(s *SearchBar) { s.delay = d }
}

func (o SearchBarOptions) ChangedHandler(f SearchBarChangedHandlerFunc) SearchBarOpt {
	return func(s *SearchBar) { s.ChangedEvent.AddHandler(handle(f)) }
}

// Background sets the field drawn behind the bar.
func (o SearchBarOptions) Background(i *image.NineSlice) SearchBarOpt {
	return func(s *SearchBar) { s.definedParams.Background = i }
}

// Icon sets the icon at the bar's start, in place of the magnifier it
// draws.
func (o SearchBarOptions) Icon(i *GraphicImage) SearchBarOpt {
	return func(s *SearchBar) { s.definedParams.Icon = i }
}

// ClearIcon sets the clear button's icon, in place of the cross it draws.
func (o SearchBarOptions) ClearIcon(i *GraphicImage) SearchBarOpt {
	return func(s *SearchBar) { s.definedParams.ClearIcon = i }
}

// ClearButtonImage sets the clear button's face; clear by default.
func (o SearchBarOptions) ClearButtonImage(i *ButtonImage) SearchBarOpt {
	return func(s *SearchBar) { s.definedParams.ClearButtonImage = i }
}

// IconColor sets the color of the icons it draws itself.
func (o SearchBarOptions) IconColor(c color.Color) SearchBarOpt {
	return func(s *SearchBar) { s.definedParams.IconColor = c }
}

// TextInput returns the text input inside.
func (s *SearchBar) TextInput() *TextInput {
	s.init.Do()
	return s.input
}

// Query returns the text typed.
func (s *SearchBar) Query() string {
	s.init.Do()
	return s.input.GetText()
}

// SetQuery sets the text and tells of the search at once.
func (s *SearchBar) SetQuery(q string) {
	s.init.Do()
	s.input.SetText(q)
	s.fire(q)
}

// Clear empties the search and tells of it at once.
func (s *SearchBar) Clear() { s.SetQuery("") }

// fire tells of search q.
func (s *SearchBar) fire(q string) {
	s.typed, s.fired, s.waited, s.pending = q, q, 0, false
	s.clear.GetWidget().SetVisibility(visibleIf(q != ""))
	s.ChangedEvent.Fire(&SearchBarChangedEventArgs{SearchBar: s, Query: q})
}

func visibleIf(shown bool) Visibility {
	if shown {
		return VisibilityShow
	}
	return VisibilityHideBlocking
}

func (s *SearchBar) GetWidget() *Widget {
	s.init.Do()
	return s.container.GetWidget()
}

func (s *SearchBar) SetLocation(rect img.Rectangle) {
	s.init.Do()
	s.container.SetLocation(rect)
}

func (s *SearchBar) PreferredSize() (int, int) {
	s.init.Do()
	return s.container.PreferredSize()
}

func (s *SearchBar) RequestRelayout() {
	s.init.Do()
	s.container.RequestRelayout()
}

func (s *SearchBar) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	s.init.Do()
	s.container.SetupInputLayer(def)
}

// GetFocusers returns the text input, and the clear button while shown.
func (s *SearchBar) GetFocusers() []Focuser {
	s.init.Do()
	return s.container.GetFocusers()
}

// PlaceChildren places the icon, the input and the clear button.
func (s *SearchBar) PlaceChildren() {
	s.init.Do()
	s.container.PlaceChildren()
}

// Validate takes the look, from the options, the theme or defaults, and
// gives it to the field, the icon and the clear button.
func (s *SearchBar) Validate() {
	s.init.Do()
	s.computeParams(s.container.GetWidget().GetTheme())
	p := s.computedParams
	s.icon.images = p.Icon
	s.clear.definedParams.Image = p.ClearButtonImage
	s.clear.definedParams.GraphicImage = p.ClearIcon
	s.container.SetBackgroundImage(p.Background) // validates the container, and the children
}

func (s *SearchBar) Render(screen *ebiten.Image) {
	s.init.Do()
	s.container.Render(screen)
}

// Update updates the input, tells of the search once the text has stood
// for Delay, and clears it on a Back its input is focused for.
func (s *SearchBar) Update(updObj *UpdateObject) {
	s.init.Do()
	s.container.Update(updObj)
	q := s.input.GetText()
	if s.input.IsFocused() && q != "" && updObj.Actions.Take(ActionBack) {
		s.Clear()
		return
	}
	if q != s.typed {
		s.typed, s.waited, s.pending = q, 0, q != s.fired
		s.clear.GetWidget().SetVisibility(visibleIf(q != ""))
		if s.pending && s.delay <= 0 {
			s.fire(q)
		}
		return
	}
	if s.pending {
		s.waited++
		if s.waited >= updatesIn(s.delay) {
			s.fire(q)
		}
	}
}

// computeParams takes the look from the options, then the theme, then
// defaults: the theme's text input field, drawn icons in its text color.
func (s *SearchBar) computeParams(theme *Theme) {
	p := SearchBarParams{}
	if theme != nil && theme.SearchBarTheme != nil {
		p = *theme.SearchBarTheme
	}
	var field *image.NineSlice
	iconColor := color.Color(color.White)
	if theme != nil && theme.TextInputTheme != nil {
		if theme.TextInputTheme.Image != nil {
			field = theme.TextInputTheme.Image.Idle
		}
		if theme.TextInputTheme.Color != nil && theme.TextInputTheme.Color.Idle != nil {
			iconColor = theme.TextInputTheme.Color.Idle
		}
	}
	d := s.definedParams
	p.Background = first(d.Background, p.Background, field, image.NewNineSliceColor(color.NRGBA{0, 0, 0, 96}))
	p.IconColor = first(d.IconColor, p.IconColor, iconColor)
	p.Icon = first(d.Icon, p.Icon)
	p.ClearIcon = first(d.ClearIcon, p.ClearIcon)
	clearFace := image.NewNineSliceColor(color.Transparent)
	p.ClearButtonImage = first(d.ClearButtonImage, p.ClearButtonImage, &ButtonImage{Idle: clearFace, Pressed: clearFace})
	if p.Icon == nil {
		p.Icon = &GraphicImage{Idle: magnifier(p.IconColor)}
	}
	if p.ClearIcon == nil {
		p.ClearIcon = &GraphicImage{Idle: cross(p.IconColor)}
	}
	s.computedParams = p
}

// magnifier draws a magnifier 16 pixels wide, once.
func magnifier(c color.Color) *ebiten.Image {
	i := ebiten.NewImage(16, 16)
	vector.StrokeCircle(i, 6.5, 6.5, 5, 2, c, true)
	vector.StrokeLine(i, 10, 10, 15, 15, 2, c, true)
	return i
}

// cross draws a cross 12 pixels wide, once.
func cross(c color.Color) *ebiten.Image {
	i := ebiten.NewImage(12, 12)
	vector.StrokeLine(i, 2, 2, 10, 10, 2, c, true)
	vector.StrokeLine(i, 10, 2, 2, 10, 2, c, true)
	return i
}

func (s *SearchBar) createWidget() {
	s.container = NewContainer(
		ContainerOpts.WidgetOpts(s.widgetOpts...),
		ContainerOpts.Layout(NewGridLayout(
			GridLayoutOpts.Columns(3),
			GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{true}),
			GridLayoutOpts.Spacing(6, 0),
			GridLayoutOpts.Padding(&Insets{Left: 8, Right: 6}),
		)),
	)
	s.widgetOpts = nil

	// The icon's and the clear button's images come as it is validated.
	middle := WidgetOpts.LayoutData(GridLayoutData{VerticalPosition: GridLayoutPositionCenter})
	s.icon = NewGraphic(GraphicOpts.Images(&GraphicImage{}), GraphicOpts.WidgetOpts(middle))
	none := image.NewNineSliceColor(color.Transparent)
	s.input = NewTextInput(append([]TextInputOpt{
		TextInputOpts.Image(&TextInputImage{Idle: none, Disabled: none}),
		TextInputOpts.WidgetOpts(middle),
	}, s.inputOpts...)...)
	s.clear = NewButton(
		ButtonOpts.Graphic(&GraphicImage{}),
		ButtonOpts.WidgetOpts(middle, WidgetOpts.MinSize(20, 20)),
		ButtonOpts.ClickedHandler(func(*ButtonClickedEventArgs) {
			s.Clear()
			s.input.Focus(true)
		}),
	)
	s.clear.GetWidget().SetVisibility(VisibilityHideBlocking)
	s.container.AddChild(s.icon, s.input, s.clear)
}
