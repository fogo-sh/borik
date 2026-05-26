package args

type JobArgs interface {
	GetImageURL() string
	ActivityName() string
}
