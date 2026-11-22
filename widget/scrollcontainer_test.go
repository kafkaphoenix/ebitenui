package widget_test

import (
	img "image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/matryer/is"
)

// ScrollVisible brings a row to the top, the middle or the bottom of the
// view, and stops at the content's end.
func TestScrollContainer_ScrollVisible(t *testing.T) {
	is := is.New(t)

	content := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	rows := make([]*simpleWidget, 0, 10)
	for range 10 {
		r := newSimpleWidget(50, 20, nil)
		rows = append(rows, r)
		content.AddChild(r)
	}
	ns := newNineSliceEmpty(t)
	s := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(content), widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{Idle: ns, Mask: ns}))
	s.SetLocation(img.Rect(0, 0, 50, 50)) // a view of 2.5 rows over 10
	render(s, t)

	for _, c := range []struct {
		row    int
		p      float64
		wantAt int // the row's top, from the view's
	}{
		{5, 0, 0},
		{5, 0.5, 15},
		{5, 1, 30},
		{9, 0, 30}, // the content's end stops it
	} {
		s.ScrollVisible(rows[c.row], c.p)
		is.True(s.ScrollTop >= 0 && s.ScrollTop <= 1) // a fraction of the scroll
		render(s, t)
		is.Equal(rows[c.row].GetWidget().Rect.Min.Y, c.wantAt) // where the row shows
	}

	s.ScrollVisible(rows[0], 0)
	render(s, t)
	outside := newSimpleWidget(50, 20, nil)
	outside.SetLocation(img.Rect(0, 120, 50, 140))
	s.ScrollVisible(outside, 0) // not in the content
	is.Equal(s.ScrollTop, 0.0)
}

// Without images of its own, a scroll container takes its theme's list
// scroll images and follows the theme; given images, it keeps them; with
// neither, it can't be laid out.
func TestScrollContainer_ThemeImages(t *testing.T) {
	is := is.New(t)

	ns := newNineSliceEmpty(t)
	dark := &widget.Theme{ListTheme: &widget.ListParams{ScrollContainerImage: &widget.ScrollContainerImage{Idle: ns, Mask: ns}}}
	light := &widget.Theme{ListTheme: &widget.ListParams{ScrollContainerImage: &widget.ScrollContainerImage{Idle: ns, Mask: ns}}}
	images := func(s *widget.ScrollContainer, theme *widget.Theme) *widget.ScrollContainerImage {
		s.GetWidget().SetTheme(theme)
		s.Validate()
		return s.Image()
	}

	s := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(widget.NewContainer()))
	is.True(images(s, dark) == dark.ListTheme.ScrollContainerImage)
	is.True(images(s, light) == light.ListTheme.ScrollContainerImage)

	own := &widget.ScrollContainerImage{Idle: ns, Mask: ns}
	s = widget.NewScrollContainer(widget.ScrollContainerOpts.Content(widget.NewContainer()), widget.ScrollContainerOpts.Image(own))
	is.True(images(s, dark) == own)

	defer func() { is.True(recover() != nil) }() // neither
	widget.NewScrollContainer(widget.ScrollContainerOpts.Content(widget.NewContainer())).Validate()
}

// Validated again, as on a theme change, a scroll container still forwards
// each event of its content once.
func TestScrollContainer_ForwardsEventsOnce(t *testing.T) {
	is := is.New(t)

	ns := newNineSliceEmpty(t)
	content := widget.NewContainer()
	s := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(content), widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{Idle: ns, Mask: ns}))
	forwarded := 0
	s.GetWidget().ToolTipEvent.AddHandler(func(any) { forwarded++ })
	for range 3 {
		s.Validate()
	}
	content.GetWidget().FireToolTipEvent(nil, true)
	event.ExecuteDeferred()
	is.Equal(forwarded, 1)
}
