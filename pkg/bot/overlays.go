package bot

import (
	_ "embed"
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

func makeOverlayCommand(name, description string, op ImageOperation[OverlayImageArgs]) Command {
	return Command{
		name:         name,
		description:  description,
		textHandler:  MakeImageOpTextCommand(op),
		slashHandler: MakeImageOpSlashCommand(op),
	}
}

func generateOverlayCommands() []Command {
	return []Command{
		makeOverlayCommand("jackpog", "Have Jack Pog an image.", MakeImageOverlayOp(jackPogImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.5,
		})),
		makeOverlayCommand("sidekeenan", "Have Keenan on the side of an image.", MakeImageOverlayOp(sideKeenanImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.5,
			RightToLeft:         true,
		})),
		makeOverlayCommand("keenanthumb", "Have Keenan thumbs-up an image.", MakeImageOverlayOp(keenanThumbImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.5,
		})),
		makeOverlayCommand("mitchpoint", "Have Mitch point at an image.", MakeImageOverlayOp(mitchPointImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 1,
		})),
		makeOverlayCommand("stevepoint", "Have Steve point at an image.", MakeImageOverlayOp(stevePointImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 1,
			RightToLeft:         true,
		})),
		makeOverlayCommand("andrewpog", "Have Andrew Pog an image.", MakeImageOverlayOp(andrewPogImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.75,
			RightToLeft:         true,
		})),
		makeOverlayCommand("matlabkid", "Have matlab kid possess an image", MakeImageOverlayOp(matlabKidImage, OverlayOptions{
			VFlip:               true,
			OverlayWidthFactor:  1.2,
			OverlayHeightFactor: 1.3,
			RightToLeft:         true,
		})),
		makeOverlayCommand("natalieclimb", "Have Natalie climb an image.", MakeImageOverlayOp(natalieClimbImage, OverlayOptions{
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 1,
		})),
		makeOverlayCommand("dennystanding", "Have Denny standing in an image.", MakeImageOverlayOp(dennyStandingImage, OverlayOptions{
			OverlayWidthFactor:  0.4,
			OverlayHeightFactor: 0.6,
		})),
		makeOverlayCommand("shinji", "Have Shinji throw at an image.", MakeImageFixedOverlayOp(shinjiThrowImage, FixedOverlayOptions{
			X:      925,
			Y:      147,
			Width:  431,
			Height: 377,
		})),
	}
}
