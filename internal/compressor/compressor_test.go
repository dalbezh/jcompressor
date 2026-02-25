package compressor

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// createTestJPEG создает тестовое JPEG изображение
func createTestJPEG(t *testing.T, path string, width, height, quality int) {
	t.Helper()

	// Создаем директорию если нужно
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Создаем простое изображение с градиентом
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			c := color.RGBA{
				R: uint8((x * 255) / width),  // #nosec G115 // safe: x/width ratio is always 0-1
				G: uint8((y * 255) / height), // #nosec G115 // safe: y/height ratio is always 0-1
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	// Сохраняем как JPEG
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("Failed to encode test JPEG: %v", err)
	}
}

// TestNew проверяет создание компрессора с различными значениями quality
func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		inputQuality int
		wantQuality  int
	}{
		{
			name:         "valid quality mid-range",
			inputQuality: 50,
			wantQuality:  50,
		},
		{
			name:         "valid quality at minimum",
			inputQuality: 1,
			wantQuality:  1,
		},
		{
			name:         "valid quality at maximum",
			inputQuality: 100,
			wantQuality:  100,
		},
		{
			name:         "quality below minimum - should clamp to 1",
			inputQuality: 0,
			wantQuality:  1,
		},
		{
			name:         "negative quality - should clamp to 1",
			inputQuality: -10,
			wantQuality:  1,
		},
		{
			name:         "quality above maximum - should clamp to 100",
			inputQuality: 101,
			wantQuality:  100,
		},
		{
			name:         "very high quality - should clamp to 100",
			inputQuality: 999,
			wantQuality:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.inputQuality)

			if c == nil {
				t.Fatal("New() returned nil")
			}

			if c.Quality() != tt.wantQuality {
				t.Errorf("New(%d).Quality() = %d, want %d", tt.inputQuality, c.Quality(), tt.wantQuality)
			}
		})
	}
}

// TestCompressor_Quality проверяет метод Quality()
func TestCompressor_Quality(t *testing.T) {
	c := New(75)
	if got := c.Quality(); got != 75 {
		t.Errorf("Quality() = %d, want 75", got)
	}
}

// TestCompressor_CompressFile проверяет сжатие файлов
func TestCompressor_CompressFile(t *testing.T) {
	// Создаем временную директорию для тестов
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	// Создаем тестовый JPEG
	createTestJPEG(t, inputPath, 100, 100, 90)

	t.Run("successful compression", func(t *testing.T) {
		c := New(80)
		err := c.CompressFile(inputPath, outputPath)
		if err != nil {
			t.Fatalf("CompressFile() unexpected error: %v", err)
		}

		// Проверяем что выходной файл создан
		if _, statErr := os.Stat(outputPath); os.IsNotExist(statErr) {
			t.Error("Output file was not created")
		}

		// Проверяем что выходной файл валидный JPEG
		f, err := os.Open(outputPath)
		if err != nil {
			t.Fatalf("Failed to open output file: %v", err)
		}
		defer f.Close()

		_, err = jpeg.Decode(f)
		if err != nil {
			t.Errorf("Output file is not a valid JPEG: %v", err)
		}
	})

	t.Run("compression reduces file size", func(t *testing.T) {
		// Создаем файл с высоким качеством
		highQualityPath := filepath.Join(tmpDir, "high.jpg")
		createTestJPEG(t, highQualityPath, 200, 200, 95)

		// Сжимаем с низким качеством
		lowQualityPath := filepath.Join(tmpDir, "low.jpg")
		c := New(20)
		if err := c.CompressFile(highQualityPath, lowQualityPath); err != nil {
			t.Fatalf("CompressFile() error: %v", err)
		}

		// Сравниваем размеры файлов
		highInfo, err := os.Stat(highQualityPath)
		if err != nil {
			t.Fatalf("Failed to stat high quality file: %v", err)
		}
		lowInfo, err := os.Stat(lowQualityPath)
		if err != nil {
			t.Fatalf("Failed to stat low quality file: %v", err)
		}

		if lowInfo.Size() >= highInfo.Size() {
			t.Errorf("Compressed file size (%d) >= original (%d)", lowInfo.Size(), highInfo.Size())
		}
	})
}

// TestCompressor_CompressFile_Errors проверяет обработку ошибок
func TestCompressor_CompressFile_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(80)

	tests := []struct {
		name      string
		setupFunc func() (inputPath, outputPath string)
		wantErr   string
	}{
		{
			name: "input file does not exist",
			setupFunc: func() (string, string) {
				return filepath.Join(tmpDir, "nonexistent.jpg"),
					filepath.Join(tmpDir, "output.jpg")
			},
			wantErr: "failed to open input file",
		},
		{
			name: "input file is not JPEG",
			setupFunc: func() (string, string) {
				path := filepath.Join(tmpDir, "test.png")
				return path, filepath.Join(tmpDir, "output.jpg")
			},
			wantErr: "input file must be a JPEG image",
		},
		{
			name: "input file is not .jpg or .jpeg extension",
			setupFunc: func() (string, string) {
				path := filepath.Join(tmpDir, "test.txt")
				_ = os.WriteFile(path, []byte("not an image"), 0644) // nolint:errcheck // test setup
				return path, filepath.Join(tmpDir, "output.jpg")
			},
			wantErr: "input file must be a JPEG image",
		},
		{
			name: "input file has invalid JPEG data",
			setupFunc: func() (string, string) {
				path := filepath.Join(tmpDir, "invalid.jpg")
				_ = os.WriteFile(path, []byte("not jpeg data"), 0644) // nolint:errcheck // test setup
				return path, filepath.Join(tmpDir, "output.jpg")
			},
			wantErr: "failed to decode JPEG image",
		},
		{
			name: "output path is invalid directory",
			setupFunc: func() (string, string) {
				inputPath := filepath.Join(tmpDir, "valid.jpg")
				createTestJPEG(t, inputPath, 50, 50, 80)
				return inputPath, "/invalid/path/that/does/not/exist/output.jpg"
			},
			wantErr: "failed to create output file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath, outputPath := tt.setupFunc()
			err := c.CompressFile(inputPath, outputPath)

			if err == nil {
				t.Fatal("CompressFile() expected error but got nil")
			}

			if tt.wantErr != "" && !contains(err.Error(), tt.wantErr) {
				t.Errorf("CompressFile() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestCompressor_Compress проверяет сжатие image.Image в байты
func TestCompressor_Compress(t *testing.T) {
	// Создаем тестовое изображение в памяти
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	t.Run("successful compression to bytes", func(t *testing.T) {
		c := New(80)
		data, err := c.Compress(img)

		if err != nil {
			t.Fatalf("Compress() unexpected error: %v", err)
		}

		if len(data) == 0 {
			t.Error("Compress() returned empty data")
		}

		// Проверяем что данные содержат JPEG magic bytes
		if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
			t.Error("Compressed data does not start with JPEG magic bytes (FF D8)")
		}
	})

	t.Run("different quality produces different sizes", func(t *testing.T) {
		highQuality := New(95)
		lowQuality := New(10)

		highData, err1 := highQuality.Compress(img)
		lowData, err2 := lowQuality.Compress(img)

		if err1 != nil || err2 != nil {
			t.Fatalf("Compress() errors: %v, %v", err1, err2)
		}

		if len(lowData) >= len(highData) {
			t.Errorf("Low quality size (%d) >= high quality size (%d)", len(lowData), len(highData))
		}
	})
}

// TestCompressJPEG проверяет функцию-обертку CompressJPEG
func TestCompressJPEG(t *testing.T) {
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	createTestJPEG(t, inputPath, 100, 100, 90)

	err := CompressJPEG(inputPath, outputPath, 75)
	if err != nil {
		t.Fatalf("CompressJPEG() unexpected error: %v", err)
	}

	// Проверяем что файл создан
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("CompressJPEG() did not create output file")
	}
}

// TestCompressor_PathCleaning проверяет очистку путей (security)
func TestCompressor_PathCleaning(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаем тестовый файл
	validPath := filepath.Join(tmpDir, "test.jpg")
	createTestJPEG(t, validPath, 50, 50, 80)

	tests := []struct {
		name       string
		inputPath  string
		outputPath string
		wantErr    bool
	}{
		{
			name:       "path with ..",
			inputPath:  filepath.Join(tmpDir, "..", filepath.Base(tmpDir), "test.jpg"),
			outputPath: filepath.Join(tmpDir, "out.jpg"),
			wantErr:    false, // filepath.Clean должен обработать
		},
		{
			name:       "path with extra slashes",
			inputPath:  filepath.Join(tmpDir, "//test.jpg"),
			outputPath: filepath.Join(tmpDir, "out.jpg"),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(80)
			err := c.CompressFile(tt.inputPath, tt.outputPath)

			if tt.wantErr && err == nil {
				t.Error("CompressFile() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CompressFile() unexpected error: %v", err)
			}
		})
	}
}

// BenchmarkCompressor_CompressFile бенчмарк сжатия файла
func BenchmarkCompressor_CompressFile(b *testing.B) {
	tmpDir := b.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	createTestJPEG(&testing.T{}, inputPath, 200, 200, 90)

	c := New(80)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "output"+string(rune(i))+".jpg")
		_ = c.CompressFile(inputPath, outputPath) // nolint:errcheck // benchmark ignores errors
	}
}

// BenchmarkCompressor_Compress бенчмарк сжатия image.Image
func BenchmarkCompressor_Compress(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	c := New(80)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Compress(img) // nolint:errcheck // benchmark ignores errors
	}
}

// BenchmarkNew бенчмарк создания компрессора
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(80)
	}
}

// Вспомогательная функция для проверки содержания строки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
