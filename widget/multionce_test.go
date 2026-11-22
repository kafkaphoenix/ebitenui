package widget_test

import (
	"testing"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/matryer/is"
)

func TestMultiOnce_Do(t *testing.T) {
	is := is.New(t)

	m := widget.MultiOnce{}

	count := 0
	f := func() {
		count++
	}

	m.Append(f)
	m.Append(f)

	m.Do()
	is.Equal(count, 2)

	m.Do()
	is.Equal(count, 2)
}
