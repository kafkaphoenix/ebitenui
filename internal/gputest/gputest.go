// Package gputest runs tests and benchmarks inside Ebitengine's game loop,
// where drawing reaches the GPU and pixels can be read back. Its own test
// files are the library's: what is drawn checked pixel by pixel, and what
// drawing costs a frame. They are built with the gpu tag and need a
// display: run them with `go test -tags gpu ./internal/gputest/`. They sit
// apart from each package's tests, which a running game loop would give
// real input and real ticks.
package gputest

import (
	"errors"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// The window the tests run in.
const (
	width  = 640
	height = 480
)

// frame is work for the next frame's Draw, and the channel it closes once
// that frame is done.
type frame struct {
	draw func(screen *ebiten.Image)
	done chan struct{}
}

var frames = make(chan frame)

// Main runs the tests in a small window, with vsync off and one tick a frame
// so frames come as fast as they are drawn, and exits with their result.
// Call it from TestMain.
func Main(m *testing.M) {
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("gputest")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	finished := make(chan int, 1)
	go func() { finished <- m.Run() }()
	g := &game{finished: finished}
	if err := ebiten.RunGameWithOptions(g, &ebiten.RunGameOptions{InitUnfocused: true}); err != nil && !errors.Is(err, ebiten.Termination) {
		panic(err)
	}
	os.Exit(g.code)
}

// Frame runs draw in the next frame's Draw, with the screen to draw on, and
// returns once that frame is done: its drawing has gone to the GPU, and
// pixels read in draw are there. draw runs on the game loop's goroutine, so
// it reports back by its results; t.Fatal belongs outside it.
func Frame(draw func(screen *ebiten.Image)) {
	f := frame{draw: draw, done: make(chan struct{})}
	frames <- f
	<-f.done
}

// game takes each frame's work from Frame and stops when the tests finish.
type game struct {
	finished <-chan int
	code     int
	drawn    *frame // drawn in the last frame, done once the next one starts
	next     *frame // to draw in this frame
}

func (g *game) Update() error {
	if g.drawn != nil {
		close(g.drawn.done)
		g.drawn = nil
	}
	select {
	case code := <-g.finished:
		g.code = code
		return ebiten.Termination
	case f := <-frames:
		g.next = &f
	default:
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.next != nil {
		g.next.draw(screen)
		g.drawn, g.next = g.next, nil
	}
}

func (*game) Layout(int, int) (int, int) { return width, height }
