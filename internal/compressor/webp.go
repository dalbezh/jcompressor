//go:build cgo && !nowebp

package compressor

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

// ErrWebPNotSupported is returned when WebP functionality is not available
// This error is only used in the no-CGO build, but declared here for API consistency
var ErrWebPNotSupported = errors.New("WebP support is not available in this build (requires CGO and libwebp)")

// ConvertToWebP converts an image to WebP format with the specified quality
func ConvertToWebP(img image.Image, outputPath string, quality int) (err error) {
	outputPath = filepath.Clean(outputPath)

	// Encode image to WebP
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(quality))
	if err != nil {
		return fmt.Errorf("failed to create webp encoder options: %w", err)
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, options); err != nil {
		return fmt.Errorf("failed to encode WebP image: %w", err)
	}

	// Write to file
	// #nosec G306 -- file permissions 0644 are intentional
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write webp output file: %w", err)
	}

	return nil
}

// CompressToWebP reads a JPEG and saves it as WebP
func CompressToWebP(inputPath, outputPath string, quality int) error {
	inputPath = filepath.Clean(inputPath)

	inputFile, err := os.Open(inputPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open input file for webp: %w", err)
	}
	defer inputFile.Close()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode image for webp: %w", err)
	}

	return ConvertToWebP(img, outputPath, quality)
}
