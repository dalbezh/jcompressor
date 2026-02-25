package compressor

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// TestCompressor_QualityBoundaries проверяет граничные значения quality
func TestCompressor_QualityBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		quality int
		wantMin int
		wantMax int
	}{
		{"negative extreme", -1000, 1, 1},
		{"negative", -1, 1, 1},
		{"zero", 0, 1, 1},
		{"minimum valid", 1, 1, 1},
		{"low", 10, 10, 10},
		{"medium", 50, 50, 50},
		{"high", 90, 90, 90},
		{"maximum valid", 100, 100, 100},
		{"above maximum", 101, 100, 100},
		{"very high", 200, 100, 100},
		{"extreme high", 999999, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.quality)
			got := c.Quality()

			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("New(%d).Quality() = %d, want between %d and %d",
					tt.quality, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestCompressor_FileExtensions проверяет обработку разных расширений
func TestCompressor_FileExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(80)

	// Создаем валидный JPEG для позитивных тестов
	validJPEG := filepath.Join(tmpDir, "valid.jpg")
	createTestJPEG(t, validJPEG, 50, 50, 90)

	tests := []struct {
		name       string
		inputExt   string
		shouldFail bool
	}{
		{".jpg", ".jpg", false},
		{".jpeg", ".jpeg", false},
		{".JPG uppercase", ".JPG", false},
		{".JPEG uppercase", ".JPEG", false},
		{".Jpg mixed case", ".Jpg", false},
		{".png", ".png", true},
		{".gif", ".gif", true},
		{".bmp", ".bmp", true},
		{".webp", ".webp", true},
		{".txt", ".txt", true},
		{"no extension", "", true},
		{".jp", ".jp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputPath string

			if !tt.shouldFail {
				// Для валидных расширений копируем валидный JPEG
				inputPath = filepath.Join(tmpDir, "test"+tt.inputExt)
				data, err := os.ReadFile(validJPEG)
				if err != nil {
					t.Fatalf("Failed to read valid JPEG: %v", err)
				}
				if err := os.WriteFile(inputPath, data, 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			} else {
				// Для невалидных создаем пустой файл
				inputPath = filepath.Join(tmpDir, "test"+tt.inputExt)
				if err := os.WriteFile(inputPath, []byte("test data"), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			outputPath := filepath.Join(tmpDir, "output_"+tt.name+".jpg")
			err := c.CompressFile(inputPath, outputPath)

			if tt.shouldFail && err == nil {
				t.Errorf("CompressFile(%s) expected error but got nil", tt.inputExt)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("CompressFile(%s) unexpected error: %v", tt.inputExt, err)
			}
		})
	}
}

// TestCompressor_ImageSizes проверяет работу с разными размерами изображений
func TestCompressor_ImageSizes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping slow test in short mode")
	}

	tmpDir := t.TempDir()
	c := New(80)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"tiny 1x1", 1, 1},
		{"small square", 10, 10},
		{"medium square", 100, 100},
		{"large square", 500, 500},
		{"wide rectangle", 800, 100},
		{"tall rectangle", 100, 800},
		{"hd 1920x1080", 1920, 1080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := filepath.Join(tmpDir, "input_"+tt.name+".jpg")
			outputPath := filepath.Join(tmpDir, "output_"+tt.name+".jpg")

			createTestJPEG(t, inputPath, tt.width, tt.height, 90)

			err := c.CompressFile(inputPath, outputPath)
			if err != nil {
				t.Errorf("CompressFile(%dx%d) failed: %v", tt.width, tt.height, err)
			}

			// Проверяем что выходной файл существует
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file not created for %dx%d", tt.width, tt.height)
			}
		})
	}
}

// TestCompressor_CompressionRatios проверяет степень сжатия
func TestCompressor_CompressionRatios(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаем исходное изображение высокого качества
	inputPath := filepath.Join(tmpDir, "input.jpg")
	createTestJPEG(t, inputPath, 200, 200, 95)

	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		t.Fatalf("Failed to stat input file: %v", err)
	}
	inputSize := inputInfo.Size()

	tests := []struct {
		name           string
		quality        int
		maxSizePercent float64 // максимальный процент от исходного размера
	}{
		{"very low quality", 10, 30},
		{"low quality", 30, 50},
		{"medium quality", 50, 70},
		{"high quality", 80, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, "output_q"+string(rune(tt.quality))+".jpg")

			c := New(tt.quality)
			if err := c.CompressFile(inputPath, outputPath); err != nil {
				t.Fatalf("CompressFile() failed: %v", err)
			}

			outputInfo, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("Failed to stat output file: %v", err)
			}
			outputSize := outputInfo.Size()

			sizePercent := float64(outputSize) / float64(inputSize) * 100

			if sizePercent > tt.maxSizePercent {
				t.Errorf("Quality %d: output size %.1f%% > max %.1f%%",
					tt.quality, sizePercent, tt.maxSizePercent)
			}

			t.Logf("Quality %d: compressed from %d to %d bytes (%.1f%%)",
				tt.quality, inputSize, outputSize, sizePercent)
		})
	}
}

// TestCompressor_Compress_ImageTypes проверяет сжатие разных типов image.Image
func TestCompressor_Compress_ImageTypes(t *testing.T) {
	c := New(80)
	width, height := 50, 50

	tests := []struct {
		createImg func() image.Image
		name      string
	}{
		{
			name: "RGBA",
			createImg: func() image.Image {
				img := image.NewRGBA(image.Rect(0, 0, width, height))
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
					}
				}
				return img
			},
		},
		{
			name: "Gray",
			createImg: func() image.Image {
				img := image.NewGray(image.Rect(0, 0, width, height))
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						img.Set(x, y, color.Gray{Y: 128})
					}
				}
				return img
			},
		},
		{
			name: "NRGBA",
			createImg: func() image.Image {
				img := image.NewNRGBA(image.Rect(0, 0, width, height))
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						img.Set(x, y, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
					}
				}
				return img
			},
		},
		{
			name: "YCbCr",
			createImg: func() image.Image {
				return image.NewYCbCr(image.Rect(0, 0, width, height), image.YCbCrSubsampleRatio444)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := tt.createImg()
			data, err := c.Compress(img)

			if err != nil {
				t.Errorf("Compress(%s) failed: %v", tt.name, err)
			}

			if len(data) == 0 {
				t.Errorf("Compress(%s) returned empty data", tt.name)
			}

			// Проверяем JPEG magic bytes
			if data[0] != 0xFF || data[1] != 0xD8 {
				t.Errorf("Compress(%s) data is not JPEG (magic bytes: %x %x)",
					tt.name, data[0], data[1])
			}
		})
	}
}

// TestCompressor_ConcurrentAccess проверяет безопасность при конкурентном доступе
func TestCompressor_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	tmpDir := t.TempDir()
	c := New(80)

	// Создаем входной файл
	inputPath := filepath.Join(tmpDir, "input.jpg")
	createTestJPEG(t, inputPath, 100, 100, 90)

	// Запускаем несколько горутин
	done := make(chan bool, 10)
	for i := range 10 {
		go func(id int) {
			outputPath := filepath.Join(tmpDir, fmt.Sprintf("output_%d.jpg", id))
			err := c.CompressFile(inputPath, outputPath)
			if err != nil {
				t.Errorf("Goroutine %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for range 10 {
		<-done
	}
}

// TestCompressor_MemoryLeaks проверяет на утечки памяти (простая проверка)
func TestCompressor_MemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Создаем изображение
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	c := New(80)

	// Многократно сжимаем
	for i := range 100 {
		_, err := c.Compress(img)
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}
	}

	// Если тест завершился без panic или OOM - всё ок
}

// TestCompressor_EdgeCases проверяет граничные случаи
func TestCompressor_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("output path is same as input", func(t *testing.T) {
		path := filepath.Join(tmpDir, "same.jpg")
		createTestJPEG(t, path, 50, 50, 90)

		c := New(50)
		err := c.CompressFile(path, path)

		// Должно работать (перезапись файла)
		if err != nil {
			t.Errorf("CompressFile(same path) failed: %v", err)
		}
	})

	t.Run("output directory does not exist but parent does", func(t *testing.T) {
		inputPath := filepath.Join(tmpDir, "input2.jpg")
		createTestJPEG(t, inputPath, 30, 30, 80)

		// Директория не существует
		outputPath := filepath.Join(tmpDir, "newdir", "output.jpg")

		c := New(80)
		err := c.CompressFile(inputPath, outputPath)

		// Должна быть ошибка (директория не создается автоматически)
		if err == nil {
			t.Error("Expected error for non-existent directory, got nil")
		}
	})
}
