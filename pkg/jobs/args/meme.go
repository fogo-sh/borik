package args

type Meme struct {
	Text     string `description:"Meme text. Use | to separate top and bottom text."`
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Meme) GetImageURL() string {
	return a.ImageURL
}

func (a Meme) ActivityName() string {
	return "Meme"
}
