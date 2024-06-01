package workflows

import "go.temporal.io/sdk/worker"

func RegisterWorkflows(worker worker.Worker) {
	worker.RegisterWorkflow(ProcessImageWorkflow)
}
