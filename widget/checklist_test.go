package widget_test

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// objectives are a mission's: two to do, one optional.
func objectives() []widget.ChecklistItem {
	return []widget.ChecklistItem{
		{Label: "Reach orbit"},
		{Label: "Survey the moons", Detail: "0/4"},
		{Label: "Photograph the rings", Optional: true},
	}
}

// newChecklist returns a checklist at 0,0 as wide as w, validated and
// drawn once.
func newChecklist(t *testing.T, w int, opts ...widget.ChecklistOpt) *widget.Checklist {
	t.Helper()
	c := widget.NewChecklist(append([]widget.ChecklistOpt{widget.ChecklistOpts.Face(loadFont(t))}, opts...)...)
	c.Validate()
	_, h := c.PreferredSize()
	c.SetLocation(img.Rect(0, 0, w, h))
	c.Render(ebiten.NewImage(w, max(h, 1)))
	return c
}

// updateChecklist runs n updates of c, the input's too, and draws it.
func updateChecklist(c *widget.Checklist, n int, u *widget.UpdateObject) {
	for range n {
		input.Update()
		c.Update(u)
		input.AfterUpdate()
		event.ExecuteDeferred()
	}
	r := c.GetWidget().Rect
	c.Render(ebiten.NewImage(max(r.Dx(), 1), max(r.Dy(), 1)))
}

// The game sets each item's state as it happens; the checklist tells of
// each change, from which state, and once as every item that isn't
// optional is done. Progress counts those; a failed one keeps the list
// from being complete.
func TestChecklist_States(t *testing.T) {
	is := is.New(t)

	var changed []widget.ChecklistItemChangedEventArgs
	completed := 0
	c := newChecklist(t, 300, widget.ChecklistOpts.Items(objectives()...),
		widget.ChecklistOpts.ItemChangedHandler(func(a *widget.ChecklistItemChangedEventArgs) { changed = append(changed, *a) }),
		widget.ChecklistOpts.CompletedHandler(func(*widget.ChecklistCompletedEventArgs) { completed++ }))
	done, total := c.Progress()
	is.Equal([]int{done, total}, []int{0, 2}) // the optional one left out
	is.True(!c.Complete())

	c.SetState(0, widget.ChecklistDone)
	c.SetState(0, widget.ChecklistDone) // no change
	c.SetDetail(1, "4/4")
	c.SetState(1, widget.ChecklistFailed)
	event.ExecuteDeferred()
	is.Equal(len(changed), 2)
	is.Equal([]any{changed[0].Index, changed[0].Previous, changed[0].Item.State}, []any{0, widget.ChecklistPending, widget.ChecklistDone})
	is.Equal([]any{changed[1].Index, changed[1].Previous, changed[1].Item.Detail}, []any{1, widget.ChecklistPending, "4/4"})
	is.True(!c.Complete()) // failed
	is.Equal(completed, 0)

	c.SetState(1, widget.ChecklistDone)
	c.SetState(2, widget.ChecklistDone) // optional: complete already
	event.ExecuteDeferred()
	is.True(c.Complete())
	is.Equal(completed, 1)
	done, total = c.Progress()
	is.Equal([]int{done, total}, []int{2, 2})

	c.SetState(0, widget.ChecklistPending)
	c.SetState(0, widget.ChecklistDone)
	event.ExecuteDeferred()
	is.Equal(completed, 2) // complete again
	is.Equal(c.Item(1).Detail, "4/4")
	is.Equal(len(c.Items()), 3)
}

// Each state has its mark, drawn when the game gives none, and its text
// color: done and failed items fade.
func TestChecklist_MarksAndColors(t *testing.T) {
	is := is.New(t)

	c := newChecklist(t, 300, widget.ChecklistOpts.Items(
		widget.ChecklistItem{Label: "a"}, widget.ChecklistItem{Label: "b", State: widget.ChecklistDone},
		widget.ChecklistItem{Label: "c", State: widget.ChecklistFailed}))
	m := c.DrawnMarks()
	is.True(m.Pending != nil && m.Done != nil && m.Failed != nil)
	is.True(m.Pending != m.Done && m.Done != m.Failed)
	is.Equal([]*ebiten.Image{c.MarkOf(0), c.MarkOf(1), c.MarkOf(2)}, []*ebiten.Image{m.Pending, m.Done, m.Failed})

	col := c.ComputedParams().Colors
	is.Equal(col.Pending, color.Color(color.White)) // no theme: white
	is.Equal(col.Done, color.Color(color.NRGBA{255, 255, 255, 128}))
	is.Equal(c.Label(0).ComputedParams().Color, col.Pending)
	is.Equal(c.Label(1).ComputedParams().Color, col.Done)
	is.Equal(c.Label(2).ComputedParams().Color, col.Failed)
	c.SetState(1, widget.ChecklistPending)
	is.Equal(c.Label(1).ComputedParams().Color, col.Pending)

	tick := ebiten.NewImage(12, 12)
	own := newChecklist(t, 300, widget.ChecklistOpts.Items(widget.ChecklistItem{Label: "a", State: widget.ChecklistDone}),
		widget.ChecklistOpts.Marks(&widget.ChecklistMarks{Done: tick}))
	is.Equal(own.MarkOf(0), tick)
	is.True(own.DrawnMarks().Pending != nil) // the others drawn
}

// The rows stack down under the title's line, each its mark, its label
// after it and its detail at its end; the checklist is as wide as its
// widest row. The progress shows at the title line's end, worded by the
// game, and follows the items.
func TestChecklist_Layout(t *testing.T) {
	is := is.New(t)

	c := newChecklist(t, 300, widget.ChecklistOpts.Items(objectives()...), widget.ChecklistOpts.Title("Jupiter"),
		widget.ChecklistOpts.Progress(func(done, total int) string { return itoa(done) + " of " + itoa(total) }))
	p := c.ComputedParams()
	rp := p.RowPadding
	_, th := c.TitleText().PreferredSize()
	first := c.RowRect(0)
	is.Equal(first.Min.Y, rp.Top+th+rp.Bottom)
	is.Equal(c.RowRect(1).Min.Y, first.Max.Y)
	is.Equal(first.Dx(), 300)
	is.Equal(c.ProgressText().Label, "0 of 2")
	is.Equal(c.TitleText().Label, "Jupiter")

	d := c.Detail(1).GetWidget().Rect
	is.Equal(d.Max.X, 300-rp.Right) // at the row's end
	l := c.Label(1).GetWidget().Rect
	is.True(l.Min.X > rp.Left)
	is.True(l.Max.X < d.Min.X)

	w, h := c.PreferredSize()
	is.Equal(h, c.RowRect(2).Max.Y)
	lw, _ := c.Label(2).PreferredSize()
	is.True(w >= rp.Left+lw+rp.Right)

	c.SetState(0, widget.ChecklistDone)
	is.Equal(c.ProgressText().Label, "1 of 2")

	c.AddItem(widget.ChecklistItem{Label: "Return"})
	_, h2 := c.PreferredSize()
	is.True(h2 > h)
	c.RemoveItem(3)
	c.RemoveItem(9) // none
	_, h3 := c.PreferredSize()
	is.Equal(h3, h)
	c.SetTitle("")
	is.True(c.TitleText() != nil) // the progress still shows
	c.SetItems(nil)
	is.Equal(c.Len(), 0)
}

func itoa(n int) string { return string(rune('0' + n)) }

// With MaxWidth the labels wrap, their rows growing taller; the checklist
// is no wider. A done item's label is struck through, each line of it,
// with StrikeDone.
func TestChecklist_WrapAndStrike(t *testing.T) {
	is := is.New(t)

	long := "Survey every moon of the planet from a low orbit"
	c := newChecklist(t, 200, widget.ChecklistOpts.MaxWidth(200), widget.ChecklistOpts.StrikeDone(),
		widget.ChecklistOpts.Items(widget.ChecklistItem{Label: long}, widget.ChecklistItem{Label: "Land"}))
	w, _ := c.PreferredSize()
	is.True(w <= 200)
	is.True(c.RowRect(0).Dy() > c.RowRect(1).Dy()) // wrapped
	is.Equal(len(c.Strikes(0)), 0)                 // not done

	c.SetState(0, widget.ChecklistDone)
	updateChecklist(c, 1, &widget.UpdateObject{})
	s := c.Strikes(0)
	is.True(len(s) > 1) // a line through each line
	l := c.Label(0).GetWidget().Rect
	for _, r := range s {
		is.True(r.In(l))
		is.True(r.Dy() >= 1)
	}
	is.True(s[1].Min.Y > s[0].Min.Y)

	plain := newChecklist(t, 200, widget.ChecklistOpts.Items(widget.ChecklistItem{Label: "Land", State: widget.ChecklistDone}))
	is.Equal(len(plain.Strikes(0)), 0) // without StrikeDone

	spaced := newChecklist(t, 200, widget.ChecklistOpts.StrikeDone(),
		widget.ChecklistOpts.Items(widget.ChecklistItem{Label: "Land  ", State: widget.ChecklistDone}))
	land := widget.FontAdvanceAt("Land", 4, loadFont(t))
	is.True(spaced.Strikes(0)[0].Dx() >= land-1 && spaced.Strikes(0)[0].Dx() <= land+1) // up to the last letter

	word := newChecklist(t, 60, widget.ChecklistOpts.MaxWidth(60),
		widget.ChecklistOpts.Items(widget.ChecklistItem{Label: "Circumnavigation"}))
	w, _ = word.PreferredSize()
	is.Equal(w, 60) // a word too long to wrap: no wider
}

// center returns r's middle.
func center(r img.Rectangle) img.Point { return img.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2) }

// clickChecklist presses the left button at p for an update and lets go
// at q for another.
func clickChecklist(c *widget.Checklist, d *uitest.Devices, p, q img.Point) {
	d.Cursor = p
	d.Mouse[ebiten.MouseButtonLeft] = true
	updateChecklist(c, 1, &widget.UpdateObject{})
	d.Cursor = q
	updateChecklist(c, 1, &widget.UpdateObject{})
	d.Mouse[ebiten.MouseButtonLeft] = false
	updateChecklist(c, 1, &widget.UpdateObject{})
}

// The player ticks and unticks a tickable item with a click, hovering it
// shows; an item that isn't tickable, a press let go on another row, or a
// disabled checklist ticks nothing.
func TestChecklist_Ticks(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	c := newChecklist(t, 300, widget.ChecklistOpts.Items(
		widget.ChecklistItem{Label: "Pack", Tickable: true}, widget.ChecklistItem{Label: "Launch"},
		widget.ChecklistItem{Label: "Wave", Tickable: true}))
	at := func(i int) img.Point { return center(c.RowRect(i)) }

	d.Cursor = at(0)
	updateChecklist(c, 2, &widget.UpdateObject{})
	is.Equal(c.Background(0), c.ComputedParams().HoverImage)
	d.Cursor = at(1)
	updateChecklist(c, 1, &widget.UpdateObject{})
	is.True(c.Background(1) == nil) // not tickable: no hover
	is.True(c.Background(0) == nil)

	clickChecklist(c, d, at(0), at(0))
	is.Equal(c.Item(0).State, widget.ChecklistDone)
	clickChecklist(c, d, at(0), at(0))
	is.Equal(c.Item(0).State, widget.ChecklistPending) // unticked
	clickChecklist(c, d, at(1), at(1))
	is.Equal(c.Item(1).State, widget.ChecklistPending) // not tickable
	clickChecklist(c, d, at(0), at(2))
	is.Equal([]widget.ChecklistState{c.Item(0).State, c.Item(2).State}, []widget.ChecklistState{widget.ChecklistPending, widget.ChecklistPending})

	c.GetWidget().Disabled = true
	clickChecklist(c, d, at(2), at(2))
	is.Equal(c.Item(2).State, widget.ChecklistPending)
}

// Focused, the checklist moves through its tickable items with Up and
// Down, leaving them at its ends, and Activate, Space or Enter ticks the
// one it is on. It takes no focus while no item is tickable.
func TestChecklist_Keys(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	c := newChecklist(t, 300, widget.ChecklistOpts.Items(
		widget.ChecklistItem{Label: "Read"}, widget.ChecklistItem{Label: "Pack", Tickable: true},
		widget.ChecklistItem{Label: "Launch"}, widget.ChecklistItem{Label: "Wave", Tickable: true}))
	is.Equal(c.TabOrder(), 0)
	c.Focus(true)
	is.True(c.IsFocused())
	is.Equal(c.FocusedRow(), 1) // the first tickable
	is.Equal(c.Background(1), c.ComputedParams().HoverImage)

	for i, s := range []struct {
		act   widget.Action
		taken bool
		at    int
	}{
		{widget.ActionUp, false, 1}, // the first: left
		{widget.ActionDown, true, 3},
		{widget.ActionDown, false, 3}, // the last: left
	} {
		a := given(s.act)
		c.Update(&widget.UpdateObject{Actions: a})
		is.Equal([]any{i, !a.Left(s.act), c.FocusedRow()}, []any{i, s.taken, s.at})
	}
	a := given(widget.ActionActivate)
	c.Update(&widget.UpdateObject{Actions: a})
	is.True(!a.Left(widget.ActionActivate))
	is.Equal(c.Item(3).State, widget.ChecklistDone)

	d.Keys[ebiten.KeySpace] = true
	updateChecklist(c, 1, &widget.UpdateObject{})
	d.Keys[ebiten.KeySpace] = false
	updateChecklist(c, 1, &widget.UpdateObject{})
	is.Equal(c.Item(3).State, widget.ChecklistPending)

	c.RemoveItem(0)
	is.Equal(c.FocusedRow(), 2) // still on "Wave"
	is.Equal(c.Item(c.FocusedRow()).Label, "Wave")

	c.Focus(false)
	a = given(widget.ActionUp)
	c.Update(&widget.UpdateObject{Actions: a})
	is.True(a.Left(widget.ActionUp)) // unfocused

	none := newChecklist(t, 300, widget.ChecklistOpts.Items(objectives()...), widget.ChecklistOpts.TabOrder(3))
	is.Equal(none.TabOrder(), -1)
	none.AddItem(widget.ChecklistItem{Label: "Note", Tickable: true})
	is.Equal(none.TabOrder(), 3)
}

// In a sequential list the first pending item that isn't optional is the
// current one, on its own background, and those after it fade, or with
// HideUpcoming don't show, until it is reached.
func TestChecklist_Sequential(t *testing.T) {
	is := is.New(t)

	steps := []widget.ChecklistItem{{Label: "Open the map"}, {Label: "Zoom", Optional: true}, {Label: "Pick a planet"}, {Label: "Launch the rocket to the moon"}}
	c := newChecklist(t, 300, widget.ChecklistOpts.Sequential(), widget.ChecklistOpts.Items(steps...))
	col := c.ComputedParams().Colors
	is.Equal(c.Current(), 0)
	is.Equal(c.Background(0), c.ComputedParams().CurrentImage)
	is.Equal(c.Label(2).ComputedParams().Color, col.Upcoming)
	c.SetState(0, widget.ChecklistDone)
	is.Equal(c.Current(), 2) // the optional one skipped
	is.Equal(c.Label(2).ComputedParams().Color, col.Pending)
	is.Equal(c.Label(3).ComputedParams().Color, col.Upcoming)

	plain := newChecklist(t, 300, widget.ChecklistOpts.Items(steps...))
	is.True(plain.Background(0) == nil)
	is.Equal(plain.Label(3).ComputedParams().Color, col.Pending)

	hide := newChecklist(t, 300, widget.ChecklistOpts.HideUpcoming(), widget.ChecklistOpts.Items(steps...))
	w, h := hide.PreferredSize()
	lw, _ := hide.Label(3).PreferredSize()
	is.True(w > lw) // as wide as the hidden steps will be
	is.Equal([]bool{hide.RowShown(0), hide.RowShown(1), hide.RowShown(3)}, []bool{true, false, false})
	hide.SetState(0, widget.ChecklistDone)
	w2, h2 := hide.PreferredSize()
	is.True(h2 > h)
	is.Equal(w2, w) // no wider as they show
	is.Equal([]bool{hide.RowShown(1), hide.RowShown(2), hide.RowShown(3)}, []bool{true, true, false})
	hide.SetState(2, widget.ChecklistDone)
	hide.SetState(3, widget.ChecklistDone)
	is.Equal(hide.Current(), -1)
	is.True(hide.RowShown(3)) // all shown once done
}

// Its look comes from the options, then the theme's checklist, then the
// theme's face and text color: the light theme darkens the rows'
// backgrounds and the marks' colors.
func TestChecklist_Theme(t *testing.T) {
	is := is.New(t)

	light := themes.GetBasicLightTheme()
	c := widget.NewChecklist(widget.ChecklistOpts.Items(objectives()...))
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	root.AddChild(c)
	root.GetWidget().SetTheme(light)
	c.Validate()
	p := c.ComputedParams()
	is.Equal(p.HoverImage, light.ChecklistTheme.HoverImage)
	is.Equal(p.Colors.DoneMark, light.ChecklistTheme.Colors.DoneMark)
	is.Equal(p.Colors.Pending, light.DefaultTextColor)
	is.Equal(p.Face, light.DefaultFace)

	red := color.NRGBA{255, 0, 0, 255}
	own := eimage.NewNineSliceColor(red)
	mine := widget.NewChecklist(widget.ChecklistOpts.Items(objectives()...), widget.ChecklistOpts.HoverImage(own),
		widget.ChecklistOpts.Colors(&widget.ChecklistColors{Done: red}))
	root.AddChild(mine)
	mine.Validate()
	is.Equal(mine.ComputedParams().HoverImage, own)
	is.Equal(mine.ComputedParams().Colors.Done, color.Color(red))
	is.Equal(mine.ComputedParams().Colors.DoneMark, light.ChecklistTheme.Colors.DoneMark) // the rest themed
}

// It tells UI.Changed of each change it shows, a hover's too, and of
// nothing while idle; an idle update allocates nothing.
func TestChecklist_Changes(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	c := newChecklist(t, 300, widget.ChecklistOpts.Items(widget.ChecklistItem{Label: "Pack", Tickable: true}))
	d.Cursor = img.Pt(400, 400)
	updateChecklist(c, 2, &widget.UpdateObject{})
	from := changes.Count()
	updateChecklist(c, 5, &widget.UpdateObject{})
	is.Equal(changes.Count(), from) // idle

	c.SetState(0, widget.ChecklistDone)
	is.True(changes.Count() > from)
	from = changes.Count()
	d.Cursor = center(c.RowRect(0))
	updateChecklist(c, 2, &widget.UpdateObject{})
	is.True(changes.Count() > from) // hovered

	from = changes.Count()
	c.SetState(0, c.Item(0).State)
	is.Equal(changes.Count(), from) // no change

	u := &widget.UpdateObject{}
	if n := testing.AllocsPerRun(100, func() {
		input.Update()
		c.Update(u)
		input.AfterUpdate()
	}); n != 0 {
		t.Errorf("an update allocated %v times", n)
	}

	d.Cursor = img.Pt(400, 400)
	updateChecklist(c, 2, &widget.UpdateObject{})
	c.GetWidget().Disabled = true
	from = changes.Count()
	d.Cursor = center(c.RowRect(0))
	updateChecklist(c, 2, &widget.UpdateObject{})
	is.Equal(changes.Count(), from) // disabled: no hover
	is.True(c.Background(0) == nil)
}
