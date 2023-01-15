package bot

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type InvertArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args InvertArgs) GetImageURL() string {
	return args.ImageURL
}

// Invert inverts an image's colours.
func Invert(ctx context.Context, wand *imagick.MagickWand, args InvertArgs) ([]*imagick.MagickWand, error) {
	wand.SetImageChannelMask(imagick.CHANNEL_RED | imagick.CHANNEL_GREEN | imagick.CHANNEL_BLUE)
	err := wand.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error inverting image: %w", err)
	}
	return []*imagick.MagickWand{wand}, nil
}
