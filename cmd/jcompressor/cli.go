package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/dalbezh/jcompressor/internal/compressor"
)

type CLIParams struct {
	InputPath  string
	OutputPath string
	Quality    int
}

var ErrHelpRequested = errors.New("help requested")

// ParseCLI parses command-line arguments from args (typically os.Args[1:]).
// It recognizes -h/--help and -q/--quality. inputPath is required, outputPath is optional.
func ParseCLI(args []string) (*CLIParams, error) {
	fs := flag.NewFlagSet("jcompressor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var help bool
	var quality int

	fs.BoolVar(&help, "h", false, "show help")
	fs.BoolVar(&help, "help", false, "show help")
	fs.IntVar(&quality, "q", 50, "JPEG quality (1-100)")
	fs.IntVar(&quality, "quality", 50, "JPEG quality (1-100)")

	fs.Usage = func() {
		// Use a fixed program name in usage output to avoid reporting untrusted
		// data (os.Args[0]) to linters like gosec (G705).
		fmt.Fprintln(os.Stderr, "Usage: jcompressor [flags] <input.jpg> [output.jpg]")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nIf output.jpg is omitted, a file with suffix _compressed will be created.")
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
	output := ""
	if len(pos) >= 2 {
		output = pos[1]
	} else {
		output = compressor.GenerateOutputPath(input)
	}

	if quality < 1 || quality > 100 {
		return nil, fmt.Errorf("quality must be between 1 and 100")
	}

	return &CLIParams{Quality: quality, InputPath: input, OutputPath: output}, nil
}
