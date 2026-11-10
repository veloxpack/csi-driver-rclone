package extsort

// Config holds configuration settings for external sorting operations.
// Pass a nil *Config to use DefaultConfig(). In a non-nil Config, a ChunkSize or
// NumWorkers below 1 and a negative ChanBuffSize or SortedChanBuffSize are replaced by
// their defaults, but a zero ChanBuffSize or SortedChanBuffSize means an unbuffered channel.
// The sorter works on its own copy, so one Config can be shared by several sorters.
type Config struct {
	// ChunkSize specifies the maximum number of records to store in each chunk
	// before writing to disk. Larger chunks use more memory but reduce I/O operations.
	// Default: 1,000,000 records. Values below 1 use the default.
	ChunkSize int

	// NumWorkers controls the maximum number of goroutines used for parallel
	// chunk sorting and merging. More workers can improve CPU utilization on multi-core systems.
	// Default: 2 workers. Values below 1 use the default.
	NumWorkers int

	// ChanBuffSize sets how many whole chunks can wait between reading the input and
	// sorting them. Each buffered chunk holds up to ChunkSize records in memory.
	// Default: 1. Zero means unbuffered; negative values use the default.
	ChanBuffSize int

	// SortedChanBuffSize sets the buffer size for the output channel that delivers
	// sorted results. Larger buffers allow more decoupling between sorting and consumption.
	// Default: 1000. Zero means unbuffered; negative values use the default.
	SortedChanBuffSize int

	// TempFilesDir specifies the directory for temporary files during sorting.
	// When empty (default), the library uses intelligent directory selection that
	// prefers disk-backed locations over potentially memory-backed filesystems
	// (like tmpfs on Linux). This helps prevent out-of-memory issues when sorting
	// datasets larger than available RAM.
	//
	// For production use with large datasets, it's recommended to explicitly set
	// this to a known disk-backed directory (such as "/var/tmp" on Unix systems)
	// to ensure optimal performance and avoid memory limitations. On Linux systems,
	// prefer "/var/tmp" over "/tmp" since "/tmp" may be mounted as tmpfs.
	//
	// Default: "" (intelligent selection).
	TempFilesDir string
}

// DefaultConfig returns a Config with sensible default values optimized for
// general-purpose external sorting. These defaults balance memory usage,
// I/O efficiency, and parallelism for typical workloads.
func DefaultConfig() *Config {
	return &Config{
		ChunkSize:          int(1e6), // 1M
		NumWorkers:         2,
		ChanBuffSize:       1,
		SortedChanBuffSize: 1000,
		TempFilesDir:       "",
	}
}

// mergeConfig returns a validated and normalized copy of c, replacing invalid values
// with defaults. If c is nil, returns DefaultConfig().
// The caller's Config is never modified, since it may be shared by other sorters.
func mergeConfig(c *Config) *Config {
	d := DefaultConfig()
	if c == nil {
		return d
	}
	merged := *c
	if merged.ChunkSize < 1 {
		merged.ChunkSize = d.ChunkSize
	}
	if merged.NumWorkers < 1 {
		merged.NumWorkers = d.NumWorkers
	}
	if merged.ChanBuffSize < 0 {
		merged.ChanBuffSize = d.ChanBuffSize
	}
	if merged.SortedChanBuffSize < 0 {
		merged.SortedChanBuffSize = d.SortedChanBuffSize
	}
	return &merged
}
