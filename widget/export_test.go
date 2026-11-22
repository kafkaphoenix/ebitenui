package widget

import (
	image2 "image"
	"testing"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// What the external tests reach inside the package.

var AllocTextSlot = allocTextSlot

type ColorSwatch = colorSwatch

var FontAdvanceAt = fontAdvanceAt
var FreeTextSlot = freeTextSlot
var HsvToRGB = hsvToRGB
var HueTrack = hueTrack
var JoinPieces = joinPieces
var RgbToHSV = rgbToHSV
var RingPixels = ringPixels
var SweepTextSlots = sweepTextSlots

type TextPage = textPage

const TextPageMaxSlot = textPageMaxSlot

var WordBounds = wordBounds

const TextPageSize = textPageSize

type TextSlot = textSlot
type ThemedPanel = themedPanel

func (x *Button) Hovering() bool                                        { return x.hovering }
func (x *Button) TextLabel() string                                     { return x.textLabel }
func (x *Caret) Visible() bool                                          { return x.visible }
func (x *Checkbox) CurrentImage() *image.NineSlice                      { return x.currentImage() }
func (x *ColorPicker) Container() *Container                            { return x.container }
func (x *ColorPicker) Slider() *Slider                                  { return x.slider }
func (x *ColorPicker) Square() *colorSquare                             { return x.square }
func (x *ColorPicker) SwatchFrames() [2][2]*image.NineSlice             { return x.swatchFrames }
func (x *ColorPicker) Wheel() *colorWheel                               { return x.wheel }
func (x *ComboButton) Content() HasWidget                               { return x.content }
func (x *Container) ChildrenForTest() []PreferredSizeLocateableWidget   { return x.children }
func (x *Container) EffectBuffer() *ebiten.Image                        { return x.effectBuffer }
func (x *Container) Widget() *Widget                                    { return x.widget }
func (x *Dialog) Contents() *themedPanel                                { return x.contents }
func (x *Graphic) GifCurrentImage() int                                 { return x.gifCurrentImage }
func (x *Graphic) ImageForTest() *ebiten.Image                          { return x.image() }
func (x *Label) Text() *Text                                            { return x.text }
func (x *List) ComputedParams() *ListParams                             { return &x.computedParams }
func (x *List) DefinedParams() *ListParams                              { return &x.definedParams }
func (x *List) FocusIndex() int                                         { return x.focusIndex }
func (x *List) HSlider() *Slider                                        { return x.hSlider }
func (x *List) Rows() *listRows                                         { return x.rows }
func (x *List) ScrollContainer() *ScrollContainer                       { return x.scrollContainer }
func (x *List) SetScrollTop(a0 float64)                                 { x.setScrollTop(a0) }
func (x *List) VSlider() *Slider                                        { return x.vSlider }
func (x *ListComboButton) Button() *SelectComboButton                   { return x.button }
func (x *ListComboButton) DefinedParams() *ListComboButtonParams        { return &x.definedParams }
func (x *ListComboButton) List() *List                                  { return x.list }
func (x *ListParams) EntrySelectedColor() *ButtonImage                  { return x.entrySelectedColor }
func (x *ListParams) EntryUnselectedColor() *ButtonImage                { return x.entryUnselectedColor }
func (x *OnScreenKeyboard) Container() *themedPanel                     { return x.container }
func (x *OnScreenKeyboard) Letters() []*Button                          { return x.letters }
func (x *OnScreenKeyboard) Preview() *Text                              { return x.preview }
func (x *Panel) EffectShaderOp() *ebiten.DrawRectShaderOptions          { return &x.effectShaderOp }
func (x *ScrollContainer) Image() *ScrollContainerImage                 { return x.image }
func (x *SelectComboButton) Button() *ComboButton                       { return x.button }
func (x *TabBook) BtnContainer() *Container                             { return x.btnContainer }
func (x *TabBook) FlipBook() *FlipBook                                  { return x.flipBook }
func (x *TabBook) TabToButton() map[*TabBookTab]*Button                 { return x.tabToButton }
func (x *TabBook) Tabs() []*TabBookTab                                  { return x.tabs }
func (x *Text) Cache() *textCache                                       { return x.cache }
func (x *Text) ComputedParams() *TextParams                             { return &x.computedParams }
func (x *Text) Glyphs() []text.Glyph                                    { return x.glyphs }
func (x *Text) Measurements() *textMeasurements                         { return &x.measurements }
func (x *TextArea) HSlider() *Slider                                    { return x.hSlider }
func (x *TextArea) ScrollContainer() *ScrollContainer                   { return x.scrollContainer }
func (x *TextArea) VSlider() *Slider                                    { return x.vSlider }
func (x *TextInput) CaretRect() image2.Rectangle                        { return x.caretRect }
func (x *TextInput) CommitIME(a0 string, a1 bool, a2 string, a3 string) { x.commitIME(a0, a1, a2, a3) }
func (x *TextInput) Compose(a0 string, a1 int)                          { x.compose(a0, a1) }
func (x *TextInput) Composer() *textinput.Composer                      { return &x.composer }
func (x *TextInput) ComputedParams() *TextInputParams                   { return &x.computedParams }
func (x *TextInput) CursorPosition() int                                { return x.cursorPosition }
func (x *TextInput) ImeSession() *textinput.SessionOptions              { return x.imeSession() }
func (x *TextInput) StateImage() *image.NineSlice                       { return x.stateImage() }
func (x *TextInput) Text() *Text                                        { return x.text }
func (x *TextInput) Wrap()                                              { x.wrappedLines() }
func (x *TextInput) WrappedLines() []string {
	lines := x.wrappedLines()
	s := make([]string, 0, len(lines))
	for _, l := range lines {
		s = append(s, l.text)
	}
	return s
}
func (x *ToolTip) Content() Containerer                                 { return x.content }
func (x *Widget) LongPressDuration() int                                { return x.longPressDuration }
func (x *Window) Layer() *input.Layer                                   { return &x.layer }
func (x *bbCodeText) Glyphs() []text.Glyph                              { return x.glyphs }
func (x *bbCodeText) Text() string                                      { return x.text }
func (x *colorSquare) Widget() *Widget                                  { return x.widget }
func (x *colorSwatch) Chosen() bool                                     { return x.chosen() }
func (x *colorSwatch) Fill() *image.NineSlice                           { return x.fill }
func (x *colorSwatch) Widget() *Widget                                  { return x.widget }
func (x *colorWheel) Widget() *Widget                                   { return x.widget }
func (x *listLabelRow) Entry() int                                      { return x.entry }
func (x *listLabelRow) Text() *Text                                     { return x.text }
func (x *listRows) Background(a0 int) *image.NineSlice                  { return x.background(a0) }
func (x *listRows) Labels() []listLabelRow                              { return x.labels }
func (x *listRows) RowRect(a0 int) image2.Rectangle                     { return x.rowRect(a0) }
func (x *listRows) RowSpan(a0 int) (top int, bottom int)                { return x.rowSpan(a0) }
func (x *listRows) Widget() *Widget                                     { return x.widget }
func (x *listRows) Widgets() map[any]PreferredSizeLocateableWidget      { return x.widgets }
func (x *textCache) Image() *ebiten.Image                               { return x.image }
func (x *textCache) Own() *ebiten.Image                                 { return x.own }
func (x *textCache) Renders() int                                       { return x.renders }
func (x *textCache) Slot() *textSlot                                    { return x.slot }
func (x *textCache) TooLarge() bool                                     { return x.tooLarge }
func (x *textMeasurements) BoundingBoxWidth() float64                   { return x.boundingBoxWidth }
func (x *textMeasurements) LineHeight() float64                         { return x.lineHeight }
func (x *textMeasurements) ProcessedLineGlyphs() [][]text.Glyph         { return x.processedLineGlyphs }
func (x *textMeasurements) ProcessedLineTexts() []string                { return x.processedLineTexts }
func (x *textMeasurements) ProcessedLineWidths() []float64              { return x.processedLineWidths }
func (x *textMeasurements) ProcessedLines() [][]*bbCodeText             { return x.processedLines }
func (x *textMeasurements) ProcessedLinesLaid() []bool                  { return x.processedLinesLaid }
func (x *textPage) Alloc(a0 int, a1 int) *textSlot                      { return x.alloc(a0, a1) }
func (x *textPage) Free(a0 *textSlot)                                   { x.free(a0) }
func (x *textPage) Shelves() []textShelf                                { return x.shelves }
func (x *textPage) Used() int                                           { return x.used }
func (x *textSlot) Freed() bool                                         { return x.freed }
func (x *textSlot) Page() *textPage                                     { return x.page }
func (x *textSlot) Rect() image2.Rectangle                              { return x.rect }
func (x *themedPanel) ChildrenForTest() []PreferredSizeLocateableWidget { return x.children }
func (x *themedPanel) ComputedParams() *PanelParams                     { return &x.computedParams }

// EmptyTextPages gives a test a pool of text pages of its own.
func EmptyTextPages(t testing.TB) {
	t.Helper()
	pages, kept := textPages.pages, textPages.kept
	textPages.pages, textPages.kept = nil, nil
	t.Cleanup(func() { textPages.pages, textPages.kept = pages, kept })
}

// TextPageCount returns how many pages the text pool has.
func TextPageCount() int { return len(textPages.pages) }

// TextPageUsed returns how many slots page i gives out.
func TextPageUsed(i int) int { return textPages.pages[i].used }

// KeepTextSlot has the sweeps look at s, as at a slot a text keeps.
func KeepTextSlot(s *textSlot) { textPages.kept = append(textPages.kept, s) }

// SetPreviousLink sets the link a text last had under the pointer.
func (x *Text) SetPreviousLink(b *bbCodeText) { x.previousLink = b }

// SetCursorPosition puts the caret before rune i.
func (x *TextInput) SetCursorPosition(i int) { x.cursorPosition = i }

// SetDragStartIndex starts a selection at rune i.
func (x *TextInput) SetDragStartIndex(i int) { x.dragStartIndex = i }

// SetLastInputText sets the text the input last reported as changed.
func (x *TextInput) SetLastInputText(s string) { x.lastInputText = s }

// UseIME has text inputs reach update and end in place of the platform's
// text input, and returns what puts the platform's back.
func UseIME(update func(*textinput.Composer) (bool, error), end func(*textinput.Composer, bool)) (restore func()) {
	oldUpdate, oldEnd := imeUpdate, imeEnd
	imeUpdate, imeEnd = update, end
	return func() { imeUpdate, imeEnd = oldUpdate, oldEnd }
}

// Handle returns the slider's handle.
func (x *Slider) Handle() *Button { return x.handle }
