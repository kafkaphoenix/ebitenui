package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

func TestRadioGroup_Active_Initial(t *testing.T) {
	is := is.New(t)

	cbs := make([]*widget.Checkbox, 0, 3)
	for range 3 {
		c := newCheckbox(t)
		cbs = append(cbs, c)
	}

	var eventArgs *widget.RadioGroupChangedEventArgs

	r := newRadioGroup(t, cbs, widget.RadioGroupOpts.ChangedHandler(func(args *widget.RadioGroupChangedEventArgs) {
		eventArgs = args
	}))

	is.Equal(r.Active(), cbs[0])
	is.Equal(eventArgs.Active, cbs[0])
	is.Equal(cbs[0].State(), widget.Checked)
}

func TestRadioGroup_ChangedEvent_User(t *testing.T) {
	is := is.New(t)

	cbs := make([]*widget.Checkbox, 0, 3)
	for range 3 {
		c := newCheckbox(t)
		cbs = append(cbs, c)
	}

	r := newRadioGroup(t, cbs)

	var eventArgs *widget.RadioGroupChangedEventArgs
	r.ChangedEvent.AddHandler(func(args any) {
		if a, ok := args.(*widget.RadioGroupChangedEventArgs); ok {
			eventArgs = a
		}
	})

	leftMouseButtonClick(cbs[1], t)

	is.Equal(r.Active(), cbs[1])
	is.Equal(eventArgs.Active, cbs[1])
	is.Equal(cbs[0].State(), widget.Unchecked)
	is.Equal(cbs[1].State(), widget.Checked)
}

func TestRadioGroup_SetActive(t *testing.T) {
	is := is.New(t)

	cbs := make([]*widget.Checkbox, 0, 3)
	for range 3 {
		c := newCheckbox(t)
		cbs = append(cbs, c)
	}

	r := newRadioGroup(t, cbs)

	var eventArgs *widget.RadioGroupChangedEventArgs
	r.ChangedEvent.AddHandler(func(args any) {
		if a, ok := args.(*widget.RadioGroupChangedEventArgs); ok {
			eventArgs = a
		}
	})

	r.SetActive(cbs[1])
	event.ExecuteDeferred()

	is.Equal(r.Active(), cbs[1])
	is.Equal(eventArgs.Active, cbs[1])
	is.Equal(cbs[0].State(), widget.Unchecked)
	is.Equal(cbs[1].State(), widget.Checked)
}

func newRadioGroup(t *testing.T, cbs []*widget.Checkbox, opts ...widget.RadioGroupOpt) *widget.RadioGroup {
	t.Helper()

	elements := make([]widget.RadioGroupElement, 0, len(cbs))

	for _, cb := range cbs {
		elements = append(elements, cb)
	}
	r := widget.NewRadioGroup(append(opts, widget.RadioGroupOpts.Elements(elements...))...)
	event.ExecuteDeferred()
	for _, c := range cbs {
		render(c, t)
	}
	return r
}
