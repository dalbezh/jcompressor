//go:build !cgo || nowebp

package compressor

import (
	"errors"
	"image"
	"image/color"
	"testing"
)

// TestWebP_NotSupported проверяет что WebP недоступен в no-CGO сборке
func TestWebP_NotSupported(t *testing.T) {
	t.Run("ConvertToWebP returns error", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		err := ConvertToWebP(img, "/tmp/output.webp", 80)

		if err == nil {
			t.Fatal("ConvertToWebP() expected error but got nil")
		}

		if !errors.Is(err, ErrWebPNotSupported) {
			t.Errorf("ConvertToWebP() error = %v, want ErrWebPNotSupported", err)
		}
	})

	t.Run("CompressToWebP returns error", func(t *testing.T) {
		err := CompressToWebP("/tmp/input.jpg", "/tmp/output.webp", 80)

		if err == nil {
			t.Fatal("CompressToWebP() expected error but got nil")
		}

		if !errors.Is(err, ErrWebPNotSupported) {
			t.Errorf("CompressToWebP() error = %v, want ErrWebPNotSupported", err)
		}
	})

	t.Run("ErrWebPNotSupported has correct message", func(t *testing.T) {
		expectedMsg := "WebP support is not available in this build"
		if !contains(ErrWebPNotSupported.Error(), expectedMsg) {
			t.Errorf("ErrWebPNotSupported.Error() = %q, want to contain %q",
				ErrWebPNotSupported.Error(), expectedMsg)
		}
	})
}

// TestWebP_StubFunctionsSignature проверяет сигнатуры функций
func TestWebP_StubFunctionsSignature(t *testing.T) {
	// Проверяем что функции имеют правильные сигнатуры
	var _ = ConvertToWebP
	var _ = CompressToWebP
	var _ = ErrWebPNotSupported
}

// TestWebP_ErrorIsComparable проверяет что ошибку можно сравнивать
func TestWebP_ErrorIsComparable(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	err1 := ConvertToWebP(img, "/tmp/test.webp", 80)
	err2 := CompressToWebP("/tmp/input.jpg", "/tmp/output.webp", 80)

	// Обе функции должны возвращать одну и ту же ошибку
	if !errors.Is(err1, ErrWebPNotSupported) {
		t.Error("ConvertToWebP() does not return ErrWebPNotSupported")
	}
	if !errors.Is(err2, ErrWebPNotSupported) {
		t.Error("CompressToWebP() does not return ErrWebPNotSupported")
	}

	// Проверяем что это именно та же ошибка
	if err1 != ErrWebPNotSupported {
		t.Error("ConvertToWebP() returned different error instance")
	}
	if err2 != ErrWebPNotSupported {
		t.Error("CompressToWebP() returned different error instance")
	}
}

// TestWebP_WithDifferentQuality проверяет что качество не влияет на ошибку
func TestWebP_WithDifferentQuality(t *testing.T) {
	qualities := []int{1, 50, 100}
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))

	for _, q := range qualities {
		t.Run("quality", func(t *testing.T) {
			err := ConvertToWebP(img, "/tmp/test.webp", q)
			if !errors.Is(err, ErrWebPNotSupported) {
				t.Errorf("ConvertToWebP(quality=%d) error = %v, want ErrWebPNotSupported", q, err)
			}
		})
	}
}

// TestWebP_WithNilImage проверяет поведение с nil изображением
func TestWebP_WithNilImage(t *testing.T) {
	err := ConvertToWebP(nil, "/tmp/test.webp", 80)
	if err == nil {
		t.Fatal("ConvertToWebP(nil) expected error but got nil")
	}
	// Должна вернуться ErrWebPNotSupported, а не паника
	if !errors.Is(err, ErrWebPNotSupported) {
		t.Errorf("ConvertToWebP(nil) error = %v, want ErrWebPNotSupported", err)
	}
}

// TestWebP_WithEmptyPaths проверяет поведение с пустыми путями
func TestWebP_WithEmptyPaths(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		outputPath string
	}{
		{"empty input", "", "/tmp/output.webp"},
		{"empty output", "/tmp/input.jpg", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompressToWebP(tt.inputPath, tt.outputPath, 80)
			if err == nil {
				t.Fatal("CompressToWebP() expected error but got nil")
			}
			// Stub всегда возвращает ErrWebPNotSupported независимо от аргументов
			if !errors.Is(err, ErrWebPNotSupported) {
				t.Errorf("CompressToWebP() error = %v, want ErrWebPNotSupported", err)
			}
		})
	}
}

// BenchmarkWebP_StubOverhead бенчмарк для измерения overhead stub функций
func BenchmarkWebP_StubOverhead(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	b.Run("ConvertToWebP", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ConvertToWebP(img, "/tmp/test.webp", 80) // nolint:errcheck // benchmark stub
		}
	})

	b.Run("CompressToWebP", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = CompressToWebP("/tmp/input.jpg", "/tmp/output.webp", 80) // nolint:errcheck // benchmark stub
		}
	})
}
