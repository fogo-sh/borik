package bot

import (
	_ "embed"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed steve_point.png
var stevePointImage []byte

type StevePointArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Flip     bool   `default:"false" description:"Have Steve pointing from the left side of the image, rather than the right side."`
}

func (args StevePointArgs) GetImageURL() string {
	return args.ImageURL
}

func StevePoint(wand *imagick.MagickWand, args StevePointArgs) ([]*imagick.MagickWand, error) {
	err := OverlayImage(
		wand,
		stevePointImage,
		OverlayOptions{
			HFlip:               args.Flip,
			VFlip:               false,
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 1,
			RightToLeft:         true,
		},
	)
	return []*imagick.MagickWand{wand}, err
}
