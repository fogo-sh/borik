package activities

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/jobs/args"
	"github.com/fogo-sh/borik/pkg/jobs/workspace"
)

//go:embed fonts/anton.ttf
var antonFontData []byte

var (
	antonFontPath string
	antonFontOnce sync.Once
	antonFontErr  error
)

const (
	memeZoneWidth   = 0.90
	memeZoneHeight  = 0.25
	memeMaxFontSize = 0.10
	memePadding     = 0.02
	memeKerning     = 0.02
	memeStrokeRatio = 1.0 / 8.0
	memeMinStroke   = 2.0
)

func Meme(ctx context.Context, jobWorkspace workspace.Workspace, opArgs OperationArgs) ([]workspace.Artifact, error) {
	wand, err := jobWorkspace.RetrieveWand(opArgs.Frame)
	if err != nil {
		return nil, err
	}

	var memeArgs args.Meme
	err = decodeOperationArgs(opArgs, &memeArgs)
	if err != nil {
		return nil, fmt.Errorf("error while decoding operation args: %w", err)
	}

	topText, bottomText := parseMemeText(memeArgs.Text)

	err = drawMemeText(wand, topText, false)
	if err != nil {
		return nil, fmt.Errorf("error drawing top text: %w", err)
	}

	err = drawMemeText(wand, bottomText, true)
	if err != nil {
		return nil, fmt.Errorf("error drawing bottom text: %w", err)
	}

	return saveFrames(jobWorkspace, wand)
}

func getMemeFontPath() (string, error) {
	antonFontOnce.Do(func() {
		fontPath := filepath.Join(os.TempDir(), "borik-anton.ttf")
		antonFontErr = os.WriteFile(fontPath, antonFontData, 0644)
		antonFontPath = fontPath
	})
	if antonFontErr != nil {
		return "", fmt.Errorf("error writing embedded Anton font to temp file: %w", antonFontErr)
	}

	return antonFontPath, nil
}

func parseMemeText(text string) (top string, bottom string) {
	parts := strings.SplitN(text, "|", 2)
	top = strings.ToUpper(strings.TrimSpace(parts[0]))
	if len(parts) > 1 {
		bottom = strings.ToUpper(strings.TrimSpace(parts[1]))
	}
	return top, bottom
}

func memeWrapText(wand *imagick.MagickWand, dw *imagick.DrawingWand, text string, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	currentLine := words[0]
	for _, word := range words[1:] {
		testLine := currentLine + " " + word
		metrics := wand.QueryFontMetrics(dw, testLine)
		if metrics != nil && metrics.TextWidth <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	return append(lines, currentLine)
}

func memeFitText(wand *imagick.MagickWand, text string, zoneWidth, zoneHeight, imgHeight float64) (float64, string, error) {
	if text == "" {
		return 0, "", nil
	}

	fontPath, err := getMemeFontPath()
	if err != nil {
		return 0, "", err
	}

	dw := imagick.NewDrawingWand()
	defer dw.Destroy()
	if err := dw.SetFont(fontPath); err != nil {
		return 0, "", fmt.Errorf("error setting font: %w", err)
	}

	maxSize := imgHeight * memeMaxFontSize
	lo := 1.0
	hi := math.Min(zoneHeight*0.9, maxSize)
	bestSize := lo
	bestText := text

	for lo <= hi {
		mid := math.Floor((lo + hi) / 2)
		if mid < 1 {
			break
		}
		dw.SetFontSize(mid)
		dw.SetTextKerning(mid * memeKerning)

		lines := memeWrapText(wand, dw, text, zoneWidth)
		joined := strings.Join(lines, "\n")
		metrics := wand.QueryMultilineFontMetrics(dw, joined)
		if metrics == nil {
			hi = mid - 1
			continue
		}

		if metrics.TextHeight <= zoneHeight && metrics.TextWidth <= zoneWidth {
			bestSize = mid
			bestText = joined
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	return bestSize, bestText, nil
}

func drawMemeText(wand *imagick.MagickWand, text string, bottom bool) error {
	if text == "" {
		return nil
	}

	fontPath, err := getMemeFontPath()
	if err != nil {
		return err
	}

	imgWidth := float64(wand.GetImageWidth())
	imgHeight := float64(wand.GetImageHeight())
	zoneWidth := imgWidth * memeZoneWidth
	zoneHeight := imgHeight * memeZoneHeight

	fontSize, wrappedText, err := memeFitText(wand, text, zoneWidth, zoneHeight, imgHeight)
	if err != nil {
		return err
	}
	if wrappedText == "" {
		return nil
	}

	strokeWidth := math.Max(fontSize*memeStrokeRatio, memeMinStroke)
	kerning := fontSize * memeKerning
	padding := imgHeight * memePadding
	xPos := imgWidth / 2

	var yPos float64
	if bottom {
		metricsDw := imagick.NewDrawingWand()
		defer metricsDw.Destroy()
		_ = metricsDw.SetFont(fontPath)
		metricsDw.SetFontSize(fontSize)
		metricsDw.SetTextKerning(kerning)
		metricsDw.SetStrokeWidth(strokeWidth)
		metrics := wand.QueryMultilineFontMetrics(metricsDw, wrappedText)
		textHeight := fontSize
		if metrics != nil {
			textHeight = metrics.TextHeight
		}
		yPos = imgHeight - padding - textHeight + fontSize
	} else {
		yPos = padding + fontSize
	}

	white := imagick.NewPixelWand()
	defer white.Destroy()
	white.SetColor("white")

	black := imagick.NewPixelWand()
	defer black.Destroy()
	black.SetColor("black")

	none := imagick.NewPixelWand()
	defer none.Destroy()
	none.SetColor("none")

	textCanvas := imagick.NewMagickWand()
	defer textCanvas.Destroy()
	if err := textCanvas.NewImage(wand.GetImageWidth(), wand.GetImageHeight(), none); err != nil {
		return fmt.Errorf("error creating text canvas: %w", err)
	}

	dw := imagick.NewDrawingWand()
	defer dw.Destroy()
	if err := dw.SetFont(fontPath); err != nil {
		return fmt.Errorf("error setting font: %w", err)
	}
	dw.SetFontSize(fontSize)
	dw.SetTextKerning(kerning)
	dw.SetTextAlignment(imagick.ALIGN_CENTER)

	dw.SetFillColor(white)
	dw.SetStrokeColor(black)
	dw.SetStrokeWidth(strokeWidth)
	dw.Annotation(xPos, yPos, wrappedText)

	dw.SetStrokeColor(none)
	dw.SetStrokeWidth(0)
	dw.Annotation(xPos, yPos, wrappedText)

	if err := textCanvas.DrawImage(dw); err != nil {
		return fmt.Errorf("error drawing text: %w", err)
	}

	return wand.CompositeImage(textCanvas, imagick.COMPOSITE_OP_OVER, true, 0, 0)
}
