package bot

import (
	_ "embed"
	"fmt"

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
	jack := imagick.NewMagickWand()
	err := jack.ReadImageBlob(jackPogImage)
	if err != nil {
		return nil, fmt.Errorf("error reading jack: %w", err)
	}

	if args.HFlip {
		err = jack.FlopImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping jack: %w", err)
		}
	}
	if args.VFlip {
		err = jack.FlipImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping jack: %w", err)
		}
	}

	inputWidth := wand.GetImageWidth()
	inputHeight := wand.GetImageHeight()

	err = ResizeMaintainAspectRatio(jack, inputWidth, inputHeight/2)
	if err != nil {
		return nil, fmt.Errorf("error resizing jack: %w", err)
	}

	jackWidth := jack.GetImageWidth()
	jackHeight := jack.GetImageHeight()

	xOffset := 0
	if args.HFlip {
		xOffset = int(inputWidth - jackWidth)
	}

	yOffset := 0
	if !args.VFlip {
		yOffset = int(inputHeight - jackHeight)
	}

	err = wand.CompositeImage(jack, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
