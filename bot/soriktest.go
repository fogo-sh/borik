package bot

import (
	"github.com/bwmarrin/discordgo"
)

var sorikTestSrc = []byte(`
"""
Reimplementation of Borik's magik command as a Sorik script
Usage: sorik run examples/magik.star --args=image_url=DESIRED_IMAGE_URL,scale=INTENSITY
"""

image_url = get_arg("image_url")
scale = float(get_arg("scale"))

image = load_image(image_url)

width, height = image.width, image.height

scaled = liquid_rescale(image, int(width / 2), int(height / 2), deltax=scale)
output = liquid_rescale(scaled, width, height, deltax=scale)
`)

type SorikTestArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image." mapstructure:"image_url"`
	Scale    string `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images." mapstructure:"scale"`
}

func (args SorikTestArgs) GetImageURL() string {
	return args.ImageURL
}

func SorikTest(message *discordgo.MessageCreate, args SorikTestArgs) {
	ExecuteSorikScript("sorik_test", sorikTestSrc, message, args)
}
