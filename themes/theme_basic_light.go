package themes

import (
	img "image"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/constantutil"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func getLightCheckbox(s shape) *widget.CheckboxImage {
	border := s.upTo(checkboxRadius).box(color.White, color.Black, 2)
	hoverBorder := s.upTo(checkboxRadius).box(color.NRGBA{214, 242, 255, 255}, color.Black, 2)
	idle := ebiten.NewImage(32, 32)
	border.Draw(idle, 32, 32, nil)
	idle9s := image.NewFixedNineSlice(idle)

	idleHover := ebiten.NewImage(32, 32)
	hoverBorder.Draw(idleHover, 32, 32, nil)
	idleHover9s := image.NewFixedNineSlice(idleHover)
	// Create checked image
	checked := ebiten.NewImage(32, 32)
	border.Draw(checked, 32, 32, nil)

	var checkmark vector.Path
	checkmark.MoveTo(28, 7)
	checkmark.LineTo(14, 24)
	checkmark.LineTo(6, 15)

	strokeOp := &vector.StrokeOptions{
		Width:    3,
		LineJoin: vector.LineJoinBevel,
	}
	drawOp := &vector.DrawPathOptions{AntiAlias: true}
	drawOp.ColorScale.ScaleWithColor(color.Black)
	vector.StrokePath(checked, &checkmark, strokeOp, drawOp)
	checked9s := image.NewFixedNineSlice(checked)

	checkedHover := ebiten.NewImage(32, 32)
	hoverBorder.Draw(checkedHover, 32, 32, nil)
	vector.StrokePath(checkedHover, &checkmark, strokeOp, drawOp)
	checkedHover9s := image.NewFixedNineSlice(checkedHover)

	// Create greyed image
	greyed := ebiten.NewImage(32, 32)
	border.Draw(greyed, 32, 32, nil)
	vector.StrokeLine(greyed, 5, 16, 27, 16, 3, color.Black, true)
	greyed9s := image.NewFixedNineSlice(greyed)

	greyedHover := ebiten.NewImage(32, 32)
	hoverBorder.Draw(greyedHover, 32, 32, nil)
	vector.StrokeLine(greyedHover, 5, 16, 27, 16, 3, color.Black, true)
	greyedHover9s := image.NewFixedNineSlice(greyedHover)

	return &widget.CheckboxImage{
		Unchecked:         idle9s,
		Checked:           checked9s,
		Greyed:            greyed9s,
		UncheckedHovered:  idleHover9s,
		CheckedHovered:    checkedHover9s,
		GreyedHovered:     greyedHover9s,
		UncheckedDisabled: idle9s,
		CheckedDisabled:   checked9s,
		GreyedDisabled:    greyed9s,
	}
}

// GetBasicLightTheme returns the basic light theme: square boxes, its own
// borders.
func GetBasicLightTheme() *widget.Theme { return newBasicLightTheme(shape{}) }

// GetBasicLightThemeWith returns the basic light theme shaped by opts: Rounded
// corners, Borders of another width.
func GetBasicLightThemeWith(opts ...ThemeOpt) *widget.Theme {
	return newBasicLightTheme(newShape(opts))
}

// newBasicLightTheme returns the basic light theme, its boxes drawn as s
// shapes them.
func newBasicLightTheme(s shape) *widget.Theme {
	// load button text font
	face := loadFont(20)
	entryColor := &widget.ListEntryColor{
		Unselected:         color.Black,
		Selected:           color.Black,
		DisabledUnselected: color.NRGBA{127, 122, 126, 255},
		DisabledSelected:   color.NRGBA{127, 122, 126, 255},

		SelectedBackground:        color.NRGBA{204, 232, 255, 255},
		SelectedFocusedBackground: color.NRGBA{214, 242, 255, 255},

		SelectingBackground:        color.NRGBA{229, 243, 255, 255},
		FocusedBackground:          color.NRGBA{229, 243, 255, 255},
		SelectingFocusedBackground: color.NRGBA{229, 243, 255, 255},
		DisabledSelectedBackground: color.NRGBA{229, 243, 255, 255},
	}

	return &widget.Theme{
		DefaultFace:      &face,
		DefaultTextColor: color.Black,
		ButtonTheme: &widget.ButtonParams{
			TextColor: &widget.ButtonTextColor{
				Idle:    color.Black,
				Hover:   color.Black,
				Pressed: color.Black,
			},
			TextFace: &face,
			Image: &widget.ButtonImage{
				Idle:    s.box(color.NRGBA{233, 231, 231, 255}, color.NRGBA{223, 220, 220, 255}, 2),
				Hover:   s.box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{197, 192, 196, 255}, 2),
				Pressed: s.box(color.NRGBA{197, 192, 196, 255}, color.NRGBA{177, 172, 176, 255}, 2),
			},
			TextPadding: &widget.Insets{
				Left:   30,
				Right:  30,
				Top:    5,
				Bottom: 5,
			},
			TextPosition: &widget.TextPositioning{
				VTextPosition: widget.TextPositionCenter,
				HTextPosition: widget.TextPositionCenter,
			},
		},
		PanelTheme: &widget.PanelParams{
			BackgroundImage: s.fill(color.NRGBA{245, 246, 247, 255}),
		},
		LabelTheme: &widget.LabelParams{
			Face: &face,
			Color: &widget.LabelColor{
				Idle:     color.Black,
				Disabled: color.NRGBA{122, 122, 122, 255},
			},
		},
		TextTheme: &widget.TextParams{
			Face:  &face,
			Color: color.Black,
		},
		TabbookTheme: &widget.TabBookParams{
			TabButton: &widget.ButtonParams{
				TextColor: &widget.ButtonTextColor{
					Idle:    color.Black,
					Hover:   color.Black,
					Pressed: color.Black,
				},
				TextFace: &face,
				Image: &widget.ButtonImage{
					Idle:    s.fill(color.NRGBA{233, 231, 231, 255}),
					Hover:   s.fill(color.NRGBA{223, 220, 220, 255}),
					Pressed: s.fill(color.NRGBA{197, 192, 196, 255}),
				},
				TextPadding: widget.NewInsetsSimple(5),
				MinSize:     &img.Point{98, 40},
			},
			TabSpacing: constantutil.ConstantToPointer(1),
		},
		TabTheme: &widget.TabParams{
			BackgroundImage: s.fill(color.NRGBA{197, 197, 197, 255}),
		},
		TextInputTheme: &widget.TextInputParams{
			Face: &face,
			Image: &widget.TextInputImage{
				Highlight: s.highlight(),
				Idle:      s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled:  s.box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
			Color: &widget.TextInputColor{
				Idle:          color.Black,
				Caret:         color.Black,
				Disabled:      color.NRGBA{122, 122, 122, 255},
				DisabledCaret: color.NRGBA{122, 122, 122, 255},
			},
			Padding: widget.NewInsetsSimple(5),
		},
		TextAreaTheme: &widget.TextAreaParams{
			Face:                   &face,
			StripBBCode:            constantutil.ConstantToPointer(true),
			ControlWidgetSpacing:   constantutil.ConstantToPointer(2),
			TextPadding:            &widget.Insets{Right: 18},
			ForegroundColor:        color.Black,
			ScrollContainerPadding: widget.NewInsetsSimple(4),
			ScrollContainerImage: &widget.ScrollContainerImage{
				Idle:     s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Mask:     s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
			},
			Slider: &widget.SliderParams{
				TrackImage: &widget.SliderTrackImage{
					Idle:     s.upTo(smallRadius).box(color.NRGBA{233, 231, 231, 255}, color.NRGBA{223, 220, 220, 255}, 2),
					Disabled: s.upTo(smallRadius).box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				},
				HandleImage: &widget.ButtonImage{
					Idle:    s.upTo(smallRadius).box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
					Hover:   s.upTo(smallRadius).box(color.NRGBA{235, 235, 235, 255}, color.NRGBA{177, 177, 177, 255}, 2),
					Pressed: s.upTo(smallRadius).box(color.NRGBA{210, 210, 210, 255}, color.NRGBA{177, 177, 177, 255}, 2),
				},
			},
		},
		ProgressBarTheme: &widget.ProgressBarParams{
			TrackPadding: widget.NewInsetsSimple(2),
			TrackImage: &widget.ProgressBarImage{
				Idle:     s.upTo(smallRadius).box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				Hover:    s.upTo(smallRadius).box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.upTo(smallRadius).box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
		},
		SliderTheme: &widget.SliderParams{
			TrackPadding:    widget.NewInsetsSimple(2),
			FixedHandleSize: constantutil.ConstantToPointer(6),
			TrackOffset:     constantutil.ConstantToPointer(0),
			PageSizeFunc: func() int {
				return 1
			},
			TrackImage: &widget.SliderTrackImage{
				Idle:     s.upTo(smallRadius).box(color.NRGBA{233, 231, 231, 255}, color.NRGBA{223, 220, 220, 255}, 2),
				Disabled: s.upTo(smallRadius).box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
			HandleImage: &widget.ButtonImage{
				Idle:    s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{51, 51, 51, 255}, 2),
				Hover:   s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
				Pressed: s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
			},
		},
		ListTheme: &widget.ListParams{
			EntryFace:                   &face,
			EntryTextPadding:            widget.NewInsetsSimple(5),
			EntryTextHorizontalPosition: constantutil.ConstantToPointer(widget.TextPositionStart),
			EntryTextVerticalPosition:   constantutil.ConstantToPointer(widget.TextPositionCenter),
			MinSize:                     &img.Point{150, 0},
			EntryColor:                  entryColor,
			EntryImage:                  s.entries(entryColor),
			ScrollContainerPadding:      widget.NewInsetsSimple(4),
			ScrollContainerImage: &widget.ScrollContainerImage{
				Idle:     s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Mask:     s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
			},
			Slider: &widget.SliderParams{
				TrackImage: &widget.SliderTrackImage{
					Idle:     s.upTo(smallRadius).box(color.NRGBA{233, 231, 231, 255}, color.NRGBA{223, 220, 220, 255}, 2),
					Disabled: s.upTo(smallRadius).box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				},
				HandleImage: &widget.ButtonImage{
					Idle:    s.upTo(smallRadius).box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
					Hover:   s.upTo(smallRadius).box(color.NRGBA{235, 235, 235, 255}, color.NRGBA{177, 177, 177, 255}, 2),
					Pressed: s.upTo(smallRadius).box(color.NRGBA{210, 210, 210, 255}, color.NRGBA{177, 177, 177, 255}, 2),
				},
			},
		},
		ListComboButtonTheme: &widget.ListComboButtonParams{
			MaxContentHeight: constantutil.ConstantToPointer(200),
			List: &widget.ListParams{
				EntryFace:                   &face,
				EntryTextPadding:            widget.NewInsetsSimple(5),
				EntryTextHorizontalPosition: constantutil.ConstantToPointer(widget.TextPositionStart),
				EntryTextVerticalPosition:   constantutil.ConstantToPointer(widget.TextPositionCenter),
				MinSize:                     &img.Point{200, 0},
				EntryColor:                  entryColor,
				EntryImage:                  s.entries(entryColor),
				ScrollContainerPadding:      widget.NewInsetsSimple(4),
				ScrollContainerImage: &widget.ScrollContainerImage{
					Idle:     s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
					Disabled: s.box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					Mask:     s.box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				},
				Slider: &widget.SliderParams{
					TrackImage: &widget.SliderTrackImage{
						Idle:     s.upTo(smallRadius).box(color.NRGBA{233, 231, 231, 255}, color.NRGBA{223, 220, 220, 255}, 2),
						Disabled: s.upTo(smallRadius).box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					},
					HandleImage: &widget.ButtonImage{
						Idle:    s.upTo(smallRadius).box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
						Hover:   s.upTo(smallRadius).box(color.NRGBA{235, 235, 235, 255}, color.NRGBA{177, 177, 177, 255}, 2),
						Pressed: s.upTo(smallRadius).box(color.NRGBA{210, 210, 210, 255}, color.NRGBA{177, 177, 177, 255}, 2),
					},
				},
			},
			Button: &widget.ButtonParams{
				TextColor: &widget.ButtonTextColor{
					Idle:    color.Black,
					Hover:   color.Black,
					Pressed: color.Black,
				},
				TextFace: &face,
				Image: &widget.ButtonImage{
					Idle:    s.box(color.NRGBA{233, 231, 231, 255}, color.NRGBA{223, 220, 220, 255}, 2),
					Hover:   s.box(color.NRGBA{223, 220, 220, 255}, color.NRGBA{197, 192, 196, 255}, 2),
					Pressed: s.box(color.NRGBA{197, 192, 196, 255}, color.NRGBA{177, 172, 176, 255}, 2),
				},
				TextPadding: &widget.Insets{
					Left:   30,
					Right:  30,
					Top:    5,
					Bottom: 5,
				},
				TextPosition: &widget.TextPositioning{
					VTextPosition: widget.TextPositionCenter,
					HTextPosition: widget.TextPositionCenter,
				},
				MinSize: &img.Point{200, 0},
			},
		},
		CheckboxTheme: &widget.CheckboxParams{
			Image: getLightCheckbox(s),
			Label: &widget.LabelParams{
				Face: &face,
				Color: &widget.LabelColor{
					Idle:     color.Black,
					Disabled: color.NRGBA{122, 122, 122, 255},
				},
			},
		},
	}
}
