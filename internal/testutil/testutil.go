package testutil

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// CreateTestJPEG создает тестовый JPEG файл с заданными параметрами
func CreateTestJPEG(t *testing.T, path string, width, height, quality int) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	img := CreateTestImage(width, height)

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}
}

// CreateTestImage создает тестовое изображение с градиентом
func CreateTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			r := uint8((x * 255) / width)                // #nosec G115 // safe: ratio is 0-1
			g := uint8((y * 255) / height)               // #nosec G115 // safe: ratio is 0-1
			b := uint8((x + y) * 255 / (width + height)) // #nosec G115 // safe: ratio is 0-1
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// CreateSolidColorImage создает изображение одного цвета
func CreateSolidColorImage(width, height int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			img.Set(x, y, c)
		}
	}

	return img
}

// CreateCheckerboardImage создает изображение в виде шахматной доски
func CreateCheckerboardImage(width, height, blockSize int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			blockX := x / blockSize
			blockY := y / blockSize

			if (blockX+blockY)%2 == 0 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}

	return img
}

// FileExists проверяет существование файла
func FileExists(t *testing.T, path string) bool {
	t.Helper()

	_, err := os.Stat(path)
	return err == nil
}

// AssertFileExists проверяет что файл существует и падает если нет
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if !FileExists(t, path) {
		t.Fatalf("File does not exist: %s", path)
	}
}

// GetFileSize возвращает размер файла в байтах
func GetFileSize(t *testing.T, path string) int64 {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", path, err)
	}

	return info.Size()
}

// CompareFileSizes сравнивает размеры двух файлов
func CompareFileSizes(t *testing.T, path1, path2 string) (size1, size2 int64) {
	t.Helper()

	size1 = GetFileSize(t, path1)
	size2 = GetFileSize(t, path2)

	return size1, size2
}

// AssertJPEGValid проверяет что файл является валидным JPEG
func AssertJPEGValid(t *testing.T, path string) {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", path, err)
	}
	defer file.Close()

	_, err = jpeg.Decode(file)
	if err != nil {
		t.Fatalf("File %s is not a valid JPEG: %v", path, err)
	}
}

// ReadJPEGImage читает JPEG файл и возвращает image.Image
func ReadJPEGImage(t *testing.T, path string) image.Image {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", path, err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode JPEG %s: %v", path, err)
	}

	return img
}

// CreateTempFile создает временный файл с заданным содержимым
func CreateTempFile(t *testing.T, dir, pattern string, content []byte) string {
	t.Helper()

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return file.Name()
}

// CleanupFiles удаляет список файлов (для defer)
func CleanupFiles(t *testing.T, paths ...string) {
	t.Helper()

	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove %s: %v", path, err)
		}
	}
}

// SkipIfShort пропускает тест если установлен флаг -short
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()

	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}
