package bot

import (
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed presidents_frame.png
var frameImage []byte

type FrameArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args FrameArgs) GetImageURL() string {
	return args.ImageURL
}

func PresidentsFrame(wand *imagick.MagickWand, args FrameArgs) ([]*imagick.MagickWand, error) {
	frame := imagick.NewMagickWand()
	if err := frame.ReadImageBlob(frameImage); err != nil {
		return nil, fmt.Errorf("error reading frame: %w", err)
	}

	frameWidth := frame.GetImageWidth()
	frameHeight := frame.GetImageHeight()
	targetWandWidth := uint(float64(frameWidth) * 0.6)
	targetWandHeight := uint(float64(frameHeight) * 0.6)
	if err := wand.ResizeImage(targetWandWidth, targetWandHeight, imagick.FILTER_LANCZOS); err != nil {
		return nil, fmt.Errorf("error resizing original image: %w", err)
	}

	wandWidth := wand.GetImageWidth()
	wandHeight := wand.GetImageHeight()
	xOffset := (int(frameWidth) - int(wandWidth)) / 2
	yOffset := (int(frameHeight) - int(wandHeight)) / 2
	if err := frame.CompositeImage(wand, imagick.COMPOSITE_OP_OVER, true, xOffset, yOffset); err != nil {
		return nil, fmt.Errorf("error compositing wand onto frame: %w", err)
	}

	return []*imagick.MagickWand{frame}, nil
}
