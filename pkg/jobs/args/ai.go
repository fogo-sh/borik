package args

type AIMetadata struct {
	Seed      int
	SessionID string
	UserID    string
}

type ImageGen struct {
	Prompt   string `description:"Prompt to generate an image for."`
	Metadata AIMetadata
}

func (a ImageGen) WithMetadata(metadata AIMetadata) ImageGen {
	a.Metadata = metadata
	return a
}

type ImageEdit struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Metadata AIMetadata
}

func (a ImageEdit) GetImageURL() string {
	return a.ImageURL
}

func (a ImageEdit) ActivityName() string {
	return "ImageEdit"
}

func (a ImageEdit) WithMetadata(metadata AIMetadata) ImageEdit {
	a.Metadata = metadata
	return a
}

type LoopEdit struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
	Metadata AIMetadata
}

func (a LoopEdit) GetImageURL() string {
	return a.ImageURL
}

func (a LoopEdit) ActivityName() string {
	return "LoopEdit"
}

func (a LoopEdit) WithMetadata(metadata AIMetadata) LoopEdit {
	a.Metadata = metadata
	return a
}

type FlipFlop struct {
	Prompt1  string `description:"First prompt to edit the image with."`
	Prompt2  string `description:"Second prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
	Metadata AIMetadata
}

func (a FlipFlop) GetImageURL() string {
	return a.ImageURL
}

func (a FlipFlop) ActivityName() string {
	return "FlipFlop"
}

func (a FlipFlop) WithMetadata(metadata AIMetadata) FlipFlop {
	a.Metadata = metadata
	return a
}

type AiZoom struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
	Steps    uint   `default:"2" description:"Number of zoom steps to perform."`
	Metadata AIMetadata
}

func (a AiZoom) GetImageURL() string {
	return a.ImageURL
}

func (a AiZoom) ActivityName() string {
	return "AiZoom"
}

func (a AiZoom) WithMetadata(metadata AIMetadata) AiZoom {
	a.Metadata = metadata
	return a
}

type AiLoopZoom struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
	Steps    uint   `default:"5" description:"Number of zoom steps to perform."`
	Metadata AIMetadata
}

func (a AiLoopZoom) GetImageURL() string {
	return a.ImageURL
}

func (a AiLoopZoom) ActivityName() string {
	return "AiLoopZoom"
}

func (a AiLoopZoom) WithMetadata(metadata AIMetadata) AiLoopZoom {
	a.Metadata = metadata
	return a
}
