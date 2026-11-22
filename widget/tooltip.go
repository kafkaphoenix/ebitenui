package widget

import (
	"image"
	"image/color"
	"time"

	e_image "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type ToolTipPosition int

const (
	// The tooltip will follow the cursor around while visible.
	ToolTipPositionCursorFollow ToolTipPosition = iota
	// The tooltip will stick to where the cursor was when the tooltip was made visible.
	ToolTipPositionCursorSticky
	// The tooltip will display based on the Widget and Content anchor settings.
	// It defaults to opening right aligned and directly under the widget.
	ToolTipPositionWidget
	// The tooltip will display based on x/y (offset is required).
	ToolTipPositionAbsolute
	// The tooltip will display based on the Widget and Content anchor settings.
	// It defaults to opening right aligned and directly under the x: 0, y: 0.
	ToolTipPositionScreen
)

type ToolTipAnchor int

const (
	// Anchor at the start of the element.
	ToolTipAnchorStart ToolTipAnchor = iota
	// Anchor in the middle of the element.
	ToolTipAnchorMiddle
	// Anchor at the end of the element.
	ToolTipAnchorEnd
)

type ToolTipDirection int

type ToolTip struct {
	Position ToolTipPosition
	// WidgetOriginVertical was renamed to AnchorOriginVertical to make the it more generic and reuse it for ToolTipPositionScreen
	AnchorOriginVertical ToolTipAnchor
	// WidgetOriginHorizontal was renamed to AnchorOriginHorizontal to make the it more generic and reuse it for ToolTipPositionScreen
	AnchorOriginHorizontal  ToolTipAnchor
	ContentOriginVertical   ToolTipAnchor
	ContentOriginHorizontal ToolTipAnchor
	Delay                   time.Duration
	Offset                  image.Point
	KeepOnHover             bool
	content                 Containerer
	window                  *Window
	visible                 bool

	state          toolTipState
	armedUpdates   int // the updates the pointer has rested since the tooltip armed
	ToolTipUpdater ToolTipUpdater
}
type ToolTipOpt func(t *ToolTip)
type ToolTipOptions struct {
}

var ToolTipOpts ToolTipOptions

type toolTipState func(*Widget) toolTipState

type ToolTipUpdater func(Containerer)

// Create a new Tooltip. This method allows you to specify
// every aspect of the displayed tooltip's content.
func NewToolTip(opts ...ToolTipOpt) *ToolTip {
	t := &ToolTip{
		Offset: image.Point{10, 10},
	}
	t.state = t.idleState()
	t.AnchorOriginHorizontal = ToolTipAnchorEnd
	t.AnchorOriginVertical = ToolTipAnchorEnd
	t.ContentOriginHorizontal = ToolTipAnchorEnd
	t.ContentOriginVertical = ToolTipAnchorStart
	for _, o := range opts {
		o(t)
	}

	t.Validate()

	t.window = NewWindow(
		WindowOpts.CloseMode(WindowCloseNone),
		WindowOpts.Contents(t.content),
	)

	return t
}

func (t *ToolTip) Validate() {
	if t.content == nil {
		panic("ToolTip: Contents is required.")
	}
}

// Create a new Text Tooltip with the following defaults:
//   - ProcessBBCode = true
//   - Padding = Top/Bottom: 5px Left/Right: 10px
//   - Delay = 800ms
//   - Offset = 0, 20
//   - ContentOriginHorizontal = ToolTipAnchorEnd
//   - ContentOriginVertical = ToolTipAnchorStart
//
// face, color and background may be nil: the tooltip then takes its theme's
// text face and color and its theme's panel background.
func NewTextToolTip(label string, face *text.Face, color color.Color, background *e_image.NineSlice) *ToolTip {
	c := &themedPanel{NewContainer(
		ContainerOpts.BackgroundImage(background),
		ContainerOpts.AutoDisableChildren(),
		ContainerOpts.Layout(NewAnchorLayout(AnchorLayoutOpts.Padding(&Insets{
			Top:    5,
			Bottom: 5,
			Left:   10,
			Right:  10,
		}))),
	)}

	textOpts := []TextOpt{TextOpts.ProcessBBCode(true), TextOpts.TextLabel(label)}
	if face != nil {
		textOpts = append(textOpts, TextOpts.TextFace(face))
	}
	if color != nil {
		textOpts = append(textOpts, TextOpts.TextColor(color))
	}
	c.AddChild(NewText(textOpts...))

	return NewToolTip(
		ToolTipOpts.Content(c),
		ToolTipOpts.Delay(800*time.Millisecond),
		ToolTipOpts.Offset(image.Point{0, 20}),
		ToolTipOpts.ContentOriginHorizontal(ToolTipAnchorStart),
		ToolTipOpts.ContentOriginVertical(ToolTipAnchorStart),
	)
}

// The container to be displayed.
func (o ToolTipOptions) Content(c Containerer) ToolTipOpt {
	return func(t *ToolTip) {
		t.content = c
	}
}

// The X/Y offsets from the Tooltip anchor point.
func (o ToolTipOptions) Offset(off image.Point) ToolTipOpt {
	return func(t *ToolTip) {
		t.Offset = off
	}
}

// The vertical position of the anchor on the widget. Only used when Position = WIDGET.
func (o ToolTipOptions) AnchorOriginVertical(anchorOriginVertical ToolTipAnchor) ToolTipOpt {
	return func(t *ToolTip) {
		t.AnchorOriginVertical = anchorOriginVertical
	}
}

// The horizontal position of the anchor on the widget. Only used when Position = WIDGET.
func (o ToolTipOptions) AnchorOriginHorizontal(anchorOriginHorizontal ToolTipAnchor) ToolTipOpt {
	return func(t *ToolTip) {
		t.AnchorOriginHorizontal = anchorOriginHorizontal
	}
}

// The vertical position of the anchor on the tooltip.
func (o ToolTipOptions) ContentOriginVertical(contentOriginVertical ToolTipAnchor) ToolTipOpt {
	return func(t *ToolTip) {
		t.ContentOriginVertical = contentOriginVertical
	}
}

// The horizontal position of the anchor on the tooltip.
func (o ToolTipOptions) ContentOriginHorizontal(contentOriginHorizontal ToolTipAnchor) ToolTipOpt {
	return func(t *ToolTip) {
		t.ContentOriginHorizontal = contentOriginHorizontal
	}
}

// Where to display the tooltip.
func (o ToolTipOptions) Position(position ToolTipPosition) ToolTipOpt {
	return func(t *ToolTip) {
		t.Position = position
	}
}

// How long to wait before displaying the tooltip.
func (o ToolTipOptions) Delay(d time.Duration) ToolTipOpt {
	return func(t *ToolTip) {
		t.Delay = d
	}
}

// A method that is called every draw call that the tooltip is visible.
// This allows you to hook into the draw loop to update the tooltip if necessary.
func (o ToolTipOptions) ToolTipUpdater(toolTipUpdater ToolTipUpdater) ToolTipOpt {
	return func(t *ToolTip) {
		t.ToolTipUpdater = toolTipUpdater
	}
}

// KeepOnHover will make it so if the user cursor is on the ToolTip it'll
// not be hidden as it does by default.
func (o ToolTipOptions) KeepOnHover(b bool) ToolTipOpt {
	return func(t *ToolTip) {
		t.KeepOnHover = b
	}
}

func (t *ToolTip) Update(parent *Widget) {
	newState := t.state(parent)
	if newState != nil {
		t.state = newState
	}
}

func (t *ToolTip) idleState() toolTipState {
	return func(parent *Widget) toolTipState {
		if input.MouseButtonPressed(ebiten.MouseButtonLeft) ||
			input.MouseButtonPressed(ebiten.MouseButtonMiddle) ||
			input.MouseButtonPressed(ebiten.MouseButtonRight) ||
			!parent.IsVisible() {
			t.visible = false
			parent.FireToolTipEvent(t.window, false)
			return nil
		}

		x, y := input.CursorPosition()
		p := image.Point{x, y}
		if !p.In(parent.Rect) {
			return nil
		}
		if !parent.EffectiveInputLayer().ActiveFor(x, y, input.LayerEventTypeAny) {
			return nil
		}

		if t.Delay <= 0 {
			return t.showingState(p)
		}

		t.armedUpdates = 0
		return t.armedState()
	}
}

// armedState waits for the pointer to rest on the widget for the delay,
// counted in updates, then shows the tooltip from the next update on.
func (t *ToolTip) armedState() toolTipState {
	return func(parent *Widget) toolTipState {
		x, y := input.CursorPosition()
		cp := image.Point{x, y}

		if input.MouseButtonPressed(ebiten.MouseButtonLeft) ||
			input.MouseButtonPressed(ebiten.MouseButtonMiddle) ||
			input.MouseButtonPressed(ebiten.MouseButtonRight) ||
			!cp.In(parent.Rect) ||
			!parent.EffectiveInputLayer().ActiveFor(x, y, input.LayerEventTypeAny) {
			if !t.KeepOnHover || !cp.In(t.content.GetWidget().Rect) {
				t.visible = false
				parent.FireToolTipEvent(t.window, false)
				return t.idleState()
			}
		}
		t.armedUpdates++
		if t.armedUpdates >= updatesIn(t.Delay) {
			return t.showingState(cp)
		}
		return nil
	}
}

func (t *ToolTip) showingState(p image.Point) toolTipState {
	return func(parent *Widget) toolTipState {
		x, y := input.CursorPosition()
		cp := image.Point{x, y}
		if input.MouseButtonPressed(ebiten.MouseButtonLeft) ||
			input.MouseButtonPressed(ebiten.MouseButtonMiddle) ||
			input.MouseButtonPressed(ebiten.MouseButtonRight) ||
			!cp.In(parent.Rect) ||
			!parent.EffectiveInputLayer().ActiveFor(x, y, input.LayerEventTypeAny) {
			if !t.KeepOnHover || !cp.In(t.content.GetWidget().Rect) {
				t.visible = false
				parent.FireToolTipEvent(t.window, false)
				return t.idleState()
			}
		}
		sx, sy := t.content.PreferredSize()

		position := p
		switch t.Position {
		case ToolTipPositionCursorSticky:
			// Where the pointer was as the tooltip armed.
		case ToolTipPositionCursorFollow:
			position = cp
		case ToolTipPositionWidget:
			position = t.processWidgetPosition(parent.Rect)
		case ToolTipPositionAbsolute:
			position = image.Point{}
		case ToolTipPositionScreen:
			position = t.processScreenPosition()
		}
		position = position.Add(t.Offset)
		position = t.processContentPosition(position, sx, sy, parent.Rect)

		if t.ToolTipUpdater != nil {
			t.ToolTipUpdater(t.content)
		}

		r := image.Rect(0, 0, sx, sy)
		r = r.Add(position)
		t.window.SetLocation(r)
		t.content.SetLocation(r)
		if !t.visible {
			parent.FireToolTipEvent(t.window, true)
			t.visible = true
		}
		return t.showingState(p)
	}
}

func (t *ToolTip) processWidgetPosition(widgetRect image.Rectangle) image.Point {
	p := image.Point{}
	switch t.AnchorOriginVertical {
	case ToolTipAnchorStart:
		switch t.AnchorOriginHorizontal {
		case ToolTipAnchorStart:
			p.X = widgetRect.Min.X
			p.Y = widgetRect.Min.Y
		case ToolTipAnchorMiddle:
			p.X = widgetRect.Min.X + (widgetRect.Dx() / 2)
			p.Y = widgetRect.Min.Y
		case ToolTipAnchorEnd:
			p.X = widgetRect.Max.X
			p.Y = widgetRect.Min.Y
		}
	case ToolTipAnchorMiddle:
		switch t.AnchorOriginHorizontal {
		case ToolTipAnchorStart:
			p.X = widgetRect.Min.X
			p.Y = widgetRect.Min.Y + (widgetRect.Dy() / 2)
		case ToolTipAnchorMiddle:
			p.X = widgetRect.Min.X + (widgetRect.Dx() / 2)
			p.Y = widgetRect.Min.Y + (widgetRect.Dy() / 2)
		case ToolTipAnchorEnd:
			p.X = widgetRect.Max.X
			p.Y = widgetRect.Min.Y + (widgetRect.Dy() / 2)
		}
	case ToolTipAnchorEnd:
		switch t.AnchorOriginHorizontal {
		case ToolTipAnchorStart:
			p.X = widgetRect.Min.X
			p.Y = widgetRect.Max.Y
		case ToolTipAnchorMiddle:
			p.X = widgetRect.Min.X + (widgetRect.Dx() / 2)
			p.Y = widgetRect.Max.Y
		case ToolTipAnchorEnd:
			p.X = widgetRect.Max.X
			p.Y = widgetRect.Max.Y
		}
	}
	return p
}

func (t *ToolTip) processScreenPosition() image.Point {
	windowSize := input.GetWindowSize()
	p := image.Point{}
	switch t.AnchorOriginHorizontal {
	case ToolTipAnchorStart:
		p.X = 0
	case ToolTipAnchorMiddle:
		p.X = windowSize.X / 2
	case ToolTipAnchorEnd:
		p.X = windowSize.X
	}
	switch t.AnchorOriginVertical {
	case ToolTipAnchorStart:
		p.Y = 0
	case ToolTipAnchorMiddle:
		p.Y = windowSize.Y / 2
	case ToolTipAnchorEnd:
		p.Y = windowSize.Y
	}
	return p
}

func (t *ToolTip) processContentPosition(p image.Point, sx, sy int, widgetRect image.Rectangle) image.Point {
	result := processContentPositionWorker(p, sx, sy, t.ContentOriginHorizontal, t.ContentOriginVertical)
	windowSize := input.GetWindowSize()
	horizontalAnchor := t.ContentOriginHorizontal
	if result.X+sx > windowSize.X {
		horizontalAnchor = ToolTipAnchorEnd
		if t.Position == ToolTipPositionWidget {
			p.X = widgetRect.Min.X
		}
		p.X -= 2 * t.Offset.X
		result = processContentPositionWorker(p, sx, sy, horizontalAnchor, t.ContentOriginVertical)
	} else if result.X < 0 {
		p.X -= 2 * t.Offset.X
		horizontalAnchor = ToolTipAnchorStart
		result = processContentPositionWorker(p, sx, sy, horizontalAnchor, t.ContentOriginVertical)
	}

	if result.Y+sy > windowSize.Y {
		if t.Position == ToolTipPositionWidget {
			p.Y = widgetRect.Min.Y
		}
		p.Y -= 2 * t.Offset.Y
		result = processContentPositionWorker(p, sx, sy, horizontalAnchor, ToolTipAnchorEnd)
	} else if result.Y < 0 {
		p.Y -= 2 * t.Offset.Y
		result = processContentPositionWorker(p, sx, sy, horizontalAnchor, ToolTipAnchorStart)
	}

	return result
}

func processContentPositionWorker(p image.Point, sx int, sy int, originHorizontal ToolTipAnchor, originVertical ToolTipAnchor) image.Point {
	switch originVertical {
	case ToolTipAnchorStart:
		switch originHorizontal {
		case ToolTipAnchorStart:
			// Do nothing
		case ToolTipAnchorMiddle:
			p.X -= (sx / 2)
		case ToolTipAnchorEnd:
			p.X -= sx
		}
	case ToolTipAnchorMiddle:
		switch originHorizontal {
		case ToolTipAnchorStart:
			p.Y -= (sy / 2)
		case ToolTipAnchorMiddle:
			p.X -= (sx / 2)
			p.Y -= (sy / 2)
		case ToolTipAnchorEnd:
			p.X -= sx
			p.Y -= (sy / 2)
		}
	case ToolTipAnchorEnd:
		switch originHorizontal {
		case ToolTipAnchorStart:
			p.Y -= sy
		case ToolTipAnchorMiddle:
			p.X -= (sx / 2)
			p.Y -= sy
		case ToolTipAnchorEnd:
			p.X -= sx
			p.Y -= sy
		}
	}
	return p
}
