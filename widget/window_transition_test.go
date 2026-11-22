package widget_test

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

var transitionPlace = image.Rect(100, 100, 200, 150)

// A window with a transition comes in from its offset, fading in, and
// shows at its place once the transition is over, wherever it is placed.
func TestWindow_TransitionIn(t *testing.T) {
	is := is.New(t)

	w := newTransitionWindow(t)
	w.SetCloseFunction(func() {}) // as a UI does when adding it

	is.Equal(shownRect(w), transitionPlace.Add(image.Pt(0, 40))) // at its offset
	is.Equal(shownOpacity(w), float32(0))                        // faded out

	stepWindow(w, 2)
	r, o := shownRect(w), shownOpacity(w)
	is.True(r.Min.Y > 100 && r.Min.Y < 140) // on its way
	is.True(o > 0 && o < 1)                 // fading in

	stepWindow(w, 2)
	is.Equal(shownRect(w), transitionPlace) // in place
	is.Equal(shownOpacity(w), float32(1))   // opaque

	elsewhere := image.Rect(10, 10, 110, 60)
	w.SetLocation(elsewhere)
	is.Equal(shownRect(w), elsewhere)
}

// Closing runs the transition backward, the window placed anew on the way
// staying off its place; it is removed, and ClosedEvent fires, at the end.
func TestWindow_TransitionOut(t *testing.T) {
	is := is.New(t)

	w := newTransitionWindow(t)
	removed, closed := 0, 0
	w.ClosedEvent.AddHandler(func(any) { closed++ })
	w.SetCloseFunction(func() { removed++ })
	stepWindow(w, 4)

	w.Close()
	is.True(w.Closing())
	stepWindow(w, 2)
	is.Equal(removed, 0) // on its way out
	moved := transitionPlace.Add(image.Pt(30, 0))
	w.SetLocation(moved)
	r := shownRect(w)
	is.True(r.Min.X == moved.Min.X && r.Min.Y > moved.Min.Y) // placed anew, still off its place

	stepWindow(w, 2)
	is.Equal(removed, 1) // out
	is.Equal(closed, 1)
}

// A window turned back on its way out returns to its place and stays.
func TestWindow_TransitionReopen(t *testing.T) {
	is := is.New(t)

	w := newTransitionWindow(t)
	removed := 0
	w.SetCloseFunction(func() { removed++ })
	stepWindow(w, 4)

	w.Close()
	stepWindow(w, 2)
	w.Reopen()
	is.True(!w.Closing())
	stepWindow(w, 4)

	is.Equal(shownRect(w), transitionPlace)
	is.Equal(removed, 0)
}

// A window comes in by its In motion and goes out by its Out motion: from
// below fading in, then to the left without fading.
func TestWindow_TransitionMotions(t *testing.T) {
	is := is.New(t)

	w := widget.NewWindow(
		widget.WindowOpts.Contents(widget.NewContainer()),
		widget.WindowOpts.Transition(widget.WindowTransition{
			Duration: 4,
			In:       widget.WindowMotion{Offset: image.Pt(0, 40), Fade: true},
			Out:      widget.WindowMotion{Offset: image.Pt(-60, 0)},
		}),
	)
	w.SetLocation(transitionPlace)
	w.SetCloseFunction(func() {})

	stepWindow(w, 2)
	r := shownRect(w)
	is.True(r.Min.X == transitionPlace.Min.X && r.Min.Y > transitionPlace.Min.Y) // rising from below
	is.True(shownOpacity(w) < 1)                                                 // fading in
	stepWindow(w, 2)

	w.Close()
	stepWindow(w, 2)
	r = shownRect(w)
	is.True(r.Min.Y == transitionPlace.Min.Y && r.Min.X < transitionPlace.Min.X) // leaving to the left
	is.Equal(shownOpacity(w), float32(1))                                        // without fading
}

// A window with a still transition shows in place, opaque, all the way.
func TestWindow_TransitionStill(t *testing.T) {
	is := is.New(t)

	w := widget.NewWindow(widget.WindowOpts.Contents(widget.NewContainer()), widget.WindowOpts.Transition(widget.WindowTransition{Duration: 4}))
	w.SetLocation(transitionPlace)
	w.SetCloseFunction(func() {})
	is.Equal(shownRect(w), transitionPlace)
	is.Equal(shownOpacity(w), float32(1))
}

func newTransitionWindow(t *testing.T) *widget.Window {
	t.Helper()
	w := widget.NewWindow(
		widget.WindowOpts.Contents(widget.NewContainer()),
		widget.WindowOpts.Transition(widget.WindowTransition{
			Duration: 4,
			In:       widget.WindowMotion{Offset: image.Pt(0, 40), Fade: true},
			Out:      widget.WindowMotion{Offset: image.Pt(0, 40), Fade: true},
		}),
	)
	w.SetLocation(transitionPlace)
	return w
}

// stepWindow updates w n times, running the events each fires.
func stepWindow(w *widget.Window, n int) {
	for range n {
		w.Update(&widget.UpdateObject{})
		event.ExecuteDeferred()
	}
}

func shownRect(w *widget.Window) image.Rectangle { return w.GetContainer().GetWidget().Rect }

func shownOpacity(w *widget.Window) float32 {
	c, _ := w.GetContainer().(*widget.Container)
	return c.Opacity()
}
