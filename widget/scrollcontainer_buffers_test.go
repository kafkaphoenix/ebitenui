package widget_test

import (
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
)

// newMaskedScroll returns a scroll container over a tall content, 50 by 50.
func newMaskedScroll() *widget.ScrollContainer {
	content := widget.NewContainer(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(50, 200)))
	return widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: eimage.NewNineSliceColor(color.Transparent),
			Mask: eimage.NewNineSliceColor(color.White),
		}),
		widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(50, 50)),
	)
}

// newBufferUI returns a UI in the dark theme over a row of children, and a
// frame of it: an update and a draw. The input starts over for it, and the
// theme the UI sets for every widget goes as the test ends.
func newBufferUI(t *testing.T, children ...widget.PreferredSizeLocateableWidget) (*ebitenui.UI, *widget.Container, func()) {
	t.Helper()
	uitest.Use(t)
	t.Cleanup(func() { widget.SetUITheme(nil) })
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	root.AddChild(children...)
	ui := &ebitenui.UI{Container: root, PrimaryTheme: themes.GetBasicDarkTheme()}
	screen := ebiten.NewImage(1000, 300)
	return ui, root, func() {
		ui.Update()
		ui.Draw(screen)
	}
}

// A scroll container drawn keeps the images it draws its content into, and
// gives them back as it stops showing: hidden, its parent hidden, removed
// from its container, or in a window gone from its UI. Shown and drawn
// again, it makes them anew.
func TestBuffers_GivenBackWhenNotShown(t *testing.T) {
	is := is.New(t)

	sc := newMaskedScroll()
	pane := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	pane.AddChild(sc)
	ui, root, frame := newBufferUI(t, pane)
	frame()
	is.True(sc.RenderBuffer().Allocated())

	sc.GetWidget().SetVisibility(widget.VisibilityHideBlocking)
	is.True(!sc.RenderBuffer().Allocated()) // at once, not a frame later
	sc.GetWidget().SetVisibility(widget.VisibilityShow)
	frame()
	is.True(sc.RenderBuffer().Allocated())

	pane.GetWidget().SetVisibility(widget.VisibilityHide) // as a window's section shown another
	is.True(!sc.RenderBuffer().Allocated())
	pane.GetWidget().SetVisibility(widget.VisibilityShow)
	frame()
	root.RemoveChild(pane)
	is.True(!sc.RenderBuffer().Allocated())

	input := widget.NewTextInput(widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 30)))
	outer := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(input),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: eimage.NewNineSliceColor(color.Transparent), Mask: eimage.NewNineSliceColor(color.White),
		}), widget.ScrollContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 40)))
	root.AddChild(outer)
	frame()
	is.True(input.RenderBuffer().Allocated())
	outer.GetWidget().SetVisibility(widget.VisibilityHide) // what it scrolls too
	is.True(!input.RenderBuffer().Allocated())

	inWindow := newMaskedScroll()
	contents := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
	contents.AddChild(inWindow)
	window := widget.NewWindow(widget.WindowOpts.Contents(contents))
	ui.AddWindow(window)
	frame()
	is.True(inWindow.RenderBuffer().Allocated())
	window.Close()
	frame() // the UI removes it as it updates
	is.True(!inWindow.RenderBuffer().Allocated())
}

// The widgets holding scroll containers or text inputs pass it on: a list,
// a text area, a table, a search bar, a tab book's tab, a flip book's page
// and a carousel's page, each hidden.
func TestBuffers_HoldersPassItOn(t *testing.T) {
	is := is.New(t)

	face := loadFont(t)
	list := widget.NewList(widget.ListOpts.Entries([]any{"a", "b"}), widget.ListOpts.EntryLabelFunc(label),
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 60))))
	area := widget.NewTextArea(widget.TextAreaOpts.Text("a log"), widget.TextAreaOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 60))))
	table := widget.NewTable(widget.TableOpts.Columns(widget.TableColumn{Header: "N", Text: label}),
		widget.TableOpts.Rows([]any{"a", "b"}), widget.TableOpts.Face(face), widget.TableOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 60)))
	search := widget.NewSearchBar(widget.SearchBarOpts.WidgetOpts(widget.WidgetOpts.MinSize(200, 30)))
	inTab, inPage, inSlide := newMaskedScroll(), newMaskedScroll(), newMaskedScroll()
	tab := widget.NewTabBookTab(widget.TabBookTabOpts.Label("Tab"),
		widget.TabBookTabOpts.ContainerOpts(widget.ContainerOpts.Layout(widget.NewRowLayout())))
	tab.AddChild(inTab)
	book := widget.NewTabBook(widget.TabBookOpts.Tabs(tab))
	flip := widget.NewFlipBook()
	carousel := widget.NewCarousel(widget.CarouselOpts.Pages(inSlide), widget.CarouselOpts.WidgetOpts(widget.WidgetOpts.MinSize(80, 80)))
	_, _, frame := newBufferUI(t, list, area, table, search, book, flip, carousel)
	flip.SetPage(inPage)
	frame()
	frame()

	for _, h := range []struct {
		name   string
		holder widget.HasWidget
		buffer *eimage.MaskedRenderBuffer
	}{
		{"list", list, list.ScrollContainer().RenderBuffer()},
		{"text area", area, area.ScrollContainer().RenderBuffer()},
		{"table", table, table.List().ScrollContainer().RenderBuffer()},
		{"search bar", search, search.TextInput().RenderBuffer()},
		{"tab book", book, inTab.RenderBuffer()},
		{"flip book", flip, inPage.RenderBuffer()},
		{"carousel", carousel, inSlide.RenderBuffer()},
	} {
		if !h.buffer.Allocated() {
			t.Fatalf("%s: drawn, it holds no images", h.name)
		}
		h.holder.GetWidget().SetVisibility(widget.VisibilityHide)
		is.Equal([]any{h.name, h.buffer.Allocated()}, []any{h.name, false})
	}
}
