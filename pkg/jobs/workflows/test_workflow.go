package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
)

func TestWorkflow(ctx workflow.Context, imageUrl string) ([]byte, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var result []byte
	err := workflow.ExecuteActivity(ctx, activities.LoadImage, imageUrl).Get(ctx, &result)

	return result, err
}
