package colorutil_test

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/ebitenui/ebitenui/utilities/colorutil"
)

// Any text parses or reports an error, never panics, and a six-digit hex
// color it parses reads back as the same color, opaque.
func FuzzHexToColor(f *testing.F) {
	for _, s := range []string{"#ff8800", "00FF00", "#fff", "", "#", "#zz0000", "#1234567", "-1", "+ff"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		c, err := colorutil.HexToColor(s)
		if err != nil {
			return
		}
		n := c.(color.NRGBA)
		if n.A != 255 {
			t.Errorf("%q gave %v, not opaque", s, n)
		}
		if h := strings.TrimPrefix(s, "#"); len(h) == 6 {
			if got := fmt.Sprintf("%02x%02x%02x", n.R, n.G, n.B); got != strings.ToLower(h) {
				t.Errorf("%q gave %v, which reads %q", s, n, got)
			}
		}
	})
}
