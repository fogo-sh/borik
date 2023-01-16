package bot

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type DeepfryArgs struct {
	ImageURL        string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	EdgeRadius      float64 `default:"100" description:"Radius of outline to draw around edges."`
	DownscaleFactor uint    `default:"2" description:"Factor to downscale the image by while processing."`
}

func (args DeepfryArgs) GetImageURL() string {
	return args.ImageURL
}

// Deepfry destroys an image via a combination of operations.
func Deepfry(ctx context.Context, wand *imagick.MagickWand, args DeepfryArgs) ([]*imagick.MagickWand, error) {
	err := wand.ResizeImage(wand.GetImageWidth()/args.DownscaleFactor, wand.GetImageHeight()/args.DownscaleFactor, imagick.FILTER_POINT)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	err = wand.ResizeImage(wand.GetImageWidth()*args.DownscaleFactor, wand.GetImageHeight()*args.DownscaleFactor, imagick.FILTER_POINT)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	wand.SetImageChannelMask(imagick.CHANNEL_RED | imagick.CHANNEL_GREEN | imagick.CHANNEL_BLUE)

	err = wand.EdgeImage(args.EdgeRadius)
	if err != nil {
		return nil, fmt.Errorf("error edge enhancing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
