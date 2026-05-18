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
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

type processedImageResult struct {
	Image     workspace.Artifact
	Format    string
	Workspace workspace.Workspace
}

func retrieveJobResult(result processedImageResult) ([]byte, error) {
	image, err := result.Workspace.Retrieve(result.Image)
	if err != nil {
		return nil, fmt.Errorf("error retrieving image: %w", err)
	}

	err = result.Workspace.Cleanup()
	if err != nil {
		log.Error().Err(err).Msg("Error cleaning up workspace")
	}

	return image, nil
}

func (b *Bot) triggerJob(
	ctx context.Context,
	workflowID string,
	imageURL string,
	job args.JobArgs,
) (string, io.Reader, error) {
	we, err := b.temporalClient.ExecuteWorkflow(
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
		}{
			ImageURL:     imageURL,
			ActivityName: job.ActivityName(),
			ActivityArgs: job,
		},
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result processedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := retrieveJobResult(result)
	if err != nil {
		return "", nil, err
	}

	return result.Format, bytes.NewBuffer(image), nil
}

func (b *Bot) triggerGenerateImage(
	ctx context.Context,
	workflowID string,
	imageGenArgs args.ImageGen,
) (string, io.Reader, error) {
	we, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		"GenerateImageWorkflow",
		imageGenArgs,
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result processedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := retrieveJobResult(result)
	if err != nil {
		return "", nil, err
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
		"ConvertVideoToGIFWorkflow",
		gifArgs,
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result processedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := retrieveJobResult(result)
	if err != nil {
		return "", nil, err
	}

	return result.Format, bytes.NewBuffer(image), nil
}

func (b *Bot) triggerAPNGToGIF(
	ctx context.Context,
	workflowID string,
	apngToGIFArgs args.APNGToGIF,
) (string, io.Reader, error) {
	we, err := b.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: config.Instance.TemporalQueueName,
		},
		"ConvertAPNGToGIFWorkflow",
		apngToGIFArgs,
	)
	if err != nil {
		return "", nil, fmt.Errorf("error executing workflow: %w", err)
	}

	var result processedImageResult
	err = we.Get(ctx, &result)
	if err != nil {
		return "", nil, fmt.Errorf("error getting workflow result: %w", err)
	}

	image, err := retrieveJobResult(result)
	if err != nil {
		return "", nil, err
	}

	return result.Format, bytes.NewBuffer(image), nil
}
