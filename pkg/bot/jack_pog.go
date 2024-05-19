package bot

import (
	_ "embed"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed jack_pog.png
var jackPogImage []byte

type JackPogArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	HFlip    bool   `default:"false" description:"Flip Jack horizontally."`
	VFlip    bool   `default:"false" description:"Flip Jack vertically."`
}

func (args JackPogArgs) GetImageURL() string {
	return args.ImageURL
}

func JackPog(wand *imagick.MagickWand, args JackPogArgs) ([]*imagick.MagickWand, error) {
	err := OverlayImage(
		wand,
		jackPogImage,
		OverlayOptions{
			HFlip:               args.HFlip,
			VFlip:               args.VFlip,
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.5,
		},
	)
	return []*imagick.MagickWand{wand}, err
}
