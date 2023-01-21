package bot

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type HueCycleArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Steps    uint   `default:"20" description:"Number of steps to do the hue shift in."`
}

func (args HueCycleArgs) GetImageURL() string {
	return args.ImageURL
}

// HueCycle cycles the hue on an image
func HueCycle(ctx context.Context, wand *imagick.MagickWand, args HueCycleArgs) ([]*imagick.MagickWand, error) {
	wands := []*imagick.MagickWand{wand}

	for i := uint(0); i < args.Steps; i++ {
		wand = wand.Clone()
		err := wand.ModulateImage(100, 100, 100+(200/float64(args.Steps)))
		if err != nil {
			return nil, fmt.Errorf("error cycling hue: %w", err)
		}
		wands = append(wands, wand)
	}

	return wands, nil
}
