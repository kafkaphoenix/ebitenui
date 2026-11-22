package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

func TestCheckbox_State_Initial(t *testing.T) {
	is := is.New(t)

	c := newCheckbox(t,
		widget.CheckboxOpts.StateChangedHandler(func(_ *widget.CheckboxChangedEventArgs) {
			is.Fail() // event fired without previous action
		}))

	is.Equal(c.State(), widget.Unchecked)
}

func TestCheckbox_State_SetInitial(t *testing.T) {
	is := is.New(t)

	c := newCheckbox(t,
		widget.CheckboxOpts.InitialState(widget.Checked),
		widget.CheckboxOpts.StateChangedHandler(func(_ *widget.CheckboxChangedEventArgs) {
			is.Fail() // event fired without previous action
		}))

	is.Equal(c.State(), widget.Checked)
}

func TestCheckbox_State_SetInitialTri(t *testing.T) {
	is := is.New(t)

	c := newCheckbox(t,
		widget.CheckboxOpts.TriState(),
		widget.CheckboxOpts.InitialState(widget.Greyed),
		widget.CheckboxOpts.StateChangedHandler(func(_ *widget.CheckboxChangedEventArgs) {
			is.Fail() // event fired without previous action
		}))

	is.Equal(c.State(), widget.Greyed)
}

func TestCheckbox_State_SetInitialPanic(t *testing.T) {
	is := is.New(t)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	newCheckbox(t,
		widget.CheckboxOpts.InitialState(widget.Greyed),
		widget.CheckboxOpts.StateChangedHandler(func(_ *widget.CheckboxChangedEventArgs) {
			is.Fail() // event fired without previous action
		}))
}

func TestCheckbox_ChangedEvent_User(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.CheckboxChangedEventArgs

	c := newCheckbox(t,
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			eventArgs = args
		}))

	leftMouseButtonClick(c, t)

	is.Equal(eventArgs.State, widget.Checked)
	is.Equal(c.State(), widget.Checked)
}

func TestCheckbox_Click(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.CheckboxChangedEventArgs

	c := newCheckbox(t,
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			eventArgs = args
		}))

	c.Click()
	event.ExecuteDeferred()

	is.Equal(eventArgs.State, widget.Checked)
	is.Equal(c.State(), widget.Checked)
}

func TestCheckbox_SetState(t *testing.T) {
	is := is.New(t)

	var eventArgs *widget.CheckboxChangedEventArgs
	numEvents := 0

	c := newCheckbox(t,
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	c.SetState(widget.Checked)
	event.ExecuteDeferred()

	is.Equal(eventArgs.State, widget.Checked)
	is.Equal(c.State(), widget.Checked)

	c.SetState(widget.Checked)
	event.ExecuteDeferred()

	is.Equal(numEvents, 1)
}

func TestCheckbox_State_Cycle(t *testing.T) {
	is := is.New(t)

	c := newCheckbox(t)
	leftMouseButtonClick(c, t)
	is.Equal(c.State(), widget.Checked)
	leftMouseButtonClick(c, t)
	is.Equal(c.State(), widget.Unchecked)
}

func TestCheckbox_State_Cycle_TriState(t *testing.T) {
	is := is.New(t)

	c := newCheckbox(t, widget.CheckboxOpts.TriState())
	leftMouseButtonClick(c, t)
	is.Equal(c.State(), widget.Checked)
	leftMouseButtonClick(c, t)
	is.Equal(c.State(), widget.Greyed)
	leftMouseButtonClick(c, t)
	is.Equal(c.State(), widget.Unchecked)
}

// A focused checkbox toggles once each time Enter or Space goes down, as the
// input package reports the keys.
func TestCheckbox_KeysToggleOncePerPress(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	c := newCheckbox(t)
	c.Focus(true)
	for i, s := range []struct {
		key   ebiten.Key
		down  bool
		state widget.CheckState
	}{
		{ebiten.KeyEnter, true, widget.Checked},
		{ebiten.KeyEnter, true, widget.Checked}, // held
		{ebiten.KeyEnter, false, widget.Checked},
		{ebiten.KeySpace, true, widget.Unchecked},
	} {
		d.Keys[s.key] = s.down
		input.Update()
		c.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
		is.Equal([]any{i, c.State()}, []any{i, s.state})
	}
}

func newCheckbox(t *testing.T, opts ...widget.CheckboxOpt) *widget.Checkbox {
	t.Helper()

	c := widget.NewCheckbox(append(opts, []widget.CheckboxOpt{
		widget.CheckboxOpts.Image(&widget.CheckboxImage{
			Unchecked: newNineSliceEmpty(t),
			Checked:   newNineSliceEmpty(t),
			Greyed:    newNineSliceEmpty(t),
		})}...)...)

	event.ExecuteDeferred()
	render(c, t)
	return c
}

func TestCheckbox_HoveredImage(t *testing.T) {
	for _, state := range []widget.CheckState{widget.Unchecked, widget.Checked, widget.Greyed} {
		is := is.New(t)
		images := newCheckboxImages(t)
		c := newCheckboxWithImages(t, images, state)
		normal, _, hovered := checkboxImagesFor(images, state)

		moveCursor(c, 10, t)
		is.Equal(c.CurrentImage(), hovered) // cursor over the checkbox

		moveCursor(c, 30, t)
		is.Equal(c.CurrentImage(), normal) // cursor gone

		c.Focus(true)
		is.Equal(c.CurrentImage(), hovered) // focused, as when tabbed to
	}
}

func TestCheckbox_DisabledImage(t *testing.T) {
	for _, state := range []widget.CheckState{widget.Unchecked, widget.Checked, widget.Greyed} {
		is := is.New(t)
		images := newCheckboxImages(t)
		normal, disabled, _ := checkboxImagesFor(images, state)

		c := newCheckboxWithImages(t, images, state)
		moveCursor(c, 10, t)
		c.GetWidget().Disabled = true
		is.Equal(c.CurrentImage(), disabled) // disabled, though hovered

		// Without disabled images: the normal image, not the hovered one.
		images.UncheckedDisabled, images.CheckedDisabled, images.GreyedDisabled = nil, nil, nil
		c = newCheckboxWithImages(t, images, state)
		moveCursor(c, 10, t)
		c.GetWidget().Disabled = true
		is.Equal(c.CurrentImage(), normal)
	}
}

// Only the disabled image for the checkbox's state counts.
func TestCheckbox_DisabledImageOfItsState(t *testing.T) {
	is := is.New(t)

	images := newCheckboxImages(t)
	images.UncheckedDisabled = nil
	c := newCheckboxWithImages(t, images, widget.Greyed)
	c.GetWidget().Disabled = true
	is.Equal(c.CurrentImage(), images.GreyedDisabled)

	images = newCheckboxImages(t)
	images.GreyedDisabled = nil
	c = newCheckboxWithImages(t, images, widget.Greyed)
	c.GetWidget().Disabled = true
	is.Equal(c.CurrentImage(), images.Greyed)
}

// newCheckboxWithImages returns a 20x20 tri-state checkbox at the origin,
// drawn with images, in state.
func newCheckboxWithImages(t *testing.T, images *widget.CheckboxImage, state widget.CheckState) *widget.Checkbox {
	t.Helper()
	c := widget.NewCheckbox(widget.CheckboxOpts.TriState(), widget.CheckboxOpts.InitialState(state), widget.CheckboxOpts.Image(images))
	event.ExecuteDeferred()
	render(c, t)
	c.SetLocation(img.Rect(0, 0, 20, 20))
	return c
}

// newCheckboxImages returns checkbox images that are all different.
func newCheckboxImages(t *testing.T) *widget.CheckboxImage {
	t.Helper()
	return &widget.CheckboxImage{
		Unchecked: newNineSliceEmpty(t), UncheckedHovered: newNineSliceEmpty(t), UncheckedDisabled: newNineSliceEmpty(t),
		Checked: newNineSliceEmpty(t), CheckedHovered: newNineSliceEmpty(t), CheckedDisabled: newNineSliceEmpty(t),
		Greyed: newNineSliceEmpty(t), GreyedHovered: newNineSliceEmpty(t), GreyedDisabled: newNineSliceEmpty(t),
	}
}

// checkboxImagesFor returns the normal, disabled and hovered images for state.
func checkboxImagesFor(i *widget.CheckboxImage, state widget.CheckState) (normal, disabled, hovered *image.NineSlice) {
	switch state {
	case widget.Unchecked:
		return i.Unchecked, i.UncheckedDisabled, i.UncheckedHovered
	case widget.Checked:
		return i.Checked, i.CheckedDisabled, i.CheckedHovered
	case widget.Greyed:
		return i.Greyed, i.GreyedDisabled, i.GreyedHovered
	}
	return nil, nil, nil
}

// moveCursor puts the cursor at x on the check box's row, y 10, and
// updates w, so that it sees the cursor enter or leave.
func moveCursor(w widget.Updater, x int, t *testing.T) {
	t.Helper()
	point(uitest.Use(t), x, 10, false)
	w.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()
}
