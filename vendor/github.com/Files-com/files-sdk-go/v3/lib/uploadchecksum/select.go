package uploadchecksum

import (
	"sync"
	"time"
)

const defaultBestAlgorithmTiePercent = 5

var (
	bestAlgorithmOnce sync.Once
	bestAlgorithm     Algorithm
)

type BestAlgorithmOptions struct {
	Candidates []Algorithm
	SampleSize int
	Iterations int
	TiePercent int
}

type AlgorithmBenchmark struct {
	Algorithm      Algorithm
	BytesProcessed int64
	Duration       time.Duration
}

func BestAlgorithmForPlatform() Algorithm {
	bestAlgorithmOnce.Do(func() {
		best, err := SelectBestAlgorithm(BestAlgorithmOptions{})
		if err != nil {
			best = ChecksumAlgorithmCRC32C
		}
		bestAlgorithm = best
	})
	return bestAlgorithm
}

func SelectBestAlgorithm(options BestAlgorithmOptions) (Algorithm, error) {
	results, err := BenchmarkAlgorithms(options)
	if err != nil {
		return "", err
	}
	tiePercent := options.TiePercent
	if tiePercent <= 0 {
		tiePercent = defaultBestAlgorithmTiePercent
	}

	best := results[0]
	for _, result := range results[1:] {
		if result.Duration < best.Duration {
			best = result
		}
	}
	if best.Algorithm == ChecksumAlgorithmCRC32 {
		for _, result := range results {
			if result.Algorithm == ChecksumAlgorithmCRC32C && withinPercent(result.Duration, best.Duration, tiePercent) {
				return ChecksumAlgorithmCRC32C, nil
			}
		}
	}
	return best.Algorithm, nil
}

func BenchmarkAlgorithms(options BestAlgorithmOptions) ([]AlgorithmBenchmark, error) {
	candidates := normalizeCandidates(options.Candidates)
	sampleSize := options.SampleSize
	if sampleSize <= 0 {
		sampleSize = 256 * 1024
	}
	iterations := options.Iterations
	if iterations <= 0 {
		iterations = 16
	}
	sample := checksumBenchmarkSample(sampleSize)
	results := make([]AlgorithmBenchmark, 0, len(candidates))
	for _, candidate := range candidates {
		duration, err := benchmarkAlgorithm(candidate, sample, iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, AlgorithmBenchmark{
			Algorithm:      candidate,
			BytesProcessed: int64(len(sample) * iterations),
			Duration:       duration,
		})
	}
	return results, nil
}

func SupportedS3Algorithms() []Algorithm {
	return []Algorithm{ChecksumAlgorithmCRC32C, ChecksumAlgorithmCRC32}
}

func normalizeCandidates(candidates []Algorithm) []Algorithm {
	if len(candidates) == 0 {
		return SupportedS3Algorithms()
	}
	out := make([]Algorithm, 0, len(candidates))
	seen := make(map[Algorithm]bool, len(candidates))
	for _, candidate := range candidates {
		algorithm := candidate.Normalize()
		if seen[algorithm] {
			continue
		}
		seen[algorithm] = true
		out = append(out, algorithm)
	}
	return out
}

func benchmarkAlgorithm(algorithm Algorithm, sample []byte, iterations int) (time.Duration, error) {
	state, err := algorithm.NewState()
	if err != nil {
		return 0, err
	}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = state.Write(sample)
	}
	_, err = state.Encoded()
	if err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func checksumBenchmarkSample(size int) []byte {
	sample := make([]byte, size)
	for i := range sample {
		sample[i] = byte((i*31 + 17) % 251)
	}
	return sample
}

func withinPercent(candidate time.Duration, best time.Duration, percent int) bool {
	if best <= 0 {
		return true
	}
	return candidate <= best+best*time.Duration(percent)/100
}
