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
	worker.RegisterActivity(Deepfry)
	worker.RegisterActivity(Divine)
}
