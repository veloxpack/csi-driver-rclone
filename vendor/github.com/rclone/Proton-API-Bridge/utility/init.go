package utility

import (
	"io"
	"log"
)

func SetupLog() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// CheckClose is a utility function used to check the return from
// Close in a defer statement.
func CheckClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}
