package bot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fogo-sh/borik/pkg/config"
)

type renderRequest struct {
	Animation string         `json:"animation"`
	Image     string         `json:"image"`
	Params    map[string]any `json:"params"`
}

type renderResponse struct {
	Frames []string `json:"frames"`
	Delay  int      `json:"delay"`
	Error  string   `json:"error,omitempty"`
}

// RenderAnimation calls the node-renderer sidecar to render an animation.
// Returns the rendered frames as PNG byte slices and the frame delay in ms.
func RenderAnimation(animation string, imageBytes []byte, params map[string]any) ([][]byte, int, error) {
	rendererURL := config.Instance.NodeRendererUrl
	if rendererURL == "" {
		return nil, 0, fmt.Errorf("node renderer URL not configured (set BORIK_NODE_RENDERER_URL)")
	}

	reqBody := renderRequest{
		Animation: animation,
		Image:     base64.StdEncoding.EncodeToString(imageBytes),
		Params:    params,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshaling render request: %w", err)
	}

	resp, err := http.Post(rendererURL+"/render", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("error calling node renderer: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading renderer response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp renderResponse
		_ = json.Unmarshal(body, &errResp)
		return nil, 0, fmt.Errorf("renderer error (status %d): %s", resp.StatusCode, errResp.Error)
	}

	var result renderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, fmt.Errorf("error parsing renderer response: %w", err)
	}

	frames := make([][]byte, len(result.Frames))
	for i, f := range result.Frames {
		decoded, err := base64.StdEncoding.DecodeString(f)
		if err != nil {
			return nil, 0, fmt.Errorf("error decoding frame %d: %w", i, err)
		}
		frames[i] = decoded
	}

	return frames, result.Delay, nil
}
