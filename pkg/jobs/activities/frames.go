package activities

import (
	"context"
	_ "embed"

	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

//go:embed images/frames/presidents_frame.png
var presidentsFrameImage []byte

//go:embed images/frames/heritage.png
var heritageFrameImage []byte

//go:embed images/frames/shinji_throw.png
var shinjiFrameImage []byte

func PresidentsFrame(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyFrame(jobWorkspace, opArgs, presidentsFrameImage, frameOptions{
		FitMode: fitModeStretch,
	})
}

func Heritage(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyFrame(jobWorkspace, opArgs, heritageFrameImage, frameOptions{})
}

func Shinji(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	return applyFrame(jobWorkspace, opArgs, shinjiFrameImage, frameOptions{
		FitMode:      fitModeFitHeight,
		PositionMode: positionModeTopLeft,
	})
}
