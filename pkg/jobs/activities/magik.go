package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type MagikArgs struct {
	Scale            float64
	WidthMultiplier  float64
	HeightMultiplier float64
}

func Magik(ctx context.Context, args OperationArgs) ([][]byte, error) {
	wand, err := loadFrame(args.Frame)
	if err != nil {
		return nil, err
	}

	var magikArgs MagikArgs
	err = decodeOperationArgs(args, &magikArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err = wand.LiquidRescaleImage(uint(float64(width)*magikArgs.WidthMultiplier), uint(float64(height)*magikArgs.HeightMultiplier), magikArgs.Scale, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	err = wand.ResizeImage(width, height, imagick.FILTER_LANCZOS)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to resize image: %w", err)
	}

	return saveFrames(wand)
}
