package bot

import (
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v2/imagick"
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
	steve := imagick.NewMagickWand()
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

	inputHeight := wand.GetImageHeight()
	inputWidth := wand.GetImageWidth()

	steve = steve.TransformImage("", fmt.Sprintf("%dx%d", inputWidth, inputHeight))

	steveWidth := steve.GetImageWidth()

	var xOffset int
	if args.Flip {
		xOffset = 0
	} else {
		xOffset = int(inputWidth - steveWidth)
	}

	err = wand.CompositeImage(steve, imagick.COMPOSITE_OP_ATOP, xOffset, 0)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
