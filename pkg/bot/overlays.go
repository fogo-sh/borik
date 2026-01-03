package bot

import (
	_ "embed"

	"github.com/nint8835/parsley"
)

//go:embed overlay_images/jack_pog.png
var jackPogImage []byte

//go:embed overlay_images/keenan_thumb.png
var keenanThumbImage []byte

//go:embed overlay_images/mitch_point.png
var mitchPointImage []byte

//go:embed overlay_images/side_keenan.png
var sideKeenanImage []byte

//go:embed overlay_images/steve_point.png
var stevePointImage []byte

//go:embed overlay_images/andrew_pog.png
var andrewPogImage []byte

//go:embed overlay_images/trans_matlab_kid.png
var matlabKidImage []byte

//go:embed overlay_images/natalie_climb.png
var natalieClimbImage []byte

//go:embed overlay_images/denny_standing.png
var dennyStandingImage []byte

//go:embed overlay_images/shinji_throw.png
var shinjiThrowImage []byte

func registerOverlayCommands(parser *parsley.Parser) {
	_ = parser.NewCommand(
		"jackpog",
		"Have Jack Pog an image.",
		MakeImageOverlayCommand(
			jackPogImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 0.5,
			},
		),
	)
	_ = parser.NewCommand(
		"sidekeenan",
		"Have Keenan on the side of an image.",
		MakeImageOverlayCommand(
			sideKeenanImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 0.5,
				RightToLeft:         true,
			},
		),
	)
	_ = parser.NewCommand(
		"keenanthumb",
		"Have Keenan thumbs-up an image.",
		MakeImageOverlayCommand(
			keenanThumbImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 0.5,
			},
		),
	)
	_ = parser.NewCommand(
		"mitchpoint",
		"Have Mitch point at an image.",
		MakeImageOverlayCommand(
			mitchPointImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 1,
			},
		),
	)
	_ = parser.NewCommand(
		"stevepoint",
		"Have Steve point at an image.",
		MakeImageOverlayCommand(
			stevePointImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 1,
				RightToLeft:         true,
			},
		),
	)

	_ = parser.NewCommand(
		"andrewpog",
		"Have Andrew Pog an image.",
		MakeImageOverlayCommand(
			andrewPogImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 0.75,
				RightToLeft:         true,
			},
		),
	)

	_ = parser.NewCommand(
		"matlabkid",
		"Have matlab kid possess an image",
		MakeImageOverlayCommand(
			matlabKidImage,
			OverlayOptions{
				VFlip:               true,
				OverlayWidthFactor:  1.2,
				OverlayHeightFactor: 1.3,
				RightToLeft:         true,
			},
		),
	)

	_ = parser.NewCommand(
		"natalieclimb",
		"Have Natalie climb an image.",
		MakeImageOverlayCommand(
			natalieClimbImage,
			OverlayOptions{
				OverlayWidthFactor:  1,
				OverlayHeightFactor: 1,
			},
		),
	)

	_ = parser.NewCommand(
		"dennystanding",
		"Have Denny standing in an image.",
		MakeImageOverlayCommand(
			dennyStandingImage,
			OverlayOptions{
				OverlayWidthFactor:  0.4,
				OverlayHeightFactor: 0.6,
			},
		),
	)
	_ = parser.NewCommand(
		"shinji",
		"Have Shinji throw at an image.",
		MakeImageFixedOverlayCommand(
			shinjiThrowImage,
			FixedOverlayOptions{
				X:      925,
				Y:      147,
				Width:  431,
				Height: 377,
			},
		),
	)
}
