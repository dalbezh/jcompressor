package main

import (
	"errors"
	"fmt"
	"os"

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

	if err := compressor.CompressJPEG(cliParams.InputPath, cliParams.OutputPath, cliParams.Quality); err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compressed %s -> %s (quality: %d)\n", cliParams.InputPath, cliParams.OutputPath, cliParams.Quality)
}
