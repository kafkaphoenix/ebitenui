package widget_test

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/matryer/is"
)

// newKeyboard returns a keyboard whose keys and preview have images and a
// face of their own, with opts, laid out.
func newKeyboard(t *testing.T, opts ...widget.OnScreenKeyboardOpt) *widget.OnScreenKeyboard {
	t.Helper()
	ns := newNineSliceEmpty(t)
	k := widget.NewOnScreenKeyboard(append([]widget.OnScreenKeyboardOpt{
		widget.OnScreenKeyboardOpts.KeyOpts(
			widget.ButtonOpts.Image(&widget.ButtonImage{Idle: ns, Pressed: ns}),
			widget.ButtonOpts.TextFace(loadFont(t)),
			widget.ButtonOpts.TextColor(&widget.ButtonTextColor{Idle: color.White}),
		),
		widget.OnScreenKeyboardOpts.PreviewOpts(widget.TextOpts.Text("", loadFont(t), color.White)),
	}, opts...)...)
	render(k.GetContainer(), t)
	return k
}

// The caller's key options win over the keyboard's own: wider padding
// makes wider keys.
func TestOnScreenKeyboard_KeyOptsWin(t *testing.T) {
	is := is.New(t)

	narrow, _ := key(t, newKeyboard(t), "1").PreferredSize()
	wide, _ := key(t, newKeyboard(t, widget.OnScreenKeyboardOpts.KeyOpts(widget.ButtonOpts.TextPadding(&widget.Insets{Left: 30, Right: 30}))), "1").PreferredSize()
	is.True(wide >= narrow+40)
}

// withChildren is a container of any type.
type withChildren interface {
	Children() []widget.PreferredSizeLocateableWidget
}

// key returns the keyboard's key labeled label.
func key(t *testing.T, k *widget.OnScreenKeyboard, label string) *widget.Button {
	t.Helper()
	var found *widget.Button
	var walk func(c withChildren)
	walk = func(c withChildren) {
		for _, ch := range c.Children() {
			switch w := ch.(type) {
			case *widget.Button:
				if w.TextLabel() == label {
					found = w
				}
			case *widget.Container:
				walk(w)
			}
		}
	}
	walk(k.GetContainer())
	if found == nil {
		t.Fatalf("no key %q", label)
	}
	return found
}

// press clicks the keys labeled labels, one after the other.
func press(t *testing.T, k *widget.OnScreenKeyboard, labels ...string) {
	t.Helper()
	for _, l := range labels {
		key(t, k, l).Click()
		event.ExecuteDeferred()
	}
}

// The keys type into the input the keyboard was opened for: characters,
// Shift switching the letters' case, Space, and Erase deleting the
// character before the caret. The line above the keys shows the input's
// text as the input does, a secure one as stars, from the moment it opens.
func TestOnScreenKeyboard_Types(t *testing.T) {
	is := is.New(t)

	k := newKeyboard(t)
	ti := newTextInput(t)
	ti.SetText("z")
	k.Open(ti, func() {})
	is.Equal(k.Preview().Label, "z")
	press(t, k, "a", "Shift", "B", "1", "Space", "Shift", "c", "Erase")
	is.Equal(ti.GetText(), "zaB1 ")
	is.Equal(k.Preview().Label, "zaB1 ")
	is.Equal(key(t, k, "q").TextLabel(), "q") // lower case again

	secure := newTextInput(t, widget.TextInputOpts.Secure(true))
	k.Open(secure, func() {})
	_, empty := k.GetContainer().PreferredSize()
	press(t, k, "x", "y")
	is.Equal(k.Preview().Label, "**")
	_, typed := k.GetContainer().PreferredSize()
	is.Equal(typed, empty) // as tall empty as typed in, so its window stays right
}

// Done submits the input and closes the keyboard, which then types
// nothing; Close closes it without submitting.
func TestOnScreenKeyboard_DoneAndClose(t *testing.T) {
	is := is.New(t)

	k := newKeyboard(t)
	submits := 0
	ti := newTextInput(t, widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) { submits++ }))
	closes := 0
	k.Open(ti, func() { closes++ })
	press(t, k, "a", "Done")
	is.Equal(closes, 1)
	is.Equal(submits, 1)
	is.True(!k.IsOpen())
	press(t, k, "b", "Erase")
	is.Equal(ti.GetText(), "a") // closed: types nothing

	k.Open(ti, func() { closes++ })
	is.True(k.IsOpen())
	k.Close()
	event.ExecuteDeferred()
	is.Equal(closes, 2)
	is.Equal(submits, 1) // not submitted
}

// Without a background of its own, the keyboard takes its theme's panel
// background and follows the theme when it changes; given one, it keeps it;
// without either, it has none.
func TestOnScreenKeyboard_ThemeBackground(t *testing.T) {
	is := is.New(t)

	dark := &widget.Theme{PanelTheme: &widget.PanelParams{BackgroundImage: newNineSliceEmpty(t)}}
	light := &widget.Theme{PanelTheme: &widget.PanelParams{BackgroundImage: newNineSliceEmpty(t)}}
	background := func(k *widget.OnScreenKeyboard) any {
		k.GetContainer().Validate()
		return k.Container().ComputedParams().BackgroundImage
	}

	k := newKeyboard(t)
	k.GetContainer().GetWidget().SetTheme(dark)
	is.True(background(k) == dark.PanelTheme.BackgroundImage)
	k.GetContainer().GetWidget().SetTheme(light)
	is.True(background(k) == light.PanelTheme.BackgroundImage)

	own := newNineSliceEmpty(t)
	k = newKeyboard(t, widget.OnScreenKeyboardOpts.ContainerOpts(widget.ContainerOpts.BackgroundImage(own)))
	k.GetContainer().GetWidget().SetTheme(dark)
	is.True(background(k) == own)

	k = newKeyboard(t)
	is.True(background(k) == (*image.NineSlice)(nil))
}

// Rows sets the character keys, row by row: another layout's, or letters a
// language needs, such as ñ or accented vowels. Shift switches every
// letter's case, those included, and leaves the rest. Rows of different
// lengths share one grid as wide as the longest, so the keys line up in
// columns; the first key takes the focus as the keyboard opens. Without
// rows, the keyboard keeps its own.
func TestOnScreenKeyboard_Rows(t *testing.T) {
	is := is.New(t)

	k := newKeyboard(t, widget.OnScreenKeyboardOpts.Rows("¿b", "ñáß")) // the shorter row first
	ti := newTextInput(t)
	k.Open(ti, func() {})
	press(t, k, "ñ", "Shift", "Á", "Ñ", "ß", "¿", "B")
	is.Equal(ti.GetText(), "ñÁÑß¿B")
	is.True(k.FirstKey() == widget.Focuser(key(t, k, "¿")))

	c := k.GetContainer()
	c.SetLocation(img.Rect(0, 0, 300, 200))
	c.RequestRelayout()
	render(c, t)
	rect := func(label string) img.Rectangle { return key(t, k, label).GetWidget().Rect }
	is.Equal([]int{rect("Ñ").Min.X, rect("Ñ").Dx()}, []int{rect("¿").Min.X, rect("¿").Dx()}) // in the first column
	is.Equal(rect("Á").Min.X, rect("B").Min.X)                                               // in the second
	is.True(rect("Ñ").Min.Y > rect("¿").Min.Y)                                               // a row down
	is.True(300-8-rect("ß").Max.X < 3)                                                       // the longest row spans the keyboard, inside its padding

	is.Equal(key(t, newKeyboard(t, widget.OnScreenKeyboardOpts.Rows()), "q").TextLabel(), "q")
}

// Each layout's rows hold every letter of the Latin alphabet and every
// digit, and no key twice, and the letters of its language.
func TestOnScreenKeyboard_Layouts(t *testing.T) {
	is := is.New(t)

	for name, l := range map[string]struct {
		rows  []string
		extra string
	}{
		"QWERTY":  {widget.KeyboardQWERTY, ""},
		"AZERTY":  {widget.KeyboardAZERTY, "éèêàçù"},
		"QWERTZ":  {widget.KeyboardQWERTZ, "üöäß"},
		"Spanish": {widget.KeyboardSpanish, "ñáéíóúü¿¡"},
	} {
		seen := map[rune]bool{}
		for _, row := range l.rows {
			for _, r := range row {
				is.True(!seen[r]) // no key twice
				seen[r] = true
			}
		}
		for _, r := range "abcdefghijklmnopqrstuvwxyz0123456789" + l.extra {
			is.Equal([]any{name, string(r), seen[r]}, []any{name, string(r), true})
		}
	}
	is.Equal(key(t, newKeyboard(t), "q").TextLabel(), "q") // QWERTY by default
	is.Equal(len(widget.NewOnScreenKeyboard().Letters()), 26)
}
