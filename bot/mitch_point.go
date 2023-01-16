package bot

import (
	"context"
	_ "embed"
	"fmt"

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

func MitchPoint(ctx context.Context, wand *imagick.MagickWand, args MitchPointArgs) ([]*imagick.MagickWand, error) {
	mitch := imagick.NewMagickWand()
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

	err = ResizeMaintainAspectRatio(mitch, inputWidth, wand.GetImageHeight())
	if err != nil {
		return nil, fmt.Errorf("error resizing mitch: %w", err)
	}

	mitchWidth := mitch.GetImageWidth()
	mitchHeight := mitch.GetImageHeight()

	var xOffset int
	if args.Flip {
		xOffset = 0
	} else {
		xOffset = int(inputWidth - mitchWidth)
	}

	yOffset := int(inputHeight - mitchHeight)

	err = wand.CompositeImage(mitch, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
