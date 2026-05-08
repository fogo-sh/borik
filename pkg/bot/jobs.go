package bot

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go.temporal.io/sdk/client"

	"github.com/rs/zerolog/log"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workflows"
)

func (b *Bot) triggerJob(ctx context.Context, workflowID string, imageURL string, job args.JobArgs) (string, io.Reader, error) {
	we, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		workflows.ProcessImageWorkflow,
		workflows.ProcessImageArgs{
			ImageURL:     imageURL,
			ActivityName: job.ActivityName(),
			ActivityArgs: job,
		},
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result workflows.ProcessedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := result.Workspace.Retrieve(result.Image)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving image: %w", err)
	}

	err = result.Workspace.Cleanup()
	if err != nil {
		log.Error().Err(err).Msg("Error cleaning up workspace")
	}

	return result.Format, bytes.NewBuffer(image), nil
}

func (b *Bot) triggerGenerateImage(ctx context.Context, workflowID string, imageGenArgs args.ImageGen) (string, io.Reader, error) {
	we, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		workflows.GenerateImageWorkflow,
		imageGenArgs,
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result workflows.ProcessedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := result.Workspace.Retrieve(result.Image)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving image: %w", err)
	}

	err = result.Workspace.Cleanup()
	if err != nil {
		log.Error().Err(err).Msg("Error cleaning up workspace")
	}

	return result.Format, bytes.NewBuffer(image), nil
}

func (b *Bot) triggerGif(ctx context.Context, workflowID string, gifArgs args.Gif) (string, io.Reader, error) {
	we, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		workflows.ConvertVideoToGIFWorkflow,
		gifArgs,
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result workflows.ProcessedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := result.Workspace.Retrieve(result.Image)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving GIF: %w", err)
	}

	err = result.Workspace.Cleanup()
	if err != nil {
		log.Error().Err(err).Msg("Error cleaning up workspace")
	}

	return result.Format, bytes.NewBuffer(image), nil
}
