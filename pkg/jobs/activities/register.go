package activities

import "go.temporal.io/sdk/worker"

func RegisterActivities(worker worker.Worker) {
	worker.RegisterActivity(LoadImage)
}
