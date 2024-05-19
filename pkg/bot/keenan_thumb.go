package bot

import (
	_ "embed"
	"fmt"

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
	keenan := imagick.NewMagickWand()
	err := keenan.ReadImageBlob(keenanThumbImage)
	if err != nil {
		return nil, fmt.Errorf("error reading keenan: %w", err)
	}

	if args.Flip {
		err = keenan.FlopImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping keenan: %w", err)
		}
	}

	inputWidth := wand.GetImageWidth()
	inputHeight := wand.GetImageHeight()

	err = ResizeMaintainAspectRatio(keenan, inputWidth, inputHeight/2)
	if err != nil {
		return nil, fmt.Errorf("error resizing keenan: %w", err)
	}

	keenanWidth := keenan.GetImageWidth()
	keenanHeight := keenan.GetImageHeight()

	xOffset := 0
	if args.Flip {
		xOffset = int(inputWidth - keenanWidth)
	}

	yOffset := int(inputHeight - keenanHeight)

	err = wand.CompositeImage(keenan, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
