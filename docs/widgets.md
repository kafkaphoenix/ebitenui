# Widgets

This page lists every widget Ebiten UI offers, says what each is for, and names the demo that shows it. It also explains how to give widgets your own look, one widget at a time or all at once with a theme.

## Trying the demos

Every widget has a small program showing it in `_examples/widget_demos`. From the `_examples` folder, run one with:

```
go run ./widget_demos/<folder>
```

The folder for each widget is named below, after "Demo". `go run ./demo` shows many widgets together.

## Making a widget

Each widget is made with a `New` function and options, and goes into a container, which lays it out:

```go
button := widget.NewButton(
	widget.ButtonOpts.Text("Play", face, &widget.ButtonTextColor{Idle: color.White}),
	widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { startGame() }),
)
root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
root.AddChild(button)
ui := &ebitenui.UI{Container: root, PrimaryTheme: themes.GetBasicDarkTheme()}
```

The game calls `ui.Update()` from its `Update` and `ui.Draw(screen)` from its `Draw`. The package documentation lists every option.

## Giving widgets your own look

A widget takes its look from three places, in this order:

1. **Its own options**, such as `ButtonOpts.Image` or `CarouselOpts.ArrowGraphics`. These win over everything else, so one widget can look different from the rest.
2. **The theme** the interface uses (`UI.PrimaryTheme`), for whatever the widget's options leave out. One theme styles every widget the same way.
3. **Defaults**, which a few widgets have for anything neither gives: the loading indicator, the carousel and the canvas draw plain white parts. Others need a theme or options, and say what is missing when they are first laid out.

### Images

Backgrounds, borders and button faces are nine-slices: images that stretch to any size while keeping their corners and edges as they are.

- `image.NewNineSliceColor(c)` is a plain color.
- `image.NewBorderedNineSliceColor(fill, border, width)` adds a border.
- `image.NewNineSliceRounded(fill, radius, border, width)` rounds the corners.
- `image.NewNineSlice(img, widths, heights)` cuts an image of your own into nine parts: the corners stay, the edges and the middle stretch.

Icons are `GraphicImage`s: an image for each state a widget can be in (`Idle`, `Hover`, `Pressed`, `Disabled`). Buttons, graphics and the carousel's arrows take them. Where a widget draws a small icon or marker itself, an option replaces it with your image: the carousel's arrows and dots, the loading spinner's dots or a spinning icon of your own, the color picker's marker and swatch frames.

### Your own theme

A theme is a `widget.Theme`: one set of looks for each kind of widget (`ButtonTheme`, `ListTheme`, `CarouselTheme`, and so on), and a default face and text color. The simplest way to make your own is to start from a basic theme and change what you want:

```go
theme := themes.GetBasicDarkTheme() // a new theme on every call: yours to change
theme.ButtonTheme.Image = &widget.ButtonImage{
	Idle:    image.NewNineSlice(buttonIdle, [3]int{8, 32, 8}, [3]int{8, 16, 8}),
	Hover:   image.NewNineSlice(buttonHover, [3]int{8, 32, 8}, [3]int{8, 16, 8}),
	Pressed: image.NewNineSlice(buttonPressed, [3]int{8, 32, 8}, [3]int{8, 16, 8}),
}
theme.CarouselTheme.PrevGraphic = &widget.GraphicImage{Idle: leftIcon}
theme.CarouselTheme.NextGraphic = &widget.GraphicImage{Idle: rightIcon}
ui := &ebitenui.UI{Container: root, PrimaryTheme: theme}
```

- `themes.GetBasicDarkTheme()` and `GetBasicLightTheme()` give the two basic themes.
- `GetBasicDarkThemeWith(themes.Rounded(6), themes.Borders(2))` gives them with rounded corners and borders.
- To follow the system's light or dark mode, give the UI both themes: `ui.DarkTheme` and `ui.LightTheme`. It switches its `PrimaryTheme` between them as the mode changes, checking twice a second.
- A theme can also be written from nothing, as a `&widget.Theme{...}` with the parts your interface uses.

Change a theme before building the interface. To switch themes later, such as from a settings menu, set `UI.PrimaryTheme` to a different theme; the interface takes it on its next update. Windows take the interface's theme unless given their own. One widget can have a theme of its own with `GetWidget().SetTheme(theme)`.

Demo: `theming`.

## The interface

**`ebitenui.UI`** is the root of an interface. Besides the root container, it holds the windows, the theme, and which widget has the focus. It also turns keys and gamepads into navigation (`GamepadFocus`, `BackHandler`).

A few of its methods help with special cases:
- `Changed` tells a game that draws only frames showing something new when widgets changed the picture by themselves.
- `Layout(width, height)` lays the interface out without drawing it, so the game can place things by its widgets before the first frame.
- `ScreenTransform` makes the pointer work on an interface drawn scaled or moved, such as a pixel-art interface drawn small and enlarged.

Demos: `demo`, `gamepad_navigation`, `touch`.

**`Container`** holds widgets and lays them out, on an optional background. **`Panel`** is a container drawn on the theme's panel background. Demo: `basic_container`.

## Layouts

A container's layout decides where its widgets go. Each widget can say how it wants to be placed with `WidgetOpts.LayoutData`.

- **`RowLayout`** puts widgets in a row or a column, with spacing between them. A widget can stretch across or be held to a maximum size. Demo: `rowlayout`.
- **`GridLayout`** puts widgets in columns and rows. Each column or row can keep its size, stretch, or take a share of the space by weight. Demo: `gridlayout`.
- **`AnchorLayout`** pins widgets to a corner, an edge or the middle, optionally stretched. Demo: `anchorlayout`.
- **`StackedLayout`** stacks widgets on top of each other, all as big as the biggest. Demo: `stackedlayout`.

## Text and images

- **`Text`** shows text in a face and a color, wrapped at a width if you give one. It can show colored words and clickable links written in BBCode, and a shadow. With `widget.SetTextCache(true)`, texts are drawn from images of themselves, which is faster for text that rarely changes. A label your game rebuilds each frame, such as a clock or a readout, can be built in a byte buffer you keep and set with `SetLabelBytes`, which makes a new string only when the text changes. Demos: `label`, `text_cache`.
- **`Label`** is a text with its own color for when it is disabled, as buttons and check boxes use. Demo: `label`.
- **`Graphic`** shows an image, a stretchable nine-slice, an animated GIF, or an image for each state (idle, hovered, pressed, disabled). Demos: `graphic_gif`, `imagebutton`.
- **`TextArea`** shows text too long for its space, scrolled with sliders, the wheel or by dragging. With `MaxLines` it keeps only its last lines, like a log. Demo: `textarea`.

## Buttons and choices

- **`Button`** is pressed with the mouse, a finger, Enter or Space, or a gamepad's A button. It can show text, an icon or both, and can stay pressed as a toggle. Demos: `button`, `button_with_graphic`, `button_with_mask`.
- **`Checkbox`** is checked or unchecked, optionally with a third, greyed state, and a label. Demos: `checkbox`, `labeled_checkbox`.
- **`RadioGroup`** makes a set of buttons or check boxes where only one is chosen at a time. It is a helper rather than a widget you place. Demo: `radiogroup`.
- **`Slider`** picks a number between a minimum and a maximum, by dragging its handle, with the keys, the wheel or clicks on its track. It can be horizontal or vertical, and inverted. Demo: `slider`.
- **`ColorPicker`** picks a color on a wheel of hues with a brightness slider, or on a square with a hue slider, and can offer preset swatches. Demo: `color_picker`.

## Typing

- **`TextInput`** is a line of text the player types. It supports selecting with the mouse, Shift and a double click, a placeholder, checking what is typed, hidden text for passwords, and the platform's input method for languages like Chinese or Japanese. With `Multiline` it takes several lines. Demos: `textinput`, `cut_copy_paste`.
- **`SearchBar`** is a field to search with: a magnifier, the text, and a button to clear it. It tells you the search once the player pauses typing, so a list or a table filters once per search, and Back clears it. Demo: `table`.
- **`OnScreenKeyboard`** is a keyboard of buttons for typing with a gamepad on a platform without a keyboard of its own; the interface opens it for a focused text input. Demo: `gamepad_navigation`.

## Lists and choosers

- **`List`** shows entries to choose from, scrolled. It only keeps widgets for the rows in view, so it handles thousands of entries. Entries can be text, rows of any widget all the same height, or a widget each of its own height. A filter shows only the entries it keeps, as a search does (`SetFilter`). Demos: `list`, `long_list`.
- **`Table`** shows rows of data in columns under a header, like a spreadsheet, and handles thousands of rows. Each column formats its cells your way, such as numbers in scientific notation or dates. A click on a column's header sorts by it, again the other way, and Shift and a click sorts by one more column; dragging a header's edge resizes a column. A filter, fed by a search bar, shows only the rows it keeps. With a rank column and a highlighted row it makes a leaderboard. Demo: `table`.
- **`ListComboButton`** shows the chosen entry on a button that opens a list to choose another. Demo: `combobox`.
- **`ComboButton`** and **`SelectComboButton`** are buttons that open any content under them. Demo: `combobox`.

## Pages and scrolling

- **`ScrollContainer`** shows content bigger than its space, scrolled with sliders, the wheel or by dragging. Demo: `scrollcontainer`.
- **`FlipBook`** shows one page of several at a time; the game switches between them. Demo: `flipbook`.
- **`TabBook`** shows pages behind a row of tabs. Demo: `tabbook`.
- **`Carousel`** shows one page at a time with arrows, dots and swipes to move between them, sliding each new page in. It can loop and move on by itself, like a loading screen. In steps mode, each page opens only once the game marks the steps before it as done, like a tutorial. Demo: `carousel`.

## Windows and popups

- **`Window`** shows content over the interface. It can block what is under it, have a title bar, be dragged and resized, close on a click outside, and fade in and out. Demo: `window`.
- **`Dialog`** is a window asking the player to choose, such as an alert or a yes/no question. Demo: `dialog`.
- **Context menus** open on a widget with a right click, a long press, the context-menu key or a gamepad's Menu button. Demo: `contextmenu`.
- **`ToolTip`** shows content while the pointer rests on a widget, or on a long press with a finger. Demo: `tooltips`.
- **`DragAndDrop`** lets the player drag something from one widget and drop it on another. Demo: `draganddrop`.

## Progress

- **`ProgressBar`** shows progress that is known, such as a download's percentage. Demo: `progressbar`.
- **`Loading`** shows that something is loading for an unknown time: a bar with a sweeping band, or a spinner, with optional tips under it. Demo: `loading`.
- **`Checklist`** shows things to do, each with a mark for its state: pending, done or failed. Use it for tasks, a mission's objectives or a tutorial's steps. The game sets each item's state as it happens, and the player ticks the items that are tickable. Done items fade and can be struck through, an item can show a count such as "3/5", and a title line can show the progress in your words. In sequential mode the current step is highlighted, and the steps after it fade, or stay hidden until it is reached. Demo: `checklist`.

## Your own drawing

- **`Canvas`** is a widget your game draws into, such as a minimap or a chart. It is laid out like any widget, what you draw stays inside it, and its events tell you where the pointer is. Demo: `canvas`.

## What every widget has

- **Focus and navigation**: Tab, the arrow keys and gamepads move the focus between widgets. The focused widget acts on the keys it understands, such as a slider on left and right. Demos: `gamepad_navigation`, `manual_focus`.
- **Events**: every widget can tell you when the pointer enters, moves over, presses, clicks, double clicks, long presses or scrolls on it, through `GetWidget()`.
- **Showing, hiding and disabling** (`SetVisibility`, `Disabled`), and **cursor shapes** over a widget. Demos: `visibility`, `cursor_shape`, `cursor_updater`.
- **Testing**: the `uitest` package plays a player in tests, clicking, typing and pressing gamepad buttons where a player would; see the README.
