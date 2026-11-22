package main

import (
	"bytes"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// Game object used by ebiten
type game struct {
	ui *ebitenui.UI
}

var buttonImage *widget.ButtonImage
var face text.Face

/*
The Grid Layout is built to position children in a rows and columns.
*/
func main() {
	buttonImage, _ = loadButtonImage()
	face, _ = loadFont(12)
	// construct a new container that serves as the root of the UI hierarchy
	rootContainer := widget.NewContainer(
		// the container will use a plain color as its background
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(3),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(30)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true, true}, []bool{true, true, true}),
		)),
	)

	btn1 := createButton("Button 1")
	rootContainer.AddChild(btn1)

	btn2 := createButton("Button 2")
	rootContainer.AddChild(btn2)

	btn3 := createButton("Button 3")
	rootContainer.AddChild(btn3)

	btn4 := createButton("Button 4")
	rootContainer.AddChild(btn4)

	btn5 := createButton("Button 5")
	rootContainer.AddChild(btn5)

	btn6 := createButton("Button 6")
	rootContainer.AddChild(btn6)

	btn7 := createButton("Button 7")
	rootContainer.AddChild(btn7)

	btn8 := createButton("Button 8")
	rootContainer.AddChild(btn8)

	btn9 := createButton("Button 9")
	rootContainer.AddChild(btn9)

	// Links are optional: without one, ChangeFocus goes to the nearest
	// focusable widget in the direction asked. A link chooses the widget
	// itself, as these do for every direction.
	btn1.AddFocus(widget.FocusEast, btn2)
	btn1.AddFocus(widget.FocusSouth, btn4)
	btn1.AddFocus(widget.FocusSoutheast, btn5)

	btn2.AddFocus(widget.FocusWest, btn1)
	btn2.AddFocus(widget.FocusEast, btn3)
	btn2.AddFocus(widget.FocusSouth, btn5)
	btn2.AddFocus(widget.FocusSouthwest, btn4)
	btn2.AddFocus(widget.FocusSoutheast, btn6)

	btn3.AddFocus(widget.FocusWest, btn2)
	btn3.AddFocus(widget.FocusSouth, btn6)
	btn3.AddFocus(widget.FocusSouthwest, btn5)

	btn4.AddFocus(widget.FocusNorth, btn1)
	btn4.AddFocus(widget.FocusEast, btn5)
	btn4.AddFocus(widget.FocusSouth, btn7)
	btn4.AddFocus(widget.FocusNortheast, btn2)
	btn4.AddFocus(widget.FocusSoutheast, btn8)

	btn5.AddFocus(widget.FocusNorthwest, btn1)
	btn5.AddFocus(widget.FocusNorth, btn2)
	btn5.AddFocus(widget.FocusNortheast, btn3)
	btn5.AddFocus(widget.FocusWest, btn4)
	btn5.AddFocus(widget.FocusEast, btn6)
	btn5.AddFocus(widget.FocusSouthwest, btn7)
	btn5.AddFocus(widget.FocusSouth, btn8)
	btn5.AddFocus(widget.FocusSoutheast, btn9)

	btn6.AddFocus(widget.FocusNorth, btn3)
	btn6.AddFocus(widget.FocusNorthwest, btn2)
	btn6.AddFocus(widget.FocusWest, btn5)
	btn6.AddFocus(widget.FocusSouthwest, btn8)
	btn6.AddFocus(widget.FocusSouth, btn9)

	btn7.AddFocus(widget.FocusNorth, btn4)
	btn7.AddFocus(widget.FocusNortheast, btn5)
	btn7.AddFocus(widget.FocusEast, btn8)

	btn8.AddFocus(widget.FocusWest, btn7)
	btn8.AddFocus(widget.FocusNorthwest, btn4)
	btn8.AddFocus(widget.FocusNorth, btn5)
	btn8.AddFocus(widget.FocusNortheast, btn6)
	btn8.AddFocus(widget.FocusEast, btn9)

	btn9.AddFocus(widget.FocusWest, btn8)
	btn9.AddFocus(widget.FocusNorthwest, btn5)
	btn9.AddFocus(widget.FocusNorth, btn6)

	// construct the UI
	ui := ebitenui.UI{
		Container: rootContainer,
		// A gamepad's D-pad moves the focus too, and its bottom face button
		// presses the focused button.
		GamepadFocus: true,
	}

	ui.SetFocusedWidget(btn5)

	if testBtn, ok := ui.GetFocusedWidget().(*widget.Button); ok {
		testBtn.Click()
	}

	// Ebiten setup
	ebiten.SetWindowSize(400, 400)
	ebiten.SetWindowTitle("Ebiten UI - Manual Focus")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := game{
		ui: &ui,
	}

	// run Ebiten main loop
	err := ebiten.RunGame(&game)
	if err != nil {
		log.Println(err)
	}
}

func createButton(label string) *widget.Button {
	// construct a button
	return widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),

		// specify the images to use
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text(label, &face, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			println(label + ": button clicked")
		}),

		// Indicate that this button should not be submitted when enter or space are pressed
		// widget.ButtonOpts.DisableDefaultKeys(),
	)

}

// Layout implements Game.
func (g *game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update implements Game.
func (g *game) Update() error {
	// update the UI
	g.ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.ui.ChangeFocus(widget.FocusNorth)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.ui.ChangeFocus(widget.FocusWest)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.ui.ChangeFocus(widget.FocusSouth)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.ui.ChangeFocus(widget.FocusEast)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.ui.ChangeFocus(widget.FocusNorthwest)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.ui.ChangeFocus(widget.FocusNortheast)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.ui.ChangeFocus(widget.FocusSouthwest)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.ui.ChangeFocus(widget.FocusSoutheast)
	}
	return nil
}

// Draw implements Ebiten's Draw method.
func (g *game) Draw(screen *ebiten.Image) {
	// draw the UI onto the screen
	g.ui.Draw(screen)
}

func loadButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})

	hover := image.NewNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255})

	pressed := image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func loadFont(size float64) (text.Face, error) {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}, nil
}
