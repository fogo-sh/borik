package activities

import (
	"context"
	_ "embed"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

//go:embed images/overlays/jack_pog.png
var jackPogImage []byte

//go:embed images/overlays/keenan_thumb.png
var keenanThumbImage []byte

//go:embed images/overlays/mitch_point.png
var mitchPointImage []byte

//go:embed images/overlays/side_keenan.png
var sideKeenanImage []byte

//go:embed images/overlays/steve_point.png
var stevePointImage []byte

//go:embed images/overlays/andrew_pog.png
var andrewPogImage []byte

//go:embed images/overlays/trans_matlab_kid.png
var matlabKidImage []byte

//go:embed images/overlays/natalie_climb.png
var natalieClimbImage []byte

//go:embed images/overlays/denny_standing.png
var dennyStandingImage []byte

func JackPog(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, jackPogImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 0.5,
	})
}

func SideKeenan(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, sideKeenanImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 0.5,
		RightToLeft:         true,
	})
}

func KeenanThumb(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, keenanThumbImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 0.5,
	})
}

func MitchPoint(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, mitchPointImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 1,
	})
}

func StevePoint(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, stevePointImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 1,
		RightToLeft:         true,
	})
}

func AndrewPog(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, andrewPogImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 0.75,
		RightToLeft:         true,
	})
}

func MatlabKid(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, matlabKidImage, overlayOptions{
		VFlip:               true,
		OverlayWidthFactor:  1.2,
		OverlayHeightFactor: 1.3,
		RightToLeft:         true,
	})
}

func NatalieClimb(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, natalieClimbImage, overlayOptions{
		OverlayWidthFactor:  1,
		OverlayHeightFactor: 1,
	})
}

func DennyStanding(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyOverlay(jobWorkspace, opArgs, dennyStandingImage, overlayOptions{
		OverlayWidthFactor:  0.4,
		OverlayHeightFactor: 0.6,
	})
}
