package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type GmagikArgs struct {
	ImageURL         string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale            float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
	Iterations       uint    `default:"5" description:"Number of iterations of magikification to run."`
	WidthMultiplier  float64 `default:"0.5" description:"Multiplier to apply to the width of the input image to produce the intermediary image."`
	HeightMultiplier float64 `default:"0.5" description:"Multiplier to apply to the height of the input image to produce the intermediary image."`
}

func (args GmagikArgs) GetImageURL() string {
	return args.ImageURL
}

// Gmagik runs content-aware scaling on an image.
func Gmagik(wand *imagick.MagickWand, args GmagikArgs) ([]*imagick.MagickWand, error) {
	var results []*imagick.MagickWand

	lastFrame := wand

	for i := uint(0); i < args.Iterations; i++ {
		newFrame, err := Magik(
			lastFrame.Clone(),
			MagikArgs{
				Scale:            args.Scale,
				WidthMultiplier:  args.WidthMultiplier,
				HeightMultiplier: args.HeightMultiplier,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("error running magik: %w", err)
		}
		lastFrame = newFrame[0]
		results = append(results, lastFrame)
	}

	return results, nil
}
