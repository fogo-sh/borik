package bot

import (
	"fmt"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type OtsuArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Invert   bool   `default:"false" description:"Invert the colors."`
}

func (args OtsuArgs) GetImageURL() string {
	return args.ImageURL
}

// Otsu turns the image black and white by applying an adaptive threshold using Otsu's method
func Otsu(wand *imagick.MagickWand, args OtsuArgs) ([]*imagick.MagickWand, error) {
	numOfPixels := 0
	histogram := map[int]int{}

	pixelIterator := wand.NewPixelIterator()
	for y := 0; y < int(wand.GetImageHeight()); y++ {
		pixels := pixelIterator.GetNextIteratorRow()
		if len(pixels) == 0 {
			break
		}

		for _, pixel := range pixels {
			gray := 0
			alpha := pixel.GetAlpha()
			if alpha != 1 {
				if alpha == 0 {
					pixel.SetColor("#FFFFFF")
					gray = 255
				}
				pixel.SetAlpha(1)
			}
			if gray != 255 {
				red := pixel.GetRed() * 255
				green := pixel.GetGreen() * 255
				blue := pixel.GetBlue() * 255
				gray = int(0.2126*red + 0.7152*green + 0.0722*blue)
				hex := fmt.Sprintf("#%02x%02x%02x", gray, gray, gray)
				pixel.SetColor(hex)
			}
			histogram[gray] += 1
			numOfPixels++
		}

		err := pixelIterator.SyncIterator()
		if err != nil {
			return nil, fmt.Errorf("error writing colours back to image when converting to grayscale: %w", err)
		}
	}
	
	sum := 0
	for i := 0; i < 256; i++ {
		sum += i * histogram[i]
	}

	sumBackground, weightBackground, weightForeground, threshold, maxVariance := 0, 0, 0, 0, 0

	for i := 0; i < 256; i++ {
		weightBackground += histogram[i]
		if weightBackground == 0 {
			continue
		}

		weightForeground = numOfPixels - weightBackground
		if weightForeground == 0 {
			continue
		}

		sumBackground += i * histogram[i]

		meanBackground := sumBackground / weightBackground
		meanForeground := (sum - sumBackground) / weightForeground
		betweenVariance := weightBackground * weightForeground * (meanBackground - meanForeground) * (meanBackground - meanForeground)
		if betweenVariance > maxVariance {
			maxVariance = betweenVariance
			threshold = i
		}
	}

	pixelIterator = wand.NewPixelIterator()
	for y := 0; y < int(wand.GetImageHeight()); y++ {
		pixels := pixelIterator.GetNextIteratorRow()
		if len(pixels) == 0 {
			break
		}

		for _, pixel := range pixels {
			red := int(pixel.GetRed() * 255)
			if (args.Invert && red > threshold) || (!args.Invert && red < threshold) {
				pixel.SetColor("#000000")
			} else {
				pixel.SetColor("#FFFFFF")
			}
		}
		err := pixelIterator.SyncIterator()
		if err != nil {
			return nil, fmt.Errorf("error writing colours back to image when converting to black and white: %w", err)
		}
	}

	return []*imagick.MagickWand{wand}, nil
}
