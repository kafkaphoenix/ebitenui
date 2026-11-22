package widget_test

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// boundRow is a row of a list of recycled rows: the entry it shows, its
// updates and draws.
type boundRow struct {
	simpleWidget
	entry            any
	updates, renders int
}

// PreferredSize is 60 by 20 pixels once bound, 5 high before: a row as
// tall as what it shows.
func (b *boundRow) PreferredSize() (int, int) {
	if b.entry == nil {
		return 60, 5
	}
	return 60, 20
}

func (b *boundRow) Update(*widget.UpdateObject) { b.updates++ }
func (b *boundRow) Render(*ebiten.Image)        { b.renders++ }

// rowsList returns a list of n entries, 0 to n-1, shown in recycled rows of
// 60 by 20 pixels, on a 200 by 300 screen, and the rows it made.
func rowsList(t *testing.T, n int, opts ...widget.ListOpt) (*widget.List, *[]*boundRow, *ebiten.Image) {
	t.Helper()
	entries := make([]any, n)
	for i := range entries {
		entries[i] = i
	}
	rows := new([]*boundRow)
	l := widget.NewList(append(opts, widget.ListOpts.Entries(entries),
		widget.ListOpts.EntryRows(
			func() widget.PreferredSizeLocateableWidget {
				r := &boundRow{simpleWidget: *newSimpleWidget(60, 20, nil)}
				*rows = append(*rows, r)
				return r
			},
			func(row widget.PreferredSizeLocateableWidget, e any) { row.(*boundRow).entry = e },
		),
		widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Disabled: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
		widget.ListOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
		clearEntryColor(),
	)...)
	l.SetLocation(img.Rect(0, 0, 200, 300))
	l.Validate()
	event.ExecuteDeferred()
	screen := ebiten.NewImage(200, 300)
	for range 2 {
		l.Render(screen)
	}
	return l, rows, screen
}

// clearEntryColor gives a list's entries clear colors.
func clearEntryColor() widget.ListOpt {
	return widget.ListOpts.EntryColor(&widget.ListEntryColor{
		Unselected: color.Transparent, Selected: color.Transparent, DisabledUnselected: color.Transparent,
		DisabledSelected: color.Transparent, SelectedBackground: color.Transparent, DisabledSelectedBackground: color.Transparent,
	})
}

// shownIn returns the entries the rows in view show, by row top.
func shownIn(rows []*boundRow) map[int]any {
	shown := map[int]any{}
	for _, r := range rows {
		if r.entry != nil {
			shown[r.GetWidget().Rect.Min.Y] = r.entry
		}
	}
	return shown
}

// EntryRows shows entries in rows of one height the list makes and binds
// to the entries they come to show, as it does labels: a list of 10 000
// keeps a screenful of rows, each where its entry's row is, with neither a
// label function nor a face. For a list's recycled custom rows, as
// VirtualList's RowFunc and BindFunc were.
func TestList_EntryRows(t *testing.T) {
	is := is.New(t)

	l, rows, screen := rowsList(t, 10_000)
	is.True(len(*rows) <= 300/20+2) // a screenful
	top, bottom := l.Rows().RowSpan(1)
	is.Equal([]int{top, bottom}, []int{20, 40}) // the first row's height
	shown := shownIn(*rows)
	is.Equal(shown[0], 0)
	is.Equal(shown[20], 1)

	l.SetScrollTop(float64(100*20) / float64(10_000*20-l.ScrollContainer().ViewRect().Dy()))
	for range 2 {
		l.Render(screen)
	}
	made := len(*rows)
	shown = shownIn(*rows)
	is.Equal(shown[0], 100) // bound to the entries now in view
	is.Equal(shown[20], 101)
	is.Equal(len(*rows), made) // no more rows

	is.True((*rows)[0].renders > 0)
	l.Update(&widget.UpdateObject{})
	is.True((*rows)[0].updates > 0) // the rows in view take updates

	clickListEntry(t, l, 101)
	is.Equal(l.SelectedEntry(), 101) // and the list selects as ever
}

// EntryRowHeight sets the rows' height, in place of the first row's.
func TestList_EntryRowHeight(t *testing.T) {
	is := is.New(t)

	l, _, _ := rowsList(t, 100, widget.ListOpts.EntryRowHeight(32))
	top, bottom := l.Rows().RowSpan(2)
	is.Equal([]int{top, bottom}, []int{64, 96})
}

// A row has EntryTextPadding around it, in its row's height.
func TestList_EntryRowsPadding(t *testing.T) {
	is := is.New(t)

	l, rows, _ := rowsList(t, 100, widget.ListOpts.EntryTextPadding(&widget.Insets{Left: 4, Top: 3, Bottom: 7}))
	top, bottom := l.Rows().RowSpan(1)
	is.Equal([]int{top, bottom}, []int{30, 60})
	for _, r := range *rows {
		if r.entry == 1 {
			rr := l.Rows().RowRect(1)
			is.Equal(r.GetWidget().Rect, img.Rect(rr.Min.X+4, rr.Min.Y+3, rr.Max.X, rr.Max.Y-7))
		}
	}
}

// UpdateEntry binds the rows again, so a row shows its entry's change.
func TestList_EntryRowsUpdateEntry(t *testing.T) {
	is := is.New(t)

	binds := 0
	entries := []any{"a", "b"}
	l := widget.NewList(widget.ListOpts.Entries(entries),
		widget.ListOpts.EntryRows(
			func() widget.PreferredSizeLocateableWidget { return newSimpleWidget(60, 20, nil) },
			func(widget.PreferredSizeLocateableWidget, any) { binds++ },
		),
		widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Disabled: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
		widget.ListOpts.SliderParams(&widget.SliderParams{TrackImage: &widget.SliderTrackImage{}, HandleImage: &widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}}),
		clearEntryColor(),
	)
	l.SetLocation(img.Rect(0, 0, 200, 300))
	l.Validate()
	screen := ebiten.NewImage(200, 300)
	for range 2 {
		l.Render(screen)
	}
	before := binds
	l.Render(screen)
	is.Equal(binds, before) // bound once
	l.UpdateEntry("a")
	l.Render(screen)
	is.True(binds > before) // bound again
}
