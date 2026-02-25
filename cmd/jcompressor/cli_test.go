package main

import (
	"errors"
	"strings"
	"testing"
)

// TestParseCLI_ValidInput проверяет корректный парсинг валидных аргументов
func TestParseCLI_ValidInput(t *testing.T) {
	tests := []struct { //nolint:govet // test struct, fieldalignment not critical
		name          string
		args          []string
		wantInputPath string
		wantOutputDir string
		wantQuality   int
		wantWebP      bool
	}{
		{
			name:          "minimum args",
			args:          []string{"input.jpg"},
			wantInputPath: "input.jpg",
			wantOutputDir: "./compressed",
			wantQuality:   50,
			wantWebP:      false,
		},
		{
			name:          "with output dir",
			args:          []string{"input.jpg", "/tmp/output"},
			wantInputPath: "input.jpg",
			wantOutputDir: "/tmp/output",
			wantQuality:   50,
			wantWebP:      false,
		},
		{
			name:          "with quality flag -q",
			args:          []string{"-q", "80", "photo.jpeg"},
			wantInputPath: "photo.jpeg",
			wantOutputDir: "./compressed",
			wantQuality:   80,
			wantWebP:      false,
		},
		{
			name:          "with quality flag --quality",
			args:          []string{"--quality", "90", "image.jpg", "./out"},
			wantInputPath: "image.jpg",
			wantOutputDir: "./out",
			wantQuality:   90,
			wantWebP:      false,
		},
		{
			name:          "with webp flag -w",
			args:          []string{"-w", "test.jpg"},
			wantInputPath: "test.jpg",
			wantOutputDir: "./compressed",
			wantQuality:   50,
			wantWebP:      true,
		},
		{
			name:          "with webp flag --webp",
			args:          []string{"--webp", "photo.jpeg"},
			wantInputPath: "photo.jpeg",
			wantOutputDir: "./compressed",
			wantQuality:   50,
			wantWebP:      true,
		},
		{
			name:          "all flags combined",
			args:          []string{"-q", "75", "-w", "input.jpg", "/output"},
			wantInputPath: "input.jpg",
			wantOutputDir: "/output",
			wantQuality:   75,
			wantWebP:      true,
		},
		{
			name:          "quality at minimum boundary",
			args:          []string{"-q", "1", "test.jpg"},
			wantInputPath: "test.jpg",
			wantOutputDir: "./compressed",
			wantQuality:   1,
			wantWebP:      false,
		},
		{
			name:          "quality at maximum boundary",
			args:          []string{"-q", "100", "test.jpg"},
			wantInputPath: "test.jpg",
			wantOutputDir: "./compressed",
			wantQuality:   100,
			wantWebP:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := ParseCLI(tt.args)
			if err != nil {
				t.Fatalf("ParseCLI() unexpected error = %v", err)
			}

			if params.InputPath != tt.wantInputPath {
				t.Errorf("InputPath = %v, want %v", params.InputPath, tt.wantInputPath)
			}
			if params.OutputDir != tt.wantOutputDir {
				t.Errorf("OutputDir = %v, want %v", params.OutputDir, tt.wantOutputDir)
			}
			if params.Quality != tt.wantQuality {
				t.Errorf("Quality = %v, want %v", params.Quality, tt.wantQuality)
			}
			if params.WebP != tt.wantWebP {
				t.Errorf("WebP = %v, want %v", params.WebP, tt.wantWebP)
			}
		})
	}
}

// TestParseCLI_InvalidInput проверяет обработку невалидных аргументов
func TestParseCLI_InvalidInput(t *testing.T) {
	tests := []struct { //nolint:govet // test struct, fieldalignment not critical
		name       string
		args       []string
		wantErrMsg string
	}{
		{
			name:       "no arguments",
			args:       []string{},
			wantErrMsg: "inputPath required",
		},
		{
			name:       "too many arguments",
			args:       []string{"input.jpg", "output", "extra"},
			wantErrMsg: "too many arguments",
		},
		{
			name:       "quality below minimum",
			args:       []string{"-q", "0", "input.jpg"},
			wantErrMsg: "quality must be between 1 and 100",
		},
		{
			name:       "quality above maximum",
			args:       []string{"-q", "101", "input.jpg"},
			wantErrMsg: "quality must be between 1 and 100",
		},
		{
			name:       "negative quality",
			args:       []string{"-q", "-10", "input.jpg"},
			wantErrMsg: "quality must be between 1 and 100",
		},
		{
			name:       "quality with invalid value",
			args:       []string{"-q", "abc", "input.jpg"},
			wantErrMsg: "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCLI(tt.args)
			if err == nil {
				t.Fatal("ParseCLI() expected error but got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("ParseCLI() error = %v, want error containing %q", err, tt.wantErrMsg)
			}
		})
	}
}

// TestParseCLI_HelpFlag проверяет обработку флага помощи
func TestParseCLI_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "short help flag", args: []string{"-h"}},
		{name: "long help flag", args: []string{"--help"}},
		{name: "help with other args", args: []string{"-h", "input.jpg"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := ParseCLI(tt.args)
			if err == nil {
				t.Fatal("ParseCLI() expected ErrHelpRequested but got nil")
			}

			if !errors.Is(err, ErrHelpRequested) {
				t.Errorf("ParseCLI() error = %v, want ErrHelpRequested", err)
			}

			if params != nil {
				t.Errorf("ParseCLI() params = %v, want nil when help is requested", params)
			}
		})
	}
}

// TestParseCLI_Defaults проверяет значения по умолчанию
func TestParseCLI_Defaults(t *testing.T) {
	params, err := ParseCLI([]string{"test.jpg"})
	if err != nil {
		t.Fatalf("ParseCLI() unexpected error = %v", err)
	}

	// Проверяем дефолтные значения
	if params.Quality != 50 {
		t.Errorf("Default Quality = %d, want 50", params.Quality)
	}
	if params.OutputDir != "./compressed" {
		t.Errorf("Default OutputDir = %q, want ./compressed", params.OutputDir)
	}
	if params.WebP != false {
		t.Errorf("Default WebP = %v, want false", params.WebP)
	}
}

// TestCLIParams_StructFields проверяет наличие всех полей в структуре
func TestCLIParams_StructFields(t *testing.T) {
	params := &CLIParams{
		InputPath: "test.jpg",
		OutputDir: "./output",
		Quality:   80,
		WebP:      true,
	}

	// Проверка что все поля доступны и имеют правильные типы
	var _ = params.InputPath
	var _ = params.OutputDir
	var _ = params.Quality
	var _ = params.WebP
}

// BenchmarkParseCLI бенчмарк парсинга CLI аргументов
func BenchmarkParseCLI(b *testing.B) {
	args := []string{"-q", "80", "-w", "input.jpg", "./output"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseCLI(args) // nolint:errcheck // benchmark ignores errors intentionally
	}
}

// BenchmarkParseCLI_Simple бенчмарк простого парсинга
func BenchmarkParseCLI_Simple(b *testing.B) {
	args := []string{"input.jpg"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseCLI(args) // nolint:errcheck // benchmark ignores errors intentionally
	}
}
