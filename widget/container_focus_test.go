package widget_test

import (
	"slices"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/matryer/is"
)

// ownContainer is a caller's container type, embedding a Container.
type ownContainer struct{ *widget.Container }

// The focus search goes into every container: a Panel, or a type of the
// caller's that embeds a Container, as into the types it knows by name; and
// into a tab book's open tab.
func TestContainer_GetFocusersSearchesEveryContainer(t *testing.T) {
	is := is.New(t)

	panel := widget.NewPanel(widget.ContainerOpts.BackgroundImage(newNineSliceEmpty(t)))
	own := ownContainer{widget.NewContainer()}
	for _, c := range []struct {
		inner widget.PreferredSizeLocateableWidget
		add   func(...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	}{{panel, panel.AddChild}, {own, own.AddChild}} {
		b := newButton(t)
		c.add(b)
		outer := widget.NewContainer()
		outer.AddChild(c.inner)
		found := outer.GetFocusers()
		is.Equal(len(found), 1)
		is.True(found[0] == widget.Focuser(b))
	}

	b := newButton(t)
	tab := widget.NewTabBookTab(widget.TabBookTabOpts.Label("Tab"))
	tab.AddChild(b)
	outer := widget.NewContainer()
	outer.AddChild(newTabBook(t, widget.TabBookOpts.Tabs(tab)))
	is.True(slices.Contains(outer.GetFocusers(), widget.Focuser(b))) // in the open tab
}
