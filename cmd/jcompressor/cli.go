package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type CLIParams struct {
	InputPath string
	OutputDir string
	Quality   int
	WebP      bool
}

var ErrHelpRequested = errors.New("help requested")

// ParseCLI parses command-line arguments from args (typically os.Args[1:]).
// It recognizes -h/--help, -q/--quality, and -w/--webp. inputPath is required, outputDir is optional.
func ParseCLI(args []string) (*CLIParams, error) {
	fs := flag.NewFlagSet("jcompressor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var help bool
	var quality int
	var webp bool

	fs.BoolVar(&help, "h", false, "show help")
	fs.BoolVar(&help, "help", false, "show help")
	fs.IntVar(&quality, "q", 50, "JPEG quality (1-100)")
	fs.IntVar(&quality, "quality", 50, "JPEG quality (1-100)")
	fs.BoolVar(&webp, "w", false, "also create WebP version")
	fs.BoolVar(&webp, "webp", false, "also create WebP version")

	fs.Usage = func() {
		// Use a fixed program name in usage output to avoid reporting untrusted
		// data (os.Args[0]) to linters like gosec (G705).
		fmt.Fprintln(os.Stderr, "Usage: jcompressor [flags] <input.jpg> [output_dir]")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nIf output_dir is omitted, files will be saved to ./compressed")
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if help {
		fs.Usage()
		return nil, ErrHelpRequested
	}

	pos := fs.Args()
	if len(pos) < 1 {
		fs.Usage()
		return nil, fmt.Errorf("inputPath required")
	}
	if len(pos) > 2 {
		return nil, fmt.Errorf("too many arguments")
	}

	input := pos[0]
	outputDir := "./compressed"
	if len(pos) >= 2 {
		outputDir = pos[1]
	}

	if quality < 1 || quality > 100 {
		return nil, fmt.Errorf("quality must be between 1 and 100")
	}

	return &CLIParams{Quality: quality, InputPath: input, OutputDir: outputDir, WebP: webp}, nil
}
