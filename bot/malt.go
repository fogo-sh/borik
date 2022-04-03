package bot

import (
	"fmt"
	"io"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type _MaltArgs struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degree   float64 `default:"45" description:"Number of degrees to rotate the image by while processing."`
}

func (args _MaltArgs) GetImageURL() string {
	return args.ImageURL
}

// Malt mixes an image via a combination of operations.
func Malt(src []byte, dest io.Writer, args _MaltArgs) error {
	wand := imagick.NewMagickWand()
	err := wand.ReadImageBlob(src)
	if err != nil {
		return fmt.Errorf("error reading image: %w", err)
	}

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err = wand.SwirlImage(args.Degree)
	if err != nil {
		return fmt.Errorf("error while attempting to swirl: %w", err)
	}

	err = wand.LiquidRescaleImage(width/2, height/2, 1, 0)
	if err != nil {
		return fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	err = wand.SwirlImage(args.Degree * -1)
	if err != nil {
		return fmt.Errorf("error while attempting to swirl: %w", err)
	}

	err = wand.LiquidRescaleImage(width, height, 1, 0)
	if err != nil {
		return fmt.Errorf("error while attempting to liquid rescale: %w", err)
	}

	_, err = dest.Write(wand.GetImageBlob())
	if err != nil {
		return fmt.Errorf("error writing image: %w", err)
	}

	return nil
}
