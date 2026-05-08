package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func ConvertVideoToGIFWorkflow(ctx workflow.Context, args args.Gif) (ProcessedImageResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	// TODO: Don't need to use the workspace for this - can use a tmpdir
	jobWorkspace, err := workspace.InitJobWorkspace(workflow.GetInfo(ctx).WorkflowExecution.ID)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error initializing job workspace: %w", err)
	}

	var outputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.ConvertVideoToGIF, jobWorkspace, args).Get(ctx, &outputArtifact)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error converting video to GIF: %w", err)
	}

	return ProcessedImageResult{
		Image:     outputArtifact,
		Workspace: jobWorkspace,
		Format:    "gif",
	}, nil
}
