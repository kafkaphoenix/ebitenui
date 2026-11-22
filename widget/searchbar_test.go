package widget_test

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// newSearchBar returns a search bar at 0,0 to 200,30, its input focused,
// drawn once, and the searches it told of.
func newSearchBar(t *testing.T, opts ...widget.SearchBarOpt) (*widget.SearchBar, *[]string) {
	t.Helper()
	got := new([]string)
	s := widget.NewSearchBar(append([]widget.SearchBarOpt{
		widget.SearchBarOpts.TextInputOpts(widget.TextInputOpts.Face(loadFont(t)),
			widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White, Caret: color.White})),
		widget.SearchBarOpts.ChangedHandler(func(a *widget.SearchBarChangedEventArgs) { *got = append(*got, a.Query) }),
	}, opts...)...)
	s.Validate()
	s.SetLocation(img.Rect(0, 0, 200, 30))
	s.Render(ebiten.NewImage(200, 30))
	s.TextInput().Focus(true)
	return s, got
}

// typeAndWait types chars, one an update, then runs n updates with nothing
// typed, as a player pausing.
func typeAndWait(s *widget.SearchBar, d *uitest.Devices, chars string, n int, actions *widget.Actions) {
	step := func() {
		input.Update()
		s.Update(&widget.UpdateObject{Actions: actions})
		input.AfterUpdate()
		event.ExecuteDeferred()
		s.Render(ebiten.NewImage(200, 30))
	}
	for _, r := range chars {
		d.Chars = []rune{r}
		step()
	}
	for range n {
		step()
	}
}

// A search bar tells of the search once the player pauses typing, 250 ms
// by default, not at every key, so a list filters once a search.
func TestSearchBar_TellsOnAPause(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	s, got := newSearchBar(t)
	pause := ebiten.TPS() / 4
	typeAndWait(s, d, "ore", pause-1, nil)
	is.Equal(len(*got), 0) // still typing, as far as it knows
	typeAndWait(s, d, "", 1, nil)
	is.Equal(*got, []string{"ore"})
	typeAndWait(s, d, "", 2*pause, nil)
	is.Equal(*got, []string{"ore"}) // once
	is.Equal(s.Query(), "ore")

	typeAndWait(s, d, "s", pause-3, nil)
	typeAndWait(s, d, "!", 3, nil) // a key in the pause starts it over
	is.Equal(*got, []string{"ore"})
	typeAndWait(s, d, "", pause, nil)
	is.Equal(*got, []string{"ore", "ores!"})

	quick, gotQuick := newSearchBar(t, widget.SearchBarOpts.Delay(0))
	typeAndWait(quick, d, "ab", 0, nil)
	is.Equal(*gotQuick, []string{"a", "ab"}) // no delay: at every key
}

// The clear button shows while there is text; pressing it, or Back while
// the input has the focus, empties the search and tells of it at once.
// Back with nothing to clear is left to the interface.
func TestSearchBar_Clears(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	s, got := newSearchBar(t, widget.SearchBarOpts.Delay(0))
	is.True(!s.ClearButton().GetWidget().IsVisible())
	typeAndWait(s, d, "ice", 0, nil)
	is.True(s.ClearButton().GetWidget().IsVisible())

	r := s.ClearButton().GetWidget().Rect
	d.Cursor = img.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
	for _, down := range []bool{true, false} {
		d.Mouse[ebiten.MouseButtonLeft] = down
		typeAndWait(s, d, "", 1, nil)
	}
	is.Equal(s.Query(), "")
	is.Equal((*got)[len(*got)-1], "")
	is.True(!s.ClearButton().GetWidget().IsVisible()) // nothing to clear

	typeAndWait(s, d, "ice", 0, nil)
	back := given(widget.ActionBack)
	typeAndWait(s, d, "", 1, back)
	is.Equal(s.Query(), "")
	is.True(!back.Left(widget.ActionBack)) // taken
	back = given(widget.ActionBack)
	typeAndWait(s, d, "", 1, back)
	is.True(back.Left(widget.ActionBack)) // nothing to clear: left
}

// SetQuery sets the search and tells of it at once, as a game restoring
// one does.
func TestSearchBar_SetQuery(t *testing.T) {
	is := is.New(t)

	s, got := newSearchBar(t)
	s.SetQuery("gas")
	event.ExecuteDeferred()
	is.Equal(s.TextInput().GetText(), "gas")
	is.Equal(*got, []string{"gas"})
}

// Its icons, clear button and field are the game's, given as options or
// by a theme; the input inside draws no field of its own.
func TestSearchBar_OwnImages(t *testing.T) {
	is := is.New(t)

	icon := &widget.GraphicImage{Idle: ebiten.NewImage(12, 12)}
	cross := &widget.GraphicImage{Idle: ebiten.NewImage(10, 10)}
	field := eimage.NewNineSliceColor(color.Black)
	s, _ := newSearchBar(t, widget.SearchBarOpts.Icon(icon), widget.SearchBarOpts.ClearIcon(cross), widget.SearchBarOpts.Background(field))
	is.True(s.Icon().ImageForTest() == icon.Idle)
	is.True(s.ClearButton().GraphicImage() == cross)
	is.True(s.ComputedParams().Background == field)

	themed := widget.NewSearchBar(widget.SearchBarOpts.TextInputOpts(widget.TextInputOpts.Face(loadFont(t)),
		widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White})))
	themed.GetWidget().SetTheme(&widget.Theme{SearchBarTheme: &widget.SearchBarParams{Icon: icon, Background: field}})
	themed.Validate()
	is.True(themed.Icon().ImageForTest() == icon.Idle)

	plain, _ := newSearchBar(t)
	is.True(plain.Icon().ImageForTest() != nil) // its own magnifier
}

// An idle search bar's update allocates nothing.
func TestSearchBar_IdleUpdateAllocatesNothing(t *testing.T) {
	uitest.Use(t)
	s, _ := newSearchBar(t)
	s.SetQuery("ice")
	u := &widget.UpdateObject{}
	if n := testing.AllocsPerRun(100, func() {
		input.Update()
		s.Update(u)
		input.AfterUpdate()
	}); n != 0 {
		t.Errorf("an update allocated %v times", n)
	}
}
