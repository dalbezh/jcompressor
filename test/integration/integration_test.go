//go:build integration
// +build integration

package integration

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// createTestJPEG создает тестовый JPEG файл
func createTestJPEG(t *testing.T, path string, width, height int) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 200, B: 255, A: 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}
}

// buildBinary собирает бинарник для тестов
func buildBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "jcompressor")
	cmd := exec.Command("go", "build", "-o", binPath, "../../cmd/jcompressor")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	return binPath
}

// TestIntegration_BasicCompression проверяет базовое сжатие
func TestIntegration_BasicCompression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputDir := filepath.Join(tmpDir, "output")

	createTestJPEG(t, inputPath, 200, 200)

	// Запускаем компрессор
	cmd := exec.Command(binPath, "-q", "80", inputPath, outputDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Command failed: %v\n%s", err, output)
	}

	// Проверяем что выходной файл создан
	expectedOutput := filepath.Join(outputDir, "input.jpg")
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Errorf("Output file not created: %s", expectedOutput)
	}

	// Проверяем что это валидный JPEG
	f, err := os.Open(expectedOutput)
	if err != nil {
		t.Fatalf("Failed to open output: %v", err)
	}
	defer f.Close()

	_, err = jpeg.Decode(f)
	if err != nil {
		t.Errorf("Output is not valid JPEG: %v", err)
	}
}

// TestIntegration_DefaultOutputDir проверяет дефолтную директорию
func TestIntegration_DefaultOutputDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)
	tmpDir := t.TempDir()

	// Меняем рабочую директорию
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	inputPath := filepath.Join(tmpDir, "test.jpg")
	createTestJPEG(t, inputPath, 100, 100)

	cmd := exec.Command(binPath, inputPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Command failed: %v\n%s", err, output)
	}

	// Проверяем что файл создан в ./compressed
	expectedOutput := filepath.Join(tmpDir, "compressed", "test.jpg")
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Errorf("Output file not created in default location: %s", expectedOutput)
	}
}

// TestIntegration_HelpFlag проверяет флаг помощи
func TestIntegration_HelpFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)

	tests := []string{"-h", "--help"}

	for _, flag := range tests {
		t.Run(flag, func(t *testing.T) {
			cmd := exec.Command(binPath, flag)
			output, err := cmd.CombinedOutput()

			// Help должен выйти с кодом 0
			if err != nil {
				t.Errorf("Help command failed: %v", err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Usage:") {
				t.Error("Help output does not contain 'Usage:'")
			}
			if !strings.Contains(outputStr, "jcompressor") {
				t.Error("Help output does not contain 'jcompressor'")
			}
		})
	}
}

// TestIntegration_InvalidArguments проверяет обработку ошибок
func TestIntegration_InvalidArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"invalid quality", []string{"-q", "999", "test.jpg"}},
		{"nonexistent file", []string{"nonexistent.jpg"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			_, err := cmd.CombinedOutput()

			// Должен завершиться с ошибкой
			if err == nil {
				t.Error("Expected command to fail but it succeeded")
			}
		})
	}
}

// TestIntegration_QualityEffect проверяет влияние качества на размер
func TestIntegration_QualityEffect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "input.jpg")
	createTestJPEG(t, inputPath, 300, 300)

	highQualityDir := filepath.Join(tmpDir, "high")
	lowQualityDir := filepath.Join(tmpDir, "low")

	// Сжимаем с высоким качеством
	cmd1 := exec.Command(binPath, "-q", "95", inputPath, highQualityDir)
	if output, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("High quality compression failed: %v\n%s", err, output)
	}

	// Сжимаем с низким качеством
	cmd2 := exec.Command(binPath, "-q", "20", inputPath, lowQualityDir)
	if output, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("Low quality compression failed: %v\n%s", err, output)
	}

	// Сравниваем размеры
	highFile := filepath.Join(highQualityDir, "input.jpg")
	lowFile := filepath.Join(lowQualityDir, "input.jpg")

	highInfo, _ := os.Stat(highFile)
	lowInfo, _ := os.Stat(lowFile)

	if lowInfo.Size() >= highInfo.Size() {
		t.Errorf("Low quality (%d bytes) >= high quality (%d bytes)",
			lowInfo.Size(), highInfo.Size())
	}
}

// TestIntegration_WebPFlag проверяет флаг WebP (должен показать ошибку в no-CGO сборке)
func TestIntegration_WebPFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "test.jpg")
	createTestJPEG(t, inputPath, 50, 50)

	cmd := exec.Command(binPath, "-webp", inputPath, tmpDir)
	output, err := cmd.CombinedOutput()

	// В no-CGO сборке должна быть ошибка
	if err == nil {
		t.Log("WebP might be supported in this build")
	} else {
		// Проверяем сообщение об ошибке
		if !strings.Contains(string(output), "WebP") {
			t.Errorf("Expected WebP error message, got: %s", output)
		}
	}
}

// TestIntegration_MultipleRuns проверяет что можно запускать многократно
func TestIntegration_MultipleRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildBinary(t)
	tmpDir := t.TempDir()

	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputDir := filepath.Join(tmpDir, "output")

	createTestJPEG(t, inputPath, 100, 100)

	// Запускаем несколько раз
	for i := 0; i < 3; i++ {
		cmd := exec.Command(binPath, "-q", "70", inputPath, outputDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Run %d failed: %v\n%s", i+1, err, output)
		}
	}

	// Проверяем что файл перезаписывается
	outputFile := filepath.Join(outputDir, "input.jpg")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file not found after multiple runs")
	}
}
