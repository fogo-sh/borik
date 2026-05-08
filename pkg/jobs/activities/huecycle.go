package activities

import (
	"context"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

func HueCycle(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var hueCycleArgs args.HueCycle
	err = decodeOperationArgs(opArgs, &hueCycleArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	wands := []*imagick.MagickWand{wand}

	for i := uint(0); i < hueCycleArgs.Steps; i++ {
		wand = wand.Clone()
		err := wand.ModulateImage(100, 100, 100+(200/float64(hueCycleArgs.Steps)))
		if err != nil {
			return nil, fmt.Errorf("error cycling hue: %w", err)
		}
		wands = append(wands, wand)
	}

	return saveFrames(jobWorkspace, wands...)
}
