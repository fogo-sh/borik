package bot

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

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
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
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
	if err != nil {
		return fmt.Errorf("error executing workflow: %w", err)
	}

	return nil
}

func (b *Bot) triggerGenerateImage(
	ctx context.Context,
	workflowID string,
	imageGenArgs args.ImageGen,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		"GenerateImageWorkflow",
		struct {
			ImageGen args.ImageGen
			Delivery delivery.Target
		}{
			ImageGen: imageGenArgs,
			Delivery: target,
		},
	)
	if err != nil {
		return fmt.Errorf("error executing workflow: %w", err)
	}

	return nil
}

func (b *Bot) triggerGif(
	ctx context.Context,
	workflowID string,
	gifArgs args.Gif,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		"ConvertVideoToGIFWorkflow",
		struct {
			Gif      args.Gif
			Delivery delivery.Target
		}{
			Gif:      gifArgs,
			Delivery: target,
		},
	)
	if err != nil {
		return fmt.Errorf("error executing workflow: %w", err)
	}

	return nil
}

func (b *Bot) triggerAPNGToGIF(
	ctx context.Context,
	workflowID string,
	apngToGIFArgs args.APNGToGIF,
	target delivery.Target,
) error {
	_, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		"ConvertAPNGToGIFWorkflow",
		struct {
			APNGToGIF args.APNGToGIF
			Delivery  delivery.Target
		}{
			APNGToGIF: apngToGIFArgs,
			Delivery:  target,
		},
	)
	if err != nil {
		return fmt.Errorf("error executing workflow: %w", err)
	}

	return nil
}
