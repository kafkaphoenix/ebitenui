package widget_test

import (
	"image"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/uitest"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/matryer/is"
)

// showToast puts a toast with opts in the dark theme on a 640 by 480
// screen, at Ebitengine's 60 updates a second, and draws a frame.
func showToast(t *testing.T, opts ...widget.ToastOpt) (*widget.Toast, *ebitenui.UI, *uitest.Driver) {
	t.Helper()
	t.Cleanup(func() { widget.SetUITheme(nil) })
	toast := widget.NewToast(opts...)
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	root.AddChild(toast)
	ui := &ebitenui.UI{Container: root, PrimaryTheme: themes.GetBasicDarkTheme()}
	drv := uitest.New(t, 640, 480, ui.Update, ui.Draw)
	drv.Frames(1)
	return toast, ui, drv
}

// A toast is hidden until shown; shown, it shows its message for its
// duration, 100 ms being 6 updates, then hides. Shown again, the new
// message takes the old one's place and its time starts over.
func TestToast_ShowsForItsDuration(t *testing.T) {
	is := is.New(t)
	toast, _, drv := showToast(t, widget.ToastOpts.Duration(100*time.Millisecond))
	is.True(!toast.GetWidget().IsVisible()) // hidden
	is.Equal(toast.Message(), "")

	toast.Show("Saved")
	drv.Frames(1)
	is.True(toast.GetWidget().IsVisible())
	is.Equal(toast.Message(), "Saved")
	is.Equal(toast.Text().Label, "Saved")

	drv.Frames(4)
	toast.Show("Gamepad connected")
	drv.Frames(5)
	is.Equal(toast.Message(), "Gamepad connected") // still showing, its time started over
	drv.Frames(1)
	is.True(!toast.GetWidget().IsVisible()) // gone after its 6 updates
	is.Equal(toast.Message(), "")
}

// A toast shows for 3 seconds by default; Hide hides it at once, and
// hidden it counts nothing on: no change is marked when its time would
// have ended.
func TestToast_DefaultDurationAndHide(t *testing.T) {
	is := is.New(t)
	toast, ui, drv := showToast(t)
	toast.Show("Saved")
	drv.Frames(179)
	is.Equal(toast.Message(), "Saved")
	drv.Frames(1)
	is.Equal(toast.Message(), "") // 180 updates

	toast.Show("Saved")
	drv.Frames(1)
	toast.Hide()
	is.True(!toast.GetWidget().IsVisible())
	is.Equal(toast.Message(), "")
	drv.Frames(1)
	ui.Changed()
	for range 180 {
		drv.Frames(1)
		is.True(!ui.Changed()) // nothing counted on
	}
}

// The padding is the room around the message.
func TestToast_Padding(t *testing.T) {
	is := is.New(t)
	toast, _, drv := showToast(t, widget.ToastOpts.Padding(widget.NewInsetsSimple(10)))
	toast.Show("Saved")
	drv.Frames(1)
	box, msg := toast.GetWidget().Rect, toast.Text().GetWidget().Rect
	is.Equal(msg.Min.Sub(box.Min), image.Pt(10, 10)) // room before
	is.Equal(box.Max.Sub(msg.Max), image.Pt(10, 10)) // and after
}

// With a fade, a toast eases in from transparent and out to it at its end,
// marking each frame it fades as changed and none while it holds or after
// it hides.
func TestToast_Fades(t *testing.T) {
	is := is.New(t)
	toast, ui, drv := showToast(t, widget.ToastOpts.Duration(200*time.Millisecond), widget.ToastOpts.Fade(50*time.Millisecond))
	toast.Show("Saved")
	opacities := make([]float32, 0, 13)
	changed := make([]bool, 0, 13)
	ui.Changed()
	for range 13 {
		drv.Frames(1)
		opacities = append(opacities, toast.Container().Opacity())
		changed = append(changed, ui.Changed())
	}
	third := float32(1) / 3
	is.Equal(opacities[:3], []float32{third, 2 * third, 1}) // easing in over 3 updates
	is.Equal(opacities[3:9], []float32{1, 1, 1, 1, 1, 1})   // holding
	is.Equal(opacities[9:11], []float32{2 * third, third})  // easing out
	is.Equal(changed[:3], []bool{true, true, true})         // in
	is.Equal(changed[3:9], []bool{false, false, false, false, false, false})
	is.Equal(changed[9:12], []bool{true, true, true}) // out, then hidden
	is.True(!toast.GetWidget().IsVisible())
	is.True(!changed[12]) // nothing once hidden

	toast.Show("Saved")
	is.Equal(toast.Container().Opacity(), float32(0)) // shown again, from transparent at once
	drv.Frames(6)
	toast.Show("Gamepad connected")
	drv.Frames(1)
	is.Equal(toast.Container().Opacity(), float32(1)) // a new message in place shows at once
}
