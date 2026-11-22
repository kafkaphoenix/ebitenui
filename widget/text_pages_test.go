package widget_test

import (
	img "image"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/matryer/is"
)

// Slots of one height share a shelf left to right, a pixel apart; a taller
// one opens a shelf below, and a full shelf sends a slot to a new one.
func TestTextPage_Shelves(t *testing.T) {
	is := is.New(t)

	p := &widget.TextPage{}
	a := p.Alloc(300, 14)
	b := p.Alloc(300, 14)
	tall := p.Alloc(300, 30)
	is.Equal(a.Rect(), img.Rect(0, 0, 300, 14))
	is.Equal(b.Rect(), img.Rect(301, 0, 601, 14))   // beside a, past the gap
	is.Equal(tall.Rect(), img.Rect(0, 16, 300, 46)) // below the 16 px shelf

	c := p.Alloc(300, 14)
	d := p.Alloc(300, 14)                         // the 14 px shelf holds three
	is.Equal(c.Rect(), img.Rect(602, 0, 902, 14)) // third on the shelf
	is.Equal(d.Rect(), img.Rect(0, 48, 300, 62))  // a new shelf below the 32 px one
	is.Equal(p.Used(), 5)
}

// Of the shelves that fit a slot, the least tall takes it.
func TestTextPage_TightestShelf(t *testing.T) {
	is := is.New(t)

	p := &widget.TextPage{}
	p.Alloc(500, 38)                            // a 40 px shelf
	is.Equal(p.Alloc(600, 30).Rect().Min.Y, 40) // too wide for it: a 32 px shelf below
	is.Equal(p.Alloc(300, 30).Rect().Min.Y, 40) // both fit; the 32 px one wins
	is.Equal(p.Alloc(300, 30).Rect().Min.Y, 0)  // the 32 px one is full
}

// A full page still gives a short slot room on a taller shelf.
func TestTextPage_AnyShelfWhenFull(t *testing.T) {
	is := is.New(t)

	p := &widget.TextPage{}
	first := p.Alloc(widget.TextPageSize-1, 250)
	for range 3 {
		p.Alloc(widget.TextPageSize-1, 250) // four 256 px shelves fill the page
	}
	is.True(p.Alloc(100, 14) == nil)
	p.Free(first)
	s := p.Alloc(100, 14)
	is.True(s != nil)
	is.Equal(s.Rect().Min.Y, 0) // on the emptied top shelf
}

// A freed slot's room goes to the next slot that fits, joined with free room
// beside it; empty shelves at the bottom are dropped, so their rows serve
// any height.
func TestTextPage_Free(t *testing.T) {
	is := is.New(t)

	p := &widget.TextPage{}
	p.Alloc(300, 14)
	b := p.Alloc(300, 14)
	c := p.Alloc(300, 14)
	p.Free(b)
	again := p.Alloc(300, 14)
	is.Equal(again.Rect(), b.Rect()) // b's room again

	p.Free(again)
	p.Free(c)                                                     // joins the room on both sides
	is.Equal(p.Alloc(722, 14).Rect(), img.Rect(301, 0, 1023, 14)) // b's, c's and the rest
	is.Equal(p.Used(), 2)

	low := p.Alloc(300, 40)
	is.Equal(low.Rect().Min.Y, 16)
	p.Free(low)
	is.Equal(len(p.Shelves()), 1)               // the empty bottom shelf went
	is.Equal(p.Alloc(300, 60).Rect().Min.Y, 16) // its rows serve a taller slot
}

// Slots too tall or too wide for a page aren't given; a full page sends
// slots to a new page, and a page left empty beyond the first is released.
func TestTextPages_Pool(t *testing.T) {
	is := is.New(t)
	widget.EmptyTextPages(t)

	is.True(widget.AllocTextSlot(100, widget.TextPageMaxSlot+1) == nil) // too tall
	is.True(widget.AllocTextSlot(widget.TextPageSize, 14) == nil)       // too wide with its gap

	perPage := widget.TextPageSize / (widget.TextPageMaxSlot + 8) // 264 px shelves
	slots := make([]*widget.TextSlot, 0, perPage)
	for range perPage {
		slots = append(slots, widget.AllocTextSlot(widget.TextPageSize-1, widget.TextPageMaxSlot))
	}
	is.Equal(widget.TextPageCount(), 1)
	extra := widget.AllocTextSlot(widget.TextPageSize-1, widget.TextPageMaxSlot)
	is.Equal(widget.TextPageCount(), 2) // the first page is full
	is.True(extra.Page() != slots[0].Page())

	widget.FreeTextSlot(extra)
	is.Equal(widget.TextPageCount(), 1) // released
	for _, s := range slots {
		widget.FreeTextSlot(s)
	}
	is.Equal(widget.TextPageCount(), 1) // the first page stays
}

// Between collections the cache doesn't go over its slots: the slot of a
// text already gone is freed by the first sweep after a collection, not
// before. Automatic collections are off, so only the test's count.
func TestTextPages_SweepOnlyAfterACollection(t *testing.T) {
	widget.EmptyTextPages(t)
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	widget.SweepTextSlots()
	s := widget.AllocTextSlot(10, 10)
	widget.KeepTextSlot(s) // its text already gone: a zero weak pointer
	widget.SweepTextSlots()
	if s.Freed() {
		t.Fatal("a slot was freed with no collection since the last sweep")
	}
	runtime.GC()
	widget.SweepTextSlots()
	if !s.Freed() {
		t.Error("the slot of a text gone wasn't freed after a collection")
	}
}
