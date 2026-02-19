package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type ResizeArgs struct {
	Width    float64 `description:"Width in pixels (absolute) or percent (e.g. 150 = 150%)."`
	Height   float64 `description:"Height in pixels (absolute) or percent (e.g. 150 = 150%)."`
	ImageURL string  `default:"" description:"Image URL to process. Leave blank to auto-find."`
	Mode     string  `default:"percent" description:"Resize mode (percent/absolute) for width/height values."`
}

func (args ResizeArgs) GetImageURL() string {
	return args.ImageURL
}

// Resize resizes an image.
func Resize(wand *imagick.MagickWand, args ResizeArgs) ([]*imagick.MagickWand, error) {
	var targetHeight, targetWidth uint
	switch args.Mode {
	case "absolute":
		targetHeight = uint(args.Height)
		targetWidth = uint(args.Width)
	case "percent":
		targetHeight = uint((args.Height / 100) * float64(wand.GetImageHeight()))
		targetWidth = uint((args.Width / 100) * float64(wand.GetImageWidth()))
	default:
		return nil, fmt.Errorf("unsupported mode: %s (must be one of percent, absolute)", args.Mode)
	}

	err := wand.ResizeImage(targetWidth, targetHeight, imagick.FILTER_POINT)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
