package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/fogo-sh/borik/pkg/jobs/activities"
	"github.com/fogo-sh/borik/pkg/jobs/delivery"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

type ProcessImageArgs struct {
	ImageURL     string
	ActivityName string
	ActivityArgs any
	Delivery     delivery.Target
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
	cancelTyping := startTypingPulse(ctx, args.Delivery)

	var jobWorkspace workspace.Workspace
	err := workflow.ExecuteActivity(
		ctx,
		activities.InitJobWorkspace,
		workflow.GetInfo(ctx).WorkflowExecution.ID,
	).Get(ctx, &jobWorkspace)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error initializing job workspace: %w", err))
		return ProcessedImageResult{}, fmt.Errorf("error initializing job workspace: %w", err)
	}
	defer cleanupWorkspace(ctx, jobWorkspace)

	var inputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.LoadImage, jobWorkspace, args.ImageURL).Get(ctx, &inputArtifact)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error loading image: %w", err))
		return ProcessedImageResult{}, fmt.Errorf("error loading image: %w", err)
	}

	var inputFrames []workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.SplitImage, jobWorkspace, inputArtifact).Get(ctx, &inputFrames)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error splitting image: %w", err))
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
			cancelTyping()
			notifyFailure(ctx, args.Delivery, fmt.Errorf("error executing activity: %w", err))
			return ProcessedImageResult{}, fmt.Errorf("error executing activity: %w", err)
		}
		results = append(results, result...)
	}

	var outputArtifact workspace.Artifact
	err = workflow.ExecuteActivity(ctx, activities.JoinImage, jobWorkspace, results).Get(ctx, &outputArtifact)
	if err != nil {
		cancelTyping()
		notifyFailure(ctx, args.Delivery, fmt.Errorf("error joining image: %w", err))
		return ProcessedImageResult{}, fmt.Errorf("error joining image: %w", err)
	}

	imageFormat := "png"
	if len(results) > 1 {
		imageFormat = "gif"
	}
	resultDelivery := args.Delivery.WithFormat(imageFormat)

	cancelTyping()
	if err := sendResult(ctx, jobWorkspace, outputArtifact, resultDelivery); err != nil {
		notifyFailure(ctx, args.Delivery, err)
		return ProcessedImageResult{}, err
	}

	return ProcessedImageResult{
		Image:     outputArtifact,
		Workspace: jobWorkspace,
		Format:    imageFormat,
	}, nil
}
