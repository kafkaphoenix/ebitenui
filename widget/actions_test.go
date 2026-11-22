package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// given returns actions giving acts.
func given(acts ...widget.Action) *widget.Actions {
	a := &widget.Actions{}
	for _, act := range acts {
		a.Give(act)
	}
	return a
}

func label(e any) string { return e.(string) }

// A focused list moves through its entries with Up and Down and leaves
// those at its ends, so a gamepad moves the focus on from there; Activate
// selects the entry it is on. Unfocused, or with its default keys off, it
// takes nothing.
func TestList_TakesActions(t *testing.T) {
	is := is.New(t)

	l := newList(t, widget.ListOpts.Entries([]any{"a", "b", "c"}), widget.ListOpts.EntryLabelFunc(label))
	a := given(widget.ActionDown)
	l.Update(&widget.UpdateObject{Actions: a})
	is.True(a.Left(widget.ActionDown)) // not focused

	l.Focus(true)
	for i, s := range []struct {
		act   widget.Action
		taken bool
		at    int
	}{
		{widget.ActionUp, false, 0}, // the first entry: left
		{widget.ActionDown, true, 1},
		{widget.ActionDown, true, 2},
		{widget.ActionDown, false, 2}, // the last: left
		{widget.ActionUp, true, 1},
	} {
		step := given(s.act)
		l.Update(&widget.UpdateObject{Actions: step})
		is.Equal([]any{i, !step.Left(s.act), l.FocusIndex()}, []any{i, s.taken, s.at})
	}

	a = given(widget.ActionActivate)
	l.Update(&widget.UpdateObject{Actions: a})
	event.ExecuteDeferred()
	is.True(!a.Left(widget.ActionActivate))
	is.Equal(l.SelectedEntry(), "b")

	off := newList(t, widget.ListOpts.Entries([]any{"a", "b"}), widget.ListOpts.EntryLabelFunc(label), widget.ListOpts.DisableDefaultKeys(true))
	off.Focus(true)
	a = given(widget.ActionDown, widget.ActionActivate)
	off.Update(&widget.UpdateObject{Actions: a})
	is.True(a.Left(widget.ActionDown) && a.Left(widget.ActionActivate)) // default keys off
}

// A focused slider moves its handle a page with the two directions of its
// orientation, also at its ends, and leaves the other two; unfocused, or
// with its default keys off, it takes nothing.
func TestSlider_TakesActions(t *testing.T) {
	is := is.New(t)

	for _, c := range []struct {
		dir               widget.Direction
		less, more, other widget.Action
	}{
		{widget.DirectionHorizontal, widget.ActionLeft, widget.ActionRight, widget.ActionUp},
		{widget.DirectionVertical, widget.ActionUp, widget.ActionDown, widget.ActionRight},
	} {
		s := newSlider(t, widget.SliderOpts.Direction(c.dir), widget.SliderOpts.MinMax(0, 20), widget.SliderOpts.InitialCurrent(0),
			widget.SliderOpts.PageSizeFunc(func() int { return 10 }))
		a := given(c.more)
		s.Update(&widget.UpdateObject{Actions: a})
		is.True(a.Left(c.more)) // not focused

		s.Focus(true)
		for i, step := range []struct {
			act     widget.Action
			taken   bool
			current int
		}{
			{c.more, true, 10},
			{c.more, true, 20},
			{c.more, true, 20}, // at the end: still taken
			{c.less, true, 10},
			{c.other, false, 10},
		} {
			a := given(step.act)
			s.Update(&widget.UpdateObject{Actions: a})
			is.Equal([]any{c.dir, i, !a.Left(step.act), s.Current}, []any{c.dir, i, step.taken, step.current}) // within its range before it draws
		}
	}

	off := newSlider(t, widget.SliderOpts.DisableDefaultKeys(true))
	off.Focus(true)
	a := given(widget.ActionRight)
	off.Update(&widget.UpdateObject{Actions: a})
	is.True(a.Left(widget.ActionRight))

	early := widget.NewSlider() // focused before its first layout
	early.Focus(true)
	early.Update(&widget.UpdateObject{Actions: a})
	is.True(a.Left(widget.ActionRight))
}

// A focused list combo button opens with Activate. Open, its list takes Up,
// Down and Activate, which selects and closes, and the combo keeps the
// directions its list leaves, so the focus stays in it. Closed, it leaves
// the directions, though the Down and Up keys still open it.
func TestListComboButton_TakesActions(t *testing.T) {
	is := is.New(t)

	l := newListComboButton(t, widget.ListComboButtonOpts.Entries([]any{"a", "b", "c"}), widget.ListComboButtonOpts.EntryLabelFunc(label, label))
	update := func(acts ...widget.Action) *widget.Actions { // drawn without validating, which closes it
		a := given(acts...)
		l.Update(&widget.UpdateObject{Actions: a})
		l.Render(ebiten.NewImage(1, 1))
		event.ExecuteDeferred()
		return a
	}
	a := update(widget.ActionActivate)
	is.True(a.Left(widget.ActionActivate) && !l.ContentVisible()) // not focused

	l.Focus(true)
	a = update(widget.ActionDown)
	is.True(a.Left(widget.ActionDown) && !l.ContentVisible()) // closed: left

	a = update(widget.ActionActivate)
	is.True(!a.Left(widget.ActionActivate) && l.ContentVisible()) // opens

	for i := range 3 { // to the last entry, and once more
		a = update(widget.ActionDown)
		is.Equal([]any{i, a.Left(widget.ActionDown)}, []any{i, false})
	}
	a = update(widget.ActionLeft)
	is.True(!a.Left(widget.ActionLeft)) // open: no direction leaves it

	a = update(widget.ActionActivate)
	is.True(!a.Left(widget.ActionActivate))
	is.Equal(l.SelectedEntry(), "c")
	is.True(!l.ContentVisible()) // selected and closed

	a = update(widget.ActionBack)
	is.True(a.Left(widget.ActionBack)) // closed: Back is the UI's
	update(widget.ActionActivate)
	a = update(widget.ActionBack)
	is.True(!a.Left(widget.ActionBack) && !l.ContentVisible()) // open: Back closes it
	is.Equal(l.SelectedEntry(), "c")                           // choosing nothing

	d := uitest.Use(t)
	d.Keys[ebiten.KeyArrowDown] = true
	for _, open := range []bool{true, false} {
		input.Update()
		update()
		input.AfterUpdate()
		is.Equal(l.ContentVisible(), open) // the Down key opens it, once
		l.SetContentVisible(false)
	}

	off := newListComboButton(t, widget.ListComboButtonOpts.Entries([]any{"a"}), widget.ListComboButtonOpts.EntryLabelFunc(label, label),
		widget.ListComboButtonOpts.DisableDefaultKeys(true))
	off.Focus(true)
	a = given(widget.ActionActivate)
	off.Update(&widget.UpdateObject{Actions: a})
	is.True(a.Left(widget.ActionActivate) && !off.ContentVisible()) // default keys off
}
