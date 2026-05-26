package activities

import (
	"github.com/bwmarrin/discordgo"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

func RegisterActivities(worker worker.Worker) {
	worker.RegisterActivity(LoadImage)
	worker.RegisterActivity(SplitImage)
	worker.RegisterActivity(JoinImage)

	worker.RegisterActivity(Magik)
	worker.RegisterActivity(Lagik)
	worker.RegisterActivity(Gmagik)
	worker.RegisterActivity(Arcweld)
	worker.RegisterActivity(Malt)
	worker.RegisterActivity(Deepfry)
	worker.RegisterActivity(Divine)
	worker.RegisterActivity(PresidentsFrame)
	worker.RegisterActivity(Heritage)
	worker.RegisterActivity(Shinji)
	worker.RegisterActivity(Waaw)
	worker.RegisterActivity(Haah)
	worker.RegisterActivity(Woow)
	worker.RegisterActivity(Hooh)
	worker.RegisterActivity(Invert)
	worker.RegisterActivity(Otsu)
	worker.RegisterActivity(Rotate)
	worker.RegisterActivity(Resize)
	worker.RegisterActivity(HueCycle)
	worker.RegisterActivity(Modulate)
	worker.RegisterActivity(Meme)
	worker.RegisterActivity(Hdr)
	worker.RegisterActivity(EGA)
	worker.RegisterActivity(TempleOS)
	worker.RegisterActivity(CGA)
	worker.RegisterActivity(JackPog)
	worker.RegisterActivity(SideKeenan)
	worker.RegisterActivity(KeenanThumb)
	worker.RegisterActivity(MitchPoint)
	worker.RegisterActivity(StevePoint)
	worker.RegisterActivity(AndrewPog)
	worker.RegisterActivity(MatlabKid)
	worker.RegisterActivity(NatalieClimb)
	worker.RegisterActivity(DennyStanding)
	worker.RegisterActivity(GenerateImage)
	worker.RegisterActivity(ImageEdit)
	worker.RegisterActivity(LoopEdit)
	worker.RegisterActivity(FlipFlop)
	worker.RegisterActivity(AiZoom)
	worker.RegisterActivity(AiLoopZoom)
	worker.RegisterActivity(ConvertVideoToGIF)
	worker.RegisterActivity(APNGToGIF)
	worker.RegisterActivity(CleanupWorkspace)
}

func RegisterDiscordActivities(worker worker.Worker, discordSession *discordgo.Session) {
	deliveryActivities := DiscordDeliveryActivities{Session: discordSession}
	worker.RegisterActivityWithOptions(deliveryActivities.SendDiscordResult, activity.RegisterOptions{
		Name: SendDiscordResultActivityName,
	})
	worker.RegisterActivityWithOptions(deliveryActivities.SendDiscordFailure, activity.RegisterOptions{
		Name: SendDiscordFailureActivityName,
	})
	worker.RegisterActivityWithOptions(deliveryActivities.SendDiscordTyping, activity.RegisterOptions{
		Name: SendDiscordTypingActivityName,
	})
}
