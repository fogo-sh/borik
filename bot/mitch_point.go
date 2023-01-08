package bot

import (
	_ "embed"
	"fmt"

	imagick6 "gopkg.in/gographics/imagick.v2/imagick"
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

func MitchPoint(wand *imagick6.MagickWand, args MitchPointArgs) ([]*imagick6.MagickWand, error) {
	mitch := imagick6.NewMagickWand()
	err := mitch.ReadImageBlob(mitchPointImage)
	if err != nil {
		return nil, fmt.Errorf("error reading mitch: %w", err)
	}

	if args.Flip {
		err = mitch.FlopImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping mitch: %w", err)
		}
	}

	inputHeight := wand.GetImageHeight()
	inputWidth := wand.GetImageWidth()

	mitch = mitch.TransformImage("", fmt.Sprintf("%dx%d", inputWidth, inputHeight))

	mitchWidth := mitch.GetImageWidth()
	mitchHeight := mitch.GetImageHeight()

	var xOffset int
	if args.Flip {
		xOffset = 0
	} else {
		xOffset = int(inputWidth - mitchWidth)
	}

	yOffset := int(inputHeight - mitchHeight)

	err = wand.CompositeImage(mitch, imagick6.COMPOSITE_OP_ATOP, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick6.MagickWand{wand}, nil
}
