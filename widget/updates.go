package widget

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// updatesIn returns how many updates last d at Ebitengine's ticks per
// second, at least one: widgets count updates rather than read the clock,
// so what they time follows the game's ticks, pauses with them, and runs
// the same in every test. Ebitengine's default rate stands in when ticks
// follow the frames (ebiten.SyncWithFPS).
func updatesIn(d time.Duration) int {
	tps := ebiten.TPS()
	if tps <= 0 {
		tps = ebiten.DefaultTPS
	}
	return max(1, int(math.Round(d.Seconds()*float64(tps))))
}
