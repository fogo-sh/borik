package bot

import (
	"fmt"

	imagick7 "gopkg.in/gographics/imagick.v3/imagick"
)

type MagikArgs struct {
	ImageURL         string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale            float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
	WidthMultiplier  float64 `default:"0.5" description:"Multiplier to apply to the width of the input image to produce the intermediary image."`
	HeightMultiplier float64 `default:"0.5" description:"Multiplier to apply to the height of the input image to produce the intermediary image."`
}

func (args MagikArgs) GetImageURL() string {
	return args.ImageURL
}

func magikHelper(wand *imagick7.MagickWand, args MagikArgs) ([]*imagick7.MagickWand, error) {
	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err := wand.LiquidRescaleImage(uint(float64(width)*args.WidthMultiplier), uint(float64(height)*args.HeightMultiplier), args.Scale, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	err = wand.ResizeImage(width, height, imagick7.FILTER_LANCZOS)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to resize image: %w", err)
	}

	return []*imagick7.MagickWand{wand}, nil
}

// Magik runs content-aware scaling on an image.
func Magik(wand *imagick7.MagickWand, args MagikArgs) ([]*imagick7.MagickWand, error) {
	return magikHelper(wand, args)
}

type LagikArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale    float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
}

func (args LagikArgs) GetImageURL() string {
	return args.ImageURL
}

// Lagik runs content-aware scaling on an image.
func Lagik(wand *imagick7.MagickWand, args LagikArgs) ([]*imagick7.MagickWand, error) {
	return magikHelper(
		wand,
		MagikArgs{
			ImageURL:         args.ImageURL,
			Scale:            args.Scale,
			WidthMultiplier:  1.5,
			HeightMultiplier: 1.5,
		},
	)
}
