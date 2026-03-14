package bot

import _ "embed"

//go:embed presidents_frame.png
var frameImage []byte

//go:embed heritage.png
var heritageImage []byte

//go:embed overlay_images/shinji_throw.png
var shinjiThrowImage []byte

func makeFrameCommand(name, description string, frameBytes []byte, options FrameOptions) Command {
	op := MakeImageFrameOp(frameBytes, options)
	return Command{
		name:         name,
		description:  description,
		textHandler:  MakeImageOpTextCommand(op),
		slashHandler: MakeImageOpSlashCommand(op),
	}
}

func generateFrameCommands() []Command {
	return []Command{
		makeFrameCommand("presidentsframe", "Apply the President's Frame to an image.", frameImage, FrameOptions{FitMode: FitModeStretch}),
		makeFrameCommand("heritage", "Turn an image into a Canadian Heritage Minute.", heritageImage, FrameOptions{}),
		makeFrameCommand("shinji", "Have Shinji throw at an image.", shinjiThrowImage, FrameOptions{FitMode: FitModeFitHeight}),
	}
}
