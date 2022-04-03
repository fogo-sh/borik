package bot

import (
	_ "embed"
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/gographics/imagick.v2/imagick"
)

//go:embed steve_point.png
var stevePointImage []byte

type _StevePointArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Flip     bool   `default:"false" description:"Have Steve pointing from the left side of the image, rather than the right side."`
}

func (args _StevePointArgs) GetImageURL() string {
	return args.ImageURL
}

func StevePoint(srcBytes []byte, destBuffer io.Writer, args _StevePointArgs) error {
	steve := imagick.NewMagickWand()
	err := steve.ReadImageBlob(stevePointImage)
	if err != nil {
		return fmt.Errorf("error reading steve: %w", err)
	}

	if args.Flip {
		err = steve.FlopImage()
		if err != nil {
			return fmt.Errorf("error flipping steve: %w", err)
		}
	}

	wand := imagick.NewMagickWand()
	err = wand.ReadImageBlob(srcBytes)
	if err != nil {
		return fmt.Errorf("error reading input image: %w", err)
	}

	inputHeight := wand.GetImageHeight()
	inputWidth := wand.GetImageWidth()

	steve = steve.TransformImage("", fmt.Sprintf("%dx%d", inputWidth, inputHeight))

	steveWidth := steve.GetImageWidth()

	var xOffset int
	if args.Flip {
		xOffset = 0
	} else {
		xOffset = int(inputWidth - steveWidth)
	}

	err = wand.CompositeImage(steve, imagick.COMPOSITE_OP_ATOP, xOffset, 0)
	if err != nil {
		return fmt.Errorf("error compositing image: %w", err)
	}

	_, err = destBuffer.Write(wand.GetImageBlob())
	if err != nil {
		return fmt.Errorf("error writing output image: %w", err)
	}
	return nil
}

func _StevePointCommand(message *discordgo.MessageCreate, args _StevePointArgs) {
	defer TypingIndicator(message)()

	PrepareAndInvokeOperation(message, args, StevePoint)
}
