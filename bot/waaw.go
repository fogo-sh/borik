package bot

import (
	"fmt"

	imagick6 "gopkg.in/gographics/imagick.v2/imagick"
)

type mirrorDirection string

const (
	mirrorDirectionVertical   mirrorDirection = "vertical"
	mirrorDirectionHorizontal mirrorDirection = "horizontal"
)

func mirrorImage(wand *imagick6.MagickWand, direction mirrorDirection, flipped bool) ([]*imagick6.MagickWand, error) {
	var desiredGravity imagick6.GravityType
	if direction == mirrorDirectionHorizontal {
		if flipped {
			desiredGravity = imagick6.GRAVITY_EAST
		} else {
			desiredGravity = imagick6.GRAVITY_WEST
		}
	} else {
		if flipped {
			desiredGravity = imagick6.GRAVITY_SOUTH
		} else {
			desiredGravity = imagick6.GRAVITY_NORTH
		}
	}

	err := wand.SetImageGravity(desiredGravity)
	if err != nil {
		return nil, fmt.Errorf("error setting gravity: %w", err)
	}

	var half *imagick6.MagickWand
	var xOffset, yOffset int

	if direction == mirrorDirectionHorizontal {
		half = wand.TransformImage("50%x100%", "")
		err = half.FlopImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping image: %w", err)
		}

		if flipped {
			xOffset = 0
			yOffset = 0
		} else {
			xOffset = int(half.GetImageWidth())
			yOffset = 0
		}
	} else {
		half = wand.TransformImage("100%x50%", "")
		err = half.FlipImage()
		if err != nil {
			return nil, fmt.Errorf("error flipping image: %w", err)
		}

		if flipped {
			xOffset = 0
			yOffset = 0
		} else {
			xOffset = 0
			yOffset = int(half.GetImageHeight())
		}
	}

	err = wand.CompositeImage(half, imagick6.COMPOSITE_OP_ATOP, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick6.MagickWand{wand}, nil

}

type WaawArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args WaawArgs) GetImageURL() string {
	return args.ImageURL
}

func Waaw(wand *imagick6.MagickWand, args WaawArgs) ([]*imagick6.MagickWand, error) {
	return mirrorImage(wand, mirrorDirectionHorizontal, true)
}

type HaahArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args HaahArgs) GetImageURL() string {
	return args.ImageURL
}

func Haah(wand *imagick6.MagickWand, args HaahArgs) ([]*imagick6.MagickWand, error) {
	return mirrorImage(wand, mirrorDirectionHorizontal, false)
}

type WoowArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args WoowArgs) GetImageURL() string {
	return args.ImageURL
}

func Woow(wand *imagick6.MagickWand, args WoowArgs) ([]*imagick6.MagickWand, error) {
	return mirrorImage(wand, mirrorDirectionVertical, false)
}

type HoohArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (args HoohArgs) GetImageURL() string {
	return args.ImageURL
}

func Hooh(wand *imagick6.MagickWand, args HoohArgs) ([]*imagick6.MagickWand, error) {
	return mirrorImage(wand, mirrorDirectionVertical, true)
}
