package args

type Modulate struct {
	ImageURL   string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Brightness float64 `default:"100" description:"Percent change in brightness. Numbers > 100 increase brightness, < 100 decreases."`
	Saturation float64 `default:"100" description:"Percent change in saturation. Numbers > 100 increase saturation, < 100 decreases."`
	Hue        float64 `default:"100" description:"Percent change in hue. Numbers > 100 rotates hue clockwise, < 100 rotates counter-clockwise."`
}

func (a Modulate) GetImageURL() string {
	return a.ImageURL
}

func (a Modulate) ActivityName() string {
	return "Modulate"
}
