package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type MaltArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degree   float64 `default:"45" description:"Number of degrees to rotate the image by while processing."`
}

func (args MaltArgs) GetImageURL() string {
	return args.ImageURL
}

// Malt mixes an image via a combination of operations.
func Malt(wand *imagick.MagickWand, args MaltArgs) ([]*imagick.MagickWand, error) {
	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err := wand.SwirlImage(args.Degree)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to swirl: %w", err)
	}

	err = wand.LiquidRescaleImage(width/2, height/2, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	err = wand.SwirlImage(args.Degree * -1)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to swirl: %w", err)
	}

	err = wand.LiquidRescaleImage(width, height, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
