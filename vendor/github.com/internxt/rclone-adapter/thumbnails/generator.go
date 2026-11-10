package thumbnails

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"strings"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// maxSourcePixels is the largest image, in pixels, a thumbnail is made from.
const maxSourcePixels = 128 << 20

// maxProgressivePixels is the largest progressive JPEG, in pixels, a
// thumbnail is made from
const maxProgressivePixels = 8 << 20

// shrinkFactor is how many times the thumbnail size fit shrinks an image to
// by averaging before interpolating it down to the thumbnail size.
const shrinkFactor = 3

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
	img, orientation, err := decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, 0, err
	}
	thumbnailBytes, err := encodeThumbnail(img, orientation, cfg)
	if err != nil {
		return nil, 0, err
	}
	return thumbnailBytes, int64(len(thumbnailBytes)), nil
}

// Generator makes a thumbnail from the image data written to it.
type Generator struct {
	pw    *io.PipeWriter
	done  chan struct{}
	thumb []byte
	err   error
}

// NewGenerator returns a Generator making a thumbnail as described by cfg.
//
// Finish must be called to release its resources.
func NewGenerator(cfg *Config) *Generator {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	pr, pw := io.Pipe()
	g := &Generator{pw: pw, done: make(chan struct{})}
	go func() {
		defer close(g.done)
		img, orientation, err := decode(pr)
		_ = pr.Close()
		if err != nil {
			g.err = err
			return
		}
		g.thumb, g.err = encodeThumbnail(img, orientation, cfg)
	}()
	return g
}

// Write passes p to the decoder.
//
// It never returns an error so a failure to make the thumbnail can't fail
// the stream the data is teed from.
func (g *Generator) Write(p []byte) (int, error) {
	_, _ = g.pw.Write(p)
	return len(p), nil
}

// Finish marks the end of the image data and returns the thumbnail as PNG
// bytes. A non-nil err says the data is incomplete for that reason, in which
// case Finish returns without waiting for a thumbnail.
func (g *Generator) Finish(err error) ([]byte, error) {
	_ = g.pw.CloseWithError(err)
	<-g.done
	if err != nil {
		return nil, err
	}
	return g.thumb, g.err
}

// decode decodes the image read from r and returns it with its EXIF
// orientation.
func decode(r io.Reader) (image.Image, int, error) {
	var header bytes.Buffer
	imgCfg, format, err := image.DecodeConfig(io.TeeReader(r, &header))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode image: %w", err)
	}
	pixels := int64(imgCfg.Width) * int64(imgCfg.Height)
	if pixels > maxSourcePixels {
		return nil, 0, fmt.Errorf("image too large for thumbnail: %dx%d", imgCfg.Width, imgCfg.Height)
	}
	orientation := 1
	if format == "jpeg" {
		var progressive bool
		orientation, progressive = jpegHeader(header.Bytes())
		if progressive && pixels > maxProgressivePixels {
			return nil, 0, fmt.Errorf("progressive JPEG too large for thumbnail: %dx%d", imgCfg.Width, imgCfg.Height)
		}
	}

	img, _, err := image.Decode(io.MultiReader(&header, r))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, orientation, nil
}

// encodeThumbnail returns an upright thumbnail of img, which has the given
// EXIF orientation, as PNG bytes.
func encodeThumbnail(img image.Image, orientation int, cfg *Config) ([]byte, error) {
	// The image is resized as stored and then turned upright, so the
	// bounds are swapped for orientations which turn it on its side.
	maxWidth, maxHeight := cfg.MaxWidth, cfg.MaxHeight
	if orientation >= 5 {
		maxWidth, maxHeight = maxHeight, maxWidth
	}
	thumb := orient(fit(img, maxWidth, maxHeight), orientation)

	var buf bytes.Buffer
	if err := png.Encode(&buf, thumb); err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}
	return buf.Bytes(), nil
}

// fit resizes src to fit within maxWidth x maxHeight, preserving aspect ratio.
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

	if srcWidth > shrinkFactor*newWidth && srcHeight > shrinkFactor*newHeight {
		if small := shrink(src, shrinkFactor*newWidth, shrinkFactor*newHeight); small != nil {
			src, srcBounds = small, small.Bounds()
		}
	}

	dst := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, srcBounds, draw.Over, nil)

	return dst
}

// shrink returns src reduced to width x height by averaging the source
// pixels each destination pixel covers, which is fast and alias free for
// the large reductions thumbnails need.
func shrink(src image.Image, width, height int) image.Image {
	switch src := src.(type) {
	case *image.YCbCr:
		r := src.Rect
		c := chromaRect(r, src.SubsampleRatio)
		if c.Dx() < width || c.Dy() < height {
			return nil
		}
		y := shrinkPlane(src.Y[src.YOffset(r.Min.X, r.Min.Y):], src.YStride, r.Dx(), r.Dy(), width, height)
		coff := src.COffset(r.Min.X, r.Min.Y)
		cb := shrinkPlane(src.Cb[coff:], src.CStride, c.Dx(), c.Dy(), width, height)
		cr := shrinkPlane(src.Cr[coff:], src.CStride, c.Dx(), c.Dy(), width, height)
		dst := image.NewNRGBA(image.Rect(0, 0, width, height))
		for i := range y {
			rr, gg, bb := color.YCbCrToRGB(y[i], cb[i], cr[i])
			dst.Pix[4*i+0] = rr
			dst.Pix[4*i+1] = gg
			dst.Pix[4*i+2] = bb
			dst.Pix[4*i+3] = 0xff
		}
		return dst
	case *image.Gray:
		r := src.Rect
		dst := image.NewGray(image.Rect(0, 0, width, height))
		dst.Pix = shrinkPlane(src.Pix[src.PixOffset(r.Min.X, r.Min.Y):], src.Stride, r.Dx(), r.Dy(), width, height)
		return dst
	}
	return nil
}

// chromaRect returns the bounds of the chroma planes of a YCbCr image with
// luma bounds r, as the image package lays them out.
func chromaRect(r image.Rectangle, s image.YCbCrSubsampleRatio) image.Rectangle {
	// dx and dy are how many luma samples share a chroma sample across and
	// down.
	dx, dy := 1, 1
	switch s {
	case image.YCbCrSubsampleRatio422:
		dx = 2
	case image.YCbCrSubsampleRatio420:
		dx, dy = 2, 2
	case image.YCbCrSubsampleRatio440:
		dy = 2
	case image.YCbCrSubsampleRatio411:
		dx = 4
	case image.YCbCrSubsampleRatio410:
		dx, dy = 4, 2
	}
	return image.Rect(r.Min.X/dx, r.Min.Y/dy, (r.Max.X+dx-1)/dx, (r.Max.Y+dy-1)/dy)
}

// shrinkPlane reduces the w x h plane of samples starting at pix, with rows
// stride apart, to dw x dh by averaging. dw and dh must be at most w and h.
func shrinkPlane(pix []uint8, stride, w, h, dw, dh int) []uint8 {
	out := make([]uint8, dw*dh)
	col := make([]int, w)
	cols := make([]uint32, dw)
	for x := range col {
		col[x] = x * dw / w
		cols[col[x]]++
	}
	sums := make([]uint32, dw)
	sy := 0
	for dy := range dh {
		clear(sums)
		y1 := (dy + 1) * h / dh
		rows := uint32(y1 - sy)
		for ; sy < y1; sy++ {
			row := pix[sy*stride : sy*stride+w]
			for x, v := range row {
				sums[col[x]] += uint32(v)
			}
		}
		for dx, sum := range sums {
			n := cols[dx] * rows
			out[dy*dw+dx] = uint8((sum + n/2) / n)
		}
	}
	return out
}
