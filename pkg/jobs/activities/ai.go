package activities

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/config"
	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

const aiEditMaxDimension uint = 896

type aiParams interface {
	SetExtraFields(map[string]any)
}

func attachAIMetadata(params aiParams, metadata args.AIMetadata) {
	params.SetExtraFields(map[string]any{
		"litellm_session_id": metadata.SessionID,
		"user":               metadata.UserID,
	})
}

func openAIClient() openai.Client {
	return openai.NewClient(
		option.WithAPIKey(config.Instance.OpenaiApiKey),
		option.WithBaseURL(config.Instance.OpenaiBaseUrl),
	)
}

func GenerateImage(ctx context.Context, jobWorkspace workspace.Workspace, imageGenArgs args.ImageGen) (workspace.Artifact, error) {
	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, imageGenArgs.Metadata.Seed)
	finalPrompt := imageGenArgs.Prompt + stableDiffusionOpts

	params := openai.ImageGenerateParams{
		Prompt:         finalPrompt,
		Size:           "512x512",
		Model:          config.Instance.OpenaiImageGenModel,
		ResponseFormat: openai.ImageGenerateParamsResponseFormatB64JSON,
	}
	attachAIMetadata(&params, imageGenArgs.Metadata)

	client := openAIClient()
	image, err := client.Images.Generate(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error generating image: %w", err)
	}

	if len(image.Data) == 0 {
		return "", fmt.Errorf("no image data returned from generation")
	}

	decodedImg, err := base64.StdEncoding.DecodeString(image.Data[0].B64JSON)
	if err != nil {
		return "", fmt.Errorf("error decoding generated image: %w", err)
	}

	artifact, err := jobWorkspace.Persist(decodedImg)
	if err != nil {
		return "", fmt.Errorf("error persisting generated image: %w", err)
	}

	return artifact, nil
}

func editImage(ctx context.Context, wand *imagick.MagickWand, imageEditArgs args.ImageEdit, mask *imagick.MagickWand) (*imagick.MagickWand, error) {
	err := shrinkMaintainAspectRatio(wand, aiEditMaxDimension, aiEditMaxDimension)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %w", err)
	}

	imageBlob, err := wand.GetImageBlob()
	if err != nil {
		return nil, fmt.Errorf("error getting image blob: %w", err)
	}
	imageReader := bytes.NewReader(imageBlob)

	var maskReader io.Reader
	if mask != nil {
		err = shrinkMaintainAspectRatio(mask, aiEditMaxDimension, aiEditMaxDimension)
		if err != nil {
			return nil, fmt.Errorf("error resizing mask: %w", err)
		}

		maskBlob, err := mask.GetImageBlob()
		if err != nil {
			return nil, fmt.Errorf("error getting mask blob: %w", err)
		}
		maskReader = bytes.NewReader(maskBlob)
	}

	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, imageEditArgs.Metadata.Seed)
	finalPrompt := imageEditArgs.Prompt + stableDiffusionOpts

	params := openai.ImageEditParams{
		Image: openai.ImageEditParamsImageUnion{
			OfFileArray: []io.Reader{imageReader},
		},
		Mask:   maskReader,
		Prompt: finalPrompt,
		Model:  config.Instance.OpenaiImageEditModel,
		Size: openai.ImageEditParamsSize(fmt.Sprintf(
			"%dx%d",
			wand.GetImageWidth(),
			wand.GetImageHeight(),
		)),
		ResponseFormat: openai.ImageEditParamsResponseFormatB64JSON,
	}
	attachAIMetadata(&params, imageEditArgs.Metadata)

	client := openAIClient()
	editedImage, err := client.Images.Edit(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error editing image: %w", err)
	}

	if len(editedImage.Data) == 0 {
		return nil, fmt.Errorf("no image data returned from edit")
	}

	decodedImg, err := base64.StdEncoding.DecodeString(editedImage.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("error decoding edited image: %w", err)
	}

	newWand := imagick.NewMagickWand()
	err = newWand.ReadImageBlob(decodedImg)
	if err != nil {
		return nil, fmt.Errorf("error reading edited image blob: %w", err)
	}

	return newWand, nil
}

func ImageEdit(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var imageEditArgs args.ImageEdit
	err = decodeOperationArgs(opArgs, &imageEditArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	editedImage, err := editImage(ctx, wand, imageEditArgs, nil)
	if err != nil {
		return nil, err
	}

	return saveFrames(jobWorkspace, editedImage)
}

func LoopEdit(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var loopEditArgs args.LoopEdit
	err = decodeOperationArgs(opArgs, &loopEditArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	editedFrames := make([]*imagick.MagickWand, 0, loopEditArgs.Steps)

	currentWand := wand
	for range loopEditArgs.Steps {
		loopEditArgs.Metadata.Seed++
		currentWand, err = editImage(ctx, currentWand, args.ImageEdit{
			Prompt:   loopEditArgs.Prompt,
			Metadata: loopEditArgs.Metadata,
		}, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return saveFrames(jobWorkspace, editedFrames...)
}

func FlipFlop(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var flipFlopArgs args.FlipFlop
	err = decodeOperationArgs(opArgs, &flipFlopArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	editedFrames := make([]*imagick.MagickWand, 0, flipFlopArgs.Steps*2+1)
	editedFrames = append(editedFrames, wand)

	currentWand := wand
	for range flipFlopArgs.Steps {
		flipFlopArgs.Metadata.Seed++
		currentWand, err = editImage(ctx, currentWand, args.ImageEdit{
			Prompt:   flipFlopArgs.Prompt1,
			Metadata: flipFlopArgs.Metadata,
		}, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)

		currentWand, err = editImage(ctx, currentWand, args.ImageEdit{
			Prompt:   flipFlopArgs.Prompt2,
			Metadata: flipFlopArgs.Metadata,
		}, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return saveFrames(jobWorkspace, editedFrames...)
}

func performAiZoomStep(ctx context.Context, wand *imagick.MagickWand, prompt string, metadata args.AIMetadata) (*imagick.MagickWand, error) {
	originalWidth := wand.GetImageWidth()
	originalHeight := wand.GetImageHeight()

	sizeMultiplier := 0.8

	if originalWidth < uint(float64(aiEditMaxDimension)*sizeMultiplier) && originalHeight < uint(float64(aiEditMaxDimension)*sizeMultiplier) {
		originalWidth = uint(float64(originalWidth) / sizeMultiplier)
		originalHeight = uint(float64(originalHeight) / sizeMultiplier)
	} else {
		err := shrinkMaintainAspectRatio(wand, uint(float64(wand.GetImageWidth())*sizeMultiplier), uint(float64(wand.GetImageHeight())*sizeMultiplier))
		if err != nil {
			return nil, fmt.Errorf("error resizing image for zoom: %w", err)
		}
	}

	canvas := imagick.NewMagickWand()

	err := canvas.SetFormat("png")
	if err != nil {
		return nil, fmt.Errorf("error setting canvas format for zoom: %w", err)
	}

	err = canvas.NewImage(originalWidth, originalHeight, imagick.NewPixelWand())
	if err != nil {
		return nil, fmt.Errorf("error creating canvas for zoom: %w", err)
	}

	err = canvas.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_ACTIVATE)
	if err != nil {
		return nil, fmt.Errorf("error setting canvas alpha channel for zoom: %w", err)
	}

	err = canvas.CompositeImage(wand, imagick.COMPOSITE_OP_OVER, false, int((originalWidth-wand.GetImageWidth())/2), int((originalHeight-wand.GetImageHeight())/2))
	if err != nil {
		return nil, fmt.Errorf("error compositing image onto canvas for zoom: %w", err)
	}

	mask := canvas.Clone()
	err = mask.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error negating mask for zoom: %w", err)
	}

	editedImage, err := editImage(ctx, canvas, args.ImageEdit{
		Prompt:   prompt,
		Metadata: metadata,
	}, mask)
	if err != nil {
		return nil, err
	}

	return editedImage, nil
}

func AiZoom(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var aiZoomArgs args.AiZoom
	err = decodeOperationArgs(opArgs, &aiZoomArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	for range aiZoomArgs.Steps {
		wand, err = performAiZoomStep(ctx, wand, aiZoomArgs.Prompt, aiZoomArgs.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return saveFrames(jobWorkspace, wand)
}

func AiLoopZoom(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var aiLoopZoomArgs args.AiLoopZoom
	err = decodeOperationArgs(opArgs, &aiLoopZoomArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	editedFrames := make([]*imagick.MagickWand, 0, aiLoopZoomArgs.Steps+1)
	editedFrames = append(editedFrames, wand)

	for range aiLoopZoomArgs.Steps {
		wand = wand.Clone()
		wand, err = performAiZoomStep(ctx, wand, aiLoopZoomArgs.Prompt, aiLoopZoomArgs.Metadata)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, wand)
	}

	minWidth := editedFrames[0].GetImageWidth()
	minHeight := editedFrames[0].GetImageHeight()
	for _, frame := range editedFrames[1:] {
		if frame.GetImageWidth() < minWidth {
			minWidth = frame.GetImageWidth()
		}
		if frame.GetImageHeight() < minHeight {
			minHeight = frame.GetImageHeight()
		}
	}

	for _, frame := range editedFrames {
		err = shrinkMaintainAspectRatio(frame, minWidth, minHeight)
		if err != nil {
			return nil, fmt.Errorf("error resizing frame for loop zoom: %w", err)
		}
	}

	return saveFrames(jobWorkspace, editedFrames...)
}
