package bot

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/config"
)

var AI_EDIT_MAX_DIMENSION uint = 896

// Prompt upscaler system prompts from flux-3-klein-image-pup - https://huggingface.co/black-forest-labs/flux-3-image-klein-pup/blob/main/system_prompts.json
var AI_T2I_UPSCALE_PROMPT string = "You are the prompt upsampler of a text-to-image model. Expand the user's short request into ONE detailed, self-contained caption: the exact text the image model receives. Output only the caption, with no greetings, explanations, refusals, questions, Markdown, headings, JSON, coordinates, ids or angle-bracket tokens, and no quotes around it. If the input is not clearly an image request, caption a visualization of its concept. The request may begin with a \"TARGET CANVAS: W:H\" line: compose for that canvas (wide spreads content sideways, tall stacks it) and never mention the ratio or the canvas.\n\nFidelity and planning: keep every explicit subject, attribute, count, spatial constraint, style cue and piece of wording; never contradict, substitute, drop or narrow the request, and keep an abstract concept abstract. Decide everything left open as ONE concrete, consistent plan; never offer alternatives or hedge. Never re-express a stated constraint through another mechanism; add camera optics (depth of field, bokeh, blur) only when asked for or implied. Without a stated style, plan an actual photograph and open with the medium convention (\"Full-color outdoor photograph, ...\", \"Full-color indoor photograph, ...\", \"Studio photograph, ...\", \"Close-up photograph, ...\"); keep \"Photorealistic\" for renders, and turn \"photorealistic\" or \"realistic\" requests for photos into the photograph convention. Give a photograph a few ordinary, non-ideal specifics drawn from the scene's own objects instead of the stock-photo ideal. Do not invent names, brands, logos or celebrities; keep user-named ones. A long request keeps its content but not its wording: drop appraisal and sales vocabulary, turn moods into observable physical facts, and add only new concrete content with positions. Time of day is lighting, not rendered text.\n\nCaption format: one continuous prose block in the generative voice, never \"this image shows\", \"we can see\" or \"is visible\". Lead with medium and style, then scene, lighting and camera, then the subjects with concrete colors, materials and state. Every person in focus gets an observable profile: ethnicity, gender, age range, skin tone (\"fair skin\", \"brown skin\"), hair and clothing. Make the layout reconstructable from words: frame-relative positions (frame-left, frame-right, upper, lower, foreground, midground, background), sentence-initial anchors (\"On the left, ...\", \"In the background, ...\"), relations (holds, faces, behind) and a consistent order; describe up to about ten instances one by one, beyond that give the exact count and arrangement. Out-of-focus or distant planes are shapes, colors and silhouettes, never identities or readable text, and the background takes about a sixth of the caption. Be definite: no hedging adverbs, no quality or emotion judgments, no closing sentence about mood; state negations as what is present. Rendered text: transcribe every wording verbatim inside double quotes in its original language and script, keep the line breaks of multi-line text, and give its typography, color and placement; author concrete, domain-correct wording when the request needs text it does not supply, and name canonical structures (calendar, keyboard, periodic table) exactly with their full layout. All prose is English; only quoted text keeps its language. Typical length is 100 to 350 words, more for complex scenes.\n\nOutput only the caption."
var AI_I2I_UPSCALE_PROMPT string = "You are the prompt upsampler of an image-editing model. You see the reference image(s), attached in order (the first image, the second image, ...), and the user's short request. Expand it into ONE detailed, self-contained instruction: the caption the editing model receives together with the same images. Output only the instruction as one continuous English prose block, with no greetings, explanations, Markdown, JSON, coordinates, ids or angle-bracket tokens. Read a vague or conversational request as the most plausible edit of the supplied images.\n\nPick the voice by task shape. When one reference stays and gets bounded changes, write an imperative editing instruction on it (\"Remove the ...\", \"Change the ... to ...\", \"Zoom in on ...\"); a whole-image transformation opens with its scope (\"Convert the entire image into ...\", \"Apply a ... filter over the entire scene\", \"Horizontally flip the entire image so that ...\", \"Shift the camera viewpoint to ...\", then where the major elements should now be). When the references are drawn into a newly composed frame (combining subjects, a new scene, a garment, object or style donor), write a generation request (\"Create a photograph of ...\", \"Create an anime-style illustration of ...\", \"Generate a new image featuring ...\") that says what to take from each image, by ordinal, and what is newly generated.\n\nKeep every explicit edit, constraint, count, position, style cue and piece of wording, and decide what is left open as one concrete plan. Ground everything in what is actually visible: identify each referent unambiguously by its real appearance, never mention content that is not there, and use every image the user attached. State every end state concretely (which color, how large, which pose, which material) and give new content enough appearance to be generated. Cover the knock-on effects of the change (shadows, reflections, occlusion, perspective, scale, lighting) and say what fills a removed region. A zoom-in names what stays in frame and which neighbors are cropped out; an outpainting continues the scene in the same medium. Describe the change, not the whole scene, and close an editing instruction with one compact clause naming what stays unchanged; transferred people keep their identity, face and accessories unless asked otherwise. Transcribe on-image text verbatim in its original language inside double quotes, with line breaks as \\n, and match the original font, size and color when replacing it. Name recognizable public figures, brands and landmarks you can see. Directions are frame-relative; use no evaluative or quality words. Typical length is 50 to 150 words.\n\nOutput only the instruction."

type AISessionMetadata struct {
	Seed      int
	SessionID string
	UserID    string
}

type AIParams interface {
	SetExtraFields(map[string]any)
}

func attachSessionMetadata(params AIParams, metadata AISessionMetadata) {
	params.SetExtraFields(map[string]any{
		"litellm_session_id": metadata.SessionID,
		"user":               metadata.UserID,
	})
}

type ImageGenArgs struct {
	Prompt        string `description:"Prompt to generate an image for."`
	UpscalePrompt bool   `description:"Whether to upscale your prompt before generating the image." default:"true"`
}

func generateImage(ctx *OperationContext, args ImageGenArgs) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	metadata := AISessionMetadata{
		SessionID: ctx.GetSourceID(),
		UserID:    ctx.GetUserID(),
	}

	if args.UpscalePrompt {
		upscalerParams := responses.ResponseNewParams{
			Instructions: param.Opt[string]{Value: AI_T2I_UPSCALE_PROMPT},
			Input: responses.ResponseNewParamsInputUnion{
				OfString: param.Opt[string]{Value: args.Prompt},
			},
			Model: Instance.config.OpenaiTextModel,
		}

		attachSessionMetadata(&upscalerParams, metadata)

		resp, err := Instance.openAiClient.Responses.New(
			context.Background(),
			upscalerParams,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate upscaled prompt")
			return
		}

		args.Prompt = resp.OutputText()
	}

	seed := rand.Int()
	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	params := openai.ImageGenerateParams{
		Prompt:         finalPrompt,
		Size:           "512x512",
		Model:          config.Instance.OpenaiImageGenModel,
		ResponseFormat: openai.ImageGenerateParamsResponseFormatB64JSON,
	}
	attachSessionMetadata(&params, metadata)

	image, err := Instance.openAiClient.Images.Generate(
		context.TODO(),
		params,
	)
	if err != nil {
		if sendErr := ctx.SendText("Error generating image: `" + err.Error() + "`"); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error response")
		}
		return
	}

	file := &discordgo.File{
		Name:        "generated.png",
		ContentType: "image/png",
		Reader:      base64.NewDecoder(base64.StdEncoding, strings.NewReader(image.Data[0].B64JSON)),
	}

	if err := ctx.SendFiles([]*discordgo.File{file}); err != nil {
		log.Error().Err(err).Msg("Failed to send generated image")
	}
}

func ImageGenTextCommand(message *discordgo.MessageCreate, args ImageGenArgs) {
	generateImage(NewOperationContextFromMessage(Instance.session, message), args)
}

func ImageGenSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args ImageGenArgs) {
	generateImage(NewOperationContextFromInteraction(session, interaction), args)
}

func editImage(
	wand *imagick.MagickWand,
	args ImageEditArgs,
	metadata AISessionMetadata,
	mask *imagick.MagickWand,
) (*imagick.MagickWand, error) {
	err := ShrinkMaintainAspectRatio(wand, AI_EDIT_MAX_DIMENSION, AI_EDIT_MAX_DIMENSION)
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
		err = ShrinkMaintainAspectRatio(mask, AI_EDIT_MAX_DIMENSION, AI_EDIT_MAX_DIMENSION)
		if err != nil {
			return nil, fmt.Errorf("error resizing mask: %w", err)
		}

		maskBlob, err := mask.GetImageBlob()
		if err != nil {
			return nil, fmt.Errorf("error getting mask blob: %w", err)
		}
		maskReader = bytes.NewReader(maskBlob)
	}

	stableDiffusionOpts := fmt.Sprintf(`<sd_cpp_extra_args>{"seed": %d}</sd_cpp_extra_args>`, metadata.Seed)
	finalPrompt := args.Prompt + stableDiffusionOpts

	params := openai.ImageEditParams{
		Image: openai.ImageEditParamsImageUnion{
			OfFileArray: []io.Reader{imageReader},
		},
		Mask:   maskReader,
		Prompt: finalPrompt,
		Model:  config.Instance.OpenaiImageEditModel,
		// OpenAI's models require one of a few specific sizes, but stable-diffusion.cpp is more flexible
		// Pass the original image size to prevent it cropping it
		Size: openai.ImageEditParamsSize(fmt.Sprintf(
			"%dx%d",
			wand.GetImageWidth(),
			wand.GetImageHeight(),
		)),
		ResponseFormat: openai.ImageEditParamsResponseFormatB64JSON,
	}
	attachSessionMetadata(&params, metadata)

	editedImage, err := Instance.openAiClient.Images.Edit(
		context.TODO(),
		params,
	)
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

type ImageEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
}

func (args ImageEditArgs) GetImageURL() string {
	return args.ImageURL
}

func ImageEdit(
	wand *imagick.MagickWand,
	args ImageEditArgs,
	metadata AISessionMetadata,
) ([]*imagick.MagickWand, error) {
	editedImage, err := editImage(wand, args, metadata, nil)
	if err != nil {
		return nil, err
	}

	return []*imagick.MagickWand{editedImage}, nil
}

type LoopEditArgs struct {
	Prompt   string `description:"Prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
}

func (args LoopEditArgs) GetImageURL() string {
	return args.ImageURL
}

func LoopEdit(wand *imagick.MagickWand, args LoopEditArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps)

	currentWand := wand
	var err error
	for range args.Steps {
		metadata.Seed++ // Increment the seed for each iteration to produce different results
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt,
		}, metadata, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
}

type FlipFlopArgs struct {
	Prompt1  string `description:"First prompt to edit the image with."`
	Prompt2  string `description:"Second prompt to edit the image with."`
	ImageURL string `default:"" description:"URL of the image to edit."`
	Steps    uint   `default:"4" description:"Number of edit iterations to perform."`
}

func (args FlipFlopArgs) GetImageURL() string {
	return args.ImageURL
}

func FlipFlop(wand *imagick.MagickWand, args FlipFlopArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps*2+1)

	editedFrames = append(editedFrames, wand)

	currentWand := wand
	var err error
	for range args.Steps {
		metadata.Seed++ // Increment the seed for each iteration to produce different results
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt1,
		}, metadata, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
		currentWand, err = editImage(currentWand, ImageEditArgs{
			Prompt: args.Prompt2,
		}, metadata, nil)
		if err != nil {
			return nil, err
		}
		editedFrames = append(editedFrames, currentWand)
	}

	return editedFrames, nil
}

type AiZoomArgs struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
	Steps    uint   `default:"2" description:"Number of zoom steps to perform."`
}

func (args AiZoomArgs) GetImageURL() string {
	return args.ImageURL
}

func performAiZoomStep(
	wand *imagick.MagickWand,
	prompt string,
	metadata AISessionMetadata,
) (*imagick.MagickWand, error) {
	originalWidth := wand.GetImageWidth()
	originalHeight := wand.GetImageHeight()

	sizeMultiplier := 0.8

	var err error

	// If the image can be used as-is without the target canvas exceeding the max size for resizing, use it as-is.
	// Otherwise, shrink it.
	maxOriginalDimension := uint(float64(AI_EDIT_MAX_DIMENSION) * sizeMultiplier)
	if originalWidth < maxOriginalDimension && originalHeight < maxOriginalDimension {
		originalWidth = uint(float64(originalWidth) / sizeMultiplier)
		originalHeight = uint(float64(originalHeight) / sizeMultiplier)
	} else {
		err = ShrinkMaintainAspectRatio(
			wand,
			uint(float64(wand.GetImageWidth())*sizeMultiplier),
			uint(float64(wand.GetImageHeight())*sizeMultiplier),
		)
		if err != nil {
			return nil, fmt.Errorf("error resizing image for zoom: %w", err)
		}
	}

	// Place the shrunk image in the center of a transparent canvas of the original size
	canvas := imagick.NewMagickWand()

	err = canvas.SetFormat("png")
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

	err = canvas.CompositeImage(
		wand,
		imagick.COMPOSITE_OP_OVER,
		false,
		int((originalWidth-wand.GetImageWidth())/2),
		int((originalHeight-wand.GetImageHeight())/2),
	)
	if err != nil {
		return nil, fmt.Errorf("error compositing image onto canvas for zoom: %w", err)
	}

	// Flip alpha of the canvas to create a mask for outpainting the edges
	mask := canvas.Clone()
	err = mask.NegateImage(false)
	if err != nil {
		return nil, fmt.Errorf("error negating mask for zoom: %w", err)
	}

	editedImage, err := editImage(canvas, ImageEditArgs{
		Prompt: prompt,
	}, metadata, mask)
	if err != nil {
		return nil, err
	}

	return editedImage, nil
}

func AiZoom(wand *imagick.MagickWand, args AiZoomArgs, metadata AISessionMetadata) ([]*imagick.MagickWand, error) {
	var err error

	for range args.Steps {
		wand, err = performAiZoomStep(wand, args.Prompt, metadata)
		if err != nil {
			return nil, err
		}
	}

	return []*imagick.MagickWand{wand}, nil
}

type AiLoopZoomArgs struct {
	ImageURL string `default:"" description:"URL of the image to edit."`
	Prompt   string `default:"Expand the image outwards." description:"Prompt to edit the image with."`
	Steps    uint   `default:"5" description:"Number of zoom steps to perform."`
}

func (args AiLoopZoomArgs) GetImageURL() string {
	return args.ImageURL
}

func AiLoopZoom(
	wand *imagick.MagickWand,
	args AiLoopZoomArgs,
	metadata AISessionMetadata,
) ([]*imagick.MagickWand, error) {
	editedFrames := make([]*imagick.MagickWand, 0, args.Steps+1)

	editedFrames = append(editedFrames, wand)

	var err error

	for range args.Steps {
		wand = wand.Clone()
		wand, err = performAiZoomStep(wand, args.Prompt, metadata)
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
		err = ShrinkMaintainAspectRatio(frame, minWidth, minHeight)
		if err != nil {
			return nil, fmt.Errorf("error resizing frame for loop zoom: %w", err)
		}
	}

	return editedFrames, nil
}
