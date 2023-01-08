package bot

import (
	"fmt"

	imagick7 "gopkg.in/gographics/imagick.v3/imagick"
)

type RotateArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degrees  float64 `default:"90" description:"Number of degrees to rotate the image by."`
}

func (args RotateArgs) GetImageURL() string {
	return args.ImageURL
}

// Rotate rotates an image.
func Rotate(wand *imagick7.MagickWand, args RotateArgs) ([]*imagick7.MagickWand, error) {
	bgWand := imagick7.NewPixelWand()
	bgWand.SetAlpha(0)

	err := wand.RotateImage(bgWand, args.Degrees)
	if err != nil {
		return nil, fmt.Errorf("error rotating image: %w", err)
	}

	return []*imagick7.MagickWand{wand}, nil
}
