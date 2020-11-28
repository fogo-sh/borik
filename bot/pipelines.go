package bot

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// ErrNoPendingPipeline occurs when attempting to add a step to a command pipeline for a user with no pending pipeline.
var ErrNoPendingPipeline = errors.New("you do not have a pending pipeline - use the createpipeline command to begin creating one")

// ErrPendingPipelineExists occurs when attempting to create a new pipeline for a user with an already existing pending pipeline.
var ErrPendingPipelineExists = errors.New("you already have an existing pending pipeline")

// ErrPipelineDoesNotExist occurs when attempting to access a pipeline that does not exist.
var ErrPipelineDoesNotExist = errors.New("specified pipeline does not exist")

// ErrInvalidPipelineName occurs when a user specifies an invalid pipeline name while saving.
var ErrInvalidPipelineName = errors.New("invalid pipeline name")

// ErrPipelineAlreadyExists occurs when a user attempts to save a pipeline with an already used name
var ErrPipelineAlreadyExists = errors.New("specified pipeline already exists")

// PipelineEntry represents a single entry in a command pipeline.
type PipelineEntry struct {
	Operation string
	Args      interface{}
}

// SavedPipeline represents a user-created pipeline.
type SavedPipeline struct {
	Creator  string
	Pipeline []PipelineEntry
}

// PipelineManager manages the saving, creation, and execution of command pipelines.
type PipelineManager struct {
	SavedPipelines   map[string]SavedPipeline
	PendingPipelines map[string][]PipelineEntry
}

// CreatePipeline creates a pending pipeline for a given user.
func (manager *PipelineManager) CreatePipeline(message *discordgo.MessageCreate) error {
	_, found := manager.PendingPipelines[message.Author.ID]
	if found {
		return ErrPendingPipelineExists
	}
	manager.PendingPipelines[message.Author.ID] = make([]PipelineEntry, 0)
	return nil
}

// AddStep adds a step to a user's pending command pipeline.
func (manager *PipelineManager) AddStep(message *discordgo.MessageCreate, operation string, args interface{}) error {
	pipeline, ok := manager.PendingPipelines[message.Author.ID]
	if !ok {
		return ErrNoPendingPipeline
	}

	manager.PendingPipelines[message.Author.ID] = append(pipeline, PipelineEntry{operation, args})

	return nil
}

// DeletePipeline deletes a pipeline.
func (manager *PipelineManager) DeletePipeline(message *discordgo.MessageCreate, name string) error {
	if name == "pending" {
		_, found := manager.PendingPipelines[message.Author.ID]
		if !found {
			return ErrPipelineDoesNotExist
		}
		delete(manager.PendingPipelines, message.Author.ID)
	}
	return nil
}

// GetPipeline returns a currently stored pipeline.
func (manager *PipelineManager) GetPipeline(message *discordgo.MessageCreate, name string) ([]PipelineEntry, error) {
	if name == "pending" {
		entry, found := manager.PendingPipelines[message.Author.ID]
		if !found {
			return make([]PipelineEntry, 0), ErrPipelineDoesNotExist
		}
		return entry, nil
	}
	savedPipeline, found := manager.SavedPipelines[name]
	if !found {
		return make([]PipelineEntry, 0), ErrPipelineDoesNotExist
	}
	return savedPipeline.Pipeline, nil
}

// SavePipeline saves a pending pipeline.
func (manager *PipelineManager) SavePipeline(message *discordgo.MessageCreate, name string) error {
	if name == "pending" {
		return ErrInvalidPipelineName
	}

	pendingPipeline, found := manager.PendingPipelines[message.Author.ID]
	if !found {
		return ErrNoPendingPipeline
	}

	_, duplicateFound := manager.SavedPipelines[name]
	if duplicateFound {
		return ErrPipelineAlreadyExists
	}

	manager.SavedPipelines[name] = SavedPipeline{
		message.Author.ID, pendingPipeline,
	}
	delete(manager.PendingPipelines, message.Author.ID)

	return nil
}

func _CreatePipelineCommand(message *discordgo.MessageCreate, args struct{}) {
	err := Instance.PipelineManager.CreatePipeline(message)
	if err != nil {
		Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror creating new pipeline: %s\n```", err.Error()))
	}
}

type _DeletePipelineArgs struct {
	PipelineName string
}

func _DeletePipelineCommand(message *discordgo.MessageCreate, args _DeletePipelineArgs) {
	err := Instance.PipelineManager.DeletePipeline(message, args.PipelineName)
	if err != nil {
		Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror deleting pipeline: %s\n```", err.Error()))
	}
}

func _DebugPipelineCommand(message *discordgo.MessageCreate, args struct{}) {
	Instance.Session.ChannelMessageSend(
		message.ChannelID,
		fmt.Sprintf("```\n%#v\n```", Instance.PipelineManager),
	)
}

type _RunPipelineArgs struct {
	PipelineName string
	ImageURL     string `default:""`
}

func _RunPipelineCommand(message *discordgo.MessageCreate, args _RunPipelineArgs) {
	pipeline, err := Instance.PipelineManager.GetPipeline(message, args.PipelineName)
	if err != nil {
		Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror running pipeline: %s\n```", err.Error()))
		return
	}

	if args.ImageURL == "" {
		var err error
		args.ImageURL, err = FindImageURL(message)
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	srcBytes, err := DownloadImage(args.ImageURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}
	destBuffer := new(bytes.Buffer)

	embed := discordgo.MessageEmbed{
		Title: "Pipeline Running",
		Color: (155 << 16) + (89 << 8) + 182,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Operation",
				Value: "Starting...",
			},
			{
				Name:  "Args",
				Value: "```go\nnil\n```",
			},
			{
				Name:  "Progress",
				Value: "0 / 0",
			},
		},
	}
	statusMsg, err := Instance.Session.ChannelMessageSendEmbed(message.ChannelID, &embed)
	if err != nil {
		panic(err)
	}

	for index, operation := range pipeline {
		embed.Fields[0].Value = operation.Operation
		embed.Fields[1].Value = fmt.Sprintf("```go\n%+v```", operation.Args)
		embed.Fields[2].Value = fmt.Sprintf("%d / %d", index+1, len(pipeline))

		_, err := Instance.Session.ChannelMessageEditEmbed(message.ChannelID, statusMsg.ID, &embed)
		if err != nil {
			log.Error().Err(err).Msg("Error editing status embed")
		}

		destBuffer = new(bytes.Buffer)
		log.Debug().Interface("operation", operation).Msg("Running operation step")
		err = _OperationMap[operation.Operation](srcBytes, destBuffer, operation.Args)

		if err != nil {
			log.Error().Err(err).Str("operation", operation.Operation).Msg("Error running pipeline operation")
			embed.Title = "Pipeline Errored"
			embed.Color = (231 << 16) + (76 << 8) + 60
			embed.Fields = append(
				embed.Fields,
				&discordgo.MessageEmbedField{
					Name:  "Error",
					Value: fmt.Sprintf("```\n%s\n```", err.Error()),
				},
			)
			_, err := Instance.Session.ChannelMessageEditEmbed(message.ChannelID, statusMsg.ID, &embed)
			if err != nil {
				log.Error().Err(err).Msg("Error editing status embed")
			}
			return
		}
		srcBytes = destBuffer.Bytes()
	}

	_, err = Instance.Session.ChannelFileSend(message.ChannelID, "test.jpeg", destBuffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send image")
		_, err = Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Failed to send resulting image: `%s`", err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("Failed to send error message")
		}
	}

	err = Instance.Session.ChannelMessageDelete(message.ChannelID, statusMsg.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete status embed")
	}
}

type _SavePipelineArgs struct {
	Name string
}

func _SavePipelineCommand(message *discordgo.MessageCreate, args _SavePipelineArgs) {
	err := Instance.PipelineManager.SavePipeline(message, args.Name)
	if err != nil {
		Instance.Session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```\nerror saving pipeline: %s\n```", err.Error()))
		return
	}
}
