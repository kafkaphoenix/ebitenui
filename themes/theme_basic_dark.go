package themes

import (
	img "image"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/constantutil"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

func getDarkCheckbox(s shape) *widget.CheckboxImage {
	border := s.upTo(checkboxRadius).box(color.NRGBA{119, 119, 119, 255}, color.White, 2)
	idle := ebiten.NewImage(32, 32)
	border.Draw(idle, 32, 32, nil)
	idle9s := image.NewFixedNineSlice(idle)

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
	drawOp.ColorScale.ScaleWithColor(color.White)
	vector.StrokePath(checked, &checkmark, strokeOp, drawOp)
	checked9s := image.NewFixedNineSlice(checked)

	// Create greyed image
	greyed := ebiten.NewImage(32, 32)
	border.Draw(greyed, 32, 32, nil)
	vector.StrokeLine(greyed, 5, 16, 27, 16, 3, color.White, true)
	greyed9s := image.NewFixedNineSlice(greyed)

	return &widget.CheckboxImage{
		Unchecked:         idle9s,
		Checked:           checked9s,
		Greyed:            greyed9s,
		UncheckedHovered:  idle9s,
		CheckedHovered:    checked9s,
		GreyedHovered:     greyed9s,
		UncheckedDisabled: idle9s,
		CheckedDisabled:   checked9s,
		GreyedDisabled:    greyed9s,
	}
}

// GetBasicDarkTheme returns the basic dark theme: square boxes, its own
// borders.
func GetBasicDarkTheme() *widget.Theme { return newBasicDarkTheme(shape{}) }

// GetBasicDarkThemeWith returns the basic dark theme shaped by opts: Rounded
// corners, Borders of another width.
func GetBasicDarkThemeWith(opts ...ThemeOpt) *widget.Theme {
	return newBasicDarkTheme(newShape(opts))
}

// newBasicDarkTheme returns the basic dark theme, its boxes drawn as s
// shapes them.
func newBasicDarkTheme(s shape) *widget.Theme {
	// load button text font
	face := loadFont(20)
	entryColor := &widget.ListEntryColor{
		Unselected:         color.White,
		Selected:           color.White,
		DisabledUnselected: color.NRGBA{127, 122, 126, 255},
		DisabledSelected:   color.NRGBA{127, 122, 126, 255},

		SelectedBackground:        color.NRGBA{40, 40, 40, 255},
		SelectedFocusedBackground: color.NRGBA{50, 50, 50, 255},

		SelectingBackground:        color.NRGBA{99, 99, 99, 255},
		FocusedBackground:          color.NRGBA{99, 99, 99, 255},
		SelectingFocusedBackground: color.NRGBA{99, 99, 99, 255},
		DisabledSelectedBackground: color.NRGBA{99, 99, 99, 255},
	}

	return &widget.Theme{
		DefaultFace:      &face,
		DefaultTextColor: color.White,
		ButtonTheme: &widget.ButtonParams{
			TextColor: &widget.ButtonTextColor{
				Idle:    color.White,
				Hover:   color.White,
				Pressed: color.White,
			},
			TextFace: &face,
			Image: &widget.ButtonImage{
				Idle:    s.box(color.NRGBA{51, 51, 51, 255}, color.NRGBA{81, 81, 81, 255}, 2),
				Hover:   s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{51, 51, 51, 255}, 2),
				Pressed: s.box(color.NRGBA{119, 119, 119, 255}, color.NRGBA{77, 77, 77, 255}, 2),
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
			BackgroundImage: s.fill(colornames.Black),
		},
		LabelTheme: &widget.LabelParams{
			Face: &face,
			Color: &widget.LabelColor{
				Idle:     color.White,
				Disabled: color.NRGBA{122, 122, 122, 255},
			},
		},
		TextTheme: &widget.TextParams{
			Face:  &face,
			Color: color.White,
		},
		TabbookTheme: &widget.TabBookParams{
			TabButton: &widget.ButtonParams{
				TextColor: &widget.ButtonTextColor{
					Idle:    color.White,
					Hover:   color.White,
					Pressed: color.White,
				},
				TextFace: &face,
				Image: &widget.ButtonImage{
					Idle:    s.fill(color.NRGBA{51, 51, 51, 255}),
					Hover:   s.fill(color.NRGBA{77, 77, 77, 255}),
					Pressed: s.fill(color.NRGBA{119, 119, 119, 255}),
				},
				TextPadding: widget.NewInsetsSimple(5),
				MinSize:     &img.Point{98, 40},
			},
			TabSpacing: constantutil.ConstantToPointer(1),
		},
		TabTheme: &widget.TabParams{
			BackgroundImage: s.fill(color.NRGBA{32, 32, 32, 255}),
		},
		TextInputTheme: &widget.TextInputParams{
			Face: &face,
			Image: &widget.TextInputImage{
				Highlight: s.highlight(),
				Idle:      s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled:  s.box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
			Color: &widget.TextInputColor{
				Idle:          color.White,
				Caret:         color.White,
				Disabled:      color.NRGBA{127, 122, 126, 255},
				DisabledCaret: color.NRGBA{127, 122, 126, 255},
			},
			Padding: widget.NewInsetsSimple(5),
		},
		TextAreaTheme: &widget.TextAreaParams{
			Face:                   &face,
			StripBBCode:            constantutil.ConstantToPointer(true),
			ControlWidgetSpacing:   constantutil.ConstantToPointer(2),
			TextPadding:            &widget.Insets{Right: 18},
			ForegroundColor:        color.White,
			ScrollContainerPadding: widget.NewInsetsSimple(4),
			ScrollContainerImage: &widget.ScrollContainerImage{
				Idle:     s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Mask:     s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
			Slider: &widget.SliderParams{
				TrackImage: &widget.SliderTrackImage{
					Idle:     s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					Disabled: s.upTo(smallRadius).box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				},
				HandleImage: &widget.ButtonImage{
					Idle:    s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{51, 51, 51, 255}, 2),
					Hover:   s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
					Pressed: s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
				},
			},
		},
		ProgressBarTheme: &widget.ProgressBarParams{
			TrackPadding: widget.NewInsetsSimple(2),
			TrackImage: &widget.ProgressBarImage{
				Idle:     s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.upTo(smallRadius).box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Hover:    s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
		},
		SliderTheme: &widget.SliderParams{
			TrackPadding:    widget.NewInsetsSimple(0),
			FixedHandleSize: constantutil.ConstantToPointer(6),
			TrackOffset:     constantutil.ConstantToPointer(0),

			PageSizeFunc: func() int {
				return 1
			},
			TrackImage: &widget.SliderTrackImage{
				Idle:     s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.upTo(smallRadius).box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
			HandleImage: &widget.ButtonImage{
				Idle:         s.upTo(smallRadius).box(color.White, color.NRGBA{177, 177, 177, 255}, 1),
				Hover:        s.upTo(smallRadius).box(color.NRGBA{235, 235, 235, 255}, color.NRGBA{177, 177, 177, 255}, 2),
				Pressed:      s.upTo(smallRadius).box(color.NRGBA{210, 210, 210, 255}, color.NRGBA{177, 177, 177, 255}, 2),
				PressedHover: s.upTo(smallRadius).box(color.NRGBA{210, 210, 210, 255}, color.NRGBA{177, 177, 177, 255}, 2),
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
				Idle:     s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Disabled: s.box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				Mask:     s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
			},
			Slider: &widget.SliderParams{
				TrackImage: &widget.SliderTrackImage{
					Idle:     s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					Disabled: s.upTo(smallRadius).box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				},
				HandleImage: &widget.ButtonImage{
					Idle:    s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{51, 51, 51, 255}, 2),
					Hover:   s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
					Pressed: s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
				},
			},
		},
		ListComboButtonTheme: &widget.ListComboButtonParams{
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
					Idle:     s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					Disabled: s.box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					Mask:     s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
				},
				Slider: &widget.SliderParams{
					TrackImage: &widget.SliderTrackImage{
						Idle:     s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{177, 177, 177, 255}, 1),
						Disabled: s.upTo(smallRadius).box(color.NRGBA{47, 47, 47, 255}, color.NRGBA{177, 177, 177, 255}, 1),
					},
					HandleImage: &widget.ButtonImage{
						Idle:    s.upTo(smallRadius).box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{51, 51, 51, 255}, 2),
						Hover:   s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
						Pressed: s.upTo(smallRadius).box(color.NRGBA{99, 99, 99, 255}, color.NRGBA{77, 77, 77, 255}, 2),
					},
				},
			},
			Button: &widget.ButtonParams{
				TextColor: &widget.ButtonTextColor{
					Idle:    color.White,
					Hover:   color.White,
					Pressed: color.White,
				},
				TextFace: &face,
				Image: &widget.ButtonImage{
					Idle:    s.box(color.NRGBA{51, 51, 51, 255}, color.NRGBA{81, 81, 81, 255}, 2),
					Hover:   s.box(color.NRGBA{77, 77, 77, 255}, color.NRGBA{51, 51, 51, 255}, 2),
					Pressed: s.box(color.NRGBA{119, 119, 119, 255}, color.NRGBA{77, 77, 77, 255}, 2),
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
			MaxContentHeight: constantutil.ConstantToPointer(200),
		},
		CheckboxTheme: &widget.CheckboxParams{
			Label: &widget.LabelParams{
				Face: &face,
				Color: &widget.LabelColor{
					Idle:     color.White,
					Disabled: color.NRGBA{122, 122, 122, 255},
				},
			},
			Image: getDarkCheckbox(s),
		},
	}
}
