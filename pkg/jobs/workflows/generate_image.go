package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/delivery"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

type GenerateImageArgs struct {
	ImageGen args.ImageGen
	Delivery delivery.Target
}

func GenerateImageWorkflow(ctx workflow.Context, args GenerateImageArgs) (ProcessedImageResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	cancelTyping := startTypingPulse(ctx, args.Delivery)

	jobWorkspace, err := workspace.InitJobWorkspace(workflow.GetInfo(ctx).WorkflowExecution.ID)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error initializing job workspace: %w", err))
		return ProcessedImageResult{}, fmt.Errorf("error initializing job workspace: %w", err)
	}

	var outputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.GenerateImage, jobWorkspace, args.ImageGen).Get(ctx, &outputArtifact)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error generating image: %w", err))
		cleanupWorkspace(ctx, jobWorkspace)
		return ProcessedImageResult{}, fmt.Errorf("error generating image: %w", err)
	}
	resultDelivery := args.Delivery.WithFormat("png")

	cancelTyping()
	if err := sendResult(ctx, jobWorkspace, outputArtifact, resultDelivery); err != nil {
		notifyFailure(ctx, args.Delivery, err)
		return ProcessedImageResult{}, err
	}

	return ProcessedImageResult{
		Image:     outputArtifact,
		Workspace: jobWorkspace,
		Format:    "png",
	}, nil
}
