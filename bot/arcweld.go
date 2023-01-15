package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type ArcweldArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args ArcweldArgs) GetImageURL() string {
	return args.ImageURL
}

// Arcweld destroys an image via a combination of operations.
func Arcweld(wand *imagick.MagickWand, args ArcweldArgs) ([]*imagick.MagickWand, error) {
	origMask := wand.SetImageChannelMask(imagick.CHANNEL_RED)
	err := wand.EvaluateImage(imagick.EVAL_OP_LEFT_SHIFT, 1)
	if err != nil {
		return nil, fmt.Errorf("error left-shifting red channel: %w", err)
	}
	wand.SetImageChannelMask(origMask)

	err = wand.ContrastStretchImage(0.3, 0.3)
	if err != nil {
		return nil, fmt.Errorf("error contrast stretching image: %w", err)
	}

	wand.SetImageChannelMask(imagick.CHANNEL_RED)
	err = wand.EvaluateImage(imagick.EVAL_OP_THRESHOLD_BLACK, 0.9)
	if err != nil {
		return nil, fmt.Errorf("error running threshold black: %w", err)
	}
	wand.SetImageChannelMask(origMask)

	err = wand.SharpenImage(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error sharpening image: %w", err)
	}

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	err = wand.LiquidRescaleImage(width/2, height/3, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error liquid rescaling: %w", err)
	}

	width = wand.GetImageWidth()
	height = wand.GetImageHeight()

	err = wand.LiquidRescaleImage(width*2, height*3, 0.4, 0)
	if err != nil {
		return nil, fmt.Errorf("error liquid rescaling: %w", err)
	}

	err = wand.ImplodeImage(0.2, imagick.INTERPOLATE_PIXEL_NEAREST_INTERPOLATE)
	if err != nil {
		return nil, fmt.Errorf("error imploding image: %w", err)
	}

	err = wand.QuantizeImage(8, imagick.COLORSPACE_RGB, 0, imagick.DITHER_METHOD_FLOYD_STEINBERG, false)
	if err != nil {
		return nil, fmt.Errorf("error quantizing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
