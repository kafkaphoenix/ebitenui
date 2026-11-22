package widget

import (
	img "image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/input"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type TextAreaParams struct {
	Face                   *text.Face
	ForegroundColor        color.Color
	TextPadding            *Insets
	TextPosition           *TextPositioning
	ControlWidgetSpacing   *int
	StripBBCode            *bool
	LinkColor              *TextLinkColor
	Slider                 *SliderParams
	ScrollContainerImage   *ScrollContainerImage
	ScrollContainerPadding *Insets
}

type TextArea struct {
	definedParams  TextAreaParams
	computedParams TextAreaParams

	containerOpts []ContainerOpt

	processBBCode        bool
	initialText          string
	verticalScrollMode   ScrollMode
	horizontalScrollMode ScrollMode
	showHorizontalSlider bool
	showVerticalSlider   bool
	dragToScroll         bool

	linkClickedFunc       LinkHandlerFunc
	linkCursorEnteredFunc LinkHandlerFunc
	linkCursorExitedFunc  LinkHandlerFunc

	init            *MultiOnce
	container       *Container
	layout          *GridLayout
	scrollContainer *ScrollContainer
	vSlider         *Slider
	hSlider         *Slider
	text            *Text
}

type ScrollMode int

const (
	// Default. Scrolling is not automatically handled.
	None ScrollMode = iota

	// The TextArea is automatically scrolled to the beginning on change.
	ScrollBeginning

	// The TextArea is automatically scrolled to the end on change.
	ScrollEnd

	// The TextArea will initially position the text at the end of the scroll area.
	PositionAtEnd
)

type TextAreaOpt func(l *TextArea)

type TextAreaEntrySelectedEventArgs struct {
	TextArea      *TextArea
	Entry         any
	PreviousEntry any
}

type TextAreaEntrySelectedHandlerFunc func(args *TextAreaEntrySelectedEventArgs)

type TextAreaOptions struct {
}

var TextAreaOpts TextAreaOptions

func NewTextArea(opts ...TextAreaOpt) *TextArea {
	l := &TextArea{
		init: &MultiOnce{},
	}

	l.init.Append(l.createWidget)

	for _, o := range opts {
		o(l)
	}

	return l
}

func (t *TextArea) Validate() {
	t.init.Do()
	t.populateComputedParams()
	if t.computedParams.ForegroundColor == nil {
		panic("TextArea: FontColor is required.")
	}
	if t.computedParams.Face == nil {
		panic("TextArea: FontFace is required.")
	}
	t.initWidget()
}

func (t *TextArea) populateComputedParams() {
	params := TextAreaParams{}

	theme := t.GetWidget().GetTheme()

	// Set theme values
	if theme != nil {
		if theme.TextAreaTheme != nil {
			params.ControlWidgetSpacing = theme.TextAreaTheme.ControlWidgetSpacing
			if theme.TextAreaTheme.Face != nil {
				params.Face = theme.TextAreaTheme.Face
			} else {
				params.Face = theme.DefaultFace
			}
			if theme.TextAreaTheme.ForegroundColor != nil {
				params.ForegroundColor = theme.TextAreaTheme.ForegroundColor
			} else {
				params.ForegroundColor = theme.DefaultTextColor
			}
			params.LinkColor = theme.TextAreaTheme.LinkColor
			params.ScrollContainerImage = theme.TextAreaTheme.ScrollContainerImage
			params.ScrollContainerPadding = theme.TextAreaTheme.ScrollContainerPadding
			if theme.TextAreaTheme.Slider != nil {
				params.Slider = &SliderParams{
					TrackImage:      theme.TextAreaTheme.Slider.TrackImage,
					Orientation:     theme.TextAreaTheme.Slider.Orientation,
					TrackPadding:    theme.TextAreaTheme.Slider.TrackPadding,
					MinHandleSize:   theme.TextAreaTheme.Slider.MinHandleSize,
					FixedHandleSize: theme.TextAreaTheme.Slider.FixedHandleSize,
					TrackOffset:     theme.TextAreaTheme.Slider.TrackOffset,
					HandleImage:     theme.TextAreaTheme.Slider.HandleImage,
					PageSizeFunc:    theme.TextAreaTheme.Slider.PageSizeFunc,
				}
			}
			params.StripBBCode = theme.TextAreaTheme.StripBBCode
			params.TextPadding = theme.TextAreaTheme.TextPadding
			params.TextPosition = theme.TextAreaTheme.TextPosition
		}
	}

	// Set definedParam values
	if t.definedParams.ControlWidgetSpacing != nil {
		params.ControlWidgetSpacing = t.definedParams.ControlWidgetSpacing
	}
	if t.definedParams.Face != nil {
		params.Face = t.definedParams.Face
	}
	if t.definedParams.ForegroundColor != nil {
		params.ForegroundColor = t.definedParams.ForegroundColor
	}
	if t.definedParams.LinkColor != nil {
		params.LinkColor = t.definedParams.LinkColor
	}
	if t.definedParams.ScrollContainerImage != nil {
		params.ScrollContainerImage = t.definedParams.ScrollContainerImage
	}
	if t.definedParams.ScrollContainerPadding != nil {
		params.ScrollContainerPadding = t.definedParams.ScrollContainerPadding
	}
	if t.definedParams.Slider != nil {
		if params.Slider == nil {
			params.Slider = &SliderParams{}
		}
		if t.definedParams.Slider.FixedHandleSize != nil {
			params.Slider.FixedHandleSize = t.definedParams.Slider.FixedHandleSize
		}
		if t.definedParams.Slider.HandleImage != nil {
			params.Slider.HandleImage = t.definedParams.Slider.HandleImage
		}
		if t.definedParams.Slider.MinHandleSize != nil {
			params.Slider.MinHandleSize = t.definedParams.Slider.MinHandleSize
		}
		if t.definedParams.Slider.TrackImage != nil {
			params.Slider.TrackImage = t.definedParams.Slider.TrackImage
		}
		if t.definedParams.Slider.TrackOffset != nil {
			params.Slider.TrackOffset = t.definedParams.Slider.TrackOffset
		}
		if t.definedParams.Slider.TrackPadding != nil {
			params.Slider.TrackPadding = t.definedParams.Slider.TrackPadding
		}
		if t.definedParams.Slider.PageSizeFunc != nil {
			params.Slider.PageSizeFunc = t.definedParams.Slider.PageSizeFunc
		}
	}
	if t.definedParams.StripBBCode != nil {
		params.StripBBCode = t.definedParams.StripBBCode
	}
	if t.definedParams.TextPadding != nil {
		params.TextPadding = t.definedParams.TextPadding
	}
	if t.definedParams.TextPosition != nil {
		params.TextPosition = t.definedParams.TextPosition
	}
	// Set defaults

	if params.TextPadding == nil {
		params.TextPadding = &Insets{}
	}
	if params.TextPosition == nil {
		params.TextPosition = &TextPositioning{}
	}
	if params.ControlWidgetSpacing == nil {
		params.ControlWidgetSpacing = new(0)
	}
	if params.StripBBCode == nil {
		params.StripBBCode = new(false)
	}
	if params.ScrollContainerPadding == nil {
		params.ScrollContainerPadding = &Insets{}
	}

	t.computedParams = params
}

// Specify the Container options for the text area.
func (o TextAreaOptions) ContainerOpts(opts ...ContainerOpt) TextAreaOpt {
	return func(l *TextArea) {
		l.containerOpts = append(l.containerOpts, opts...)
	}
}

// Specify the images for the scroll container.
func (o TextAreaOptions) ScrollContainerImage(image *ScrollContainerImage) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.ScrollContainerImage = image
	}
}

// Specify the padding for the scroll container.
func (o TextAreaOptions) ScrollContainerPadding(padding *Insets) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.ScrollContainerPadding = padding
	}
}

// Specify the options for the scroll bars.
func (o TextAreaOptions) SliderParams(sliderParams *SliderParams) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.Slider = sliderParams
	}
}

// Specify spacing between the text container and scrollbars.
func (o TextAreaOptions) ControlWidgetSpacing(s int) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.ControlWidgetSpacing = &s
	}
}

// Show the horizontal scrollbar.
func (o TextAreaOptions) ShowHorizontalScrollbar() TextAreaOpt {
	return func(l *TextArea) {
		l.showHorizontalSlider = true
	}
}

// Show the vertical scrollbar.
// DragToScroll lets the pointer scroll the text by dragging it, as
// ScrollContainerOpts.DragToScroll does; the scrollbars follow.
func (o TextAreaOptions) DragToScroll() TextAreaOpt {
	return func(l *TextArea) {
		l.dragToScroll = true
	}
}

func (o TextAreaOptions) ShowVerticalScrollbar() TextAreaOpt {
	return func(l *TextArea) {
		l.showVerticalSlider = true
	}
}

// Set how vertical scrolling should be handled.
func (o TextAreaOptions) VerticalScrollMode(scrollMode ScrollMode) TextAreaOpt {
	return func(l *TextArea) {
		l.verticalScrollMode = scrollMode
	}
}

// Set how horizontal scrolling should be handled.
func (o TextAreaOptions) HorizontalScrollMode(scrollMode ScrollMode) TextAreaOpt {
	return func(l *TextArea) {
		l.horizontalScrollMode = scrollMode
	}
}

// Set the font face for this text area.
func (o TextAreaOptions) FontFace(f *text.Face) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.Face = f
	}
}

// Set the default color for the text area.
func (o TextAreaOptions) FontColor(color color.Color) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.ForegroundColor = color
	}
}

// Set how far from the edges of the textarea the text should be set.
func (o TextAreaOptions) TextPadding(i Insets) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.TextPadding = &i
	}
}

// Set the positioning of the text within the text area.
func (o TextAreaOptions) TextPosition(textPosition TextPositioning) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.TextPosition = &textPosition
	}
}

// Set the initial Text for the text area.
func (o TextAreaOptions) Text(initialText string) TextAreaOpt {
	return func(l *TextArea) {
		l.initialText = initialText
	}
}

// This option tells the textarea object to process BBCodes.
//
// Currently the system supports the following BBCodes:
//   - color - [color=#FFFFFF] text [/color] - defines a color code for the enclosed text
//   - link - [link=id arg1:value1 ... argX:valueX] text [/link] - defines a clickable section of text,
//     that will trigger a callback.
func (o TextAreaOptions) ProcessBBCode(processBBCode bool) TextAreaOpt {
	return func(l *TextArea) {
		l.processBBCode = processBBCode
	}
}

// Set whether or not the text area should automatically strip out BBCodes from being displayed.
func (o TextAreaOptions) StripBBCode(stripBBCode bool) TextAreaOpt {
	return func(l *TextArea) {
		l.definedParams.StripBBCode = &stripBBCode
	}
}

// This option sets the idle and hover color for text that is wrapped in a
// [link][/link] bbcode.
//
// Note: this is only used if ProcessBBCode is true.
func (o TextAreaOptions) LinkColor(linkColor *TextLinkColor) TextAreaOpt {
	return func(t *TextArea) {
		t.definedParams.LinkColor = linkColor
	}
}

// Defines the handler to be called when a BBCode defined link is clicked.
//
// Note: this is only used if ProcessBBCode is true.
func (o TextAreaOptions) LinkClickedEvent(linkClickedFunc LinkHandlerFunc) TextAreaOpt {
	return func(l *TextArea) {
		l.linkClickedFunc = linkClickedFunc
	}
}

// Defines the handler to be called when the cursor enters a BBCode defined link.
//
// Note: this is only used if ProcessBBCode is true.
func (o TextAreaOptions) LinkCursorEnteredEvent(linkCursorEnteredFunc LinkHandlerFunc) TextAreaOpt {
	return func(l *TextArea) {
		l.linkCursorEnteredFunc = linkCursorEnteredFunc
	}
}

// Defines the handler to be called when the cursor enters a BBCode defined link.
//
// Note: this is only used if ProcessBBCode is true.
func (o TextAreaOptions) LinkCursorExitedEvent(linkCursorExitedFunc LinkHandlerFunc) TextAreaOpt {
	return func(l *TextArea) {
		l.linkCursorExitedFunc = linkCursorExitedFunc
	}
}

func (t *TextArea) GetWidget() *Widget {
	t.init.Do()
	return t.container.GetWidget()
}

func (t *TextArea) PreferredSize() (int, int) {
	t.init.Do()
	w, h := t.container.PreferredSize()

	if t.container.widget != nil && h < t.container.widget.MinHeight {
		h = t.container.widget.MinHeight
	}
	if t.container.widget != nil && w < t.container.widget.MinWidth {
		w = t.container.widget.MinWidth
	}
	return w, h
}

func (t *TextArea) SetLocation(rect img.Rectangle) {
	t.init.Do()
	t.container.GetWidget().Rect = rect
}

func (t *TextArea) RequestRelayout() {
	t.init.Do()
	t.container.RequestRelayout()
}

func (t *TextArea) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	t.init.Do()
	t.container.SetupInputLayer(def)
}

func (t *TextArea) GetFocusers() []Focuser {
	t.init.Do()
	var result []Focuser
	if t.hSlider != nil && t.hSlider.tabOrder != -1 {
		result = append(result, t.hSlider)
	}
	if t.vSlider != nil && t.vSlider.tabOrder != -1 {
		result = append(result, t.vSlider)
	}
	return result
}

func (t *TextArea) Render(screen *ebiten.Image) {
	t.init.Do()

	d := t.container.GetWidget().Disabled

	if t.vSlider != nil {
		t.vSlider.DrawTrackDisabled = d
	}
	if t.hSlider != nil {
		t.hSlider.DrawTrackDisabled = d
	}
	t.text.MaxWidth = float64(t.container.GetWidget().Rect.Dx())
	t.scrollContainer.GetWidget().Disabled = d
	t.container.Render(screen)
}

func (t *TextArea) Update(updObj *UpdateObject) {
	t.init.Do()
	if t.container != nil {
		t.container.Update(updObj)
	}
}

func (t *TextArea) createWidget() {
	var cols int
	if t.showVerticalSlider {
		cols = 2
	} else {
		cols = 1
	}
	t.layout = NewGridLayout(
		GridLayoutOpts.Columns(cols),
		GridLayoutOpts.Stretch([]bool{true, false}, []bool{true, false}))

	t.container = NewContainer(
		append([]ContainerOpt{
			ContainerOpts.WidgetOpts(WidgetOpts.TrackHover(true)),
			ContainerOpts.Layout(t.layout),
		}, t.containerOpts...,
		)...)

	t.text = NewText(TextOpts.TextLabel(t.initialText))
}

func (t *TextArea) initWidget() {
	currentText := t.GetText()
	t.container.RemoveChildren()
	t.layout.columnSpacing = *t.computedParams.ControlWidgetSpacing
	t.layout.rowSpacing = *t.computedParams.ControlWidgetSpacing

	content := NewContainer(
		ContainerOpts.Layout(NewRowLayout(
			RowLayoutOpts.Direction(DirectionVertical))),
		ContainerOpts.AutoDisableChildren())

	t.text = NewText(
		TextOpts.Text(currentText, t.computedParams.Face, t.computedParams.ForegroundColor),
		TextOpts.Padding(t.computedParams.TextPadding),
		TextOpts.Position(t.computedParams.TextPosition.HTextPosition, t.computedParams.TextPosition.VTextPosition),
		TextOpts.ProcessBBCode(t.processBBCode),
		TextOpts.StripBBCode(*t.computedParams.StripBBCode),
		TextOpts.LinkColor(t.computedParams.LinkColor),
		TextOpts.LinkClickedHandler(t.linkClickedFunc),
		TextOpts.LinkCursorEnteredHandler(t.linkCursorEnteredFunc),
		TextOpts.LinkCursorExitedHandler(t.linkCursorExitedFunc),
	)
	content.AddChild(t.text)
	t.text.widget.parent = t.container.GetWidget()

	scrollOpts := []ScrollContainerOpt{
		ScrollContainerOpts.Content(content),
		ScrollContainerOpts.StretchContentWidth(),
		ScrollContainerOpts.Image(t.computedParams.ScrollContainerImage),
		ScrollContainerOpts.Padding(t.computedParams.ScrollContainerPadding),
	}
	if t.dragToScroll {
		scrollOpts = append(scrollOpts, ScrollContainerOpts.DragToScroll())
	}
	t.scrollContainer = NewScrollContainer(scrollOpts...)

	if t.dragToScroll {
		t.scrollContainer.ScrollMovedEvent.AddHandler(func(args any) {
			if a, ok := args.(*ScrollContainerScrollMovedEventArgs); ok {
				if t.vSlider != nil {
					t.vSlider.Current = int(math.Round(a.Top * 1000))
				}
				if t.hSlider != nil {
					t.hSlider.Current = int(math.Round(a.Left * 1000))
				}
			}
		})
	}
	t.container.AddChild(t.scrollContainer)

	var sliderOpts []SliderOpt
	if t.computedParams.Slider != nil {
		if t.computedParams.Slider.FixedHandleSize != nil {
			sliderOpts = append(sliderOpts, SliderOpts.FixedHandleSize(*t.computedParams.Slider.FixedHandleSize))
		} else {
			sliderOpts = append(sliderOpts, SliderOpts.FixedHandleSize(0))
		}
		if t.computedParams.Slider.HandleImage != nil {
			sliderOpts = append(sliderOpts, SliderOpts.HandleImage(t.computedParams.Slider.HandleImage))
		}
		if t.computedParams.Slider.TrackImage != nil {
			sliderOpts = append(sliderOpts, SliderOpts.TrackImage(t.computedParams.Slider.TrackImage))
		}
		if t.computedParams.Slider.MinHandleSize != nil {
			sliderOpts = append(sliderOpts, SliderOpts.MinHandleSize(*t.computedParams.Slider.MinHandleSize))
		}
		if t.computedParams.Slider.TrackOffset != nil {
			sliderOpts = append(sliderOpts, SliderOpts.TrackOffset(*t.computedParams.Slider.TrackOffset))
		}
		if t.computedParams.Slider.TrackPadding != nil {
			sliderOpts = append(sliderOpts, SliderOpts.TrackPadding(t.computedParams.Slider.TrackPadding))
		}
	}

	if t.showVerticalSlider {
		pageSizeFunc := func() int {
			return int(math.Round(float64(t.scrollContainer.ViewRect().Dy()) / float64(content.GetWidget().Rect.Dy()) * 1000))
		}

		t.vSlider = NewSlider(append(sliderOpts,
			SliderOpts.Orientation(DirectionVertical),
			SliderOpts.MinMax(0, 1000),
			SliderOpts.PageSizeFunc(pageSizeFunc),
			SliderOpts.ChangedHandler(func(args *SliderChangedEventArgs) {
				current := args.Slider.Current
				if pageSizeFunc() >= 1000 {
					current = 0
					if t.verticalScrollMode == ScrollEnd || t.verticalScrollMode == PositionAtEnd {
						current = 1000
					}
				}
				t.scrollContainer.ScrollTop = float64(current) / 1000
			}),
		)...)

		if t.verticalScrollMode == ScrollEnd || t.verticalScrollMode == PositionAtEnd {
			t.vSlider.Current = t.vSlider.Max
		}
		t.container.AddChild(t.vSlider)

		t.scrollContainer.widget.ScrolledEvent.AddHandler(func(args any) {
			if a, ok := args.(*ScrolledEventArgs); ok {
				p := max(pageSizeFunc()/3, 1)
				t.vSlider.Current -= int(math.Round(a.Y * float64(p)))
			}
		})
	}

	if t.showHorizontalSlider {
		pageSizeFunc := func() int {
			return int(math.Round(float64(t.scrollContainer.ViewRect().Dx()) / float64(content.GetWidget().Rect.Dx()) * 1000))
		}

		t.hSlider = NewSlider(append(sliderOpts,
			SliderOpts.Orientation(DirectionHorizontal),
			SliderOpts.MinMax(0, 1000),
			SliderOpts.PageSizeFunc(pageSizeFunc),
			SliderOpts.ChangedHandler(func(args *SliderChangedEventArgs) {
				current := args.Slider.Current
				if pageSizeFunc() >= 1000 {
					current = 0
					if t.horizontalScrollMode == ScrollEnd || t.horizontalScrollMode == PositionAtEnd {
						current = 1000
					}
				}
				t.scrollContainer.ScrollLeft = float64(current) / 1000
			}),
		)...)

		if t.horizontalScrollMode == ScrollEnd || t.horizontalScrollMode == PositionAtEnd {
			t.hSlider.Current = t.hSlider.Max
		}
		t.container.AddChild(t.hSlider)
	}
}

func (t *TextArea) PrependText(value string) {
	t.init.Do()
	t.SetText(value + t.text.Label)
}

func (t *TextArea) AppendText(value string) {
	t.init.Do()
	t.SetText(t.text.Label + value)
}

func (t *TextArea) SetText(value string) {
	t.init.Do()
	t.text.Label = value

	if t.showHorizontalSlider && t.hSlider != nil {
		scrollTo(t.hSlider, t.horizontalScrollMode)
	}
	if t.showVerticalSlider && t.vSlider != nil {
		scrollTo(t.vSlider, t.verticalScrollMode)
	}
}

// scrollTo moves a slider to its start or end as the mode asks on a change
// of text, or leaves it where it is: PositionAtEnd places the text only as
// the area is made.
func scrollTo(s *Slider, mode ScrollMode) {
	switch mode {
	case ScrollBeginning:
		s.Current = 0
	case ScrollEnd:
		s.Current = s.Max
	case None, PositionAtEnd:
	}
}

func (t *TextArea) GetText() string {
	t.init.Do()
	return t.text.Label
}
