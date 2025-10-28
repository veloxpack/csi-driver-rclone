package lib

import (
	"errors"
	"io"
)

func CopyAt(dst io.WriterAt, writeOff int64, src io.Reader) (written int64, err error) {
	size := int64(32 * 1024)
	buf := make([]byte, size)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.WriteAt(buf[0:nr], writeOff+written)
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return
}
