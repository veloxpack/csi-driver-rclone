package lib

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ErrInvalidMeterParameter is the error thrown when a parameter is invalid.
var ErrInvalidMeterParameter = errors.New("meter invalid parameter")

// MinMeterResolution MinResolution is the minimum time resolution to measure bit rate.
const MinMeterResolution time.Duration = time.Millisecond * 100

// Meter measures the latest data transfer amount.
type Meter struct {
	resolution          time.Duration
	sample              time.Duration
	cur, last, first    *MeterItem
	started, closed     bool
	startedAt, closedAt time.Time
	totalBytes          uint64
	mu                  sync.RWMutex
}

// MeterItem is an element of linked list for the meter.
type MeterItem struct {
	vol        float64
	start      time.Duration // must be a multiple of resolution
	end        time.Time
	next, prev *MeterItem
}

// NewMeter creates a meter with specified resolution and sample duration.
// sample must be an integral multiple of Resolution
func NewMeter(resolution, sample time.Duration) (*Meter, error) {
	switch {
	case resolution <= 0:
		return nil, fmt.Errorf("%w: resolution %d <= 0", ErrInvalidMeterParameter, resolution)
	case resolution < MinMeterResolution:
		return nil, fmt.Errorf("%w: resolution %d < minMeterResolution", ErrInvalidMeterParameter, resolution)
	case sample <= 0:
		return nil, fmt.Errorf("%w: sample %d <= 0", ErrInvalidMeterParameter, sample)
	}
	n := int(sample / resolution)
	if n < 2 {
		return nil, fmt.Errorf("%w: too small sample duration %s (at least %s)", ErrInvalidMeterParameter, sample, resolution*2)
	}
	m := &Meter{
		resolution: resolution,
		sample:     sample,
		cur:        &MeterItem{},
	}

	tail := m.cur
	for i := 0; i < n; i++ {
		tail.next = &MeterItem{prev: tail}
		tail = tail.next
	}
	m.cur.prev = tail
	tail.next = m.cur

	return m, nil
}

// Start starts measuring the data transfer.
func (m *Meter) Start(tc time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return
	}
	m.started, m.startedAt = true, tc
	m.cur.start, m.cur.end = 0, m.startedAt.Add(m.resolution)
}

// Close stops measuring the data transfer.
func (m *Meter) Close(tc time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.closed, m.closedAt = true, tc
}

// Record records the data transfer into the meter.
func (m *Meter) Record(tc time.Time, b uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tc.Before(m.cur.end) {
		m.cur.vol += float64(b)
		m.totalBytes += b
		return
	}
	m.last, m.cur = m.cur, m.cur.next
	switch m.first {
	case nil:
		m.first = m.last
	case m.cur:
		m.first = m.first.next
	}
	m.cur.start = tc.Sub(m.startedAt) / m.resolution * m.resolution
	m.cur.end = m.startedAt.Add(m.cur.start + m.resolution)
	m.cur.vol = float64(b)
	m.totalBytes += b
}

// bpscoef is a coefficient used for bit rate calculation.
const bpscoef = 8 * float64(time.Second)

// BitRate returns the bit rate in the last sample period.
func (m *Meter) BitRate(tc time.Time) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.last == nil {
		return 0
	}

	newest := m.last
	sampleEnd := m.cur.start
	if !tc.Before(m.cur.end) { // tc is after the cur
		newest = m.cur
		sampleEnd = tc.Sub(m.startedAt) / m.resolution * m.resolution
	}
	sampleStart := sampleEnd - m.sample
	sampleWidth := m.sample
	if sampleStart < 0 {
		sampleStart = 0
		sampleWidth = sampleEnd
	}

	var sum float64
	for i := newest; sampleStart <= i.start; i = i.prev {
		sum += i.vol
		if i == m.first {
			break
		}
	}
	// fmt.Printf("DEBUG: %f * 8 / %s\n", sum, sampleWidth)
	return uint64(sum * bpscoef / float64(sampleWidth))
}

// Total returns the data transfer amount, elapsed time, and bit rate
// in the entire period from Start to Close.
func (m *Meter) Total(tc time.Time) (uint64, time.Duration, float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.last == nil {
		return 0, 0, 0
	}
	if m.closed {
		tc = m.closedAt
	}
	b, d := m.totalBytes, tc.Sub(m.startedAt)
	switch {
	case b == 0:
		return 0, d, 0
	case d == 0:
		return b, d, math.Inf(+1)
	}
	return b, d, float64(b) * bpscoef / float64(d)
}
