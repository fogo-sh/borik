package activities

import (
	"math"

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
