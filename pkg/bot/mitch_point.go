package bot

import (
	_ "embed"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed mitch_point.png
var mitchPointImage []byte

type MitchPointArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Flip     bool   `default:"false" description:"Have Mitch pointing from the left side of the image, rather than the right side."`
}

func (args MitchPointArgs) GetImageURL() string {
	return args.ImageURL
}

func MitchPoint(wand *imagick.MagickWand, args MitchPointArgs) ([]*imagick.MagickWand, error) {
	err := OverlayImage(
		wand,
		mitchPointImage,
		OverlayOptions{
			HFlip:               args.Flip,
			VFlip:               false,
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 1,
		},
	)
	return []*imagick.MagickWand{wand}, err
}
