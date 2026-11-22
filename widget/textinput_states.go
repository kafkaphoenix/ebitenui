package widget

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
)

// The pointer over a text input, and the left button pressed on it, as a
// button tells of them, and the images and colors they show. For
// ebitenui/ebitenui#341.

type TextInputPressedEventArgs struct {
	TextInput        *TextInput
	OffsetX, OffsetY int
}

type TextInputReleasedEventArgs struct {
	TextInput        *TextInput
	Inside           bool
	OffsetX, OffsetY int
}

type TextInputClickedEventArgs struct {
	TextInput        *TextInput
	OffsetX, OffsetY int
}

type TextInputHoverEventArgs struct {
	TextInput        *TextInput
	Entered          bool
	OffsetX, OffsetY int
	DiffX, DiffY     int
}

type (
	TextInputPressedHandlerFunc     func(args *TextInputPressedEventArgs)
	TextInputReleasedHandlerFunc    func(args *TextInputReleasedEventArgs)
	TextInputClickedHandlerFunc     func(args *TextInputClickedEventArgs)
	TextInputCursorHoverHandlerFunc func(args *TextInputHoverEventArgs)
)

// PressedHandler is called as the left button goes down on the input.
func (o TextInputOptions) PressedHandler(f TextInputPressedHandlerFunc) TextInputOpt {
	return func(t *TextInput) { t.PressedEvent.AddHandler(handle(f)) }
}

// ReleasedHandler is called as the left button pressed on the input comes
// up, inside it or not.
func (o TextInputOptions) ReleasedHandler(f TextInputReleasedHandlerFunc) TextInputOpt {
	return func(t *TextInput) { t.ReleasedEvent.AddHandler(handle(f)) }
}

// ClickedHandler is called as the left button pressed on the input comes
// up on it.
func (o TextInputOptions) ClickedHandler(f TextInputClickedHandlerFunc) TextInputOpt {
	return func(t *TextInput) { t.ClickedEvent.AddHandler(handle(f)) }
}

// CursorEnteredHandler is called as the pointer comes over the input.
func (o TextInputOptions) CursorEnteredHandler(f TextInputCursorHoverHandlerFunc) TextInputOpt {
	return func(t *TextInput) { t.CursorEnteredEvent.AddHandler(handle(f)) }
}

// CursorMovedHandler is called as the pointer moves over the input.
func (o TextInputOptions) CursorMovedHandler(f TextInputCursorHoverHandlerFunc) TextInputOpt {
	return func(t *TextInput) { t.CursorMovedEvent.AddHandler(handle(f)) }
}

// CursorExitedHandler is called as the pointer leaves the input.
func (o TextInputOptions) CursorExitedHandler(f TextInputCursorHoverHandlerFunc) TextInputOpt {
	return func(t *TextInput) { t.CursorExitedEvent.AddHandler(handle(f)) }
}

// handle returns an event handler calling f with its arguments of type A.
func handle[A any](f func(*A)) func(any) {
	return func(args any) {
		if a, ok := args.(*A); ok {
			f(a)
		}
	}
}

// pointerOpts are the widget options that track the pointer and the left
// button for the states and tell of them.
func (t *TextInput) pointerOpts() []WidgetOpt {
	return []WidgetOpt{
		WidgetOpts.CursorEnterHandler(func(args *CursorEnterEventArgs) {
			if t.widget.Disabled {
				return
			}
			t.hovering = true
			t.CursorEnteredEvent.Fire(&TextInputHoverEventArgs{TextInput: t, Entered: true, OffsetX: args.OffsetX, OffsetY: args.OffsetY})
		}),
		WidgetOpts.CursorMoveHandler(func(args *CursorMoveEventArgs) {
			t.CursorMovedEvent.Fire(&TextInputHoverEventArgs{TextInput: t, OffsetX: args.OffsetX, OffsetY: args.OffsetY, DiffX: args.DiffX, DiffY: args.DiffY})
		}),
		WidgetOpts.CursorExitHandler(func(args *CursorExitEventArgs) {
			if !t.hovering {
				return
			}
			t.hovering = false
			t.CursorExitedEvent.Fire(&TextInputHoverEventArgs{TextInput: t, OffsetX: args.OffsetX, OffsetY: args.OffsetY})
		}),
		WidgetOpts.MouseButtonPressedHandler(func(args *MouseButtonPressedEventArgs) {
			if t.widget.Disabled || args.Button != ebiten.MouseButtonLeft {
				return
			}
			t.pressing = true
			t.PressedEvent.Fire(&TextInputPressedEventArgs{TextInput: t, OffsetX: args.OffsetX, OffsetY: args.OffsetY})
		}),
		WidgetOpts.MouseButtonReleasedHandler(func(args *MouseButtonReleasedEventArgs) {
			if t.pressing && !t.widget.Disabled && args.Button == ebiten.MouseButtonLeft {
				t.ReleasedEvent.Fire(&TextInputReleasedEventArgs{TextInput: t, Inside: args.Inside, OffsetX: args.OffsetX, OffsetY: args.OffsetY})
			}
			t.pressing = false
		}),
		WidgetOpts.MouseButtonClickedHandler(func(args *MouseButtonClickedEventArgs) {
			if !t.widget.Disabled && args.Button == ebiten.MouseButtonLeft {
				t.ClickedEvent.Fire(&TextInputClickedEventArgs{TextInput: t, OffsetX: args.OffsetX, OffsetY: args.OffsetY})
			}
		}),
	}
}

// stateImage returns the image for the input's state: Disabled, or
// PressedDisabled while pressed; Hover while focused or hovered, or
// PressedHover while pressed too; Pressed while pressed on it; otherwise,
// and for a state without an image, Idle.
func (t *TextInput) stateImage() *image.NineSlice {
	i := t.computedParams.Image
	pressed := t.pressing && t.hovering
	switch {
	case t.widget.Disabled && i.Disabled != nil:
		if pressed && i.PressedDisabled != nil {
			return i.PressedDisabled
		}
		return i.Disabled
	case (t.focused || t.hovering) && i.Hover != nil:
		if pressed && i.PressedHover != nil {
			return i.PressedHover
		}
		return i.Hover
	case pressed && i.Pressed != nil:
		return i.Pressed
	}
	return i.Idle
}
