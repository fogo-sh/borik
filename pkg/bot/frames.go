package bot

import _ "embed"

//go:embed images/frames/presidents_frame.png
var frameImage []byte

//go:embed images/frames/heritage.png
var heritageImage []byte

//go:embed images/frames/shinji_throw.png
var shinjiThrowImage []byte

func makeFrameCommand(name, description string, op ImageOperation[FrameArgs]) Command {
	return Command{
		name:         name,
		description:  description,
		textHandler:  MakeImageOpTextCommand(op),
		slashHandler: MakeImageOpSlashCommand(op),
	}
}

func generateFrameCommands() []Command {
	return []Command{
		makeFrameCommand(
			"presidentsframe",
			"Apply the President's Frame to an image.",
			MakeImageFrameOp(frameImage, FrameOptions{
				FitMode: FitModeStretch,
			})),
		makeFrameCommand(
			"heritage",
			"Turn an image into a Canadian Heritage Minute.",
			MakeImageFrameOp(heritageImage, FrameOptions{})),
		makeFrameCommand(
			"shinji",
			"Have Shinji throw at an image.",
			MakeImageFrameOp(shinjiThrowImage, FrameOptions{
				FitMode:      FitModeFitHeight,
				PositionMode: PositionModeTopLeft,
			}),
		),
	}
}
