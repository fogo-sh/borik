package args

type Arcweld struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
}

func (a Arcweld) GetImageURL() string {
	return a.ImageURL
}

func (a Arcweld) ActivityName() string {
	return "Arcweld"
}
