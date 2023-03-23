package bot

import (
	"context"
	_ "embed"
	"fmt"

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

func SideKeenan(ctx context.Context, wand *imagick.MagickWand, args SideKeenanArgs) ([]*imagick.MagickWand, error) {
	sideKeenan := imagick.NewMagickWand()
	err := sideKeenan.ReadImageBlob(sideKeenanImage)
	if err != nil {
		return nil, fmt.Errorf("error reading side keenan: %w", err)
	}

	if args.Flip {
		err = sideKeenan.FlopImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping side keenan: %w", err)
		}
	}

	inputWidth := wand.GetImageWidth()
	inputHeight := wand.GetImageHeight()

	err = ResizeMaintainAspectRatio(sideKeenan, inputWidth, inputHeight/2)
	if err != nil {
		return nil, fmt.Errorf("error resizing side keenan: %w", err)
	}

	sideKeenanWidth := sideKeenan.GetImageWidth()
	sideKeenanHeight := sideKeenan.GetImageHeight()

	xOffset := int(inputWidth - sideKeenanWidth)
	if args.Flip {
		xOffset = 0
	}

	yOffset := int(inputHeight - sideKeenanHeight)

	err = wand.CompositeImage(sideKeenan, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
