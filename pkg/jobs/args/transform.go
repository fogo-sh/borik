package args

type Rotate struct {
	ImageURL string  `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Degrees  float64 `default:"90" description:"Number of degrees to rotate the image by."`
}

func (a Rotate) GetImageURL() string {
	return a.ImageURL
}

func (a Rotate) ActivityName() string {
	return "Rotate"
}

type Resize struct {
	Width    float64 `description:"Width in pixels (absolute) or percent (e.g. 150 = 150%)."`
	Height   float64 `description:"Height in pixels (absolute) or percent (e.g. 150 = 150%)."`
	ImageURL string  `default:"" description:"Image URL to process. Leave blank to auto-find."`
	Mode     string  `default:"percent" description:"Resize mode (percent/absolute) for width/height values."`
}

func (a Resize) GetImageURL() string {
	return a.ImageURL
}

func (a Resize) ActivityName() string {
	return "Resize"
}
