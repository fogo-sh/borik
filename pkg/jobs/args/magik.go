package args

type Magik struct {
	ImageURL         string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale            float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
	WidthMultiplier  float64 `default:"0.5" description:"Multiplier to apply to the width of the input image to produce the intermediary image."`
	HeightMultiplier float64 `default:"0.5" description:"Multiplier to apply to the height of the input image to produce the intermediary image."`
}

func (a Magik) GetImageURL() string {
	return a.ImageURL
}

func (a Magik) ActivityName() string {
	return "Magik"
}

type Lagik struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale    float64 `default:"1" description:"Scale of the lagikification. Larger numbers produce more destroyed images."`
}

func (a Lagik) GetImageURL() string {
	return a.ImageURL
}

func (a Lagik) ActivityName() string {
	return "Lagik"
}

type Gmagik struct {
	ImageURL         string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Scale            float64 `default:"1" description:"Scale of the magikification. Larger numbers produce more destroyed images."`
	Iterations       uint    `default:"5" description:"Number of iterations of magikification to run."`
	WidthMultiplier  float64 `default:"0.5" description:"Multiplier to apply to the width of the input image to produce the intermediary image."`
	HeightMultiplier float64 `default:"0.5" description:"Multiplier to apply to the height of the input image to produce the intermediary image."`
}

func (a Gmagik) GetImageURL() string {
	return a.ImageURL
}

func (a Gmagik) ActivityName() string {
	return "Gmagik"
}
