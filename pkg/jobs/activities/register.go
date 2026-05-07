package activities

import "go.temporal.io/sdk/worker"

func RegisterActivities(worker worker.Worker) {
	worker.RegisterActivity(LoadImage)
	worker.RegisterActivity(SplitImage)
	worker.RegisterActivity(JoinImage)

	worker.RegisterActivity(Magik)
	worker.RegisterActivity(Lagik)
	worker.RegisterActivity(Gmagik)
	worker.RegisterActivity(Arcweld)
	worker.RegisterActivity(Malt)
	worker.RegisterActivity(Deepfry)
	worker.RegisterActivity(Divine)
	worker.RegisterActivity(PresidentsFrame)
	worker.RegisterActivity(Heritage)
	worker.RegisterActivity(Shinji)
	worker.RegisterActivity(Waaw)
	worker.RegisterActivity(Haah)
	worker.RegisterActivity(Woow)
	worker.RegisterActivity(Hooh)
	worker.RegisterActivity(Invert)
	worker.RegisterActivity(Otsu)
	worker.RegisterActivity(Rotate)
	worker.RegisterActivity(Resize)
}
