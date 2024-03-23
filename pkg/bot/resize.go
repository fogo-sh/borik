package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type ResizeArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Mode     string  `default:"percent" description:"Mode of resizing to do. Controls how the width and height arguments are handled. Must be percent or absolute."`
	Width    float64 `description:"Target width. In absolute mode this should be the target width in pixels. For percent mode, this should be a percentage represented as a whole number (for example, 150 == 150%)"`
	Height   float64 `description:"Target height. In absolute mode this should be the target height in pixels. For percent mode, this should be a percentage represented as a whole number (for example, 150 == 150%)"`
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
