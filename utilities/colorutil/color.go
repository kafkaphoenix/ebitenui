package colorutil

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

func HexToColor(h string) (color.Color, error) {
	h, _ = strings.CutPrefix(h, "#")
	u, err := strconv.ParseUint(h, 16, 0)
	if err != nil {
		return nil, fmt.Errorf("colorutil: %q is not a hex color: %w", h, err)
	}

	return color.NRGBA{
		R: uint8(u & 0xff0000 >> 16),
		G: uint8(u & 0xff00 >> 8),
		B: uint8(u & 0xff),
		A: uint8(255),
	}, nil
}
