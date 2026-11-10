package thumbnails

import (
	"bytes"
	"encoding/binary"
	"image"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// jpegHeader reads the header of the JPEG whose leading bytes are b up to its
// frame header. It returns the EXIF orientation (1-8), or 1 if there is none,
// and whether the image is progressive.
func jpegHeader(b []byte) (orientation int, progressive bool) {
	orientation = 1
	if len(b) < 2 || b[0] != 0xff || b[1] != 0xd8 {
		return orientation, false
	}
	for i := 2; i+4 <= len(b) && b[i] == 0xff; {
		marker := b[i+1]
		if marker >= 0xc0 && marker <= 0xcf && marker != 0xc4 && marker != 0xc8 && marker != 0xcc {
			return orientation, marker&3 == 2
		}
		if marker == 0xda {
			break
		}
		length := int(binary.BigEndian.Uint16(b[i+2:]))
		end := i + 2 + length
		if length < 2 || end > len(b) {
			break
		}
		if seg := b[i+4 : end]; marker == 0xe1 && bytes.HasPrefix(seg, []byte("Exif\x00\x00")) {
			orientation = exifOrientation(seg[6:])
		}
		i = end
	}
	return orientation, false
}

// exifOrientation returns the orientation tag from the first IFD of the
// TIFF structured EXIF data t, or 1 if it has none or t is malformed.
func exifOrientation(t []byte) int {
	if len(t) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	ifd := int(order.Uint32(t[4:]))
	if ifd < 8 || ifd+2 > len(t) {
		return 1
	}
	n := int(order.Uint16(t[ifd:]))
	for i := range n {
		e := ifd + 2 + 12*i
		if e+12 > len(t) {
			break
		}
		if order.Uint16(t[e:]) == 0x0112 && order.Uint16(t[e+2:]) == 3 {
			if o := int(order.Uint16(t[e+8:])); o >= 1 && o <= 8 {
				return o
			}
			break
		}
	}
	return 1
}

// orient returns img transformed for display according to the EXIF
// orientation o.
func orient(img image.Image, o int) image.Image {
	if o <= 1 || o > 8 {
		return img
	}
	b := img.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())
	var s2d f64.Aff3
	switch o {
	case 2: // mirrored horizontally
		s2d = f64.Aff3{-1, 0, w, 0, 1, 0}
	case 3: // rotated 180
		s2d = f64.Aff3{-1, 0, w, 0, -1, h}
	case 4: // mirrored vertically
		s2d = f64.Aff3{1, 0, 0, 0, -1, h}
	case 5: // mirrored horizontally and rotated 270 clockwise
		s2d = f64.Aff3{0, 1, 0, 1, 0, 0}
	case 6: // rotated 90 clockwise
		s2d = f64.Aff3{0, -1, h, 1, 0, 0}
	case 7: // mirrored horizontally and rotated 90 clockwise
		s2d = f64.Aff3{0, -1, h, -1, 0, w}
	case 8: // rotated 270 clockwise
		s2d = f64.Aff3{0, 1, 0, -1, 0, w}
	}
	minX, minY := float64(b.Min.X), float64(b.Min.Y)
	s2d[2] -= s2d[0]*minX + s2d[1]*minY
	s2d[5] -= s2d[3]*minX + s2d[4]*minY

	dw, dh := b.Dx(), b.Dy()
	if o >= 5 {
		dw, dh = dh, dw
	}
	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	xdraw.NearestNeighbor.Transform(dst, s2d, img, b, xdraw.Src, nil)
	return dst
}
