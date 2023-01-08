package bot

import (
	_ "embed"
	"fmt"

	imagick7 "gopkg.in/gographics/imagick.v3/imagick"
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

func StevePoint(wand *imagick7.MagickWand, args StevePointArgs) ([]*imagick7.MagickWand, error) {
	steve := imagick7.NewMagickWand()
	err := steve.ReadImageBlob(stevePointImage)
	if err != nil {
		return nil, fmt.Errorf("error reading steve: %w", err)
	}

	if args.Flip {
		err = steve.FlopImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping steve: %w", err)
		}
	}

	inputWidth := wand.GetImageWidth()

	err = ResizeMaintainAspectRatio(steve, inputWidth, wand.GetImageHeight())
	if err != nil {
		return nil, fmt.Errorf("error resizing steve: %w", err)
	}

	steveWidth := steve.GetImageWidth()

	var xOffset int
	if args.Flip {
		xOffset = 0
	} else {
		xOffset = int(inputWidth - steveWidth)
	}

	err = wand.CompositeImage(steve, imagick7.COMPOSITE_OP_ATOP, true, xOffset, 0)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick7.MagickWand{wand}, nil
}
