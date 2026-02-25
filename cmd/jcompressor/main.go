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

	// Валидация и очистка пути для предотвращения path traversal
	outputDir := filepath.Clean(cliParams.OutputDir)
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving output directory path: %v\n", err)
		os.Exit(1)
	}

	// Создаем output directory если не существует
	// #nosec G301 G703 -- path is cleaned and validated, permissions are intentional
	if err := os.MkdirAll(absOutputDir, 0755); err != nil { // #nosec G301
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	inputFileName := filepath.Base(cliParams.InputPath)
	jpegOutputPath := filepath.Join(absOutputDir, inputFileName)

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
		webpOutputPath := filepath.Join(absOutputDir, webpFileName)

		if err := compressor.CompressToWebP(cliParams.InputPath, webpOutputPath, cliParams.Quality); err != nil {
			// Проверяем, не является ли это ошибкой "WebP не поддерживается"
			if errors.Is(err, compressor.ErrWebPNotSupported) {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				fmt.Fprintf(os.Stderr, "Note: To enable WebP support, rebuild with CGO_ENABLED=1 and libwebp installed\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error creating WebP: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created WebP %s -> %s (quality: %d)\n", cliParams.InputPath, webpOutputPath, cliParams.Quality)
	}
}
