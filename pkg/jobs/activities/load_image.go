package activities

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

func LoadImage(ctx context.Context, imageUrl string) ([]byte, error) {
	resp, err := http.Get(imageUrl)
	if err != nil {
		return nil, fmt.Errorf("error downloading image: %w", err)
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)

	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error copying image to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}
