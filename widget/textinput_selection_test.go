package widget_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"

	"github.com/matryer/is"
)

func TestSelectedText_NonASCII_RuneIndexing(t *testing.T) {
	const want = "Ünïcödé Ship Name"

	tin := widget.NewTextInput()
	tin.SetText(want)
	tin.SelectAll()

	if got := tin.SelectedText(); got != want {
		t.Fatalf("SelectedText() after SelectAll() on %q = %q, want %q", want, got, want)
	}
}

func TestDeleteSelectedText_NonASCII_RuneIndexing(t *testing.T) {
	const text = "Ünïcödé Ship Name"

	tin := widget.NewTextInput()
	tin.SetText(text)
	tin.SelectAll()
	tin.DeleteSelectedText()

	if got := tin.GetText(); got != "" {
		t.Fatalf("GetText() after SelectAll()+DeleteSelectedText() = %q, want empty", got)
	}
}

func TestDeleteSelectedText_PositionalNotSubstringSearch(t *testing.T) {
	tin := widget.NewTextInput()
	tin.SetText("aXa")

	tin.SetDragStartIndex(2)
	tin.SetCursorPosition(3)

	if got := tin.SelectedText(); got != "a" {
		t.Fatalf("SelectedText() for [2:3] of \"aXa\" = %q, want %q", got, "a")
	}

	tin.DeleteSelectedText()

	const want = "aX"
	if got := tin.GetText(); got != want {
		t.Fatalf("GetText() after deleting [2:3] of \"aXa\" = %q, want %q", got, want)
	}
}

func TestDeleteSelectedText_PositionalDeleteAlsoCorrectsCursor(t *testing.T) {
	tin := widget.NewTextInput()
	tin.SetText("aXa")
	tin.SetDragStartIndex(2)
	tin.SetCursorPosition(3)

	tin.DeleteSelectedText()

	if tin.CursorPosition() != 2 {
		t.Fatalf("cursorPosition after deleting [2:3] of \"aXa\" = %d, want 2", tin.CursorPosition())
	}
	tin.Insert("Y")
	if got := tin.GetText(); got != "aXY" {
		t.Fatalf("GetText() after positional delete + Insert(\"Y\") = %q, want %q", got, "aXY")
	}
}

// A selection deleted from the middle of the text keeps the text after it,
// whichever way the selection was made, and leaves the caret where it
// began; typing over a selection replaces it.
func TestDeleteSelectedText_InTheMiddle(t *testing.T) {
	is := is.New(t)

	ti := newTextInput(t)
	for _, sel := range [][2]int{{5, 8}, {8, 5}} { // the second "año", forwards and backwards
		ti.SetText("año, año!")
		ti.SetDragStartIndex(sel[0])
		ti.SetCursorPosition(sel[1])
		is.Equal(ti.SelectedText(), "año")
		ti.DeleteSelectedText()
		is.Equal(ti.GetText(), "año, !")
		is.Equal(ti.CursorPosition(), 5)
	}

	ti.SetText("niño")
	ti.SetDragStartIndex(1)
	ti.SetCursorPosition(3)
	ti.Insert("ñ")
	is.Equal(ti.GetText(), "nño")
}

// monoInput is a focused text input of a 7-pixel monospace face holding
// s, at 0,0 to 200,20, so that index i lies at x = 7i.
func monoInput(t *testing.T, s string) *widget.TextInput {
	t.Helper()
	face := text.Face(text.NewGoXFace(basicfont.Face7x13))
	ti := widget.NewTextInput(widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White, Disabled: color.White, Caret: color.White}))
	ti.SetText(s)
	ti.GetWidget().Rect = image.Rect(0, 0, 200, 20)
	event.ExecuteDeferred()
	render(ti, t)
	ti.Focus(true)
	return ti
}

// pointAt moves the pointer to x with the left button down or not and
// runs an update of ti.
func pointAt(ti *widget.TextInput, d *uitest.Devices, x int, down bool) {
	point(d, x, 10, down)
	ti.Update(&widget.UpdateObject{})
	input.AfterUpdate()
	event.ExecuteDeferred()
}

// A double click selects the word it lands on as the button goes down the
// second time, and keeps it while the button drags and as it comes up; a
// third click places the caret. For ebitenui/ebitenui#350.
func TestTextInput_DoubleClickSelectsAWord(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := monoInput(t, "hello brave world")
	pointAt(ti, d, 57, true) // between r and a
	pointAt(ti, d, 57, false)
	is.Equal(ti.SelectedText(), "")
	pointAt(ti, d, 57, true)
	is.Equal(ti.SelectedText(), "brave") // as it goes down
	pointAt(ti, d, 120, true)
	is.Equal(ti.SelectedText(), "brave") // dragged
	pointAt(ti, d, 120, false)
	is.Equal(ti.SelectedText(), "brave") // let go

	pointAt(ti, d, 57, true)
	pointAt(ti, d, 57, false)
	is.Equal(ti.SelectedText(), "") // a third click
	is.Equal(ti.CursorPosition(), 8)

	for range ebiten.TPS() {
		pointAt(ti, d, 57, false)
	}
	pointAt(ti, d, 0, true)
	pointAt(ti, d, 35, true)
	pointAt(ti, d, 35, false)
	is.Equal(ti.SelectedText(), "hello") // a drag selects again
}

// Clicks too far apart in time make no double click.
func TestTextInput_DoubleClickTooSlow(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := monoInput(t, "hello brave world")
	pointAt(ti, d, 57, true)
	pointAt(ti, d, 57, false)
	for range ebiten.TPS() {
		pointAt(ti, d, 57, false)
	}
	pointAt(ti, d, 57, true)
	pointAt(ti, d, 57, false)
	is.Equal(ti.SelectedText(), "")
	is.Equal(ti.CursorPosition(), 8)
}

// A click with Shift held selects from the caret to where it lands.
func TestTextInput_ShiftClick(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := monoInput(t, "hello brave world")
	pointAt(ti, d, 42, true) // before b
	pointAt(ti, d, 42, false)
	d.Keys[ebiten.KeyShift] = true
	pointAt(ti, d, 77, true) // after e
	pointAt(ti, d, 77, false)
	is.Equal(ti.SelectedText(), "brave")
}

// The arrow keys, Home and End with Shift held select as they move the
// caret; without Shift they end a selection, at its start going left and
// at its end going right, and then move the caret.
func TestTextInput_ShiftArrows(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := monoInput(t, "hello")
	ti.SetCursorPosition(1)
	d.Keys[ebiten.KeyShift] = true
	holdKey(ti, d, ebiten.KeyArrowRight, 1)
	holdKey(ti, d, ebiten.KeyArrowRight, 1)
	is.Equal(ti.SelectedText(), "el")
	d.Keys[ebiten.KeyShift] = false
	holdKey(ti, d, ebiten.KeyArrowLeft, 1)
	is.Equal(ti.SelectedText(), "")
	is.Equal(ti.CursorPosition(), 1) // the old selection's start
	holdKey(ti, d, ebiten.KeyArrowRight, 1)
	is.Equal(ti.CursorPosition(), 2)

	d.Keys[ebiten.KeyShift] = true
	holdKey(ti, d, ebiten.KeyEnd, 1)
	is.Equal(ti.SelectedText(), "llo")
	d.Keys[ebiten.KeyShift] = false
	holdKey(ti, d, ebiten.KeyArrowRight, 1)
	is.Equal(ti.SelectedText(), "")
	is.Equal(ti.CursorPosition(), 5) // the old selection's end
	d.Keys[ebiten.KeyShift] = true
	holdKey(ti, d, ebiten.KeyHome, 1)
	is.Equal(ti.SelectedText(), "hello")
}

// A word is a run of letters, digits and underscores; runs of spaces and
// of punctuation count as words too.
func TestTextInput_WordBounds(t *testing.T) {
	tests := []struct {
		text          string
		index         int
		expectedStart int
		expectedEnd   int
	}{
		{"hello brave world", 8, 6, 11},
		{"hello brave world", 6, 6, 11},
		{"hello  world", 6, 5, 7},
		{"a...b", 2, 1, 4},
		{"snake_case name", 3, 0, 10},
		{"hello world", 11, 6, 11},
		{"", 0, 0, 0},
		{"abc123 def", 2, 0, 6},
		{"a. .b", 2, 2, 3},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			is := is.New(t)

			start, end := widget.WordBounds([]rune(test.text), test.index)

			is.Equal(start, test.expectedStart)
			is.Equal(end, test.expectedEnd)
		})
	}
}
