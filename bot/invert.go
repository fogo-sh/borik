package bot

import (
	"fmt"

	imagick6 "gopkg.in/gographics/imagick.v2/imagick"
)

type InvertArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args InvertArgs) GetImageURL() string {
	return args.ImageURL
}

// Invert inverts an image's colours.
func Invert(wand *imagick6.MagickWand, _ InvertArgs) ([]*imagick6.MagickWand, error) {
	err := wand.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error inverting image: %w", err)
	}
	return []*imagick6.MagickWand{wand}, nil
}
