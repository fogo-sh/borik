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

type ConvertVideoToGIFArgs struct {
	Gif      args.Gif
	Delivery delivery.Target
}

func ConvertVideoToGIFWorkflow(ctx workflow.Context, args ConvertVideoToGIFArgs) (ProcessedImageResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	cancelTyping := startTypingPulse(ctx, args.Delivery)

	// TODO: Don't need to use the workspace for this - can use a tmpdir
	jobWorkspace, err := workspace.InitJobWorkspace(workflow.GetInfo(ctx).WorkflowExecution.ID)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error initializing job workspace: %w", err))
		return ProcessedImageResult{}, fmt.Errorf("error initializing job workspace: %w", err)
	}

	var outputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.ConvertVideoToGIF, jobWorkspace, args.Gif).Get(ctx, &outputArtifact)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error converting video to GIF: %w", err))
		cleanupWorkspace(ctx, jobWorkspace)
		return ProcessedImageResult{}, fmt.Errorf("error converting video to GIF: %w", err)
	}
	resultDelivery := args.Delivery.WithFormat("gif")

	cancelTyping()
	if err := sendResult(ctx, jobWorkspace, outputArtifact, resultDelivery); err != nil {
		notifyFailure(ctx, args.Delivery, err)
		return ProcessedImageResult{}, err
	}

	return ProcessedImageResult{
		Image:     outputArtifact,
		Workspace: jobWorkspace,
		Format:    "gif",
	}, nil
}
