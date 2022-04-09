package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type WaawArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args WaawArgs) GetImageURL() string {
	return args.ImageURL
}

func Waaw(wand *imagick.MagickWand, args WaawArgs) ([]*imagick.MagickWand, error) {
	err := wand.SetImageGravity(imagick.GRAVITY_EAST)
	if err != nil {
		return nil, fmt.Errorf("error setting gravity: %w", err)
	}

	rightHalf := wand.TransformImage("50%x100%", "")

	err = rightHalf.FlopImage()
	if err != nil {
		return nil, fmt.Errorf("error flipping image: %w", err)
	}

	err = wand.CompositeImage(rightHalf, imagick.COMPOSITE_OP_ATOP, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}

type HaahArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args HaahArgs) GetImageURL() string {
	return args.ImageURL
}

func Haah(wand *imagick.MagickWand, args HaahArgs) ([]*imagick.MagickWand, error) {
	leftHalf := wand.TransformImage("50%x100%", "")

	err := leftHalf.FlopImage()
	if err != nil {
		return nil, fmt.Errorf("error flipping image: %w", err)
	}

	halfWidth := leftHalf.GetImageWidth()

	err = wand.CompositeImage(leftHalf, imagick.COMPOSITE_OP_ATOP, int(halfWidth), 0)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}

type WoowArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args WoowArgs) GetImageURL() string {
	return args.ImageURL
}

func Woow(wand *imagick.MagickWand, args WoowArgs) ([]*imagick.MagickWand, error) {
	topHalf := wand.TransformImage("100%x50%", "")

	err := topHalf.FlipImage()
	if err != nil {
		return nil, fmt.Errorf("error flipping image: %w", err)
	}

	halfHeight := topHalf.GetImageHeight()

	err = wand.CompositeImage(topHalf, imagick.COMPOSITE_OP_ATOP, 0, int(halfHeight))
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}

type HoohArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args HoohArgs) GetImageURL() string {
	return args.ImageURL
}

func Hooh(wand *imagick.MagickWand, args HoohArgs) ([]*imagick.MagickWand, error) {
	err := wand.SetImageGravity(imagick.GRAVITY_SOUTH)
	if err != nil {
		return nil, fmt.Errorf("error setting gravity: %w", err)
	}

	bottomHalf := wand.TransformImage("100%x50%", "")

	err = bottomHalf.FlipImage()
	if err != nil {
		return nil, fmt.Errorf("error flipping image: %w", err)
	}

	err = wand.CompositeImage(bottomHalf, imagick.COMPOSITE_OP_ATOP, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
