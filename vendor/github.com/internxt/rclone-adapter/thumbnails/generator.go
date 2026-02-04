package thumbnails

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"strings"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// supportedFormats maps file extensions to whether they support thumbnail generation
var supportedFormats = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"webp": true,
	"gif":  true,
	"tiff": true,
	"tif":  true,
}

// IsSupportedFormat checks if the given file extension supports thumbnail generation
func IsSupportedFormat(ext string) bool {
	normalized := strings.ToLower(strings.TrimPrefix(ext, "."))
	return supportedFormats[normalized]
}

// Generate creates a thumbnail from the provided image data.
// It resizes the image to fit within maxWidth x maxHeight while preserving aspect ratio,
// and returns the thumbnail as PNG bytes.
func Generate(imageData []byte, cfg *Config) ([]byte, int64, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	thumb := fit(img, cfg.MaxWidth, cfg.MaxHeight)

	var buf bytes.Buffer
	if err := png.Encode(&buf, thumb); err != nil {
		return nil, 0, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	thumbnailBytes := buf.Bytes()
	return thumbnailBytes, int64(len(thumbnailBytes)), nil
}

// fit resizes src to fit within maxWidth x maxHeight, preserving aspect ratio,
// using high-quality Catmull-Rom interpolation (similar to Lanczos).
func fit(src image.Image, maxWidth, maxHeight int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	if srcWidth <= maxWidth && srcHeight <= maxHeight {
		return src
	}

	ratio := float64(srcWidth) / float64(srcHeight)
	newWidth := maxWidth
	newHeight := int(float64(newWidth) / ratio)

	if newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = int(float64(newHeight) * ratio)
	}

	if newWidth <= 0 {
		newWidth = 1
	}

	if newHeight <= 0 {
		newHeight = 1
	}

	dst := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, srcBounds, draw.Over, nil)

	return dst
}
