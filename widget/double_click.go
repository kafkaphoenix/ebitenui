package widget

import (
	"image"
	"time"

	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
)

// DoubleClickSettings tell which two clicks of a button on a widget make a
// double click: the second lets go within Interval of the first, which is
// counted in updates at Ebitengine's TPS, and lands within Distance pixels
// of it along each axis, or TouchDistance for a tap, as a finger lands less
// precisely. An Interval of 0 makes none.
type DoubleClickSettings struct {
	Interval                time.Duration
	Distance, TouchDistance int
}

// DoubleClick is the settings of every widget's double clicks: by default
// 500 ms and 4 pixels, as Windows has them, and 24 pixels for taps.
// Ebitengine doesn't tell the platform's own; a game may offer the player
// its own. Set it on the game's goroutine.
var DoubleClick = DoubleClickSettings{Interval: 500 * time.Millisecond, Distance: 4, TouchDistance: 24}

// MouseButtonDoubleClickedEventArgs are the arguments of a double click,
// fired after the second click's MouseButtonClickedEvent.
type MouseButtonDoubleClickedEventArgs struct {
	Widget *Widget
	Button ebiten.MouseButton

	// OffsetX and OffsetY are where the second click was, from the
	// widget's Rect.
	OffsetX, OffsetY int
}

// MouseButtonDoubleClickedHandlerFunc handles double clicks.
type MouseButtonDoubleClickedHandlerFunc func(args *MouseButtonDoubleClickedEventArgs)

// MouseButtonDoubleClickedHandler configures a widget with a double click
// handler f: two clicks of a button, or two taps, close in time and place
// (DoubleClick). A third click starts a new pair. For
// ebitenui/ebitenui#211.
func (o WidgetOptions) MouseButtonDoubleClickedHandler(f MouseButtonDoubleClickedHandlerFunc) WidgetOpt {
	return func(w *Widget) {
		w.MouseButtonDoubleClickedEvent.AddHandler(func(args any) {
			if arg, ok := args.(*MouseButtonDoubleClickedEventArgs); ok {
				f(arg)
			}
		})
	}
}

// firstClick is a click a second one of its button may make a double click
// of.
type firstClick struct {
	pending bool
	button  ebiten.MouseButton
	at      image.Point
	updates int // the updates since it
}

// ageFirstClick counts an update since the first click, which a press of
// its button off the widget, or the interval passing, forgets.
func (w *Widget) ageFirstClick(inside bool) {
	f := &w.firstClick
	if !f.pending {
		return
	}
	f.updates++
	if f.updates > updatesIn(DoubleClick.Interval) || !inside && input.MouseButtonJustPressed(f.button) {
		f.pending = false
	}
}

// pairs reports whether a click or a press of button at p, a tap if
// touch, follows the first click close enough to make a double click, as
// a widget that acts on the second press asks.
func (w *Widget) pairs(button ebiten.MouseButton, p image.Point, touch bool) bool {
	f := &w.firstClick
	d := DoubleClick.Distance
	if touch {
		d = DoubleClick.TouchDistance
	}
	return f.pending && f.button == button && abs(p.X-f.at.X) <= d && abs(p.Y-f.at.Y) <= d
}

// clicked makes a click of button at p, a tap if touch, the second of a
// double click when it follows a first one close enough, and otherwise
// the first.
func (w *Widget) clicked(button ebiten.MouseButton, p image.Point, touch bool) {
	f := &w.firstClick
	if w.pairs(button, p, touch) {
		f.pending = false
		off := p.Sub(w.Rect.Min)
		w.MouseButtonDoubleClickedEvent.Fire(&MouseButtonDoubleClickedEventArgs{
			Widget: w, Button: button, OffsetX: off.X, OffsetY: off.Y,
		})
		return
	}
	*f = firstClick{pending: true, button: button, at: p}
}
