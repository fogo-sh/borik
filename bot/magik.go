package bot

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v2/imagick"
)

type MagikArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale    float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
}

func (args MagikArgs) GetImageURL() string {
	return args.ImageURL
}

// Magik runs content-aware scaling on an image.
func Magik(wand *imagick.MagickWand, args MagikArgs) ([]*imagick.MagickWand, error) {
	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	log.Debug().
		Uint("src_width", width).
		Uint("src_height", height).
		Uint("dest_width", width/2).
		Uint("dest_height", height/2).
		Msg("Liquid rescaling image")
	err := wand.LiquidRescaleImage(width/2, height/2, args.Scale, 0)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	log.Debug().
		Uint("dest_width", width).
		Uint("dest_height", height).
		Uint("src_width", width/2).
		Uint("src_height", height/2).
		Msg("Returning image to original size")
	err = wand.ResizeImage(width, height, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		return nil, fmt.Errorf("error while attempting to resize image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
