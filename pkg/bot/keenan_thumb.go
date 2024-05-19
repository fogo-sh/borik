package bot

import (
	_ "embed"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed keenan_thumb.png
var keenanThumbImage []byte

type KeenanThumbArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Flip     bool   `default:"false" description:"Flip Keenan the other way."`
}

func (args KeenanThumbArgs) GetImageURL() string {
	return args.ImageURL
}

func KeenanThumb(wand *imagick.MagickWand, args KeenanThumbArgs) ([]*imagick.MagickWand, error) {
	err := OverlayImage(
		wand,
		keenanThumbImage,
		OverlayOptions{
			HFlip:               args.Flip,
			VFlip:               false,
			OverlayWidthFactor:  1,
			OverlayHeightFactor: 0.5,
		},
	)
	return []*imagick.MagickWand{wand}, err
}
