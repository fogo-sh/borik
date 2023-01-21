package bot

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type ModulateArgs struct {
	ImageURL   string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Brightness float64 `default:"100"`
	Saturation float64 `default:"100"`
	Hue        float64 `default:"100"`
}

func (args ModulateArgs) GetImageURL() string {
	return args.ImageURL
}

// Modulate allows modifying of the brightness, saturation, and hue of an image
func Modulate(ctx context.Context, wand *imagick.MagickWand, args ModulateArgs) ([]*imagick.MagickWand, error) {
	err := wand.ModulateImage(args.Brightness, args.Saturation, args.Hue)
	if err != nil {
		return nil, fmt.Errorf("error modulating image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
