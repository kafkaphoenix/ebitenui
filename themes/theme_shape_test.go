package themes_test

import (
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/themes"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// The basic themes, each built with options.
var shapedGetters = map[string]func(...themes.ThemeOpt) *widget.Theme{
	"dark":  themes.GetBasicDarkThemeWith,
	"light": themes.GetBasicLightThemeWith,
}

// corners returns how much of a nine-slice its corners take, across and
// down: twice its radius or its border, the wider, as drawn; 0 for a plain
// fill.
func corners(n *image.NineSlice) [2]int {
	w, h := n.MinSize()
	return [2]int{w, h}
}

// The boxes a theme draws, by what they are: bordered ones, plain fills,
// and small parts.
func boxes(th *widget.Theme) (bordered, filled, small []*image.NineSlice) {
	bordered = []*image.NineSlice{
		th.ButtonTheme.Image.Idle, th.ButtonTheme.Image.Pressed,
		th.TextInputTheme.Image.Idle, th.TextAreaTheme.ScrollContainerImage.Idle,
		th.ListTheme.ScrollContainerImage.Idle, th.ListComboButtonTheme.Button.Image.Idle,
	}
	filled = []*image.NineSlice{
		th.TabbookTheme.TabButton.Image.Idle, th.TabTheme.BackgroundImage, th.PanelTheme.BackgroundImage,
	}
	small = []*image.NineSlice{
		th.SliderTheme.HandleImage.Idle, th.SliderTheme.TrackImage.Idle,
		th.ListTheme.Slider.HandleImage.Idle, th.ProgressBarTheme.TrackImage.Idle,
	}
	return bordered, filled, small
}

// Without options a theme is as it was: square, with its own borders,
// plain fills, list entries drawn by their colors and the inputs' own
// selection.
func TestThemes_Unshaped(t *testing.T) {
	is := is.New(t)

	for name, get := range shapedGetters {
		for _, th := range []*widget.Theme{get(), map[string]*widget.Theme{"dark": themes.GetBasicDarkTheme(), "light": themes.GetBasicLightTheme()}[name]} {
			bordered, filled, _ := boxes(th)
			is.Equal(corners(bordered[0]), [2]int{4, 4})      // the buttons' 2 px border
			is.Equal(corners(bordered[2]), [2]int{2, 2})      // the inputs' 1 px border
			is.Equal(corners(filled[0]), [2]int{0, 0})        // a plain fill
			is.True(th.ListTheme.EntryImage == nil)           // the colors draw the entries
			is.True(th.TextInputTheme.Image.Highlight == nil) // the inputs' own
		}
	}
}

// Rounded rounds every box: bordered ones and plain fills to the radius,
// small parts to 3 px at most; a border wider than the radius keeps its
// width. List entries get rounded backgrounds of their colors, and the
// inputs a rounded selection.
func TestThemes_Rounded(t *testing.T) {
	is := is.New(t)

	for name, get := range shapedGetters {
		th := get(themes.Rounded(6))
		bordered, filled, small := boxes(th)
		for i, n := range bordered {
			is.Equal([]any{name, "bordered", i, corners(n)}, []any{name, "bordered", i, [2]int{12, 12}})
		}
		for i, n := range filled {
			is.Equal([]any{name, "filled", i, corners(n)}, []any{name, "filled", i, [2]int{12, 12}})
		}
		for i, n := range small {
			is.Equal([]any{name, "small", i, corners(n)}, []any{name, "small", i, [2]int{6, 6}})
		}

		for _, entries := range []*widget.ListEntryImage{th.ListTheme.EntryImage, th.ListComboButtonTheme.List.EntryImage} {
			for _, n := range []*image.NineSlice{entries.Selected, entries.SelectedFocused, entries.Focused, entries.Selecting, entries.SelectingFocused, entries.DisabledSelected} {
				is.Equal(corners(n), [2]int{12, 12})
			}
			is.True(entries.Unselected == nil) // unselected entries stay clear
		}
		is.Equal(corners(th.TextInputTheme.Image.Highlight), [2]int{6, 6}) // a small part

		is.Equal(corners(get(themes.Rounded(1)).ButtonTheme.Image.Idle), [2]int{4, 4})  // the 2 px border, wider
		is.Equal(corners(get(themes.Rounded(-3)).ButtonTheme.Image.Idle), [2]int{4, 4}) // square
	}
}

// Borders draws every border the width given, also with Rounded, and none
// with 0; plain fills stay without one.
func TestThemes_Borders(t *testing.T) {
	is := is.New(t)

	for name, get := range shapedGetters {
		bordered, filled, small := boxes(get(themes.Borders(5)))
		for i, n := range append(bordered, small...) {
			is.Equal([]any{name, i, corners(n)}, []any{name, i, [2]int{10, 10}})
		}
		is.Equal(corners(filled[0]), [2]int{0, 0}) // no border

		bordered, _, _ = boxes(get(themes.Borders(0)))
		for i, n := range bordered {
			is.Equal([]any{name, i, corners(n)}, []any{name, i, [2]int{0, 0}}) // a plain fill
		}

		is.Equal(corners(get(themes.Rounded(4), themes.Borders(8)).ButtonTheme.Image.Idle), [2]int{16, 16})
		is.Equal(corners(get(themes.Rounded(9), themes.Borders(2)).ButtonTheme.Image.Idle), [2]int{18, 18})
	}
}

// The system theme with options is the same theme for the same mode and
// options, so setting it every frame re-validates nothing; without options
// it is the unshaped one.
func TestThemes_SystemShaped(t *testing.T) {
	is := is.New(t)

	opts := []themes.ThemeOpt{themes.Rounded(6)}
	first, again := themes.GetBasicSystemThemeWith(opts...), themes.GetBasicSystemThemeWith(opts...)
	is.True(first == again)
	first, again = themes.GetBasicSystemThemeWith(themes.Rounded(6)), themes.GetBasicSystemThemeWith(themes.Rounded(6))
	is.True(first == again) // the same options, given anew
	is.True(themes.GetBasicSystemThemeWith(themes.Rounded(6)) != themes.GetBasicSystemThemeWith(themes.Rounded(7)))
	is.True(themes.GetBasicSystemThemeWith() == themes.GetBasicSystemTheme())

	rounded := themes.NewShape([]themes.ThemeOpt{themes.Rounded(6)})
	dark, light := themes.ShapedThemeFor(ebiten.ColorModeDark, rounded), themes.ShapedThemeFor(ebiten.ColorModeLight, rounded)
	is.True(dark != light)
	is.True(dark.DefaultTextColor != light.DefaultTextColor) // each its mode's
	is.Equal(corners(dark.ButtonTheme.Image.Idle), [2]int{12, 12})
}

// Each entry state's rounded background is a fill of that state's color.
func TestThemes_EntryImageColors(t *testing.T) {
	is := is.New(t)

	shade := func(n uint8) color.Color { return color.NRGBA{n, n, n, 255} }
	c := &widget.ListEntryColor{
		FocusedBackground: shade(1), SelectingBackground: shade(2), SelectedBackground: shade(3),
		DisabledSelectedBackground: shade(4), SelectedFocusedBackground: shade(5), SelectingFocusedBackground: shade(6),
	}
	filled := map[*image.NineSlice]color.Color{}
	fill := func(c color.Color) *image.NineSlice {
		n := &image.NineSlice{}
		filled[n] = c
		return n
	}
	e := themes.EntryImages(c, fill)
	is.Equal([]color.Color{filled[e.Focused], filled[e.Selecting], filled[e.Selected], filled[e.DisabledSelected], filled[e.SelectedFocused], filled[e.SelectingFocused]},
		[]color.Color{shade(1), shade(2), shade(3), shade(4), shade(5), shade(6)})
}
