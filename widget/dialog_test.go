package widget_test

import (
	img "image"
	"image/color"
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/matryer/is"
)

func newDialog(t *testing.T, opts ...widget.DialogOpt) (*widget.Dialog, *[]widget.DialogChoiceEventArgs, *bool) {
	t.Helper()

	var choices []widget.DialogChoiceEventArgs
	closed := false
	ns := newNineSliceEmpty(t)
	d := widget.NewDialog(append([]widget.DialogOpt{
		widget.DialogOpts.Title("Dust storm"),
		widget.DialogOpts.Message("A storm is coming over the landing site. Recall the rover?"),
		widget.DialogOpts.Choices("Recall", "Keep going"),
		widget.DialogOpts.MaxWidth(200),
		widget.DialogOpts.TitleText(loadFont(t), color.White),
		widget.DialogOpts.MessageText(loadFont(t), color.White),
		widget.DialogOpts.ButtonParams(&widget.ButtonParams{
			Image:     &widget.ButtonImage{Idle: ns, Pressed: ns},
			TextFace:  loadFont(t),
			TextColor: &widget.ButtonTextColor{Idle: color.White},
		}),
		widget.DialogOpts.ChoiceHandler(func(args *widget.DialogChoiceEventArgs) { choices = append(choices, *args) }),
	}, opts...)...)
	d.Window().SetCloseFunction(func() { closed = true })
	return d, &choices, &closed
}

// The dialog holds its title, its message wrapped at MaxWidth, and a
// button per choice under them, in order.
func TestDialog_Contents(t *testing.T) {
	is := is.New(t)

	d, _, _ := newDialog(t)
	render(d.Contents(), t)

	is.Equal(len(d.Contents().ChildrenForTest()), 4)
	message, ok := d.Contents().ChildrenForTest()[1].(*widget.Text)
	is.True(ok)
	is.True(len(message.Measurements().ProcessedLines()) > 1) // wrapped
	_, ok = d.Contents().ChildrenForTest()[3].(*widget.Button)
	is.True(ok)
}

// A choice is reported with its index and label, and closes the dialog.
func TestDialog_Choice(t *testing.T) {
	is := is.New(t)

	d, choices, closed := newDialog(t)
	render(d.Contents(), t)
	leftMouseButtonClick(d.Contents().ChildrenForTest()[3].(*widget.Button), t)

	is.Equal(len(*choices), 1)
	is.Equal((*choices)[0].Choice, 1)
	is.Equal((*choices)[0].Label, "Keep going")
	is.True(*closed)
}

// With KeepOpen, a choice leaves the dialog open.
func TestDialog_KeepOpen(t *testing.T) {
	is := is.New(t)

	d, choices, closed := newDialog(t, widget.DialogOpts.KeepOpen())
	render(d.Contents(), t)
	leftMouseButtonClick(d.Contents().ChildrenForTest()[2].(*widget.Button), t)

	is.Equal(len(*choices), 1)
	is.True(!*closed)
}

// A dialog's window takes the options given through WindowOpts: with a
// transition, a choice starts it on its way out.
func TestDialog_WindowOpts(t *testing.T) {
	is := is.New(t)

	d, _, closed := newDialog(t, widget.DialogOpts.WindowOpts(
		widget.WindowOpts.Transition(widget.WindowTransition{Duration: 4, Out: widget.WindowMotion{Offset: img.Pt(0, 20)}}),
	))
	render(d.Contents(), t)
	leftMouseButtonClick(d.Contents().ChildrenForTest()[2].(*widget.Button), t)

	is.True(d.Window().Closing()) // easing out
	is.True(!*closed)             // not yet removed
}

// CenterIn places the dialog at its preferred size in the middle of an area.
func TestDialog_CenterIn(t *testing.T) {
	is := is.New(t)

	d, _, _ := newDialog(t)
	d.CenterIn(img.Rect(100, 50, 500, 450))

	w, h := d.Contents().PreferredSize()
	r := d.Window().GetContainer().GetWidget().Rect
	is.Equal(r.Size(), img.Pt(w, h))
	is.Equal(r.Min.Add(r.Max), img.Pt(600, 500)) // centered: twice the area's middle
}

// Without a background of its own, the dialog draws its theme's panel
// background and follows the theme; given one, it keeps it.
func TestDialog_ThemeBackground(t *testing.T) {
	is := is.New(t)

	dark := &widget.Theme{PanelTheme: &widget.PanelParams{BackgroundImage: newNineSliceEmpty(t)}}
	light := &widget.Theme{PanelTheme: &widget.PanelParams{BackgroundImage: newNineSliceEmpty(t)}}
	background := func(d *widget.Dialog, theme *widget.Theme) any { // validated through its window, as the UI does
		w := d.Window().GetContainer()
		w.GetWidget().SetTheme(theme)
		w.Validate()
		return d.Contents().ComputedParams().BackgroundImage
	}

	d, _, _ := newDialog(t)
	is.True(background(d, dark) == dark.PanelTheme.BackgroundImage)
	is.True(background(d, light) == light.PanelTheme.BackgroundImage)

	own := newNineSliceEmpty(t)
	d, _, _ = newDialog(t, widget.DialogOpts.BackgroundImage(own))
	is.True(background(d, dark) == own)
}
