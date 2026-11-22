package widget_test

import (
	img "image"
	"image/color"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui/event"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/internal/changes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// newCarousel returns a carousel of n pages, 100 by 50 pixels, at 0,0 to
// 300,200, drawn once, and its pages.
func newCarousel(t *testing.T, n int, opts ...widget.CarouselOpt) (*widget.Carousel, []*simpleWidget) {
	t.Helper()
	pages := make([]*simpleWidget, n)
	ws := make([]widget.PreferredSizeLocateableWidget, n)
	for i := range pages {
		pages[i] = newSimpleWidget(100, 50, nil)
		ws[i] = pages[i]
	}
	c := widget.NewCarousel(append([]widget.CarouselOpt{widget.CarouselOpts.Pages(ws...)}, opts...)...)
	c.Validate()
	c.SetLocation(img.Rect(0, 0, 300, 200))
	c.Render(ebiten.NewImage(300, 200))
	return c, pages
}

// updateCarousel runs n updates of c, the input's too, and draws it.
func updateCarousel(c *widget.Carousel, n int) {
	for range n {
		input.Update()
		c.Update(&widget.UpdateObject{})
		input.AfterUpdate()
		event.ExecuteDeferred()
	}
	c.Render(ebiten.NewImage(300, 200))
}

// A carousel shows one page at a time, in the page area above its dots;
// Next and Previous move along, stopping at the ends unless it loops, and
// it tells of each move. For loading screens and tutorials.
func TestCarousel_Pages(t *testing.T) {
	is := is.New(t)

	var moves [][2]int
	c, pages := newCarousel(t, 3, widget.CarouselOpts.SlideTime(0),
		widget.CarouselOpts.PageChangedHandler(func(a *widget.CarouselPageChangedEventArgs) {
			moves = append(moves, [2]int{a.Previous, a.Page})
		}))
	is.Equal(c.Page(), 0)
	is.Equal(pages[0].GetWidget().Rect, c.PageRect())
	is.Equal(c.PageRect(), img.Rect(0, 0, 300, 180)) // the dots' row under it

	c.Next()
	event.ExecuteDeferred()
	is.Equal(c.Page(), 1)
	c.Render(ebiten.NewImage(300, 200))
	is.Equal(pages[1].GetWidget().Rect, c.PageRect())
	c.Previous()
	c.Previous() // the start
	is.Equal(c.Page(), 0)
	c.SetPage(2)
	c.Next() // the end
	event.ExecuteDeferred()
	is.Equal(c.Page(), 2)
	is.Equal(moves, [][2]int{{0, 1}, {1, 0}, {0, 2}})

	loop, _ := newCarousel(t, 3, widget.CarouselOpts.Loop())
	loop.Previous()
	is.Equal(loop.Page(), 2)
	loop.Next()
	is.Equal(loop.Page(), 0)
}

// A move slides the new page in over SlideTime, marking a change each
// update so a game that skips unchanged frames draws it; a still carousel
// marks none.
func TestCarousel_Slides(t *testing.T) {
	is := is.New(t)

	c, _ := newCarousel(t, 3, widget.CarouselOpts.SlideTime(500*time.Millisecond))
	from := changes.Count()
	updateCarousel(c, 10)
	is.Equal(changes.Count(), from) // still
	c.Next()
	is.True(c.Sliding())
	from = changes.Count()
	updateCarousel(c, ebiten.TPS()/2)
	is.True(!c.Sliding())
	is.True(changes.Count()-from >= uint64(ebiten.TPS()/2))
	from = changes.Count()
	updateCarousel(c, 10)
	is.Equal(changes.Count(), from) // still again
}

// AutoAdvance moves on every while, round to the start with Loop, and
// waits while the pointer is over the carousel.
func TestCarousel_AutoAdvance(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	d.Cursor = img.Pt(400, 400) // away
	c, _ := newCarousel(t, 2, widget.CarouselOpts.AutoAdvance(time.Second), widget.CarouselOpts.Loop(), widget.CarouselOpts.SlideTime(0))
	updateCarousel(c, ebiten.TPS()-1)
	is.Equal(c.Page(), 0)
	updateCarousel(c, 1)
	is.Equal(c.Page(), 1)
	updateCarousel(c, ebiten.TPS())
	is.Equal(c.Page(), 0) // round again

	d.Cursor = img.Pt(150, 100) // over it
	updateCarousel(c, 3*ebiten.TPS())
	is.Equal(c.Page(), 0)
	d.Cursor = img.Pt(400, 400)
	updateCarousel(c, ebiten.TPS()+1) // an update for the leaving to be told, as events are
	is.Equal(c.Page(), 1)
}

// In Steps mode a page is reached only once the steps before it are done:
// the game marks each as its condition is met. Going back is always
// allowed; with AdvanceWhenDone, a step done moves on by itself.
func TestCarousel_Steps(t *testing.T) {
	is := is.New(t)

	c, _ := newCarousel(t, 3, widget.CarouselOpts.Steps(), widget.CarouselOpts.SlideTime(0))
	c.Next()
	is.Equal(c.Page(), 0) // step 0 not done
	is.True(c.NextButton().GetWidget().Disabled)
	c.SetStepDone(0, true)
	is.True(c.StepDone(0))
	c.Render(ebiten.NewImage(300, 200))
	is.True(!c.NextButton().GetWidget().Disabled)
	c.SetPage(2)
	is.Equal(c.Page(), 0) // step 1 not done
	c.Next()
	is.Equal(c.Page(), 1)
	c.Previous()
	is.Equal(c.Page(), 0)
	c.SetStepDone(1, true)
	c.SetPage(2)
	is.Equal(c.Page(), 2)

	later, _ := newCarousel(t, 3, widget.CarouselOpts.Steps(), widget.CarouselOpts.SlideTime(0))
	later.SetStepDone(1, true)
	later.SetPage(2)
	is.Equal(later.Page(), 0) // step 0, before it, not done

	auto, _ := newCarousel(t, 3, widget.CarouselOpts.Steps(), widget.CarouselOpts.AdvanceWhenDone(), widget.CarouselOpts.SlideTime(0))
	auto.SetStepDone(0, true)
	is.Equal(auto.Page(), 1)
	auto.Previous()
	auto.SetStepDone(1, true) // not the step shown: no move
	is.Equal(auto.Page(), 0)
}

// clickAt presses the left button at x, y for an update and lets go for
// another.
func clickAt(c *widget.Carousel, d *uitest.Devices, x, y int) {
	d.Cursor = img.Pt(x, y)
	d.Mouse[ebiten.MouseButtonLeft] = true
	updateCarousel(c, 1)
	d.Mouse[ebiten.MouseButtonLeft] = false
	updateCarousel(c, 1)
}

// The arrows and the dots move it: the right arrow on, the left back, a
// dot to its page; the arrows sit at the page area's sides.
func TestCarousel_ArrowsAndDots(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	c, _ := newCarousel(t, 3, widget.CarouselOpts.SlideTime(0))
	next, prev := c.NextButton().GetWidget().Rect, c.PrevButton().GetWidget().Rect
	is.Equal(next, img.Rect(300-4-24, 90-16, 300-4, 90+16))
	is.Equal(prev, img.Rect(4, 90-16, 4+24, 90+16))
	center := func(r img.Rectangle) img.Point { return img.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2) }

	is.True(c.PrevButton().GetWidget().Disabled) // nothing before the first page
	p := center(next)
	clickAt(c, d, p.X, p.Y)
	is.Equal(c.Page(), 1)
	is.True(!c.PrevButton().GetWidget().Disabled)
	p = center(prev)
	clickAt(c, d, p.X, p.Y)
	is.Equal(c.Page(), 0)
	p = center(c.DotRect(2))
	clickAt(c, d, p.X, p.Y)
	is.Equal(c.Page(), 2)
	is.Equal(c.DotRect(1).Max.Y, 194) // in the row under the pages: 180 + 6 + 8

	none, _ := newCarousel(t, 3, widget.CarouselOpts.NoArrows(), widget.CarouselOpts.NoDots())
	is.Equal(none.PageRect(), img.Rect(0, 0, 300, 200))
	is.True(none.NextButton() == nil)
}

// drag drags the pointer from x0 to x1 at y 100 over four updates.
func drag(c *widget.Carousel, d *uitest.Devices, x0, x1 int) {
	d.Cursor = img.Pt(x0, 100)
	d.Mouse[ebiten.MouseButtonLeft] = true
	updateCarousel(c, 1)
	for _, f := range []int{1, 2, 3} {
		d.Cursor = img.Pt(x0+(x1-x0)*f/3, 100)
		updateCarousel(c, 1)
	}
	d.Mouse[ebiten.MouseButtonLeft] = false
	updateCarousel(c, 1)
}

// A swipe across more than a fifth of the page moves to the page that way;
// a shorter one springs back, as does one toward a page not reached yet.
func TestCarousel_Swipe(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	c, _ := newCarousel(t, 3, widget.CarouselOpts.SlideTime(0))
	drag(c, d, 200, 100) // 100 left
	is.Equal(c.Page(), 1)
	drag(c, d, 100, 140) // 40 right: short
	is.Equal(c.Page(), 1)
	drag(c, d, 100, 200)
	is.Equal(c.Page(), 0)

	steps, pages := newCarousel(t, 3, widget.CarouselOpts.Steps(), widget.CarouselOpts.SlideTime(0))
	d.Cursor = img.Pt(200, 100)
	d.Mouse[ebiten.MouseButtonLeft] = true
	updateCarousel(steps, 1)
	d.Cursor = img.Pt(100, 100)
	updateCarousel(steps, 1)
	is.Equal(pages[0].GetWidget().Rect.Min.X, 0) // the page doesn't follow toward a step not reached
	d.Mouse[ebiten.MouseButtonLeft] = false
	updateCarousel(steps, 1)
	is.Equal(steps.Page(), 0) // step 0 not done

	still, _ := newCarousel(t, 3, widget.CarouselOpts.NoSwipe(), widget.CarouselOpts.SlideTime(0))
	drag(still, d, 200, 50)
	is.Equal(still.Page(), 0)
}

// The page shown takes updates; the others don't.
func TestCarousel_UpdatesThePageShown(t *testing.T) {
	is := is.New(t)

	a, b := &boundRow{simpleWidget: *newSimpleWidget(10, 10, nil)}, &boundRow{simpleWidget: *newSimpleWidget(10, 10, nil)}
	c := widget.NewCarousel(widget.CarouselOpts.Pages(a, b))
	c.Validate()
	c.SetLocation(img.Rect(0, 0, 300, 200))
	updateCarousel(c, 3)
	is.Equal([]int{a.updates, b.updates}, []int{3, 0})
}

// Its preferred size is its largest page's, and the dots' row under it.
func TestCarousel_PreferredSize(t *testing.T) {
	is := is.New(t)

	c := widget.NewCarousel(widget.CarouselOpts.Pages(newSimpleWidget(100, 50, nil), newSimpleWidget(120, 40, nil)))
	c.Validate()
	w, h := c.PreferredSize()
	is.Equal([]int{w, h}, []int{120, 50 + 6 + 8 + 6})
}

// An update of a still carousel allocates nothing.
func TestCarousel_IdleUpdateAllocatesNothing(t *testing.T) {
	uitest.Use(t)
	c, _ := newCarousel(t, 3)
	updateCarousel(c, 1)
	u := &widget.UpdateObject{}
	if n := testing.AllocsPerRun(100, func() {
		input.Update()
		c.Update(u)
		input.AfterUpdate()
	}); n != 0 {
		t.Errorf("an update allocated %v times", n)
	}
}

// The arrows and the focusable widgets of the page shown take the focus.
func TestCarousel_Focusers(t *testing.T) {
	is := is.New(t)

	b := widget.NewButton(widget.ButtonOpts.Image(&widget.ButtonImage{Idle: newNineSliceEmpty(t), Pressed: newNineSliceEmpty(t)}))
	c := widget.NewCarousel(widget.CarouselOpts.Pages(b, newSimpleWidget(10, 10, nil)))
	c.Validate()
	fs := c.GetFocusers()
	is.Equal(len(fs), 3)
	is.True(fs[0] == c.PrevButton() && fs[1] == c.NextButton() && fs[2] == b)
}

// A theme gives a carousel its arrows' and dots' look.
func TestCarousel_Theme(t *testing.T) {
	is := is.New(t)

	arrow := &widget.ButtonImage{Idle: eimage.NewNineSliceColor(color.Black), Pressed: eimage.NewNineSliceColor(color.White)}
	c := widget.NewCarousel(widget.CarouselOpts.Pages(newSimpleWidget(10, 10, nil)))
	c.GetWidget().SetTheme(&widget.Theme{CarouselTheme: &widget.CarouselParams{ArrowImage: arrow, DotColor: color.Black}})
	c.Validate()
	is.True(c.ComputedParams().ArrowImage == arrow)
	is.Equal(c.ComputedParams().DotColor, color.Black)
}
