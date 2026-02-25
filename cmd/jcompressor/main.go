package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dalbezh/jcompressor/internal/compressor"
)

func main() {
	cliParams, err := ParseCLI(os.Args[1:])
	if err != nil {
		if errors.Is(err, ErrHelpRequested) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Создаем output directory если не существует
	if err := os.MkdirAll(cliParams.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	inputFileName := filepath.Base(cliParams.InputPath)
	jpegOutputPath := filepath.Join(cliParams.OutputDir, inputFileName)

	// Сжимаем JPEG
	if err := compressor.CompressJPEG(cliParams.InputPath, jpegOutputPath, cliParams.Quality); err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compressed %s -> %s (quality: %d)\n", cliParams.InputPath, jpegOutputPath, cliParams.Quality)

	// Если нужно создать WebP
	if cliParams.WebP {
		// Формируем путь к WebP файлу (заменяем расширение на .webp)
		ext := filepath.Ext(inputFileName)
		webpFileName := strings.TrimSuffix(inputFileName, ext) + ".webp"
		webpOutputPath := filepath.Join(cliParams.OutputDir, webpFileName)

		if err := compressor.CompressToWebP(cliParams.InputPath, webpOutputPath, cliParams.Quality); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating WebP: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created WebP %s -> %s (quality: %d)\n", cliParams.InputPath, webpOutputPath, cliParams.Quality)
	}
}
