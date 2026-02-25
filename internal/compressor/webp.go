package compressor

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/chai2010/webp"
)

// ConvertToWebP converts an image to WebP format with the specified quality
func ConvertToWebP(img image.Image, outputPath string, quality int) (err error) {
	outputPath = filepath.Clean(outputPath)

	outputFile, err := os.Create(outputPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create webp output file: %w", err)
	}
	defer closeFile(outputFile, &err)

	// Convert quality (1-100) to WebP quality
	if err := webp.Encode(outputFile, img, &webp.Options{
		Lossless: false,
		Quality:  float32(quality),
	}); err != nil {
		return fmt.Errorf("failed to encode WebP image: %w", err)
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
