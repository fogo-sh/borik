package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

type ProcessImageArgs struct {
	ImageURL     string
	ActivityName string
	ActivityArgs any
}

type ProcessedImageResult struct {
	Image     workspace.Artifact
	Format    string
	Workspace workspace.Workspace
}

func ProcessImageWorkflow(ctx workflow.Context, args ProcessImageArgs) (ProcessedImageResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 1,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	jobWorkspace, err := workspace.InitJobWorkspace(workflow.GetInfo(ctx).WorkflowExecution.ID)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error initializing job workspace: %w", err)
	}

	var inputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.LoadImage, jobWorkspace, args.ImageURL).Get(ctx, &inputArtifact)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error loading image: %w", err)
	}

	var inputFrames []workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.SplitImage, jobWorkspace, inputArtifact).Get(ctx, &inputFrames)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error splitting image: %w", err)
	}

	var futures []workflow.Future
	for _, frame := range inputFrames {
		futures = append(futures, workflow.ExecuteActivity(ctx, args.ActivityName, jobWorkspace, activities.OperationArgs{
			Frame: frame,
			Args:  args.ActivityArgs,
		}))
	}

	var results []workspace.Artifact
	for _, future := range futures {
		var result []workspace.Artifact
		err := future.Get(ctx, &result)
		if err != nil {
			return ProcessedImageResult{}, fmt.Errorf("error executing activity: %w", err)
		}
		results = append(results, result...)
	}

	var outputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.JoinImage, jobWorkspace, results).Get(ctx, &outputArtifact)
	if err != nil {
		return ProcessedImageResult{}, fmt.Errorf("error joining image: %w", err)
	}

	imageFormat := "png"
	if len(results) > 1 {
		imageFormat = "gif"
	}

	return ProcessedImageResult{
		Image:     outputArtifact,
		Workspace: jobWorkspace,
		Format:    imageFormat,
	}, nil
}
