package activities

import (
	"fmt"
	"math"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func decodeOperationArgs(args OperationArgs, targetPtr any) error {
	// On Temporal decoding, the args come through as a map[string]any, rather than our desired type
	mapStruct := args.Args.(map[string]any)

	return mapstructure.Decode(mapStruct, targetPtr)
}

func resizeMaintainAspectRatio(wand *imagick.MagickWand, width uint, height uint) error {
	inputHeight := float64(wand.GetImageHeight())
	inputWidth := float64(wand.GetImageWidth())

	widthMagFactor := float64(width) / inputWidth
	heightMagFactor := float64(height) / inputHeight

	minFactor := math.Min(widthMagFactor, heightMagFactor)

	targetWidth := inputWidth * minFactor
	targetHeight := inputHeight * minFactor

	return wand.ScaleImage(uint(targetWidth), uint(targetHeight))
}

func shrinkMaintainAspectRatio(wand *imagick.MagickWand, width uint, height uint) error {
	inputHeight := float64(wand.GetImageHeight())
	inputWidth := float64(wand.GetImageWidth())

	if inputWidth <= float64(width) && inputHeight <= float64(height) {
		return nil
	}

	return resizeMaintainAspectRatio(wand, width, height)
}

type fitMode int

const (
	fitModeFit fitMode = iota
	fitModeStretch
	fitModeFitHeight
)

type positionMode int

const (
	positionModeCentered positionMode = iota
	positionModeTopLeft
)

type frameOptions struct {
	FitMode      fitMode
	PositionMode positionMode
}

func applyFrame(jobWorkspace workspace.Workspace, opArgs OperationArgs, frameBytes []byte, options frameOptions) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	frame := imagick.NewMagickWand()
	defer frame.Destroy()

	if err := frame.ReadImageBlob(frameBytes); err != nil {
		return nil, fmt.Errorf("error reading frame: %w", err)
	}

	openX, openY, openW, openH, err := findTransparentOpeningRect(frame)
	if err != nil {
		return nil, fmt.Errorf("error finding frame opening: %w", err)
	}

	result, err := frameImage(wand, frame, openX, openY, openW, openH, options)
	if err != nil {
		return nil, err
	}

	return saveFrames(jobWorkspace, result...)
}

func findTransparentOpeningRect(frame *imagick.MagickWand) (x, y, width, height int, err error) {
	analysis := frame.Clone()
	defer analysis.Destroy()

	if err := analysis.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_EXTRACT); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error extracting alpha channel: %w", err)
	}
	if err := analysis.NegateImage(false); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error negating alpha mask: %w", err)
	}
	if err := analysis.TrimImage(0); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error trimming: %w", err)
	}

	_, _, ox, oy, err := analysis.GetImagePage()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error getting trimmed region geometry: %w", err)
	}

	return ox, oy, int(analysis.GetImageWidth()), int(analysis.GetImageHeight()), nil
}

func frameImage(wand *imagick.MagickWand, frame *imagick.MagickWand, openX, openY, openW, openH int, options frameOptions) ([]*imagick.MagickWand, error) {
	switch options.FitMode {
	case fitModeStretch:
		if err := wand.ResizeImage(uint(openW), uint(openH), imagick.FILTER_LANCZOS); err != nil {
			return nil, fmt.Errorf("error resizing image: %w", err)
		}
	case fitModeFit:
		if err := resizeMaintainAspectRatio(wand, uint(openW), uint(openH)); err != nil {
			return nil, fmt.Errorf("error resizing image: %w", err)
		}
	case fitModeFitHeight:
		scale := float64(openH) / float64(wand.GetImageHeight())
		newW := uint(float64(wand.GetImageWidth()) * scale)
		if err := wand.ResizeImage(newW, uint(openH), imagick.FILTER_LANCZOS); err != nil {
			return nil, fmt.Errorf("error resizing image: %w", err)
		}
	}

	bg := imagick.NewMagickWand()
	defer bg.Destroy()

	bgColor := imagick.NewPixelWand()
	defer bgColor.Destroy()

	bgColor.SetColor("white")
	if err := bg.NewImage(uint(openW), uint(openH), bgColor); err != nil {
		return nil, fmt.Errorf("error creating background: %w", err)
	}

	x, y := 0, 0
	if options.PositionMode == positionModeCentered {
		x = (openW - int(wand.GetImageWidth())) / 2
		y = (openH - int(wand.GetImageHeight())) / 2
	}
	if err := bg.CompositeImage(wand, imagick.COMPOSITE_OP_OVER, true, x, y); err != nil {
		return nil, fmt.Errorf("error compositing image onto background: %w", err)
	}

	if err := frame.CompositeImage(bg, imagick.COMPOSITE_OP_DST_OVER, true, openX, openY); err != nil {
		return nil, fmt.Errorf("error compositing background onto frame: %w", err)
	}

	return []*imagick.MagickWand{frame}, nil
}

type mirrorDirection string

const (
	mirrorDirectionVertical   mirrorDirection = "vertical"
	mirrorDirectionHorizontal mirrorDirection = "horizontal"
)

func applyMirror(jobWorkspace workspace.Workspace, opArgs OperationArgs, direction mirrorDirection, flipped bool) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	result, err := mirrorImage(wand, direction, flipped)
	if err != nil {
		return nil, err
	}

	return saveFrames(jobWorkspace, result...)
}

func mirrorImage(wand *imagick.MagickWand, direction mirrorDirection, flipped bool) ([]*imagick.MagickWand, error) {
	var desiredGravity imagick.GravityType
	if direction == mirrorDirectionHorizontal {
		if flipped {
			desiredGravity = imagick.GRAVITY_EAST
		} else {
			desiredGravity = imagick.GRAVITY_WEST
		}
	} else {
		if flipped {
			desiredGravity = imagick.GRAVITY_SOUTH
		} else {
			desiredGravity = imagick.GRAVITY_NORTH
		}
	}

	err := wand.SetImageGravity(desiredGravity)
	if err != nil {
		return nil, fmt.Errorf("error setting gravity: %w", err)
	}

	var half *imagick.MagickWand
	var xOffset, yOffset int

	if direction == mirrorDirectionHorizontal {
		half = wand.Clone()
		err = half.CropImageToTiles("50%x100%")
		if err != nil {
			return nil, fmt.Errorf("error cropping image half: %w", err)
		}

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
		half = wand.Clone()
		err = half.CropImageToTiles("100%x50%")
		if err != nil {
			return nil, fmt.Errorf("error cropping image half: %w", err)
		}

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
	defer half.Destroy()

	err = wand.CompositeImage(half, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
	if err != nil {
		return nil, fmt.Errorf("error compositing image: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}

type overlayOptions struct {
	HFlip bool
	VFlip bool

	OverlayWidthFactor  float64
	OverlayHeightFactor float64

	RightToLeft bool
}

type overlayArgs struct {
	HFlip bool
	VFlip bool
}

func applyOverlay(jobWorkspace workspace.Workspace, opArgs OperationArgs, overlayImage []byte, initialOptions overlayOptions) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var args overlayArgs
	err = decodeOperationArgs(opArgs, &args)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	options := initialOptions
	if args.HFlip {
		options.HFlip = !options.HFlip
	}
	if args.VFlip {
		options.VFlip = !options.VFlip
	}

	err = overlayImageOnWand(wand, overlayImage, options)
	if err != nil {
		return nil, err
	}

	return saveFrames(jobWorkspace, wand)
}

func overlayImageOnWand(wand *imagick.MagickWand, overlay []byte, options overlayOptions) error {
	overlayWand := imagick.NewMagickWand()
	defer overlayWand.Destroy()

	err := overlayWand.ReadImageBlob(overlay)
	if err != nil {
		return fmt.Errorf("error reading overlay: %w", err)
	}

	if options.HFlip {
		err = overlayWand.FlopImage()
		if err != nil {
			return fmt.Errorf("error flipping overlay horizontally: %w", err)
		}
	}
	if options.VFlip {
		err = overlayWand.FlipImage()
		if err != nil {
			return fmt.Errorf("error flipping overlay vertically: %w", err)
		}
	}

	inputWidth := wand.GetImageWidth()
	inputHeight := wand.GetImageHeight()

	err = resizeMaintainAspectRatio(
		overlayWand,
		uint(float64(inputWidth)*options.OverlayWidthFactor),
		uint(float64(inputHeight)*options.OverlayHeightFactor),
	)
	if err != nil {
		return fmt.Errorf("error resizing overlay: %w", err)
	}

	overlayWidth := overlayWand.GetImageWidth()
	overlayHeight := overlayWand.GetImageHeight()

	if options.HFlip {
		options.RightToLeft = !options.RightToLeft
	}

	xOffset := 0
	if options.RightToLeft {
		xOffset = int(inputWidth - overlayWidth)
	}

	yOffset := 0
	if !options.VFlip {
		yOffset = int(inputHeight - overlayHeight)
	}

	return wand.CompositeImage(overlayWand, imagick.COMPOSITE_OP_ATOP, true, xOffset, yOffset)
}
