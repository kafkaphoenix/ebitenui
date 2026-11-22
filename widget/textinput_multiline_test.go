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
	"github.com/matryer/is"
	"golang.org/x/image/font/basicfont"
)

// linesInput is a focused multiline text input of a 7 by 13 pixel
// monospace face holding s, at 0,0 to w,h, drawn once: index i of a line
// lies at x = 7i, and line n at y = 13n.
func linesInput(t *testing.T, s string, w, h int, opts ...widget.TextInputOpt) *widget.TextInput {
	t.Helper()
	face := text.Face(text.NewGoXFace(basicfont.Face7x13))
	ti := widget.NewTextInput(append(opts, widget.TextInputOpts.Face(&face), widget.TextInputOpts.Multiline(true),
		widget.TextInputOpts.Color(&widget.TextInputColor{Idle: color.White, Disabled: color.White, Caret: color.White}))...)
	ti.SetText(s)
	ti.GetWidget().Rect = image.Rect(0, 0, w, h)
	event.ExecuteDeferred()
	render(ti, t)
	ti.Focus(true)
	return ti
}

// A multiline input wraps its text at its width, breaking after a space,
// and at each newline; a word wider than the input breaks where it must.
func TestTextInput_MultilineWraps(t *testing.T) {
	is := is.New(t)

	ti := linesInput(t, "hello brave world\n\nabcdefghijklmn", 70, 100) // 10 letters a line
	is.Equal(ti.WrappedLines(), []string{"hello ", "brave ", "world", "", "abcdefghij", "klmn"})
	ti.SetText("")
	is.Equal(ti.WrappedLines(), []string{""})
	ti.SetText("one\n")
	is.Equal(ti.WrappedLines(), []string{"one", ""})
	ti.GetWidget().Rect = image.Rect(0, 0, 21, 100) // 3 letters
	is.Equal(ti.WrappedLines(), []string{"one", ""})
	ti.GetWidget().Rect = image.Rect(0, 0, 5, 100) // narrower than a letter: one a line
	is.Equal(ti.WrappedLines(), []string{"o", "n", "e", ""})
}

// Enter starts a new line, rather than submitting; a single-line input
// still submits.
func TestTextInput_MultilineEnter(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	submits := 0
	ti := linesInput(t, "ab", 200, 100, widget.TextInputOpts.SubmitHandler(func(*widget.TextInputChangedEventArgs) { submits++ }))
	ti.SetCursorPosition(2)
	holdKey(ti, d, ebiten.KeyEnter, 1)
	is.Equal(ti.GetText(), "ab\n")
	is.Equal(submits, 0)

	single := monoInput(t, "ab")
	single.SubmitEvent.AddHandler(func(any) { submits++ })
	holdKey(single, d, ebiten.KeyEnter, 1)
	event.ExecuteDeferred()
	is.Equal(submits, 1)
}

// Up and Down move the caret to the line above or below, keeping its
// column where the line is long enough; Home and End move it to its
// line's ends. In a single-line input, Up and Down go to the text's ends.
func TestTextInput_MultilineCaretMoves(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := linesInput(t, "one\ntwo\nthree\nab", 200, 100)
	ti.SetCursorPosition(2) // on|e
	for _, s := range []struct {
		key ebiten.Key
		at  int
	}{
		{ebiten.KeyArrowDown, 6},  // tw|o
		{ebiten.KeyArrowDown, 10}, // th|ree
		{ebiten.KeyEnd, 13},       // three|
		{ebiten.KeyArrowDown, 16}, // ab|, the column held to the line
		{ebiten.KeyArrowDown, 16}, // the last line
		{ebiten.KeyArrowUp, 10},
		{ebiten.KeyHome, 8},
		{ebiten.KeyArrowUp, 4},
		{ebiten.KeyArrowUp, 0},
		{ebiten.KeyArrowUp, 0}, // the first line
	} {
		holdKey(ti, d, s.key, 1)
		is.Equal(ti.CursorPosition(), s.at)
	}

	single := monoInput(t, "hello")
	single.SetCursorPosition(2)
	holdKey(single, d, ebiten.KeyArrowDown, 1)
	is.Equal(single.CursorPosition(), 5)
	holdKey(single, d, ebiten.KeyArrowUp, 1)
	is.Equal(single.CursorPosition(), 0)
}

// Shift with Up or Down selects across lines.
func TestTextInput_MultilineShiftSelects(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := linesInput(t, "one\ntwo", 200, 100)
	ti.SetCursorPosition(1)
	d.Keys[ebiten.KeyShift] = true
	holdKey(ti, d, ebiten.KeyArrowDown, 1)
	is.Equal(ti.SelectedText(), "ne\nt")
}

// A click places the caret on the line it lands on, nearest between two
// letters; above the text, at its start, below it, at its end.
func TestTextInput_MultilineClick(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	ti := linesInput(t, "one\ntwo\nthree", 200, 100)
	for _, s := range []struct {
		x, y, at int
	}{
		{10, 18, 5},  // t|wo
		{30, 30, 12}, // thre|e
		{100, 5, 3},  // past the end of one
		{10, 60, 9},  // under the last line: on it
	} {
		pointAtXY(ti, d, s.x, s.y, true)
		pointAtXY(ti, d, s.x, s.y, false)
		is.Equal(ti.CursorPosition(), s.at)
	}
}

// pointAtXY moves the pointer to x, y with the left button down or not and
// runs an update of ti.
func pointAtXY(ti *widget.TextInput, d *uitest.Devices, x, y int, down bool) {
	d.Cursor = image.Pt(x, y)
	d.Mouse[ebiten.MouseButtonLeft] = down
	input.Update()
	ti.Update(&widget.UpdateObject{})
	input.AfterUpdate()
	event.ExecuteDeferred()
}

// A multiline input scrolls to keep the caret in view, and its caret is
// drawn on its line.
func TestTextInput_MultilineScrollsToTheCaret(t *testing.T) {
	is := is.New(t)

	ti := linesInput(t, "1\n2\n3\n4\n5", 100, 26) // two lines in view
	ti.SetCursorPosition(9)
	ti.Render(ebiten.NewImage(100, 26))
	is.Equal(ti.CaretRect().Min.Y, 13) // the last line, at the bottom of the view
	ti.SetCursorPosition(0)
	ti.Render(ebiten.NewImage(100, 26))
	is.Equal(ti.CaretRect().Min.Y, 0)
	ti.SetCursorPosition(4) // 3
	ti.Render(ebiten.NewImage(100, 26))
	is.Equal(ti.CaretRect().Min.Y, 13)
}

// Clearing on submit puts the caret at the start of the emptied text.
func TestTextInput_MultilineClearOnSubmit(t *testing.T) {
	is := is.New(t)

	ti := linesInput(t, "one\ntwo", 200, 100, widget.TextInputOpts.ClearOnSubmit(true))
	ti.SetCursorPosition(6)
	ti.Submit()
	is.Equal(ti.GetText(), "")
	is.Equal(ti.CursorPosition(), 0)
}

// A theme can make text inputs multiline.
func TestTextInput_MultilineFromTheTheme(t *testing.T) {
	is := is.New(t)

	on := true
	ti := widget.NewTextInput()
	ti.GetWidget().SetTheme(&widget.Theme{DefaultFace: loadFont(t), TextInputTheme: &widget.TextInputParams{
		Multiline: &on, Color: &widget.TextInputColor{Idle: color.White},
	}})
	ti.Validate()
	is.True(*ti.ComputedParams().Multiline)
}

// An unchanged multiline input wraps its text once, not every draw: the
// lines it keeps cost nothing to ask for again.
func TestTextInput_MultilineWrapsOnce(t *testing.T) {
	is := is.New(t)

	ti := linesInput(t, "hello brave world\nand more", 70, 100)
	ti.Wrap()
	is.Equal(testing.AllocsPerRun(10, ti.Wrap), 0.0)
}
