// Package tempfile provides an abstraction for creating virtual temporary files
// that are mapped to sections of a single physical file on disk. This design minimizes
// file descriptor usage while supporting efficient sequential writes and concurrent reads.
//
// The package supports two main workflows:
//  1. Write data sequentially to multiple virtual files using FileWriter
//  2. Read data back from any virtual file section using TempReader
//
// The implementation handles cross-platform differences in file cleanup behavior,
// with automatic cleanup on Unix systems and explicit cleanup on Windows.
package tempfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// file IO buffer size for each file
const fileBufferSize = 1 << 16 // 64k

// filename prefix for files put in temp directory
var mergeFilenamePrefix = fmt.Sprintf("extsort_%d_", os.Getpid())

// FileWriter provides sequential writing to virtual temporary file sections.
// Each "virtual file" corresponds to a section of the underlying physical file,
// allowing multiple logical files to share a single file descriptor and reduce
// system resource usage during external sorting operations.
type FileWriter struct {
	file         *os.File
	bufWriter    *bufio.Writer
	sections     []int64
	needsCleanup bool // true if manual cleanup is needed (Windows)
}

type fileReader struct {
	file         *os.File
	sections     []int64
	readers      []*bufio.Reader
	needsCleanup bool   // true if manual cleanup is needed (Windows)
	filename     string // filename for cleanup
}

// New creates a new FileWriter for virtual temporary files in the specified directory.
// If dir is empty, the OS default temporary directory is used (e.g., /tmp on Unix).
// The function attempts automatic cleanup on Unix systems by unlinking the file immediately,
// while Windows requires explicit cleanup when the FileWriter is closed.
func New(dir string) (*FileWriter, error) {
	var w FileWriter
	var err error
	w.file, err = os.CreateTemp(dir, mergeFilenamePrefix)
	if err != nil {
		return nil, err
	}

	// Try immediate unlink for automatic cleanup (works on Unix)
	// If it fails (likely Windows), we'll do manual cleanup later
	if err = os.Remove(w.file.Name()); err != nil {
		w.needsCleanup = true // Manual cleanup needed
	}

	w.bufWriter = bufio.NewWriterSize(w.file, fileBufferSize)
	w.sections = make([]int64, 0, 10)

	return &w, nil
}

// Size returns the total number of virtual file sections created.
// This includes the current section being written plus all completed sections.
func (w *FileWriter) Size() int {
	// we add one because we only write to the sections when we are done
	return len(w.sections) + 1
}

// Name returns the full filesystem path of the underlying physical temporary file.
// This is primarily useful for debugging and logging purposes.
func (w *FileWriter) Name() string {
	return w.file.Name()
}

// Close terminates the FileWriter, flushes any buffered data, closes the underlying file,
// and removes it from disk if manual cleanup is required. This operation is irreversible
// and should only be called when abandoning the temporary file (e.g., on error).
// Use Save() instead to transition from writing to reading.
func (w *FileWriter) Close() error {
	filename := w.file.Name()
	err := w.file.Close()
	w.sections = nil
	w.bufWriter = nil

	// Only attempt manual cleanup if needed (Windows case)
	if w.needsCleanup {
		if removeErr := os.Remove(filename); removeErr != nil && err == nil {
			err = removeErr
		}
	}

	return err
}

// Write appends data to the current virtual file section.
// Data is buffered for efficiency and will be flushed when Next() or Save() is called.
func (w *FileWriter) Write(p []byte) (int, error) {
	return w.bufWriter.Write(p)
}

// WriteString appends a string to the current virtual file section.
// This is more efficient than Write() for string data as it avoids byte slice conversion.
func (w *FileWriter) WriteString(s string) (int, error) {
	return w.bufWriter.WriteString(s)
}

// Next finalizes the current virtual file section and prepares for writing the next section.
// It flushes buffered data and records the section boundary for later reading.
// Returns the file offset where the next section will begin.
func (w *FileWriter) Next() (int64, error) {
	// save offsets
	err := w.bufWriter.Flush()
	if err != nil {
		return 0, err
	}
	pos, err := w.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	w.sections = append(w.sections, pos)

	return pos, nil
}

// Save finalizes all virtual file sections and returns a TempReader for accessing the data.
// After calling Save(), the FileWriter can no longer be used for writing.
// The returned TempReader allows concurrent access to any virtual file section.
func (w *FileWriter) Save() (TempReader, error) {
	_, err := w.Next()
	if err != nil {
		return nil, err
	}
	err = w.file.Sync()
	if err != nil {
		return nil, err
	}

	if w.needsCleanup {
		// Windows case: close file and reopen for reading
		filename := w.file.Name()
		err = w.file.Close()
		if err != nil {
			return nil, err
		}
		return newTempReader(filename, w.sections, w.needsCleanup)
	} else {
		// Unix case: file is unlinked, reuse the same file handle
		return newTempReaderFromFile(w.file, w.sections, w.needsCleanup)
	}
}

func newTempReader(filename string, sections []int64, needsCleanup bool) (*fileReader, error) {
	// create TempReader by opening file by name
	var err error
	var r fileReader
	r.file, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	r.sections = sections
	r.readers = make([]*bufio.Reader, len(r.sections))
	r.needsCleanup = needsCleanup
	r.filename = filename

	offset := int64(0)
	for i, end := range r.sections {
		section := io.NewSectionReader(r.file, offset, end-offset)
		offset = end
		r.readers[i] = bufio.NewReaderSize(section, fileBufferSize)
	}

	return &r, nil
}

func newTempReaderFromFile(file *os.File, sections []int64, needsCleanup bool) (*fileReader, error) {
	// create TempReader by reusing existing file handle
	var r fileReader
	r.file = file
	r.sections = sections
	r.readers = make([]*bufio.Reader, len(r.sections))
	r.needsCleanup = needsCleanup
	r.filename = file.Name()

	offset := int64(0)
	for i, end := range r.sections {
		section := io.NewSectionReader(r.file, offset, end-offset)
		offset = end
		r.readers[i] = bufio.NewReaderSize(section, fileBufferSize)
	}

	return &r, nil
}

func (r *fileReader) Close() error {
	r.readers = nil
	err := r.file.Close()

	// Only attempt manual cleanup if needed (Windows case)
	if r.needsCleanup {
		if removeErr := os.Remove(r.filename); removeErr != nil && err == nil {
			err = removeErr
		}
	}

	return err
}

func (r *fileReader) Size() int {
	return len(r.readers)
}

func (r *fileReader) Read(i int) *bufio.Reader {
	if i < 0 || i >= len(r.readers) {
		panic("tempfile: read request out of range")
	}
	return r.readers[i]
}
