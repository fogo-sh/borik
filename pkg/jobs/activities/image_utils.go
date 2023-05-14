package activities

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/gographics/imagick.v3/imagick"
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

func SplitImage(ctx context.Context, imageBytes []byte) ([][]byte, error) {
	input := imagick.NewMagickWand()
	defer input.Destroy()

	err := input.ReadImageBlob(imageBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading image: %w", err)
	}

	var resultBytes [][]byte

	for i := uint(0); i < input.GetNumberImages(); i++ {
		input.SetIteratorIndex(int(i))
		resultBytes = append(resultBytes, input.GetImageBlob())
	}

	return resultBytes, nil
}

func JoinImage(ctx context.Context, imageBytes [][]byte) ([]byte, error) {
	output := imagick.NewMagickWand()
	defer output.Destroy()

	for _, image := range imageBytes {
		frame := imagick.NewMagickWand()
		err := frame.ReadImageBlob(image)
		if err != nil {
			return nil, fmt.Errorf("error reading frame: %w", err)
		}

		err = output.AddImage(frame)
		if err != nil {
			return nil, fmt.Errorf("error adding frame to output: %w", err)
		}
	}

	output.ResetIterator()

	if len(imageBytes) > 1 {
		err := output.SetImageFormat("GIF")
		if err != nil {
			return nil, fmt.Errorf("error setting output format: %w", err)
		}

		err = output.SetImageDelay(10)
		if err != nil {
			return nil, fmt.Errorf("error setting output framerate: %w", err)
		}
	} else {
		err := output.SetImageFormat("PNG")
		if err != nil {
			return nil, fmt.Errorf("error setting output format: %w", err)
		}
	}

	err := output.ResetImagePage("0x0+0+0")
	if err != nil {
		return nil, fmt.Errorf("error repaging output: %w", err)
	}

	output = output.DeconstructImages()

	return output.GetImagesBlob(), nil
}

func loadFrame(frameBytes []byte) (*imagick.MagickWand, error) {
	frame := imagick.NewMagickWand()
	err := frame.ReadImageBlob(frameBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading frame: %w", err)
	}

	return frame, nil
}

func saveFrames(frames ...*imagick.MagickWand) ([][]byte, error) {
	var resultBytes [][]byte

	for _, frame := range frames {
		resultBytes = append(resultBytes, frame.GetImageBlob())
	}

	return resultBytes, nil
}
