package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	imagick6 "gopkg.in/gographics/imagick.v2/imagick"
)

type graphicsFormat struct {
	Name       string
	Palette    []string
	Resolution [2]uint
}

var cgaPalette = []string{
	"#000000",
	"#0000AA",
	"#00AA00",
	"#00AAAA",
	"#AA0000",
	"#AA00AA",
	"#AA5500",
	"#AAAAAA",
	"#555555",
	"#5555FF",
	"#55FF55",
	"#55FFFF",
	"#FF5555",
	"#FF55FF",
	"#FFFF55",
	"#FFFFFF",
}

var graphicsFormats = []graphicsFormat{
	{
		Name:       "EGA",
		Palette:    cgaPalette,
		Resolution: [2]uint{640, 350},
	},
	{
		Name:       "TempleOS",
		Palette:    cgaPalette,
		Resolution: [2]uint{640, 480},
	},
	{
		Name:       "CGA",
		Palette:    cgaPalette,
		Resolution: [2]uint{160, 100},
	},
}

func getPaletteImage(palette []string) (*imagick6.MagickWand, error) {
	paletteWand := imagick6.NewMagickWand()
	err := paletteWand.SetSize(uint(len(palette)), 1)
	if err != nil {
		return nil, fmt.Errorf("error resizing palette image: %w", err)
	}

	err = paletteWand.ReadImage("xc:BLACK")
	if err != nil {
		return nil, fmt.Errorf("error loading default image: %w", err)
	}

	pixelIterator := paletteWand.NewPixelIterator()

	for index, pixel := range pixelIterator.GetCurrentIteratorRow() {
		pixel.SetColor(palette[index])
	}

	err = pixelIterator.SyncIterator()
	if err != nil {
		return nil, fmt.Errorf("error writing colours back to image: %w", err)
	}

	return paletteWand, nil
}

func convertGraphicsFormat(wand *imagick6.MagickWand, format graphicsFormat, dither bool) ([]*imagick6.MagickWand, error) {
	paletteWand, err := getPaletteImage(format.Palette)
	if err != nil {
		return nil, fmt.Errorf("error getting format palette: %w", err)
	}

	wand = wand.TransformImage("", fmt.Sprintf("%d!x%d!", format.Resolution[0], format.Resolution[1]))

	ditherMethod := imagick6.DITHER_METHOD_NO
	if dither {
		ditherMethod = imagick6.DITHER_METHOD_FLOYD_STEINBERG
	}

	err = wand.RemapImage(paletteWand, ditherMethod)
	if err != nil {
		return nil, fmt.Errorf("error remapping image palette: %w", err)
	}

	return []*imagick6.MagickWand{wand}, nil
}

type graphicsFormatArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Dither   bool   `default:"false" description:"Whether the final image should be dithered."`
}

func (args graphicsFormatArgs) GetImageURL() string {
	return args.ImageURL
}

func MakeGraphicsFormatOpCommand(format graphicsFormat) func(*discordgo.MessageCreate, graphicsFormatArgs) {
	return MakeImageOpCommand(func(wand *imagick6.MagickWand, args graphicsFormatArgs) ([]*imagick6.MagickWand, error) {
		return convertGraphicsFormat(wand, format, args.Dither)
	})
}

func registerGraphicsFormatCommands(parser *parsley.Parser) {
	for _, format := range graphicsFormats {
		_ = parser.NewCommand(
			strings.ToLower(format.Name),
			fmt.Sprintf("Convert an image to %s graphics", format.Name),
			MakeGraphicsFormatOpCommand(format),
		)
	}
}
