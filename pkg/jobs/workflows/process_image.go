package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
)

type ProcessImageArgs struct {
	ImageURL     string
	ActivityName string
	ActivityArgs any
}

type ProcessedImageResult struct {
	Image  []byte
	Format string
}

func ProcessImageWorkflow(ctx workflow.Context, args ProcessImageArgs) (ProcessedImageResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	var imageBytes []byte
	err := workflow.ExecuteActivity(ctx, activities.LoadImage, args.ImageURL).Get(ctx, &imageBytes)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error loading image: %w", err)
	}

	var frames [][]byte
	err = workflow.ExecuteActivity(ctx, activities.SplitImage, imageBytes).Get(ctx, &frames)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error splitting image: %w", err)
	}

	var futures []workflow.Future
	for _, frame := range frames {
		futures = append(futures, workflow.ExecuteActivity(ctx, args.ActivityName, activities.OperationArgs{
			Frame: frame,
			Args:  args.ActivityArgs,
		}))
	}

	var results [][]byte
	for _, future := range futures {
		var result [][]byte
		err := future.Get(ctx, &result)
		if err != nil {
			return ProcessedImageResult{}, fmt.Errorf("error executing activity: %w", err)
		}
		results = append(results, result...)
	}

	var outputBytes []byte
	err = workflow.ExecuteActivity(ctx, activities.JoinImage, results).Get(ctx, &outputBytes)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error joining image: %w", err)
	}

	imageFormat := "png"
	if len(results) > 1 {
		imageFormat = "gif"
	}

	return ProcessedImageResult{
		Image:  outputBytes,
		Format: imageFormat,
	}, nil
}
