package bot

import (
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed profiles/2020_profile.icc
var icc2020Profile []byte

type HdrArgs struct {
	ImageURL   string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Multiply   float64 `default:"1.5" description:"Multiplier for pixel values. Higher values produce brighter, more saturated results."`
	GammaExponent float64 `default:"0.9" description:"Exponent for gamma power curve. Lower values brighten midtones more."`
}

func (args HdrArgs) GetImageURL() string {
	return args.ImageURL
}

// Hdr applies aggressive color boosting to an image, producing an over-processed HDR look.
// Converts to linear RGB, auto-adjusts gamma, multiplies and power-curves pixel values,
// converts back to sRGB, and applies a Display P3 ICC profile.
func Hdr(wand *imagick.MagickWand, args HdrArgs) ([]*imagick.MagickWand, error) {
	// Convert to linear RGB colorspace
	err := wand.TransformImageColorspace(imagick.COLORSPACE_RGB)
	if err != nil {
		return nil, fmt.Errorf("error converting to RGB colorspace: %w", err)
	}

	// Auto-gamma correction
	err = wand.AutoGammaImage()
	if err != nil {
		return nil, fmt.Errorf("error applying auto-gamma: %w", err)
	}

	// Multiply pixel values (boost intensity)
	err = wand.EvaluateImage(imagick.EVAL_OP_MULTIPLY, args.Multiply)
	if err != nil {
		return nil, fmt.Errorf("error evaluating multiply: %w", err)
	}

	// Apply power curve (brighten midtones)
	err = wand.EvaluateImage(imagick.EVAL_OP_POW, args.GammaExponent)
	if err != nil {
		return nil, fmt.Errorf("error evaluating pow: %w", err)
	}

	// Convert back to sRGB
	err = wand.TransformImageColorspace(imagick.COLORSPACE_SRGB)
	if err != nil {
		return nil, fmt.Errorf("error converting to sRGB colorspace: %w", err)
	}

	// Set 16-bit depth
	err = wand.SetImageDepth(16)
	if err != nil {
		return nil, fmt.Errorf("error setting image depth: %w", err)
	}

	// Apply ICC profile
	err = wand.ProfileImage("icc", icc2020Profile)
	if err != nil {
		return nil, fmt.Errorf("error applying ICC profile: %w", err)
	}

	return []*imagick.MagickWand{wand}, nil
}
