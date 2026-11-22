package main

import (
	"cmp"
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Two tables in the theme's look. At the left, bodies of the solar system:
// a search bar filters them by name, a click on a header sorts by its
// column and again the other way, Shift and a click adds a column to the
// sort, and dragging a header's edge resizes its column; the masses are in
// scientific notation, as the game formats them. At the right, a
// leaderboard: a rank column, sorted by score, the player's row
// highlighted.
func main() {
	theme := themes.GetBasicDarkTheme()

	type body struct {
		name   string
		kind   string
		mass   float64 // kg
		radius float64 // km
	}
	bodies := []any{
		&body{"Mercury", "planet", 3.30e23, 2440}, &body{"Venus", "planet", 4.87e24, 6052},
		&body{"Earth", "planet", 5.97e24, 6371}, &body{"Moon", "moon", 7.35e22, 1737},
		&body{"Mars", "planet", 6.42e23, 3390}, &body{"Phobos", "moon", 1.07e16, 11},
		&body{"Ceres", "dwarf planet", 9.38e20, 470}, &body{"Jupiter", "planet", 1.90e27, 69911},
		&body{"Io", "moon", 8.93e22, 1822}, &body{"Europa", "moon", 4.80e22, 1561},
		&body{"Ganymede", "moon", 1.48e23, 2634}, &body{"Callisto", "moon", 1.08e23, 2410},
		&body{"Saturn", "planet", 5.68e26, 58232}, &body{"Titan", "moon", 1.35e23, 2575},
		&body{"Uranus", "planet", 8.68e25, 25362}, &body{"Neptune", "planet", 1.02e26, 24622},
		&body{"Triton", "moon", 2.14e22, 1353}, &body{"Pluto", "dwarf planet", 1.30e22, 1188},
	}
	b := func(r any) *body { return r.(*body) }
	bodyTable := widget.NewTable(
		widget.TableOpts.Columns(
			widget.TableColumn{Header: "Name", Text: func(r any) string { return b(r).name },
				Compare: func(x, y any) int { return cmp.Compare(b(x).name, b(y).name) }, Width: 110},
			widget.TableColumn{Header: "Kind", Text: func(r any) string { return b(r).kind },
				Compare: func(x, y any) int { return cmp.Compare(b(x).kind, b(y).kind) }, Width: 110},
			widget.TableColumn{Header: "Mass (kg)", Text: func(r any) string { return strconv.FormatFloat(b(r).mass, 'e', 2, 64) },
				Compare: func(x, y any) int { return cmp.Compare(b(x).mass, b(y).mass) }, Width: 120, Align: widget.TextPositionEnd},
			widget.TableColumn{Header: "Radius (km)", Text: func(r any) string { return strconv.FormatFloat(b(r).radius, 'f', 0, 64) },
				Compare: func(x, y any) int { return cmp.Compare(b(x).radius, b(y).radius) }, Width: 130, Align: widget.TextPositionEnd},
		),
		widget.TableOpts.Rows(bodies),
		widget.TableOpts.SortBy(widget.TableSort{Column: 2, Descending: true}),
		widget.TableOpts.WidgetOpts(widget.WidgetOpts.MinSize(490, 300)),
	)
	search := widget.NewSearchBar(
		widget.SearchBarOpts.TextInputOpts(widget.TextInputOpts.Placeholder("Search bodies")),
		widget.SearchBarOpts.ChangedHandler(func(a *widget.SearchBarChangedEventArgs) {
			q := strings.ToLower(a.Query)
			bodyTable.SetFilter(func(r any) bool { return strings.Contains(strings.ToLower(b(r).name), q) })
		}),
		widget.SearchBarOpts.WidgetOpts(widget.WidgetOpts.MinSize(490, 32)),
	)

	type score struct {
		name   string
		points int
	}
	scores := []any{&score{"Ada", 9120}, &score{"You", 7450}, &score{"Grace", 8800}, &score{"Linus", 6100}, &score{"Hedy", 7990}}
	s := func(r any) *score { return r.(*score) }
	board := widget.NewTable(
		widget.TableOpts.RankColumn("#"),
		widget.TableOpts.Columns(
			widget.TableColumn{Header: "Player", Text: func(r any) string { return s(r).name }, Width: 110},
			widget.TableColumn{Header: "Score", Text: func(r any) string { return strconv.Itoa(s(r).points) },
				Compare: func(x, y any) int { return cmp.Compare(s(x).points, s(y).points) }, Width: 90, Align: widget.TextPositionEnd},
		),
		widget.TableOpts.Rows(scores),
		widget.TableOpts.SortBy(widget.TableSort{Column: 1, Descending: true}),
		widget.TableOpts.Highlight(func(r any) bool { return s(r).name == "You" }),
		widget.TableOpts.WidgetOpts(widget.WidgetOpts.MinSize(260, 300)),
	)

	left := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(8))))
	left.AddChild(search, bodyTable)
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)), widget.RowLayoutOpts.Spacing(20))))
	root.AddChild(left, board)

	ui := &ebitenui.UI{Container: root, PrimaryTheme: theme}
	ebiten.SetWindowSize(850, 400)
	ebiten.SetWindowTitle("Ebiten UI - Table")
	if err := ebiten.RunGame(&game{ui: ui}); err != nil {
		log.Println(err)
	}
}

type game struct{ ui *ebitenui.UI }

func (g *game) Update() error {
	g.ui.Update()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) { g.ui.Draw(screen) }

func (g *game) Layout(w, h int) (int, int) { return w, h }
