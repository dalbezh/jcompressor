package main

import (
	"fmt"
	"os"

	"github.com/dalbezh/jcompressor/internal/compressor"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.jpg> [output.jpg]\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := ""

	if len(os.Args) >= 3 {
		outputPath = os.Args[2]
	} else {
		// Если выходной файл не указан, создаём имя с суффиксом _compressed
		outputPath = compressor.GenerateOutputPath(inputPath)
	}

	if err := compressor.CompressJPEG(inputPath, outputPath, 50); err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compressed %s -> %s (quality: 50)\n", inputPath, outputPath)
}
