package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

type AiTranslateArgs struct {
	Text   string `description:"The text to translate."`
	Source string `default:"English" description:"The language to translate from."`
	Target string `default:"LinkedIn Speak" description:"The language to translate to."`
}

func aiTranslate(ctx *OperationContext, args AiTranslateArgs) {
	resp, err := Instance.openAiClient.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Instructions: param.Opt[string]{Value: "You are a Translator you are to respond with only the resultant text of translation based on your chat history"},
			Input: responses.ResponseNewParamsInputUnion{
				OfString: param.Opt[string]{Value: fmt.Sprintf(
					"Sure I'll translate the text \"%s\" from %s to %s.",
					args.Text,
					args.Source,
					args.Target,
				)},
			},
			Model: Instance.config.OpenaiTextModel,
		},
	)
	if err != nil {
		ctx.SendText("Error calling OpenAI API: " + err.Error())
		return
	}

	err = ctx.SendText(resp.OutputText())
	if err != nil {
		ctx.SendText("Error sending translation: " + err.Error())
	}
}

func AiTranslateCommand(message *discordgo.MessageCreate, args AiTranslateArgs) {
	aiTranslate(NewOperationContextFromMessage(Instance.session, message), args)
}

func AiTranslateSlashCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate, args AiTranslateArgs) {
	aiTranslate(NewOperationContextFromInteraction(session, interaction), args)
}
