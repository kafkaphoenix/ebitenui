package widget_test

import (
	img "image"
	"image/color"
	"strconv"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

func TestList_SelectedEntry_Initial(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	list := newList(t,
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return result
		}),

		widget.ListOpts.EntrySelectedHandler(func(_ *widget.ListEntrySelectedEventArgs) {
			is.Fail() // event fired without previous action
		}))

	is.Equal(list.SelectedEntry(), nil)
}

func TestList_NoSliderOpts(t *testing.T) {
	entries := []any{"first", "second", "third"}
	_ = widget.NewList(
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return result
		}),

		widget.ListOpts.EntrySelectedHandler(func(_ *widget.ListEntrySelectedEventArgs) {
		}),
		widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
			Idle:     newNineSliceEmpty(t),
			Disabled: newNineSliceEmpty(t),
			Mask:     newNineSliceEmpty(t),
		}),

		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.HideVerticalSlider(),

		widget.ListOpts.EntryFontFace(loadFont(t)),

		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Unselected:                 color.Transparent,
			Selected:                   color.Transparent,
			DisabledUnselected:         color.Transparent,
			DisabledSelected:           color.Transparent,
			SelectedBackground:         color.Transparent,
			DisabledSelectedBackground: color.Transparent,
		}),
	)
}

func TestList_SetSelectedEntry(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	var eventArgs *widget.ListEntrySelectedEventArgs
	numEvents := 0

	list := newList(t,
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return result
		}),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	list.SetSelectedEntry(entries[1])
	event.ExecuteDeferred()

	is.Equal(eventArgs.Entry, entries[1])
	is.Equal(list.SelectedEntry(), entries[1])

	list.SetSelectedEntry(entries[1])
	event.ExecuteDeferred()

	is.Equal(numEvents, 1)
}

func TestList_UpdateEntry(t *testing.T) {
	is := is.New(t)

	type listEntry struct {
		label string
	}

	entries := []any{
		&listEntry{label: "first"},
		&listEntry{label: "second"},
		&listEntry{label: "third"},
	}

	list := newList(t,
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			return e.(*listEntry).label
		}),
	)

	entry := entries[1].(*listEntry)

	is.Equal(shownLabel(list, 1), "second")

	screen := ebiten.NewImage(200, 300)
	entry.label = "updated"
	list.Render(screen)
	is.Equal(shownLabel(list, 1), "second") // a row shows what its entry was as it was bound

	list.UpdateEntry(entry)
	list.Render(screen)
	is.Equal(shownLabel(list, 1), "updated")
}

func TestList_EntrySelectedEvent_IncludesList(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	var eventArgs *widget.ListEntrySelectedEventArgs
	var selectedEntry any

	list := newList(t,
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return result
		}),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			eventArgs = args
			selectedEntry = args.List.SelectedEntry()
		}))

	list.SetSelectedEntry(entries[1])
	event.ExecuteDeferred()

	is.Equal(eventArgs.List, list)
	is.Equal(selectedEntry, entries[1])
}

func TestList_EntrySelectedEvent_User(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	var eventArgs *widget.ListEntrySelectedEventArgs
	numEvents := 0

	list := newList(t,
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return result
		}),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	clickListEntry(t, list, 1)

	is.Equal(eventArgs.Entry, entries[1])
	is.Equal(list.SelectedEntry(), entries[1])

	clickListEntry(t, list, 1)

	is.Equal(numEvents, 1)
}

func TestList_EntrySelectedEvent_User_AllowReselect(t *testing.T) {
	is := is.New(t)

	entries := []any{"first", "second", "third"}

	var eventArgs *widget.ListEntrySelectedEventArgs
	numEvents := 0

	list := newList(t,
		widget.ListOpts.Entries(entries),

		widget.ListOpts.EntryLabelFunc(func(e any) string {
			result, _ := e.(string)
			return result
		}),

		widget.ListOpts.AllowReselect(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			eventArgs = args
			numEvents++
		}))

	clickListEntry(t, list, 1)

	is.Equal(eventArgs.Entry, entries[1])
	is.Equal(list.SelectedEntry(), entries[1])

	clickListEntry(t, list, 1)

	is.Equal(eventArgs.Entry, entries[1])
	is.Equal(eventArgs.PreviousEntry, entries[1])
	is.Equal(list.SelectedEntry(), entries[1])

	is.Equal(numEvents, 2)
}

func TestList_Entries_SortedOnCreate(t *testing.T) {
	is := is.New(t)

	list := newList(t,
		widget.ListOpts.Entries([]any{"third", "first", "second"}),
		widget.ListOpts.EntrySortFunc(func(a, b any) int {
			left := a.(string)
			right := b.(string)
			switch {
			case left < right:
				return -1
			case left > right:
				return 1
			default:
				return 0
			}
		}),
		widget.ListOpts.EntryLabelFunc(func(e any) string {
			return e.(string)
		}),
	)

	is.Equal(list.Entries()[0], "first")
	is.Equal(list.Entries()[1], "second")
	is.Equal(list.Entries()[2], "third")
	is.Equal(shownLabel(list, 0), "first")
	is.Equal(shownLabel(list, 1), "second")
	is.Equal(shownLabel(list, 2), "third")
}

func TestList_AddEntry_Resorts(t *testing.T) {
	is := is.New(t)

	list := newList(t,
		widget.ListOpts.Entries([]any{2, 4}),
		widget.ListOpts.EntrySortFunc(func(a, b any) int {
			return a.(int) - b.(int)
		}),
		widget.ListOpts.EntryLabelFunc(func(e any) string {
			return strconv.Itoa(e.(int))
		}),
	)

	list.AddEntry(3)
	render(list, t)

	is.Equal(list.Entries(), []any{2, 3, 4})
	is.Equal([]string{shownLabel(list, 0), shownLabel(list, 1), shownLabel(list, 2)}, []string{"2", "3", "4"})
}

func newList(t *testing.T, opts ...widget.ListOpt) *widget.List {
	t.Helper()

	l := widget.NewList(append(opts, []widget.ListOpt{
		widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
			Idle:     newNineSliceEmpty(t),
			Disabled: newNineSliceEmpty(t),
			Mask:     newNineSliceEmpty(t),
		}),

		widget.ListOpts.SliderParams(&widget.SliderParams{
			TrackImage: &widget.SliderTrackImage{},
			HandleImage: &widget.ButtonImage{
				Idle:    newNineSliceEmpty(t),
				Pressed: newNineSliceEmpty(t),
			}}),

		widget.ListOpts.EntryFontFace(loadFont(t)),

		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Unselected:                 color.Transparent,
			Selected:                   color.Transparent,
			DisabledUnselected:         color.Transparent,
			DisabledSelected:           color.Transparent,
			SelectedBackground:         color.Transparent,
			DisabledSelectedBackground: color.Transparent,
		}),
	}...)...)

	l.SetLocation(img.Rect(0, 0, 200, 300))
	event.ExecuteDeferred()
	render(l, t)
	render(l, t) // the rows in view, laid out
	return l
}

// shownLabel returns the label entry i's row shows, "" when it isn't in
// view.
func shownLabel(l *widget.List, i int) string {
	for _, r := range l.Rows().Labels() {
		if r.Entry() == i {
			return r.Text().Label
		}
	}
	return ""
}

// clickListEntry clicks the middle of entry i's row.
func clickListEntry(t *testing.T, l *widget.List, i int) {
	t.Helper()
	top, bottom := l.Rows().RowSpan(i)
	w := l.Rows().Widget()
	w.MouseButtonClickedEvent.Fire(&widget.MouseButtonClickedEventArgs{Widget: w, Button: ebiten.MouseButtonLeft, OffsetY: (top + bottom) / 2})
	event.ExecuteDeferred()
}

// Entries can be widgets, built per entry, with neither a label function nor
// a font: each shows inside its entry's button, clicking the button selects
// the entry, and UpdateEntry builds the widget again.
func TestList_EntryWidgets(t *testing.T) {
	is := is.New(t)

	built := map[any]int{}
	widgets := map[any]*simpleWidget{}
	var selected any
	l := widget.NewList(
		widget.ListOpts.Entries([]any{"a", "b"}),
		widget.ListOpts.EntryWidgetFunc(func(e any) widget.PreferredSizeLocateableWidget {
			built[e]++
			widgets[e] = newSimpleWidget(30, 15, nil)
			return widgets[e]
		}),
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) { selected = args.Entry }),
		widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
		widget.ListOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Unselected: color.Transparent, Selected: color.Transparent, DisabledUnselected: color.Transparent,
			DisabledSelected: color.Transparent, SelectedBackground: color.Transparent, DisabledSelectedBackground: color.Transparent,
		}),
	)
	l.SetLocation(img.Rect(0, 0, 100, 100))
	event.ExecuteDeferred()
	render(l, t)

	is.Equal(len(built), 2)
	is.True(!widgets["b"].GetWidget().Rect.Empty() && widgets["b"].GetWidget().Rect.In(l.Rows().RowRect(1))) // inside its row

	clickListEntry(t, l, 1)
	is.Equal(selected, "b") // selected by clicking
	clickListEntry(t, l, 0)
	is.Equal(selected, "a") // the entry clicked, row by row

	l.UpdateEntry("b")
	is.Equal(built["b"], 2) // built again
}

// newEntryImages returns entry images, each its own.
func newEntryImages(t *testing.T) *widget.ListEntryImage {
	t.Helper()
	return &widget.ListEntryImage{
		Unselected: newNineSliceEmpty(t), DisabledUnselected: newNineSliceEmpty(t),
		Focused: newNineSliceEmpty(t), Selecting: newNineSliceEmpty(t),
		Selected: newNineSliceEmpty(t), DisabledSelected: newNineSliceEmpty(t),
		SelectedFocused: newNineSliceEmpty(t), SelectingFocused: newNineSliceEmpty(t),
	}
}

// isEntryImages checks that a list draws its entries with images, each state
// its own.
func isEntryImages(is *is.I, l *widget.List, images *widget.ListEntryImage) {
	is.Helper()
	unselected, selected := l.ComputedParams().EntryUnselectedColor(), l.ComputedParams().EntrySelectedColor()
	is.True(unselected.Idle == images.Unselected)
	is.True(unselected.Disabled == images.DisabledUnselected)
	is.True(unselected.Hover == images.Focused)
	is.True(unselected.Pressed == images.Selecting)
	is.True(selected.Idle == images.Selected)
	is.True(selected.Disabled == images.DisabledSelected)
	is.True(selected.Hover == images.SelectedFocused)
	is.True(selected.Pressed == images.SelectingFocused)
}

// Entry images draw the entries' backgrounds, rounded or bordered, in place
// of their colors, each state its own: the selected entry's button takes the
// selected ones. A state without an image keeps its color, and an
// unselected entry without one stays clear.
func TestList_EntryImages(t *testing.T) {
	is := is.New(t)

	images := newEntryImages(t)
	l := newList(t, widget.ListOpts.EntryImage(images), widget.ListOpts.Entries([]any{"a", "b"}),
		widget.ListOpts.EntryLabelFunc(func(e any) string { return e.(string) }))
	isEntryImages(is, l, images)
	l.SetSelectedEntry("b")
	is.True(l.Rows().Background(1) == images.Selected)   // the selected entry
	is.True(l.Rows().Background(0) == images.Unselected) // the others

	focused := color.NRGBA{1, 2, 3, 255}
	partial := newList(t, widget.ListOpts.EntryImage(&widget.ListEntryImage{Selected: images.Selected}), widget.ListOpts.Entries([]any{"a"}),
		widget.ListOpts.EntryLabelFunc(func(e any) string { return e.(string) }))
	colors := *partial.DefinedParams().EntryColor
	colors.SelectedFocusedBackground = focused
	partial.DefinedParams().EntryColor = &colors
	partial.Validate()
	is.True(partial.ComputedParams().EntrySelectedColor().Idle == images.Selected)
	is.True(partial.ComputedParams().EntrySelectedColor().Hover == image.NewNineSliceColor(focused)) // its color
	is.True(partial.ComputedParams().EntryUnselectedColor().Idle == image.NewNineSliceColor(color.Transparent))
}

// A theme's entry images reach a list given neither colors nor images, and
// its combo button entry images such a list combo button's list. Images
// given show in their place, and so do colors given: not the theme's images
// over them.
func TestList_EntryImagesFromTheTheme(t *testing.T) {
	is := is.New(t)

	l := newList(t, widget.ListOpts.Entries([]any{"a"}), widget.ListOpts.EntryLabelFunc(label))
	c := newListComboButton(t, widget.ListComboButtonOpts.Entries([]any{"a"}), widget.ListComboButtonOpts.EntryLabelFunc(label, label))
	colors := *l.DefinedParams().EntryColor // the theme's, none given
	l.DefinedParams().EntryColor, c.DefinedParams().List.EntryColor = nil, nil

	themed, comboThemed := newEntryImages(t), newEntryImages(t)
	theme := &widget.Theme{
		ListTheme:            &widget.ListParams{EntryColor: &colors, EntryImage: themed},
		ListComboButtonTheme: &widget.ListComboButtonParams{List: &widget.ListParams{EntryColor: &colors, EntryImage: comboThemed}},
	}
	l.GetWidget().SetTheme(theme)
	l.Validate()
	isEntryImages(is, l, themed)
	c.GetWidget().SetTheme(theme)
	c.Validate()
	isEntryImages(is, c.List(), comboThemed) // a combo's list takes the combo's

	given := newEntryImages(t)
	l.DefinedParams().EntryImage, c.DefinedParams().List.EntryImage = given, given
	l.Validate()
	c.Validate()
	isEntryImages(is, l, given)
	isEntryImages(is, c.List(), given)

	selected := color.NRGBA{1, 2, 3, 255}
	givenColors := colors
	givenColors.SelectedBackground = selected
	l.DefinedParams().EntryImage, l.DefinedParams().EntryColor = nil, &givenColors
	c.DefinedParams().List.EntryImage, c.DefinedParams().List.EntryColor = nil, &givenColors
	l.Validate()
	c.Validate()
	is.True(l.ComputedParams().EntrySelectedColor().Idle == image.NewNineSliceColor(selected)) // the colors given
	is.True(c.List().ComputedParams().EntrySelectedColor().Idle == image.NewNineSliceColor(selected))
}

// With SlidersWhenNeeded, the vertical slider shows only while the entries
// overflow the list, and keeps its room while hidden.
func TestList_SlidersWhenNeeded(t *testing.T) {
	is := is.New(t)

	l := newList(t, widget.ListOpts.SlidersWhenNeeded(), widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.Entries([]any{"first"}), widget.ListOpts.EntryLabelFunc(func(e any) string { return e.(string) }))
	l.SetLocation(img.Rect(0, 0, 200, 120))
	l.RequestRelayout() // as a parent laying it out does
	screen := ebiten.NewImage(200, 120)
	l.Render(screen) // laid out
	l.Render(screen)
	is.Equal(l.VSlider().GetWidget().GetVisibility(), widget.VisibilityHideBlocking) // one entry fits

	for i := range 20 {
		l.AddEntry("entry " + strconv.Itoa(i))
	}
	l.Render(screen)
	is.Equal(l.VSlider().GetWidget().GetVisibility(), widget.VisibilityShow) // twenty-one do not
}

// Entries wider than the list (upstream #325): a move of the focus scrolls
// to the focused entry, showing its start, and stops once there, so the view
// settles instead of shaking sideways, and the wheel and the sliders scroll
// the list freely afterwards instead of being pulled back to the entry.
func TestList_WideEntriesScroll(t *testing.T) {
	is := is.New(t)

	entries := make([]any, 20)
	for i := range entries {
		entries[i] = "A very long entry, much wider than the list, number " + strconv.Itoa(i)
	}
	l := newList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryLabelFunc(label))
	l.SetLocation(img.Rect(0, 0, 200, 100))
	l.RequestRelayout()
	screen := ebiten.NewImage(200, 100)
	frames := func(n int) {
		for range n {
			l.Render(screen)
			event.ExecuteDeferred()
		}
	}
	frames(3)
	sc := l.ScrollContainer()
	is.True(sc.ContentRect().Dx() > 2*sc.ViewRect().Dx()) // the entries are wider than the view

	l.Focus(true)
	for range 6 {
		l.Update(&widget.UpdateObject{Actions: given(widget.ActionDown)}) // to an entry below the view
	}
	frames(30)
	view, focused := sc.ViewRect(), l.Rows().RowRect(6)
	is.True(view.Max.Y-focused.Max.Y <= 1 && focused.Max.Y <= view.Max.Y) // as little as shows it: its end at the view's
	is.Equal(focused.Min.X, view.Min.X)                                   // its start
	top, left := sc.ScrollTop, sc.ScrollLeft
	frames(5)
	is.Equal([]float64{sc.ScrollTop, sc.ScrollLeft}, []float64{top, left}) // settled

	sc.GetWidget().ScrolledEvent.Fire(&widget.ScrolledEventArgs{Widget: sc.GetWidget(), Y: -3}) // the wheel, down
	l.HSlider().Current = 500                                                                   // the slider, sideways
	event.ExecuteDeferred()
	frames(5)
	is.True(sc.ScrollTop > top)  // still down
	is.Equal(sc.ScrollLeft, 0.5) // still sideways

	for range 6 {
		l.Update(&widget.UpdateObject{Actions: given(widget.ActionUp)}) // to the first entry, above the view
	}
	frames(30)
	view, first := sc.ViewRect(), l.Rows().RowRect(0)
	is.Equal([]int{first.Min.Y, first.Min.X}, []int{view.Min.Y, view.Min.X}) // its start, at the view's
}

// With SelectFocus, which selects the focused entry as the list draws, the
// wheel still scrolls the list away from that entry.
func TestList_SelectFocusLetsTheWheelScroll(t *testing.T) {
	is := is.New(t)

	entries := make([]any, 20)
	for i := range entries {
		entries[i] = strconv.Itoa(i)
	}
	l := newList(t, widget.ListOpts.Entries(entries), widget.ListOpts.EntryLabelFunc(label), widget.ListOpts.SelectFocus())
	l.SetLocation(img.Rect(0, 0, 200, 100))
	l.RequestRelayout()
	screen := ebiten.NewImage(200, 100)
	frames := func(n int) {
		for range n {
			l.Render(screen)
			event.ExecuteDeferred()
		}
	}
	frames(3)
	sc := l.ScrollContainer()
	sc.GetWidget().ScrolledEvent.Fire(&widget.ScrolledEventArgs{Widget: sc.GetWidget(), Y: -3}) // the wheel, down
	event.ExecuteDeferred()
	frames(5)
	is.True(sc.ScrollTop > 0)        // still down
	is.Equal(l.SelectedEntry(), "0") // the focused entry, selected
}

// The rows show what the entries are: duplicates are dropped, entries set
// anew show at once and grow the content, asking the window to lay out
// again, and a scroll binds the rows
// coming into view to their entries.
func TestList_RowsFollowTheEntries(t *testing.T) {
	is := is.New(t)

	l := newList(t, widget.ListOpts.Entries([]any{"a", "b", "a"}), widget.ListOpts.EntryLabelFunc(label))
	is.Equal(l.Entries(), []any{"a", "b"})
	screen := ebiten.NewImage(200, 300)
	frames := func() {
		for range 2 {
			l.Render(screen)
		}
	}

	entries := make([]any, 60)
	for i := range entries {
		entries[i] = "entry " + strconv.Itoa(i)
	}
	h := l.ScrollContainer().ContentRect().Dy()
	for range 3 { // the first frames bind rows, and the rows grew wider: done
		l.Update(&widget.UpdateObject{})
		frames()
	}
	still := &widget.UpdateObject{}
	l.Update(still)
	is.True(!still.RelayoutRequested) // nothing changed
	l.SetEntries(entries)
	upd := &widget.UpdateObject{}
	l.Update(upd)
	is.True(upd.RelayoutRequested) // so a window sized by the list follows
	frames()
	is.Equal(shownLabel(l, 0), "entry 0")
	is.True(l.ScrollContainer().ContentRect().Dy() > h) // grown

	l.SetScrollTop(1)
	frames()
	is.Equal(shownLabel(l, 59), "entry 59")
}
