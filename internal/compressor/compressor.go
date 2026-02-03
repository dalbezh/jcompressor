package compressor

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
)

// Compressor handles JPEG image compression.
type Compressor struct {
	quality int
}

func New(quality int) *Compressor {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	return &Compressor{quality: quality}
}

// CompressFile compresses a JPEG file and saves the result.
// If outputPath is empty, the compressed file will be saved with "_compressed" suffix.
func (c *Compressor) CompressFile(inputPath, outputPath string) error {
	// Validate input file extension
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext != ".jpg" && ext != ".jpeg" {
		return fmt.Errorf("input file must be a JPEG image (got %s)", ext)
	}

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Decode JPEG image
	img, err := jpeg.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode JPEG image: %w", err)
	}

	// Generate output path if not specified
	if outputPath == "" {
		outputPath = c.generateOutputPath(inputPath)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode with specified quality
	options := &jpeg.Options{Quality: c.quality}
	if err := jpeg.Encode(outputFile, img, options); err != nil {
		return fmt.Errorf("failed to encode JPEG image: %w", err)
	}

	return nil
}

// Compress compresses an image.Image and returns the result as bytes.
func (c *Compressor) Compress(img image.Image) ([]byte, error) {
	var buf strings.Builder
	options := &jpeg.Options{Quality: c.quality}

	// We need a Writer that collects bytes
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- jpeg.Encode(pw, img, options)
		pw.Close()
	}()

	data := make([]byte, 0)
	buf2 := make([]byte, 1024)
	for {
		n, err := pr.Read(buf2)
		if n > 0 {
			data = append(data, buf2[:n]...)
		}
		if err != nil {
			break
		}
	}

	if err := <-errCh; err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	_ = buf // unused, keeping for clarity
	return data, nil
}

// generateOutputPath generates an output file path by adding "_compressed" suffix.
func (c *Compressor) generateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return fmt.Sprintf("%s_compressed%s", base, ext)
}

// Quality returns the current quality setting.
func (c *Compressor) Quality() int {
	return c.quality
}

// CompressJPEG is a convenience function to compress a JPEG file with the specified quality.
func CompressJPEG(inputPath, outputPath string, quality int) error {
	c := New(quality)
	return c.CompressFile(inputPath, outputPath)
}

// GenerateOutputPath generates an output file path by adding "_compressed" suffix.
func GenerateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return fmt.Sprintf("%s_compressed%s", base, ext)
}
