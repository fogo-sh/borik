package bot

import (
	"fmt"
	"io"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type _ArcweldArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args _ArcweldArgs) GetImageURL() string {
	return args.ImageURL
}

// Arcweld destroys an image via a combination of operations.
func Arcweld(src []byte, dest io.Writer, args _ArcweldArgs) error {
	wand := imagick.NewMagickWand()
	err := wand.ReadImageBlob(src)
	if err != nil {
		return fmt.Errorf("error reading image: %w", err)
	}

	err = wand.EvaluateImageChannel(imagick.CHANNEL_RED, imagick.EVAL_OP_LEFT_SHIFT, 1)
	if err != nil {
		return fmt.Errorf("error left-shifting red channel: %w", err)
	}

	err = wand.ContrastStretchImage(0.3, 0.3)
	if err != nil {
		return fmt.Errorf("error contrast stretching image: %w", err)
	}

	err = wand.EvaluateImageChannel(imagick.CHANNEL_RED, imagick.EVAL_OP_THRESHOLD_BLACK, 0.9)
	if err != nil {
		return fmt.Errorf("error running threshold black: %w", err)
	}

	err = wand.SharpenImage(0, 0)
	if err != nil {
		return fmt.Errorf("error sharpening image: %w", err)
	}

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err = wand.LiquidRescaleImage(width/2, height/3, 1, 0)
	if err != nil {
		return fmt.Errorf("error liquid rescaling: %w", err)
	}

	width = wand.GetImageWidth()
	height = wand.GetImageHeight()

	err = wand.LiquidRescaleImage(width*2, height*3, 0.4, 0)
	if err != nil {
		return fmt.Errorf("error liquid rescaling: %w", err)
	}

	err = wand.ImplodeImage(0.2)
	if err != nil {
		return fmt.Errorf("error imploding image: %w", err)
	}

	err = wand.QuantizeImage(8, imagick.COLORSPACE_RGB, 0, true, false)
	if err != nil {
		return fmt.Errorf("error quantizing image: %w", err)
	}

	_, err = dest.Write(wand.GetImageBlob())
	if err != nil {
		return fmt.Errorf("error writing image: %w", err)
	}

	return nil
}
