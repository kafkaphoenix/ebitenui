package widget

import (
	"math"
	"strings"

	img "image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// A multiline text input: its text wrapped at its width and at each
// newline, worked out again only when the text shown, the width or the
// face changes. For ebitenui/ebitenui#340.

// Multiline makes the input take lines: Enter starts a new line rather
// than submitting, the text wraps at the input's width, breaking after a
// space where it can, Up and Down move the caret between lines and Home
// and End to its line's ends, and the input scrolls down to keep the
// caret in view.
func (o TextInputOptions) Multiline(multiline bool) TextInputOpt {
	return func(t *TextInput) {
		t.definedParams.Multiline = &multiline
	}
}

// wrappedLine is a line of the text as drawn: its text, and the index of
// its first rune in the text shown.
type wrappedLine struct {
	text  string
	start int
}

// wrapKey is what a text's lines are wrapped from.
type wrapKey struct {
	text  string
	width int
	face  *text.Face
}

// wrap is the text shown wrapped into lines, as one label for the input's
// Text, and the height of a line.
type wrap struct {
	key     wrapKey
	lines   []wrappedLine
	label   string
	lineH   float64
	wrapped bool
}

func (t *TextInput) multiline() bool {
	return t.computedParams.Multiline != nil && *t.computedParams.Multiline
}

// wrappedLines returns the text shown wrapped into lines at the input's
// width, worked out again when the text, the width or the face changes.
func (t *TextInput) wrappedLines() []wrappedLine {
	w := &t.wrap
	k := wrapKey{t.shown(), t.contentWidth(), t.computedParams.Face}
	if !w.wrapped || k != w.key {
		w.key, w.wrapped = k, true
		w.lines = wrapText(w.lines[:0], k.text, k.face, k.width)
		parts := make([]string, len(w.lines))
		for i, l := range w.lines {
			parts[i] = l.text
		}
		w.label = strings.Join(parts, "\n")
		_, w.lineH = text.Measure(" ", *k.face, 0)
	}
	return w.lines
}

// contentWidth returns the width the text has, the input's less its
// padding.
func (t *TextInput) contentWidth() int {
	if t.widget == nil {
		return 0
	}
	return t.widget.Rect.Dx() - t.computedParams.Padding.Left - t.computedParams.Padding.Right
}

// wrapText appends to dst the lines of s that fit in maxWidth pixels,
// breaking after a space where it can and at each newline; a word wider
// than maxWidth breaks where it must. With no width, it breaks only at
// newlines.
func wrapText(dst []wrappedLine, s string, face *text.Face, maxWidth int) []wrappedLine {
	runes := []rune(s)
	for start := 0; start <= len(runes); {
		end := start
		for end < len(runes) && runes[end] != '\n' {
			end++
		}
		dst = wrapLine(dst, runes, start, end, face, maxWidth)
		start = end + 1
	}
	return dst
}

// wrapLine appends the lines runes[start:end] wraps into, a line of the
// text without its newline: one empty line for an empty one.
func wrapLine(dst []wrappedLine, runes []rune, start, end int, face *text.Face, maxWidth int) []wrappedLine {
	if start == end {
		return append(dst, wrappedLine{start: start})
	}
	line := string(runes[start:end])
	for from := start; from < end; {
		to := end
		if maxWidth > 0 {
			to = lineBreak(runes, line, start, from, end, face, maxWidth)
		}
		dst = append(dst, wrappedLine{text: string(runes[from:to]), start: from})
		from = to
	}
	return dst
}

// lineBreak returns where the wrapped line from rune from of runes ends:
// after the last space of what fits in maxWidth pixels, or where it
// doesn't fit when there is none, a rune at least. Widths are measured
// within line, runes[start:end], as it is shaped.
func lineBreak(runes []rune, line string, start, from, end int, face *text.Face, maxWidth int) int {
	left := fontAdvanceAt(line, from-start, face)
	fit := from
	for fit < end && fontAdvanceAt(line, fit+1-start, face)-left <= maxWidth {
		fit++
	}
	fit = max(fit, from+1)
	if fit < end {
		at := fit
		for at > from && runes[at-1] != ' ' {
			at--
		}
		if at > from {
			fit = at
		}
	}
	return fit
}

// lineAt returns the line the rune index cursor is on and its column in
// it.
func lineAt(lines []wrappedLine, cursor int) (int, int) {
	for i := len(lines) - 1; i >= 0; i-- {
		if cursor >= lines[i].start {
			return i, min(cursor-lines[i].start, len([]rune(lines[i].text)))
		}
	}
	return 0, 0
}

// CursorMoveUp moves the caret of a multiline input to the line above, in
// the same column where that line is long enough.
func (t *TextInput) CursorMoveUp() {
	t.init.Do()
	t.moveLines(-1)
}

// CursorMoveDown moves the caret of a multiline input to the line below,
// in the same column where that line is long enough.
func (t *TextInput) CursorMoveDown() {
	t.init.Do()
	t.moveLines(1)
}

func (t *TextInput) moveLines(by int) {
	if !t.multiline() {
		return
	}
	lines := t.wrappedLines()
	line, col := lineAt(lines, t.cursorPosition)
	if to := line + by; to >= 0 && to < len(lines) {
		t.cursorPosition = lines[to].start + min(col, len([]rune(lines[to].text)))
	}
	t.caret.ResetBlinking()
}

// moveUp and moveDown are Up's and Down's commands: between lines in a
// multiline input, to the text's ends in a single-line one.
func (t *TextInput) moveUp() {
	if t.multiline() {
		t.CursorMoveUp()
	} else {
		t.CursorMoveStart()
	}
}

func (t *TextInput) moveDown() {
	if t.multiline() {
		t.CursorMoveDown()
	} else {
		t.CursorMoveEnd()
	}
}

// lineEnds returns the rune indexes where the caret's line starts and
// ends.
func (t *TextInput) lineEnds() (int, int) {
	lines := t.wrappedLines()
	line, _ := lineAt(lines, t.cursorPosition)
	l := lines[line]
	return l.start, l.start + len([]rune(l.text))
}

// lineIndex returns where in the text x, y, within the text's area tr,
// points: on the line it lands on, or the nearest, the nearest place
// between two letters.
func (t *TextInput) lineIndex(x, y int, tr img.Rectangle) int {
	lines := t.wrappedLines()
	line := int(math.Floor(float64(y-tr.Min.Y-t.scrollOffsetY) / t.wrap.lineH))
	line = min(max(line, 0), len(lines)-1)
	return lines[line].start + fontStringIndex([]rune(lines[line].text), t.computedParams.Face, x-tr.Min.X)
}

// drawLines draws a multiline input's text, its selection or composition
// and its caret, scrolled down to keep the caret in view.
func (t *TextInput) drawLines(screen *ebiten.Image) {
	rect := t.widget.Rect
	pad := t.computedParams.Padding
	tr := rect.Add(img.Point{pad.Left, pad.Top})
	lines := t.wrappedLines()
	lineH := t.wrap.lineH
	lineY := func(i int) int { return int(math.Round(float64(i) * lineH)) }

	line, col := lineAt(lines, t.shownCaret())
	cx := fontAdvanceAt(lines[line].text, col, t.computedParams.Face)
	cy := lineY(line)
	view := rect.Dy() - pad.Top - pad.Bottom
	if bottom := lineY(line + 1); bottom > view-t.scrollOffsetY {
		t.scrollOffsetY = view - bottom
	}
	if cy < -t.scrollOffsetY {
		t.scrollOffsetY = -cy
	}

	if t.focused {
		if t.dragStartIndex != -1 {
			t.highlightLines(screen, tr, lines, min(t.dragStartIndex, t.cursorPosition), max(t.dragStartIndex, t.cursorPosition))
		}
		if t.composition != "" {
			t.highlightLines(screen, tr, lines, t.cursorPosition, t.cursorPosition+len([]rune(t.composition)))
		}
	}

	t.text.SetLocation(tr.Add(img.Point{0, t.scrollOffsetY}))
	if t.inputText != "" {
		t.text.Label = t.wrap.label
	} else {
		t.text.Label = t.placeholderText
	}
	t.setTextColor()
	t.text.Render(screen)

	if t.focused {
		t.setCaretColor()
		at := tr.Add(img.Point{cx, cy + t.scrollOffsetY})
		t.caret.SetLocation(at)
		t.caretRect = img.Rect(at.Min.X, at.Min.Y, at.Min.X+max(1, t.caret.Width), at.Min.Y+t.caret.Height)
		t.caret.Render(screen)
	}
}

// highlightLines draws the highlight under the runes from to to of the
// text shown, a band on each line they cover.
func (t *TextInput) highlightLines(screen *ebiten.Image, tr img.Rectangle, lines []wrappedLine, from, to int) {
	face := t.computedParams.Face
	first, firstCol := lineAt(lines, from)
	last, lastCol := lineAt(lines, to)
	lineH := t.wrap.lineH
	for i := first; i <= last; i++ {
		x0, x1 := 0, fontAdvanceAt(lines[i].text, len([]rune(lines[i].text)), face)
		if i == first {
			x0 = fontAdvanceAt(lines[i].text, firstCol, face)
		}
		if i == last {
			x1 = fontAdvanceAt(lines[i].text, lastCol, face)
		}
		if x1 <= x0 {
			continue
		}
		y0 := int(math.Round(float64(i)*lineH)) + t.scrollOffsetY
		h := int(math.Round(float64(i+1)*lineH)) + t.scrollOffsetY - y0
		t.computedParams.Image.Highlight.Draw(screen, x1-x0, h, func(opts *ebiten.DrawImageOptions) {
			opts.GeoM.Translate(float64(tr.Min.X+x0), float64(tr.Min.Y+y0))
		})
	}
}
