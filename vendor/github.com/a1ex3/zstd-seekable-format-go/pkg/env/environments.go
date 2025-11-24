package env

// WEnvironment can be used to inject a custom file writer that is different from normal WriteCloser.
// This is useful when, for example there is a custom chunking code.
type WEnvironment interface {
	// WriteFrame is called each time frame is encoded and needs to be written upstream.
	WriteFrame(p []byte) (n int, err error)
	// WriteSeekTable is called on Close to flush the seek table.
	WriteSeekTable(p []byte) (n int, err error)
}

// REnvironment can be used to inject a custom file reader that is different from normal ReadSeeker.
// This is useful when, for example there is a custom chunking code.
type REnvironment interface {
	// GetFrameByIndex returns the compressed frame by its index.
	GetFrameByIndex(index FrameOffsetEntry) ([]byte, error)
	// ReadFooter returns buffer whose last 9 bytes are interpreted as a `Seek_Table_Footer`.
	ReadFooter() ([]byte, error)
	// ReadSkipFrame returns the full Seek Table Skippable frame
	// including the `Skippable_Magic_Number` and `Frame_Size`.
	ReadSkipFrame(skippableFrameOffset int64) ([]byte, error)
}
