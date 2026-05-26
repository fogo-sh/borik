package args

type Otsu struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Invert   bool   `default:"false" description:"Invert the colors."`
}

func (a Otsu) GetImageURL() string {
	return a.ImageURL
}

func (a Otsu) ActivityName() string {
	return "Otsu"
}
