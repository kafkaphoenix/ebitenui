package widget

import (
	img "image"
	"image/color"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/matryer/is"
)

// A text tooltip given no face, color or background takes its theme's
// text face and color and panel background; given ones win.
func TestTextToolTip_ThemeFallback(t *testing.T) {
	is := is.New(t)

	ns := newNineSliceEmpty(t)
	theme := &Theme{TextTheme: &TextParams{Face: loadFont(t), Color: color.White}, PanelTheme: &PanelParams{BackgroundImage: ns}}
	looks := func(tip *ToolTip) (*Text, any) {
		tip.content.GetWidget().SetTheme(theme)
		tip.content.Validate()
		var background any // none unless drawn on a themed panel
		if p, ok := tip.content.(*themedPanel); ok {
			background = p.computedParams.BackgroundImage
		}
		return tip.content.Children()[0].(*Text), background
	}

	txt, background := looks(NewTextToolTip("tip", nil, nil, nil))
	is.True(txt.computedParams.Face == theme.TextTheme.Face)
	is.Equal(txt.computedParams.Color, color.White)
	is.True(background == ns)

	shared := *loadFont(t)
	var face text.Face = &text.GoTextFace{Source: shared.(*text.GoTextFace).Source, Size: 30}
	own := newNineSliceEmpty(t)
	txt, background = looks(NewTextToolTip("tip", &face, color.Black, own))
	is.True(txt.computedParams.Face == &face)
	is.Equal(txt.computedParams.Color, color.Black)
	is.True(background == own)
}

// A list combo button's own widget gets its updates: a tooltip on it shows
// while the pointer rests on the button.
func TestListComboButton_ToolTip(t *testing.T) {
	is := is.New(t)

	d := uitest.Use(t)
	input.SetupInputLayersWithDeferred(nil) // drops layers other tests left
	tip := NewToolTip(ToolTipOpts.Content(NewContainer()))
	l := newListComboButton(t, ListComboButtonOpts.Entries([]any{"a"}), ListComboButtonOpts.EntryLabelFunc(label, label),
		ListComboButtonOpts.WidgetOpts(WidgetOpts.ToolTip(tip)))
	l.SetLocation(img.Rect(0, 0, 100, 30))
	shown := 0
	l.GetWidget().ToolTipEvent.AddHandler(func(a any) {
		if a.(*WidgetToolTipEventArgs).Show {
			shown++
		}
	})

	point(d, 50, 15, false)
	for range 2 {
		l.Update(&UpdateObject{})
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
	input.SetupInputLayersWithDeferred(nil) // drops layers other tests left
	tip := NewToolTip(ToolTipOpts.Content(NewContainer()), ToolTipOpts.Delay(500*time.Millisecond))
	l := newListComboButton(t, ListComboButtonOpts.Entries([]any{"a"}), ListComboButtonOpts.EntryLabelFunc(label, label),
		ListComboButtonOpts.WidgetOpts(WidgetOpts.ToolTip(tip)))
	l.SetLocation(img.Rect(0, 0, 100, 30))
	shown := 0
	l.GetWidget().ToolTipEvent.AddHandler(func(a any) {
		if a.(*WidgetToolTipEventArgs).Show {
			shown++
		}
	})
	update := func(n int) int {
		for range n {
			l.Update(&UpdateObject{})
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
