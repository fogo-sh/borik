package activities

import (
	"context"
	"fmt"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func Modulate(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var modulateArgs args.Modulate
	err = decodeOperationArgs(opArgs, &modulateArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	err = wand.ModulateImage(modulateArgs.Brightness, modulateArgs.Saturation, modulateArgs.Hue)
	if err != nil {
		return nil, fmt.Errorf("error modulating image: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}
