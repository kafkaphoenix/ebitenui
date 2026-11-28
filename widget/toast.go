package widget

import (
	img "image"
	"time"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/hajimehoshi/ebiten/v2"
)

// defaultToastDuration is how long a toast shows with no Duration set.
const defaultToastDuration = 3 * time.Second

// Toast shows a short message for a while, then hides: a gamepad plugged
// in, a game saved, a setting applied. Show shows a message in place of
// any showing, for the toast's Duration, fading in and out over its Fade
// when it has one. Its time is counted in updates at Ebitengine's TPS, so
// it pauses with the game's ticks; it works only while showing, and hidden
// it neither counts nor marks changes.
//
// A toast is a container of the message's text, hidden until shown. The
// game places it, as at the bottom of the screen with an AnchorLayoutData
// among its ContainerOpts, and gives it its background (ContainerOpts),
// padding (Padding) and text's look (TextOpts).
type Toast struct {
	duration, fade time.Duration
	padding        *Insets
	containerOpts  []ContainerOpt
	textOpts       []TextOpt

	init      *MultiOnce
	container *Container
	text      *Text
	message   string
	left      int // updates it still shows
	shown     int // updates it has shown
}

type ToastOpt func(t *Toast)

type ToastOptions struct{}

// ToastOpts holds the options of NewToast.
var ToastOpts ToastOptions

// NewToast returns a toast, hidden.
func NewToast(opts ...ToastOpt) *Toast {
	t := &Toast{duration: defaultToastDuration, init: &MultiOnce{}}
	t.init.Append(t.createWidget)
	for _, o := range opts {
		o(t)
	}
	return t
}

// Duration sets how long a message shows, fades included; by default 3
// seconds.
func (o ToastOptions) Duration(d time.Duration) ToastOpt {
	return func(t *Toast) { t.duration = d }
}

// Fade sets how long a message takes to fade in as it shows, and out as
// its time ends; by default 0, shown and hidden at once.
func (o ToastOptions) Fade(d time.Duration) ToastOpt {
	return func(t *Toast) { t.fade = d }
}

// Padding sets the room around the message.
func (o ToastOptions) Padding(i *Insets) ToastOpt {
	return func(t *Toast) { t.padding = i }
}

// ContainerOpts configures the toast's container: its background, widget
// options and layout data. The toast sets its layout.
func (o ToastOptions) ContainerOpts(opts ...ContainerOpt) ToastOpt {
	return func(t *Toast) { t.containerOpts = append(t.containerOpts, opts...) }
}

// TextOpts configures the message's text: its face, color and position.
// The toast sets its label.
func (o ToastOptions) TextOpts(opts ...TextOpt) ToastOpt {
	return func(t *Toast) { t.textOpts = append(t.textOpts, opts...) }
}

// Show shows message for the toast's duration, in place of any message
// showing: a toast hidden fades in, one showing shows the new message at
// once, its time starting over.
func (t *Toast) Show(message string) {
	t.init.Do()
	fade := t.fadeUpdates()
	if t.left == 0 || fade == 0 {
		t.shown = 0
	} else {
		t.shown = fade // already in
	}
	t.message, t.left = message, updatesIn(t.duration)
	t.text.Label = message
	if fade > 0 {
		t.container.SetOpacity(t.opacity())
	}
	t.container.GetWidget().SetVisibility(VisibilityShow)
}

// Hide hides the toast at once.
func (t *Toast) Hide() {
	t.init.Do()
	t.message, t.left = "", 0
	t.container.GetWidget().SetVisibility(VisibilityHide)
}

// Message returns the message showing, or "" while the toast is hidden.
func (t *Toast) Message() string { return t.message }

// Container returns the toast's container.
func (t *Toast) Container() *Container {
	t.init.Do()
	return t.container
}

// Text returns the message's text.
func (t *Toast) Text() *Text {
	t.init.Do()
	return t.text
}

// GetWidget implements HasWidget.
func (t *Toast) GetWidget() *Widget {
	t.init.Do()
	return t.container.GetWidget()
}

// PreferredSize implements PreferredSizer.
func (t *Toast) PreferredSize() (int, int) {
	t.init.Do()
	return t.container.PreferredSize()
}

// SetLocation implements Locateable.
func (t *Toast) SetLocation(rect img.Rectangle) {
	t.init.Do()
	t.container.SetLocation(rect)
}

func (t *Toast) Validate() {
	t.init.Do()
	t.container.Validate()
}

// RequestRelayout implements Relayoutable.
func (t *Toast) RequestRelayout() {
	t.init.Do()
	t.container.RequestRelayout()
}

// SetupInputLayer implements InputLayerer.
func (t *Toast) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	t.init.Do()
	t.container.SetupInputLayer(def)
}

// Render implements Renderer.
func (t *Toast) Render(screen *ebiten.Image) {
	t.init.Do()
	t.container.Render(screen)
}

// Update counts an update of the message showing, fading it, and hides it
// once its time is up; while hidden it does nothing.
func (t *Toast) Update(updObj *UpdateObject) {
	t.init.Do()
	t.container.Update(updObj)
	if t.left == 0 {
		return
	}
	t.left--
	t.shown++
	if t.left == 0 {
		t.Hide()
		changes.Mark() // gone by itself
		return
	}
	if t.fadeUpdates() > 0 {
		if o := t.opacity(); o != t.container.Opacity() {
			t.container.SetOpacity(o)
			changes.Mark()
		}
	}
}

// fadeUpdates returns how many updates a fade lasts, 0 for none.
func (t *Toast) fadeUpdates() int {
	if t.fade <= 0 {
		return 0
	}
	return updatesIn(t.fade)
}

// opacity returns the opacity the message shows at: rising over the fade
// from its start, falling over the fade to its end.
func (t *Toast) opacity() float32 {
	f := float32(t.fadeUpdates())
	return min(1, float32(t.shown)/f, float32(t.left)/f)
}

func (t *Toast) createWidget() {
	t.text = NewText(t.textOpts...)
	t.container = NewContainer(append(t.containerOpts, ContainerOpts.Layout(NewRowLayout(RowLayoutOpts.Padding(t.padding))))...)
	t.container.AddChild(t.text)
	t.container.GetWidget().SetVisibility(VisibilityHide)
	t.containerOpts, t.textOpts = nil, nil
}
