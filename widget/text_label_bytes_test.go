package widget_test

import (
	"testing"

	"github.com/matryer/is"
)

// A label set from bytes reads them, and is laid out anew when they
// differ; set from the same bytes again, as a readout rebuilt each frame,
// it allocates nothing and keeps its layout.
func TestText_SetLabelBytes(t *testing.T) {
	is := is.New(t)
	tw := newText(t, "12:00")
	buf := []byte("12:01")
	tw.SetLabelBytes(buf)
	is.Equal(tw.Label, "12:01")
	render(tw, t)
	is.Equal(tw.Measurements().Label(), "12:01") // laid out anew

	buf[4] = '2' // the buffer is not kept
	is.Equal(tw.Label, "12:01")

	same := []byte("12:01")
	is.Equal(testing.AllocsPerRun(10, func() { tw.SetLabelBytes(same) }), 0.0)
	is.Equal(tw.Label, "12:01")
}
