package bot

import (
	"fmt"

	imagick7 "gopkg.in/gographics/imagick.v3/imagick"
)

type InvertArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args InvertArgs) GetImageURL() string {
	return args.ImageURL
}

// Invert inverts an image's colours.
func Invert(wand *imagick7.MagickWand, _ InvertArgs) ([]*imagick7.MagickWand, error) {
	wand.SetImageChannelMask(imagick7.CHANNEL_RED | imagick7.CHANNEL_GREEN | imagick7.CHANNEL_BLUE)
	err := wand.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error inverting image: %w", err)
	}
	return []*imagick7.MagickWand{wand}, nil
}
