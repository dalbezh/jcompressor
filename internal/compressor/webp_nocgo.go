//go:build !cgo || nowebp

package compressor

import (
	"errors"
	"image"
)

var ErrWebPNotSupported = errors.New("WebP support is not available in this build (requires CGO and libwebp)")

// ConvertToWebP returns an error indicating WebP is not supported
func ConvertToWebP(img image.Image, outputPath string, quality int) error {
	return ErrWebPNotSupported
}

// CompressToWebP returns an error indicating WebP is not supported
func CompressToWebP(inputPath, outputPath string, quality int) error {
	return ErrWebPNotSupported
}
