package bot

import (
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v2/imagick"
)

// Magik runs content-aware scaling on an image.
func Magik(src []byte, dest io.Writer) error {
	wand := imagick.NewMagickWand()
	wand.ReadImageBlob(src)

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()

	log.Debug().
		Uint("src_width", width).
		Uint("src_height", height).
		Uint("dest_width", width/2).
		Uint("dest_height", height/2).
		Msg("Liquid rescaling image")
	wand.LiquidRescaleImage(uint(width/2), uint(height/2), 1, 0)

	log.Debug().
		Uint("dest_width", width).
		Uint("dest_height", height).
		Uint("src_width", width/2).
		Uint("src_height", height/2).
		Msg("Returning image to original size")
	err := wand.ResizeImage(width, height, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		return fmt.Errorf("error while attempting to resize image: %w", err)
	}
	dest.Write(wand.GetImageBlob())

	return nil
}
