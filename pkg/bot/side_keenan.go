package bot

import (
	_ "embed"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed side_keenan.png
var sideKeenanImage []byte

type SideKeenanArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Flip     bool   `default:"false" description:"Flip Side Keenan the other way."`
}

func (args SideKeenanArgs) GetImageURL() string {
	return args.ImageURL
}

func SideKeenan(wand *imagick.MagickWand, args SideKeenanArgs) ([]*imagick.MagickWand, error) {
	err := OverlayImage(
		wand,
		sideKeenanImage,
		OverlayOptions{
			HFlip:               args.Flip,
			VFlip:               false,
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.5,
		},
	)
	return []*imagick.MagickWand{wand}, err
}
