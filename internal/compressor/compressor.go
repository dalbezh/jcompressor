package compressor

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
)

func closeFile(f *os.File, err *error) {
	if cerr := f.Close(); cerr != nil && *err == nil {
		*err = fmt.Errorf("failed to close file: %w", cerr)
	}
}

type Compressor struct {
	quality int
}

// Создаёт Compressor. Качество ограничивается диапазоном 1-100.
func New(quality int) *Compressor {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	return &Compressor{quality: quality}
}

// CompressFile сжимает JPEG файл. Если outputPath пустой — добавляет суффикс "_compressed".
func (c *Compressor) CompressFile(inputPath, outputPath string) (err error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext != ".jpg" && ext != ".jpeg" {
		return fmt.Errorf("input file must be a JPEG image (got %s)", ext)
	}

	inputPath = filepath.Clean(inputPath)

	inputFile, err := os.Open(inputPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer closeFile(inputFile, &err)

	img, err := jpeg.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode JPEG image: %w", err)
	}

	if outputPath == "" {
		outputPath = c.generateOutputPath(inputPath)
	}
	outputPath = filepath.Clean(outputPath)

	outputFile, err := os.Create(outputPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer closeFile(outputFile, &err)

	options := &jpeg.Options{Quality: c.quality}
	if err := jpeg.Encode(outputFile, img, options); err != nil {
		return fmt.Errorf("failed to encode JPEG image: %w", err)
	}

	return nil
}

// Compress image.Image and return bytes.
func (c *Compressor) Compress(img image.Image) ([]byte, error) {
	options := &jpeg.Options{Quality: c.quality}

	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		encErr := jpeg.Encode(pw, img, options)
		if cerr := pw.Close(); cerr != nil && encErr == nil {
			encErr = cerr
		}
		errCh <- encErr
	}()

	data := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, readErr := pr.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	if err := <-errCh; err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return data, nil
}

func (c *Compressor) generateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return fmt.Sprintf("%s_compressed%s", base, ext)
}

func (c *Compressor) Quality() int {
	return c.quality
}

// Обёртка для сжатия файла.
func CompressJPEG(inputPath, outputPath string, quality int) error {
	c := New(quality)
	return c.CompressFile(inputPath, outputPath)
}

// GenerateOutputPath генерирует путь с суффиксом "_compressed".
func GenerateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return fmt.Sprintf("%s_compressed%s", base, ext)
}
