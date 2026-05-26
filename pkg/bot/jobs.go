package bot

import (
	"context"
	"fmt"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/delivery"
)

func (b *Bot) triggerJob(
	ctx context.Context,
	workflowID string,
	imageURL string,
	job args.JobArgs,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		discordWorkflowStartOptions(workflowID),
		"ProcessImageWorkflow",
		struct {
			ImageURL     string
			ActivityName string
			ActivityArgs any
			Delivery     delivery.Target
		}{
			ImageURL:     imageURL,
			ActivityName: job.ActivityName(),
			ActivityArgs: job,
			Delivery:     target,
		},
	)
	return handleExecuteWorkflowError(err)
}

func (b *Bot) triggerGenerateImage(
	ctx context.Context,
	workflowID string,
	imageGenArgs args.ImageGen,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		discordWorkflowStartOptions(workflowID),
		"GenerateImageWorkflow",
		struct {
			ImageGen args.ImageGen
			Delivery delivery.Target
		}{
			ImageGen: imageGenArgs,
			Delivery: target,
		},
	)
	return handleExecuteWorkflowError(err)
}

func (b *Bot) triggerGif(
	ctx context.Context,
	workflowID string,
	gifArgs args.Gif,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		discordWorkflowStartOptions(workflowID),
		"ConvertVideoToGIFWorkflow",
		struct {
			Gif      args.Gif
			Delivery delivery.Target
		}{
			Gif:      gifArgs,
			Delivery: target,
		},
	)
	return handleExecuteWorkflowError(err)
}

func (b *Bot) triggerAPNGToGIF(
	ctx context.Context,
	workflowID string,
	apngToGIFArgs args.APNGToGIF,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		discordWorkflowStartOptions(workflowID),
		"ConvertAPNGToGIFWorkflow",
		struct {
			APNGToGIF args.APNGToGIF
			Delivery  delivery.Target
		}{
			APNGToGIF: apngToGIFArgs,
			Delivery:  target,
		},
	)
	return handleExecuteWorkflowError(err)
}

func discordWorkflowStartOptions(workflowID string) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: config.Instance.TemporalQueueName,

		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,

		WorkflowExecutionErrorWhenAlreadyStarted: true,
	}
}

func handleExecuteWorkflowError(err error) error {
	if err == nil || temporal.IsWorkflowExecutionAlreadyStartedError(err) {
		return nil
	}

	return fmt.Errorf("error executing workflow: %w", err)
}
