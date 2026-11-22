package widget_test

import (
	img "image"
	"image/color"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/matryer/is"
)

// A text tooltip given no face, color or background takes its theme's
// text face and color and panel background; given ones win.
func TestTextToolTip_ThemeFallback(t *testing.T) {
	is := is.New(t)

	ns := newNineSliceEmpty(t)
	theme := &widget.Theme{TextTheme: &widget.TextParams{Face: loadFont(t), Color: color.White}, PanelTheme: &widget.PanelParams{BackgroundImage: ns}}
	looks := func(tip *widget.ToolTip) (*widget.Text, any) {
		tip.Content().GetWidget().SetTheme(theme)
		tip.Content().Validate()
		var background any // none unless drawn on a themed panel
		if p, ok := tip.Content().(*widget.ThemedPanel); ok {
			background = p.ComputedParams().BackgroundImage
		}
		return tip.Content().Children()[0].(*widget.Text), background
	}

	txt, background := looks(widget.NewTextToolTip("tip", nil, nil, nil))
	is.True(txt.ComputedParams().Face == theme.TextTheme.Face)
	is.Equal(txt.ComputedParams().Color, color.White)
	is.True(background == ns)

	shared := *loadFont(t)
	var face text.Face = &text.GoTextFace{Source: shared.(*text.GoTextFace).Source, Size: 30}
	own := newNineSliceEmpty(t)
	txt, background = looks(widget.NewTextToolTip("tip", &face, color.Black, own))
	is.True(txt.ComputedParams().Face == &face)
	is.Equal(txt.ComputedParams().Color, color.Black)
	is.True(background == own)
}

// A list combo button's own widget gets its updates: a tooltip on it shows
// while the pointer rests on the button.
func TestListComboButton_ToolTip(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	tip := widget.NewToolTip(widget.ToolTipOpts.Content(widget.NewContainer()))
	l := newListComboButton(t, widget.ListComboButtonOpts.Entries([]any{"a"}), widget.ListComboButtonOpts.EntryLabelFunc(label, label),
		widget.ListComboButtonOpts.WidgetOpts(widget.WidgetOpts.ToolTip(tip)))
	l.SetLocation(img.Rect(0, 0, 100, 30))
	shown := 0
	l.GetWidget().ToolTipEvent.AddHandler(func(a any) {
		if a.(*widget.ToolTipEventArgs).Show {
			shown++
		}
	})

	point(d, 50, 15, false)
	for range 2 {
		l.Update(&widget.UpdateObject{})
		event.ExecuteDeferred()
	}
	is.Equal(shown, 1)
}

// A tooltip with a delay shows once the pointer has rested on its widget
// that long, counted in updates: 500 ms is 30 at 60 a second, between the
// update that sees the pointer arrive and the one that shows it. Moving
// off starts it over.
func TestToolTip_DelayInUpdates(t *testing.T) {
	d := uitest.Use(t)
	tip := widget.NewToolTip(widget.ToolTipOpts.Content(widget.NewContainer()), widget.ToolTipOpts.Delay(500*time.Millisecond))
	l := newListComboButton(t, widget.ListComboButtonOpts.Entries([]any{"a"}), widget.ListComboButtonOpts.EntryLabelFunc(label, label),
		widget.ListComboButtonOpts.WidgetOpts(widget.WidgetOpts.ToolTip(tip)))
	l.SetLocation(img.Rect(0, 0, 100, 30))
	shown := 0
	l.GetWidget().ToolTipEvent.AddHandler(func(a any) {
		if a.(*widget.ToolTipEventArgs).Show {
			shown++
		}
	})
	update := func(n int) int {
		for range n {
			l.Update(&widget.UpdateObject{})
			event.ExecuteDeferred()
		}
		return shown
	}

	point(d, 50, 15, false)
	if got := update(31); got != 0 {
		t.Fatalf("shown %d times in the 31 updates since the pointer came, want none yet", got)
	}
	if got := update(1); got != 1 {
		t.Fatalf("shown %d times in the 32 updates since the pointer came, want once", got)
	}
	point(d, 500, 500, false)
	update(1)
	point(d, 50, 15, false)
	if got := update(31); got != 1 {
		t.Errorf("shown %d times in the 31 updates since the pointer came back, want no new show yet", got)
	}
}
