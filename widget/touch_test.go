package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// touchButton returns a button at (20, 20)-(120, 60), away from where an
// absent mouse points, counting its clicks and context menus opened, on a
// touch screen or a desktop with one, and step, which runs an update with
// the fingers given down.
func touchButton(t *testing.T, touchscreen bool) (b *widget.Button, clicks, menus *int, step func(fingers ...img.Point)) {
	t.Helper()
	d := uitest.Use(t)
	d.Touchscreen = touchscreen
	clicks, menus = new(int), new(int)
	b = newButton(t, widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { *clicks++ }),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.ContextMenu(widget.NewContainer())))
	b.SetLocation(img.Rect(20, 20, 120, 60))
	b.GetWidget().ContextMenuEvent.AddHandler(func(any) { *menus++ })
	step = func(fingers ...img.Point) {
		d.Touches = fingers
		input.Update()
		b.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
	}
	return b, clicks, menus, step
}

// A tap clicks, released where the finger lifts, on a desktop with a touch
// screen too. On a touch screen the pointer then leaves the screen, so the
// tapped widget doesn't stay hovered.
func TestTouch_TapLeavesNoHover(t *testing.T) {
	is := is.New(t)

	_, clicks, _, step := touchButton(t, false)
	step(img.Pt(70, 40))
	step()
	is.Equal(*clicks, 1) // on a desktop

	b, clicks, menus, step := touchButton(t, true)
	at := img.Pt(70, 40)
	step(at)
	step(at)
	step() // lifted: released where it was
	is.Equal(*clicks, 1)
	is.True(b.Hovering()) // the pointer over it, up
	step()                // the next update
	x, y := input.CursorPosition()
	is.Equal([]int{x, y}, []int{-1, -1})
	is.True(!b.Hovering())
	is.Equal(*menus, 0)
}

// Held still for the long press's time, a touch opens the widget's context
// menu, as a right click does, clicks nothing when it lifts, and stays as a
// hover, so a tooltip shows, until the next touch ends.
func TestTouch_LongPress(t *testing.T) {
	is := is.New(t)

	b, clicks, menus, step := touchButton(t, true)
	at := img.Pt(70, 40)
	for range b.GetWidget().LongPressDuration() + 1 {
		step(at)
	}
	is.Equal(*menus, 1)
	step()
	step()
	is.Equal(*clicks, 0)
	is.True(b.Hovering()) // still hovered

	step(at) // the next touch, a tap on it
	step()
	step()
	is.True(!b.Hovering())
	is.Equal(*clicks, 1)
}

// A second finger makes the touch a gesture: lifting the fingers clicks
// nothing, and holding them opens no context menu.
func TestTouch_SecondFingerCancels(t *testing.T) {
	is := is.New(t)

	b, clicks, menus, step := touchButton(t, true)
	at, other := img.Pt(70, 40), img.Pt(100, 50)
	step(at)
	for range b.GetWidget().LongPressDuration() + 1 {
		step(at, other)
	}
	step(at) // one lifts first
	step()
	is.Equal(*clicks, 0)
	is.Equal(*menus, 0)

	step(at) // a tap again is a tap
	step()
	is.Equal(*clicks, 1)
}

// A gesture ends the update after its fingers lift: on a desktop with a
// touch screen, the mouse clicks again after a pan, and after a long press
// a second finger joined.
func TestTouch_MouseClicksAfterAGesture(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	clicks := 0
	b := newButton(t, widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { clicks++ }))
	b.SetLocation(img.Rect(20, 20, 120, 60))
	update := func() {
		input.Update()
		b.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
	}
	d.Touches = []img.Point{{200, 200}, {300, 200}} // a pan, away from the button
	update()
	d.Touches = nil
	update()
	is.True(input.TouchGesture()) // the update they lift in
	update()
	is.True(!input.TouchGesture())
	click := func() {
		d.Cursor = img.Pt(70, 40)
		d.Mouse[ebiten.MouseButtonLeft] = true
		update()
		d.Mouse[ebiten.MouseButtonLeft] = false
		update()
	}
	click()
	is.Equal(clicks, 1)

	at := img.Pt(70, 40)
	for range b.GetWidget().LongPressDuration() + 1 {
		d.Touches = []img.Point{at} // a long press: the hover kept as it lifts
		update()
	}
	d.Touches = []img.Point{at, {100, 50}}
	update()
	d.Touches = nil
	update()
	update()
	is.True(!input.TouchGesture())
	click()
	is.Equal(clicks, 2)
}

// The mouse's long press keeps its meaning: it opens no context menu, and
// the release still clicks.
func TestTouch_MouseLongPressUnchanged(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	clicks, menus := 0, 0
	b := newButton(t, widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { clicks++ }),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.ContextMenu(widget.NewContainer())))
	b.SetLocation(img.Rect(0, 0, 100, 40))
	b.GetWidget().ContextMenuEvent.AddHandler(func(any) { menus++ })
	for range b.GetWidget().LongPressDuration() + 2 {
		point(d, 50, 20, true)
		b.Update(&widget.UpdateObject{})
		event.ExecuteDeferred()
	}
	point(d, 50, 20, false)
	b.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()
	is.Equal(menus, 0)
	is.Equal(clicks, 1)
}
