package widget_test

import (
	"bufio"
	"image"
	"image/color"
	"slices"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/matryer/is"
)

func TestText_Lines(t *testing.T) {
	is := is.New(t)

	for _, label := range []string{"", "one", "one\ntwo", "one\r\ntwo", "one\n", "one\n\ntwo", "\n"} {
		tw := newText(t, label)

		var want []string
		s := bufio.NewScanner(strings.NewReader(label))
		for s.Scan() {
			want = append(want, s.Text())
		}
		is.Equal(tw.Measurements().ProcessedLineTexts(), want) // lines of label
	}
}

func TestText_WrappedLinesKeepEveryCharacter(t *testing.T) {
	is := is.New(t)

	label := "a wrapped label with enough words in it to run over several lines"
	tw := newText(t, label, widget.TextOpts.MaxWidth(120))

	is.True(len(tw.Measurements().ProcessedLineTexts()) > 1)
	is.Equal(strings.Join(tw.Measurements().ProcessedLineTexts(), ""), label)
}

func TestText_LinkHover(t *testing.T) {
	is := is.New(t)

	var entered []string
	tw := newText(t, "go [link=home]there[/link] now",
		widget.TextOpts.ProcessBBCode(true),
		widget.TextOpts.LinkCursorEnteredHandler(func(args *widget.LinkEventArgs) {
			entered = append(entered, args.ID)
		}))
	tw.SetLocation(image.Rect(0, 0, 400, 100))

	// Move the link right; it is hovered where it is now drawn.
	tw.Label = "go further [link=home]there[/link] now"
	render(tw, t)

	face := *loadFont(t)
	before, _ := text.Measure("go further ", face, 0)
	link, _ := text.Measure("there", face, 0)
	point(uitest.Use(t), int(before+link/2), int(tw.Measurements().LineHeight()/2), false)

	tw.Update(&widget.UpdateObject{})
	event.ExecuteDeferred()

	is.Equal(entered, []string{"home"}) // link entered
}

func TestText_GlyphsFollowText(t *testing.T) {
	is := is.New(t)

	small := *loadFont(t)
	goFace, ok := small.(*text.GoTextFace)
	is.True(ok)
	var big text.Face = &text.GoTextFace{Source: goFace.Source, Size: 30}

	screen := ebiten.NewImage(400, 200)
	draw := func(tw *widget.Text) {
		tw.Validate()
		tw.Render(screen)
	}

	plain := newText(t, "first line\nsecond line")
	bbcode := newText(t, "a [color=#ff0000]red[/color] word", widget.TextOpts.ProcessBBCode(true))
	for _, change := range []func(tw *widget.Text){
		func(*widget.Text) {},
		func(tw *widget.Text) { tw.Label = "another label\nwith [color=#00ff00]two[/color] lines" },
		func(tw *widget.Text) { tw.SetFace(&big) },
	} {
		change(plain)
		draw(plain)
		total := 0
		for i, line := range plain.Measurements().ProcessedLineTexts() {
			want := visibleGlyphs(line, *plain.ComputedParams().Face)
			is.True(slices.Equal(plain.Measurements().ProcessedLineGlyphs()[i], want)) // line glyphs
			total += len(want)
		}
		is.Equal(len(plain.Glyphs()), total) // the buffer holds only the current glyphs

		change(bbcode)
		draw(bbcode)
		for _, pieces := range bbcode.Measurements().ProcessedLines() {
			for _, piece := range pieces {
				is.True(slices.Equal(piece.Glyphs(), visibleGlyphs(piece.Text(), *bbcode.ComputedParams().Face))) // piece glyphs
			}
		}
	}
}

func TestText_LaysOutOnlyLinesDrawn(t *testing.T) {
	is := is.New(t)

	tw := newText(t, "first line\nsecond line") // drawn on a 1x1 screen

	is.Equal(tw.Measurements().ProcessedLinesLaid(), []bool{true, false})
	is.Equal(tw.Measurements().ProcessedLineGlyphs()[1], nil) // the line below the screen
}

func TestText_FaceChangedInPlace(t *testing.T) {
	is := is.New(t)

	goFace, ok := (*loadFont(t)).(*text.GoTextFace)
	is.True(ok)
	face := *goFace
	var tf text.Face = &face
	tw := widget.NewText(widget.TextOpts.Text("Hello", &tf, color.White))
	render(tw, t)
	before := tw.Measurements().BoundingBoxWidth()

	face.Size *= 2 // a new size, without SetFace
	render(tw, t)

	is.True(tw.Measurements().BoundingBoxWidth() > 1.5*before)                                    // measured again
	is.True(slices.Equal(tw.Measurements().ProcessedLineGlyphs()[0], visibleGlyphs("Hello", tf))) // laid out again
}

func TestText_SetFaceLaysOutAgain(t *testing.T) {
	is := is.New(t)

	limited := text.NewLimitedFace(*loadFont(t))
	limited.AddUnicodeRange('a', 'b') // 'c' falls back to the missing glyph
	var tf text.Face = limited
	tw := widget.NewText(widget.TextOpts.Text("abc", &tf, color.White))
	render(tw, t)
	before := slices.Clone(tw.Measurements().ProcessedLineGlyphs()[0])

	limited.AddUnicodeRange('c', 'c') // the same metrics
	tw.SetFace(&tf)
	render(tw, t)

	got := tw.Measurements().ProcessedLineGlyphs()[0]
	is.True(slices.Equal(got, visibleGlyphs("abc", tf))) // laid out again
	is.True(!slices.Equal(got, before))                  // with the real 'c'
}

// A link and the punctuation after it move to the next line together (#271).
func TestText_WrapKeepsALinkWithItsPunctuation(t *testing.T) {
	is := is.New(t)

	face := *loadFont(t)
	fits, _ := text.Measure("Something about a very important thing", face, 0)
	dot, _ := text.Measure(".", face, 0)
	tw := newText(t, "Something about a very important [link=foo]thing[/link].",
		widget.TextOpts.ProcessBBCode(true), widget.TextOpts.MaxWidth(fits+dot/2))

	lines := make([]string, 0, len(tw.Measurements().ProcessedLines()))
	for _, pieces := range tw.Measurements().ProcessedLines() {
		lines = append(lines, widget.JoinPieces(pieces))
	}
	is.Equal(lines, []string{"Something about a very important ", "thing."})
}

// Text without spaces, as Chinese and Japanese are written, wraps between
// its characters, within MaxWidth, keeping every one.
func TestText_WrapWithoutSpaces(t *testing.T) {
	is := is.New(t)

	const label = "日本語の文章は空白なしで折り返します"
	face := *loadFont(t)
	five, _ := text.Measure("日本語の文", face, 0)
	tw := newText(t, label, widget.TextOpts.MaxWidth(five+1))

	is.True(len(tw.Measurements().ProcessedLineTexts()) > 2) // wrapped
	for _, w := range tw.Measurements().ProcessedLineWidths() {
		is.True(w < five+1) // line within MaxWidth
	}
	is.Equal(strings.Join(tw.Measurements().ProcessedLineTexts(), ""), label) // every character kept
}

// A text takes its theme's shadow unless it sets its own, and SetShadow
// replaces it.
func TestText_Shadow(t *testing.T) {
	is := is.New(t)

	themed := &widget.TextShadow{Color: color.Black, OffsetX: 1, OffsetY: 1}
	parent := widget.NewContainer()
	parent.GetWidget().SetTheme(&widget.Theme{TextTheme: &widget.TextParams{Shadow: themed}})

	plain := newText(t, "a")
	own := newText(t, "b", widget.TextOpts.Shadow(color.White, 2, 3))
	parent.AddChild(plain, own)
	plain.Validate()
	own.Validate()

	is.Equal(plain.ComputedParams().Shadow, themed) // the theme's
	is.Equal(*own.ComputedParams().Shadow, widget.TextShadow{Color: color.White, OffsetX: 2, OffsetY: 3})

	own.SetShadow(nil)
	is.Equal(own.ComputedParams().Shadow, nil) // removed
}

// visibleGlyphs returns the glyphs text.AppendGlyphs lays out for s, without
// those that have no image.
func visibleGlyphs(s string, face text.Face) []text.Glyph {
	var glyphs []text.Glyph
	for _, g := range text.AppendGlyphs(nil, s, face, nil) {
		if g.Image != nil {
			glyphs = append(glyphs, g)
		}
	}
	return glyphs
}

func newText(t *testing.T, label string, opts ...widget.TextOpt) *widget.Text {
	t.Helper()

	tw := widget.NewText(append(opts, widget.TextOpts.Text(label, loadFont(t), color.White))...)
	render(tw, t)
	return tw
}

// A link scrolled out of a scroll container's view is neither hovered nor
// clicked, though the cursor is where it is laid out.
func TestText_LinkOutOfViewTakesNoInput(t *testing.T) {
	is := is.New(t)

	var entered, clicked []string
	tw := widget.NewText(
		widget.TextOpts.Text("[link=away]there[/link]", loadFont(t), color.White),
		widget.TextOpts.ProcessBBCode(true),
		widget.TextOpts.LinkCursorEnteredHandler(func(args *widget.LinkEventArgs) { entered = append(entered, args.ID) }),
		widget.TextOpts.LinkClickedHandler(func(args *widget.LinkEventArgs) { clicked = append(clicked, args.ID) }),
	)
	content := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))
	content.AddChild(newSimpleWidget(100, 200, nil), tw)
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{Idle: newNineSliceEmpty(t), Mask: newNineSliceEmpty(t)}),
	)
	scroll.SetLocation(image.Rect(0, 0, 100, 50))
	render(scroll, t)
	d := uitest.Use(t) // before the layers, which it starts over
	input.SetupInputLayersWithDeferred([]input.Layerer{scroll})
	press := func() { // and let go
		r := tw.GetWidget().Rect
		for _, down := range []bool{true, false} {
			point(d, r.Min.X+5, r.Min.Y+r.Dy()/2, down)
			tw.Update(&widget.UpdateObject{})
			event.ExecuteDeferred()
		}
	}

	press()
	is.Equal(entered, nil) // out of view: not hovered
	is.Equal(clicked, nil) // out of view: not clicked

	scroll.ScrollTop = 1
	render(scroll, t)
	press()
	is.Equal(entered, []string{"away"}) // in view: hovered
	is.Equal(clicked, []string{"away"}) // in view: clicked
}

// A text whose label changes its size asks for the layout to be done
// again, through its parent's update; one whose size stays does not.
func TestText_RelayoutWhenResized(t *testing.T) {
	is := is.New(t)

	tw := newText(t, "short")
	parent := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	parent.AddChild(tw)
	parent.Validate()
	relayout := func() bool {
		u := &widget.UpdateObject{}
		parent.Update(u)
		return u.RelayoutRequested
	}
	relayout() // added: lays out anyway

	tw.Label = "a much longer label"
	is.True(relayout())  // grown
	is.True(!relayout()) // the same since
	tw.Label = "a much longer label"
	is.True(!relayout()) // set to what it was
	tw.MaxWidth = 60
	is.True(relayout()) // wrapped onto more lines
}
