package lib

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// AdaptiveConcurrencyDefaultInitialTarget is the default starting target used by NewAdaptiveConcurrencyManager.
	AdaptiveConcurrencyDefaultInitialTarget = 8
	// AdaptiveConcurrencyDefaultMinTarget is the default minimum target for config-based managers.
	AdaptiveConcurrencyDefaultMinTarget = 2
	// AdaptiveConcurrencyDefaultGrowEvery is the default success count between growth steps.
	AdaptiveConcurrencyDefaultGrowEvery = 8
	// AdaptiveConcurrencyDefaultGrowStep is the default normal growth step.
	AdaptiveConcurrencyDefaultGrowStep = 1
	// AdaptiveConcurrencyDefaultFailureShrinkPercent is the default shrink percentage after failure.
	AdaptiveConcurrencyDefaultFailureShrinkPercent = 50
	// AdaptiveConcurrencyDefaultBackPressureShrinkPercent is the default shrink percentage after backpressure.
	AdaptiveConcurrencyDefaultBackPressureShrinkPercent = 50
	// AdaptiveConcurrencyDefaultBackPressurePause is the default pause after backpressure.
	AdaptiveConcurrencyDefaultBackPressurePause = time.Second
)

type AdaptiveConcurrencySample struct {
	Success      bool
	Duration     time.Duration
	Bytes        int64
	StatusCode   int
	BackPressure bool
	RetryAfter   time.Duration
}

type AdaptiveConcurrencyManager struct {
	// Coordination state.
	mu      sync.Mutex
	wg      sync.WaitGroup
	notify  chan struct{}
	waiters int
	done    chan struct{}
	running int

	// Target bounds and growth controls.
	target                      int
	max                         int
	min                         int
	growthCeiling               int
	growthCeilingProbeBytes     int64
	growthCeilingProbeSuccesses int
	growthCeilingProbeRate      float64
	peakTarget                  int
	peakRunning                 int
	throughputFloor             int
	successes                   int
	growEvery                   int
	growStep                    int
	sqrtGrowth                  bool

	// Failure, back-pressure, and retry controls.
	failureShrink      int
	backPressureShrink int
	backPressurePause  time.Duration
	pauseUntil         time.Time

	// Throughput window and probe controls.
	throughputWindow                int
	throughputMinGain               int
	throughputShrink                int
	throughputHold                  int
	throughputProbeMin              int
	throughputProbeFloor            int
	throughputProbeFloorRate        float64
	throughputProbePlateau          int
	throughputProbeMinGainPerTarget float64
	throughputProbeLossTolerance    int
	throughputProbeStartTarget      int
	throughputProbeMisses           int
	probePending                    bool
	throughputHoldRemaining         int

	// Latency controls.
	latencyFloor           int
	latencyShrink          int
	latencyQueueHigh       float64
	latencyGrowthQueueHigh float64

	// Aggregate counters exposed through snapshots and upload telemetry.
	runningWorkers                     int32
	successTotal                       int
	failureTotal                       int
	growTotal                          int
	shrinkTotal                        int
	backPressureTotal                  int
	retryAfterTotal                    int
	throughputBackoffTotal             int
	throughputProbeMissTotal           int
	throughputProbeEfficiencyMissTotal int
	latencyBackoffTotal                int
	latencyGrowthSuppressionTotal      int
	durationTotal                      time.Duration
	bytesTotal                         int64

	// Current throughput and latency sample state.
	windowSamples                    int
	windowBytes                      int64
	windowDuration                   time.Duration
	windowStartedAt                  time.Time
	lastThroughput                   float64
	bestThroughput                   float64
	lastThroughputProbeGain          float64
	lastThroughputProbeTargetDelta   int
	lastThroughputProbeGainPerTarget float64
	minDurationPerByte               float64
	lastDurationPerByte              float64
	lastQueueEstimate                float64
}

type AdaptiveConcurrencyManagerWithSample interface {
	ConcurrencyManager
	DoneWithSample(AdaptiveConcurrencySample)
}

type AdaptiveConcurrencySnapshot struct {
	Target                                  int
	Max                                     int
	GrowthCeiling                           int
	GrowthCeilingUnlocked                   bool
	GrowthCeilingProbeSuccessThreshold      int
	Running                                 int
	PeakTarget                              int
	PeakRunning                             int
	SuccessTotal                            int
	FailureTotal                            int
	GrowTotal                               int
	ShrinkTotal                             int
	BackPressureTotal                       int
	RetryAfterTotal                         int
	ThroughputBackoffTotal                  int
	ThroughputProbeMissTotal                int
	ThroughputProbeEfficiencyMissTotal      int
	LatencyBackoffTotal                     int
	LatencyGrowthSuppressionTotal           int
	BytesTotal                              int64
	AverageDuration                         time.Duration
	LastThroughputBytesPerSecond            float64
	BestThroughputBytesPerSecond            float64
	LastThroughputProbeGainPercent          float64
	LastThroughputProbeTargetDelta          int
	LastThroughputProbeGainPerTargetPercent float64
	LastQueueEstimate                       float64
	MinDurationPerByte                      float64
	LastDurationPerByte                     float64
}

type AdaptiveConcurrencyConfig struct {
	MaxConcurrency                         int
	InitialTarget                          int
	MinTarget                              int
	GrowthCeiling                          int
	GrowthCeilingProbeBytes                int64
	GrowthCeilingProbeSuccesses            int
	GrowthCeilingProbeRate                 float64
	ThroughputFloor                        int
	GrowEvery                              int
	GrowStep                               int
	SqrtGrowth                             bool
	FailureShrinkPercent                   int
	BackPressureShrinkPercent              int
	BackPressurePause                      time.Duration
	ThroughputWindow                       int
	ThroughputMinGainPercent               int
	ThroughputShrinkPercent                int
	ThroughputHoldWindows                  int
	ThroughputProbeMinWindows              int
	ThroughputProbeFloor                   int
	ThroughputProbeFloorRate               float64
	ThroughputProbePlateauTarget           int
	ThroughputProbeMinGainPerTargetPercent float64
	ThroughputProbeLossTolerancePercent    int
	LatencyFloor                           int
	LatencyShrinkPercent                   int
	LatencyQueueHigh                       float64
	LatencyGrowthQueueHigh                 float64
}

// AdaptiveConcurrencyDefaults contains the primary concurrency settings callers
// usually need when initializing an adaptive manager.
type AdaptiveConcurrencyDefaults struct {
	MaxConcurrency int
	InitialTarget  int
}

// DefaultAdaptiveConcurrencyDefaults returns the SDK's default adaptive
// concurrency settings with the supplied maximum concurrency.
func DefaultAdaptiveConcurrencyDefaults(maxConcurrency int) AdaptiveConcurrencyDefaults {
	return AdaptiveConcurrencyDefaults{
		MaxConcurrency: maxConcurrency,
		InitialTarget:  AdaptiveConcurrencyDefaultInitialTarget,
	}
}

// Config returns a full manager config for these defaults.
func (d AdaptiveConcurrencyDefaults) Config() AdaptiveConcurrencyConfig {
	initialTarget := d.InitialTarget
	if initialTarget <= 0 {
		initialTarget = AdaptiveConcurrencyDefaultInitialTarget
	}
	return AdaptiveConcurrencyConfig{
		MaxConcurrency:            d.MaxConcurrency,
		InitialTarget:             initialTarget,
		GrowEvery:                 AdaptiveConcurrencyDefaultGrowEvery,
		GrowStep:                  AdaptiveConcurrencyDefaultGrowStep,
		FailureShrinkPercent:      AdaptiveConcurrencyDefaultFailureShrinkPercent,
		BackPressureShrinkPercent: AdaptiveConcurrencyDefaultBackPressureShrinkPercent,
		BackPressurePause:         AdaptiveConcurrencyDefaultBackPressurePause,
	}
}

func NewAdaptiveConcurrencyManager(maxConcurrency int) *AdaptiveConcurrencyManager {
	return NewAdaptiveConcurrencyManagerWithInitial(maxConcurrency, AdaptiveConcurrencyDefaultInitialTarget)
}

// DefaultAdaptiveConcurrencyConfig returns the default manager config with the
// supplied maximum concurrency. Callers can override individual fields before
// passing it to NewAdaptiveConcurrencyManagerWithConfig.
func DefaultAdaptiveConcurrencyConfig(maxConcurrency int) AdaptiveConcurrencyConfig {
	return DefaultAdaptiveConcurrencyDefaults(maxConcurrency).Config()
}

func NewAdaptiveConcurrencyManagerWithDefaults(defaults AdaptiveConcurrencyDefaults) *AdaptiveConcurrencyManager {
	return NewAdaptiveConcurrencyManagerWithConfig(defaults.Config())
}

func NewAdaptiveConcurrencyManagerWithInitial(maxConcurrency int, initialTarget int) *AdaptiveConcurrencyManager {
	config := DefaultAdaptiveConcurrencyConfig(maxConcurrency)
	config.InitialTarget = initialTarget
	return NewAdaptiveConcurrencyManagerWithConfig(config)
}

func NewAdaptiveConcurrencyManagerWithConfig(config AdaptiveConcurrencyConfig) *AdaptiveConcurrencyManager {
	maxConcurrency := config.MaxConcurrency
	maxConcurrency = max(maxConcurrency, 1)
	initialTarget := min(max(config.InitialTarget, 1), maxConcurrency)
	minTarget := config.MinTarget
	if minTarget <= 0 {
		minTarget = min(AdaptiveConcurrencyDefaultMinTarget, maxConcurrency)
	}
	minTarget = min(max(minTarget, 1), maxConcurrency)
	growthCeiling := config.GrowthCeiling
	if growthCeiling < 0 {
		growthCeiling = 0
	}
	if growthCeiling > 0 {
		growthCeiling = min(max(growthCeiling, minTarget), maxConcurrency)
		if growthCeiling >= maxConcurrency {
			growthCeiling = 0
		}
	}
	growthCeilingProbeBytes := config.GrowthCeilingProbeBytes
	if growthCeilingProbeBytes < 0 {
		growthCeilingProbeBytes = 0
	}
	growthCeilingProbeSuccesses := config.GrowthCeilingProbeSuccesses
	if growthCeilingProbeSuccesses < 0 {
		growthCeilingProbeSuccesses = 0
	}
	growthCeilingProbeRate := config.GrowthCeilingProbeRate
	if growthCeilingProbeRate < 0 {
		growthCeilingProbeRate = 0
	}
	throughputFloor := config.ThroughputFloor
	if throughputFloor <= 0 {
		throughputFloor = minTarget
	}
	throughputFloor = min(max(throughputFloor, minTarget), maxConcurrency)
	growEvery := max(config.GrowEvery, 1)
	growStep := max(config.GrowStep, 1)
	failureShrink := config.FailureShrinkPercent
	if failureShrink <= 0 {
		failureShrink = AdaptiveConcurrencyDefaultFailureShrinkPercent
	}
	backPressureShrink := config.BackPressureShrinkPercent
	if backPressureShrink < 0 {
		backPressureShrink = 0
	}
	throughputWindow := config.ThroughputWindow
	if throughputWindow < 0 {
		throughputWindow = 0
	}
	throughputMinGain := config.ThroughputMinGainPercent
	if throughputMinGain < 0 {
		throughputMinGain = 0
	}
	throughputShrink := config.ThroughputShrinkPercent
	if throughputShrink < 0 {
		throughputShrink = 0
	}
	throughputHold := config.ThroughputHoldWindows
	if throughputHold < 0 {
		throughputHold = 0
	}
	throughputProbeMin := config.ThroughputProbeMinWindows
	if throughputProbeMin <= 0 {
		throughputProbeMin = 1
	}
	throughputProbeFloor := config.ThroughputProbeFloor
	if throughputProbeFloor < 0 {
		throughputProbeFloor = 0
	}
	throughputProbeFloor = min(throughputProbeFloor, maxConcurrency)
	throughputProbeFloorRate := config.ThroughputProbeFloorRate
	if throughputProbeFloorRate < 0 {
		throughputProbeFloorRate = 0
	}
	throughputProbePlateau := config.ThroughputProbePlateauTarget
	if throughputProbePlateau < 0 {
		throughputProbePlateau = 0
	}
	throughputProbePlateau = min(throughputProbePlateau, maxConcurrency)
	throughputProbeMinGainPerTarget := config.ThroughputProbeMinGainPerTargetPercent
	if throughputProbeMinGainPerTarget < 0 {
		throughputProbeMinGainPerTarget = 0
	}
	throughputProbeLossTolerance := config.ThroughputProbeLossTolerancePercent
	if throughputProbeLossTolerance < 0 {
		throughputProbeLossTolerance = 0
	}
	latencyShrink := config.LatencyShrinkPercent
	if latencyShrink < 0 {
		latencyShrink = 0
	}
	latencyFloor := config.LatencyFloor
	if latencyFloor <= 0 {
		latencyFloor = minTarget
	}
	latencyFloor = min(max(latencyFloor, minTarget), maxConcurrency)
	latencyQueueHigh := config.LatencyQueueHigh
	if latencyQueueHigh < 0 {
		latencyQueueHigh = 0
	}
	latencyGrowthQueueHigh := config.LatencyGrowthQueueHigh
	if latencyGrowthQueueHigh < 0 {
		latencyGrowthQueueHigh = 0
	}
	return &AdaptiveConcurrencyManager{
		target:                          initialTarget,
		max:                             maxConcurrency,
		min:                             minTarget,
		growthCeiling:                   growthCeiling,
		growthCeilingProbeBytes:         growthCeilingProbeBytes,
		growthCeilingProbeSuccesses:     growthCeilingProbeSuccesses,
		growthCeilingProbeRate:          growthCeilingProbeRate,
		peakTarget:                      initialTarget,
		throughputFloor:                 throughputFloor,
		growEvery:                       growEvery,
		growStep:                        growStep,
		sqrtGrowth:                      config.SqrtGrowth,
		failureShrink:                   min(failureShrink, 100),
		backPressureShrink:              min(backPressureShrink, 100),
		backPressurePause:               config.BackPressurePause,
		throughputWindow:                throughputWindow,
		throughputMinGain:               throughputMinGain,
		throughputShrink:                min(throughputShrink, 100),
		throughputHold:                  throughputHold,
		throughputProbeMin:              throughputProbeMin,
		throughputProbeFloor:            throughputProbeFloor,
		throughputProbeFloorRate:        throughputProbeFloorRate,
		throughputProbePlateau:          throughputProbePlateau,
		throughputProbeMinGainPerTarget: throughputProbeMinGainPerTarget,
		throughputProbeLossTolerance:    throughputProbeLossTolerance,
		latencyFloor:                    latencyFloor,
		latencyShrink:                   min(latencyShrink, 100),
		latencyQueueHigh:                latencyQueueHigh,
		latencyGrowthQueueHigh:          latencyGrowthQueueHigh,
		notify:                          make(chan struct{}),
		done:                            make(chan struct{}),
	}
}

func (a *AdaptiveConcurrencyManager) Wait() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.waitForCapacityLocked(context.Background()) {
		a.acquireLocked()
	}
}

func (a *AdaptiveConcurrencyManager) WaitWithContext(ctx context.Context) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.waitForCapacityLocked(ctx) {
		return false
	}
	a.acquireLocked()
	return true
}

func (a *AdaptiveConcurrencyManager) Done() {
	a.DoneWithSample(AdaptiveConcurrencySample{Success: true})
}

func (a *AdaptiveConcurrencyManager) DoneNeutral() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.releaseLocked()
}

func (a *AdaptiveConcurrencyManager) DoneWithSample(sample AdaptiveConcurrencySample) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if sample.Duration > 0 {
		a.durationTotal += sample.Duration
	}
	if sample.Bytes > 0 {
		a.bytesTotal += sample.Bytes
	}
	backPressure := sample.BackPressure || sample.StatusCode == 429 || sample.StatusCode == 503 || sample.StatusCode == 504
	if backPressure {
		a.backPressureTotal++
		if sample.RetryAfter > 0 {
			a.retryAfterTotal++
			a.pauseUntil = time.Now().Add(sample.RetryAfter)
		} else if a.backPressurePause > 0 {
			pauseUntil := time.Now().Add(a.backPressurePause)
			if a.pauseUntil.Before(pauseUntil) {
				a.pauseUntil = pauseUntil
			}
		}
	}
	if sample.Success {
		a.successTotal++
		if backPressure {
			a.successes = 0
			a.probePending = false
			a.throughputProbeStartTarget = 0
			a.throughputProbeMisses = 0
			if a.shrinkLocked(a.backPressureShrink) {
				a.shrinkTotal++
			}
		} else {
			a.recordThroughputLocked(sample)
			a.successes++
			growthCeiling := a.growthCeilingLocked()
			if a.successes >= a.growEvery && a.target < growthCeiling && !a.probePending && a.throughputHoldRemaining == 0 {
				if a.shouldSuppressGrowthForLatencyLocked() {
					a.latencyGrowthSuppressionTotal++
					a.successes = 0
				} else {
					previousTarget := a.target
					a.target = min(growthCeiling, a.target+a.growStepLocked())
					if a.target > a.peakTarget {
						a.peakTarget = a.target
					}
					a.growTotal++
					a.successes = 0
					if a.shouldProbeGrowthLocked() {
						a.probePending = true
						a.throughputProbeStartTarget = previousTarget
						a.throughputProbeMisses = 0
					}
				}
			}
		}
	} else {
		a.failureTotal++
		a.successes = 0
		a.probePending = false
		a.throughputProbeStartTarget = 0
		a.throughputProbeMisses = 0
		if a.shrinkLocked(a.failureShrink) {
			a.shrinkTotal++
		}
	}
	a.releaseLocked()
}

func (a *AdaptiveConcurrencyManager) releaseLocked() {
	a.running--
	atomic.AddInt32(&a.runningWorkers, -1)
	a.wg.Done()
	a.signalDoneLocked()
	a.signalLocked()
}

func (a *AdaptiveConcurrencyManager) shrinkLocked(percent int) bool {
	if percent <= 0 || a.target <= a.min {
		return false
	}
	next := a.target * (100 - percent) / 100
	if next >= a.target {
		next = a.target - 1
	}
	a.target = max(a.min, next)
	return true
}

func (a *AdaptiveConcurrencyManager) growStepLocked() int {
	if !a.sqrtGrowth {
		return a.growStep
	}
	return max(a.growStep, int(math.Sqrt(float64(max(a.target, 1)))))
}

func (a *AdaptiveConcurrencyManager) growthCeilingLocked() int {
	if a.growthCeilingUnlockedLocked() {
		return a.max
	}
	return a.growthCeiling
}

func (a *AdaptiveConcurrencyManager) growthCeilingUnlockedLocked() bool {
	if a.growthCeiling <= 0 {
		return true
	}
	if !a.growthCeilingProbeWorkReadyLocked() {
		return false
	}
	if a.growthCeilingProbeRate > 0 && a.bestThroughput < a.growthCeilingProbeRate {
		return false
	}
	return true
}

func (a *AdaptiveConcurrencyManager) growthCeilingProbeWorkReadyLocked() bool {
	bytesGateEnabled := a.growthCeilingProbeBytes > 0
	successGateEnabled := a.growthCeilingProbeSuccesses > 0
	if !bytesGateEnabled && !successGateEnabled {
		return true
	}
	if bytesGateEnabled && a.bytesTotal >= a.growthCeilingProbeBytes {
		return true
	}
	return successGateEnabled && a.successTotal >= a.growthCeilingProbeSuccesses
}

func (a *AdaptiveConcurrencyManager) shouldProbeGrowthLocked() bool {
	if a.throughputWindow <= 0 || a.throughputMinGain <= 0 || a.throughputShrink <= 0 {
		return false
	}
	if a.throughputProbeFloor <= 0 || a.target >= a.throughputProbeFloor {
		return true
	}
	if a.bestThroughput <= 0 {
		return false
	}
	if a.throughputProbeFloorRate <= 0 {
		return false
	}
	return a.bestThroughput < a.throughputProbeFloorRate
}

func (a *AdaptiveConcurrencyManager) shrinkThroughputLocked(percent int) bool {
	floor := a.throughputBackoffFloorLocked()
	if percent <= 0 || a.target <= floor {
		return false
	}
	next := a.target * (100 - percent) / 100
	if next >= a.target {
		next = a.target - 1
	}
	a.target = max(floor, next)
	return true
}

func (a *AdaptiveConcurrencyManager) throughputBackoffFloorLocked() int {
	floor := a.throughputFloor
	if a.throughputProbeFloor > floor && a.throughputProbeFloorRate > 0 && a.bestThroughput >= a.throughputProbeFloorRate {
		floor = a.throughputProbeFloor
	}
	return min(max(floor, a.throughputFloor), a.max)
}

func (a *AdaptiveConcurrencyManager) shrinkLatencyLocked(percent int) bool {
	if percent <= 0 || a.target <= a.latencyFloor {
		return false
	}
	next := a.target * (100 - percent) / 100
	if next >= a.target {
		next = a.target - 1
	}
	a.target = max(a.latencyFloor, next)
	return true
}

func (a *AdaptiveConcurrencyManager) recordThroughputLocked(sample AdaptiveConcurrencySample) {
	if a.throughputWindow <= 0 || sample.Bytes <= 0 {
		return
	}
	now := time.Now()
	if sample.Duration > 0 {
		startedAt := now.Add(-sample.Duration)
		if a.windowStartedAt.IsZero() || startedAt.Before(a.windowStartedAt) {
			a.windowStartedAt = startedAt
		}
	} else if a.windowStartedAt.IsZero() {
		a.windowStartedAt = now
	}
	a.windowSamples++
	a.windowBytes += sample.Bytes
	a.windowDuration += sample.Duration
	if a.windowSamples < a.throughputWindow {
		return
	}

	rate := a.windowThroughputLocked(now)
	durationPerByte := a.windowDurationPerByteLocked()
	a.recordLatencyLocked(durationPerByte)
	a.windowSamples = 0
	a.windowBytes = 0
	a.windowDuration = 0
	a.windowStartedAt = time.Time{}
	if rate <= 0 {
		return
	}
	a.lastThroughput = rate

	if a.throughputHoldRemaining > 0 {
		a.throughputHoldRemaining--
	}
	if a.bestThroughput <= 0 {
		a.bestThroughput = rate
		a.throughputProbeMisses = 0
		return
	}
	if !a.probePending {
		if a.shouldBackOffForLatencyLocked(rate) {
			if a.shrinkLatencyLocked(a.latencyShrink) {
				a.shrinkTotal++
				a.latencyBackoffTotal++
				a.successes = 0
				a.throughputProbeMisses = 0
				a.throughputHoldRemaining = a.throughputHold
			}
			return
		}
		if rate > a.bestThroughput {
			a.bestThroughput = rate
		}
		return
	}

	previousBest := a.bestThroughput
	requiredGain := a.requiredThroughputProbeGainPercentLocked()
	a.recordThroughputProbeGainLocked(rate, previousBest)
	required := previousBest * (1 + requiredGain/100)
	a.probePending = false
	if rate >= required {
		a.bestThroughput = rate
		a.throughputProbeMisses = 0
		a.throughputProbeStartTarget = 0
		return
	}
	if a.shouldAcceptNeutralThroughputProbeLocked(rate, previousBest) {
		a.throughputProbeMisses = 0
		a.throughputProbeStartTarget = 0
		return
	}
	if requiredGain > float64(a.throughputMinGain) && a.lastThroughputProbeGain >= float64(a.throughputMinGain) {
		a.throughputProbeEfficiencyMissTotal++
	}
	a.throughputProbeMisses++
	a.throughputProbeMissTotal++
	if a.throughputProbeMisses < a.throughputProbeMin {
		a.probePending = true
		return
	}
	a.throughputProbeStartTarget = 0
	if a.shrinkThroughputLocked(a.throughputShrink) {
		a.shrinkTotal++
		a.throughputBackoffTotal++
		a.successes = 0
		a.throughputHoldRemaining = a.throughputHold
	}
	a.throughputProbeMisses = 0
}

func (a *AdaptiveConcurrencyManager) requiredThroughputProbeGainPercentLocked() float64 {
	required := float64(a.throughputMinGain)
	if a.throughputProbePlateau <= 0 || a.target <= a.throughputProbePlateau || a.throughputProbeMinGainPerTarget <= 0 {
		return required
	}
	plateauRequired := float64(a.throughputProbeTargetDeltaLocked()) * a.throughputProbeMinGainPerTarget
	if plateauRequired > required {
		return plateauRequired
	}
	return required
}

func (a *AdaptiveConcurrencyManager) shouldAcceptNeutralThroughputProbeLocked(rate float64, previousBest float64) bool {
	if a.throughputProbeLossTolerance <= 0 || previousBest <= 0 {
		return false
	}
	if a.growthCeiling <= 0 || a.target <= a.growthCeiling || !a.growthCeilingUnlockedLocked() {
		return false
	}
	if a.throughputProbePlateau > 0 && a.target > a.throughputProbePlateau {
		return false
	}
	return rate >= previousBest*(1-float64(a.throughputProbeLossTolerance)/100)
}

func (a *AdaptiveConcurrencyManager) throughputProbeTargetDeltaLocked() int {
	if a.throughputProbePlateau > 0 && a.target > a.throughputProbePlateau {
		startTarget := a.throughputProbeStartTarget
		if startTarget <= 0 || startTarget < a.throughputProbePlateau {
			startTarget = a.throughputProbePlateau
		}
		if startTarget < a.target {
			return a.target - startTarget
		}
		return 1
	}
	if a.throughputProbeStartTarget > 0 && a.target > a.throughputProbeStartTarget {
		return a.target - a.throughputProbeStartTarget
	}
	return 1
}

func (a *AdaptiveConcurrencyManager) recordThroughputProbeGainLocked(rate float64, previousBest float64) {
	if previousBest <= 0 {
		a.lastThroughputProbeGain = 0
		a.lastThroughputProbeTargetDelta = 0
		a.lastThroughputProbeGainPerTarget = 0
		return
	}
	gain := (rate/previousBest - 1) * 100
	targetDelta := a.throughputProbeTargetDeltaLocked()
	a.lastThroughputProbeGain = gain
	a.lastThroughputProbeTargetDelta = targetDelta
	a.lastThroughputProbeGainPerTarget = gain / float64(targetDelta)
}

func (a *AdaptiveConcurrencyManager) windowThroughputLocked(now time.Time) float64 {
	if a.windowBytes <= 0 {
		return 0
	}
	wallDuration := now.Sub(a.windowStartedAt)
	if wallDuration > 0 {
		return float64(a.windowBytes) / wallDuration.Seconds()
	}
	if a.windowDuration > 0 {
		return float64(a.windowBytes) / a.windowDuration.Seconds()
	}
	return 0
}

func (a *AdaptiveConcurrencyManager) windowDurationPerByteLocked() float64 {
	if a.windowBytes <= 0 || a.windowDuration <= 0 {
		return 0
	}
	return a.windowDuration.Seconds() / float64(a.windowBytes)
}

func (a *AdaptiveConcurrencyManager) recordLatencyLocked(durationPerByte float64) {
	if durationPerByte <= 0 {
		return
	}
	a.lastDurationPerByte = durationPerByte
	if a.minDurationPerByte <= 0 || durationPerByte < a.minDurationPerByte {
		a.minDurationPerByte = durationPerByte
	}
	if a.minDurationPerByte <= 0 || durationPerByte <= a.minDurationPerByte {
		a.lastQueueEstimate = 0
		return
	}
	a.lastQueueEstimate = float64(a.target) * (1 - a.minDurationPerByte/durationPerByte)
}

func (a *AdaptiveConcurrencyManager) shouldBackOffForLatencyLocked(rate float64) bool {
	if a.latencyShrink <= 0 || a.latencyQueueHigh <= 0 || a.lastQueueEstimate <= a.latencyQueueHigh || a.bestThroughput <= 0 {
		return false
	}
	if a.throughputProbeFloor > 0 && a.target <= a.throughputProbeFloor && a.throughputProbeFloorRate > 0 && a.bestThroughput >= a.throughputProbeFloorRate {
		return false
	}
	allowedDrop := 1 - float64(max(a.throughputMinGain, 1))/100
	return rate < a.bestThroughput*allowedDrop
}

func (a *AdaptiveConcurrencyManager) shouldSuppressGrowthForLatencyLocked() bool {
	return a.latencyGrowthQueueHigh > 0 && a.lastQueueEstimate > a.latencyGrowthQueueHigh && a.bestThroughput > 0
}

func (a *AdaptiveConcurrencyManager) canAcquireLocked() bool {
	if a.running >= a.target {
		return false
	}
	if a.pauseUntil.IsZero() {
		return true
	}
	if time.Now().Before(a.pauseUntil) {
		return false
	}
	a.pauseUntil = time.Time{}
	return true
}

func (a *AdaptiveConcurrencyManager) waitForCapacityLocked(ctx context.Context) bool {
	for !a.canAcquireLocked() {
		notify := a.notifyLocked()
		wait := time.Duration(0)
		if !a.pauseUntil.IsZero() && a.running < a.target {
			if pauseWait := time.Until(a.pauseUntil); pauseWait > 0 {
				wait = pauseWait
			}
		}
		a.waiters++
		a.mu.Unlock()
		canceled := false
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				canceled = true
			case <-notify:
			case <-timer.C:
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		} else {
			select {
			case <-ctx.Done():
				canceled = true
			case <-notify:
			}
		}
		a.mu.Lock()
		a.waiters--
		if canceled {
			return false
		}
	}
	return ctx.Err() == nil
}

func (a *AdaptiveConcurrencyManager) acquireLocked() {
	a.running++
	if a.running > a.peakRunning {
		a.peakRunning = a.running
	}
	atomic.AddInt32(&a.runningWorkers, 1)
	a.wg.Add(1)
}

func (a *AdaptiveConcurrencyManager) notifyLocked() chan struct{} {
	if a.notify == nil {
		a.notify = make(chan struct{})
	}
	return a.notify
}

func (a *AdaptiveConcurrencyManager) doneLocked() chan struct{} {
	if a.done == nil {
		a.done = make(chan struct{})
	}
	return a.done
}

func (a *AdaptiveConcurrencyManager) signalLocked() {
	if a.waiters == 0 {
		return
	}
	close(a.notifyLocked())
	a.notify = make(chan struct{})
}

func (a *AdaptiveConcurrencyManager) signalDoneLocked() {
	close(a.doneLocked())
	a.done = make(chan struct{})
}

func (a *AdaptiveConcurrencyManager) WaitAllDone() {
	a.wg.Wait()
}

func (a *AdaptiveConcurrencyManager) RunningCount() int {
	return int(atomic.LoadInt32(&a.runningWorkers))
}

func (a *AdaptiveConcurrencyManager) WaitForADone() bool {
	return a.WaitForADoneWithContext(context.Background())
}

func (a *AdaptiveConcurrencyManager) WaitForADoneWithContext(ctx context.Context) bool {
	a.mu.Lock()
	if a.running == 0 {
		a.mu.Unlock()
		return false
	}
	done := a.doneLocked()
	a.mu.Unlock()

	select {
	case <-ctx.Done():
		return false
	case <-done:
		return true
	}
}

func (a *AdaptiveConcurrencyManager) NewSubWorker() ConcurrencyManager {
	return &AdaptiveConcurrencySubWorker{parent: a}
}

func (a *AdaptiveConcurrencyManager) Target() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.target
}

func (a *AdaptiveConcurrencyManager) Max() int {
	return a.max
}

func (a *AdaptiveConcurrencyManager) Snapshot() AdaptiveConcurrencySnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	var averageDuration time.Duration
	completed := a.successTotal + a.failureTotal
	if completed > 0 {
		averageDuration = a.durationTotal / time.Duration(completed)
	}
	return AdaptiveConcurrencySnapshot{
		Target:                                  a.target,
		Max:                                     a.max,
		GrowthCeiling:                           a.growthCeiling,
		GrowthCeilingUnlocked:                   a.growthCeilingUnlockedLocked(),
		GrowthCeilingProbeSuccessThreshold:      a.growthCeilingProbeSuccesses,
		Running:                                 a.running,
		PeakTarget:                              a.peakTarget,
		PeakRunning:                             a.peakRunning,
		SuccessTotal:                            a.successTotal,
		FailureTotal:                            a.failureTotal,
		GrowTotal:                               a.growTotal,
		ShrinkTotal:                             a.shrinkTotal,
		BackPressureTotal:                       a.backPressureTotal,
		RetryAfterTotal:                         a.retryAfterTotal,
		ThroughputBackoffTotal:                  a.throughputBackoffTotal,
		ThroughputProbeMissTotal:                a.throughputProbeMissTotal,
		ThroughputProbeEfficiencyMissTotal:      a.throughputProbeEfficiencyMissTotal,
		LatencyBackoffTotal:                     a.latencyBackoffTotal,
		LatencyGrowthSuppressionTotal:           a.latencyGrowthSuppressionTotal,
		BytesTotal:                              a.bytesTotal,
		AverageDuration:                         averageDuration,
		LastThroughputBytesPerSecond:            a.lastThroughput,
		BestThroughputBytesPerSecond:            a.bestThroughput,
		LastThroughputProbeGainPercent:          a.lastThroughputProbeGain,
		LastThroughputProbeTargetDelta:          a.lastThroughputProbeTargetDelta,
		LastThroughputProbeGainPerTargetPercent: a.lastThroughputProbeGainPerTarget,
		LastQueueEstimate:                       a.lastQueueEstimate,
		MinDurationPerByte:                      a.minDurationPerByte,
		LastDurationPerByte:                     a.lastDurationPerByte,
	}
}

type AdaptiveConcurrencySubWorker struct {
	parent       *AdaptiveConcurrencyManager
	wg           sync.WaitGroup
	runningCount int32
	mu           sync.Mutex
	notify       chan struct{}
}

func (s *AdaptiveConcurrencySubWorker) Wait() {
	s.parent.Wait()
	s.mu.Lock()
	s.wg.Add(1)
	atomic.AddInt32(&s.runningCount, 1)
	s.mu.Unlock()
}

func (s *AdaptiveConcurrencySubWorker) WaitWithContext(ctx context.Context) bool {
	if s.parent.WaitWithContext(ctx) {
		s.mu.Lock()
		s.wg.Add(1)
		atomic.AddInt32(&s.runningCount, 1)
		s.mu.Unlock()
		return true
	}
	return false
}

func (s *AdaptiveConcurrencySubWorker) Done() {
	s.DoneWithSample(AdaptiveConcurrencySample{Success: true})
}

func (s *AdaptiveConcurrencySubWorker) DoneNeutral() {
	s.done(s.parent.DoneNeutral)
}

func (s *AdaptiveConcurrencySubWorker) DoneWithSample(sample AdaptiveConcurrencySample) {
	s.done(func() {
		s.parent.DoneWithSample(sample)
	})
}

func (s *AdaptiveConcurrencySubWorker) done(parentDone func()) {
	s.mu.Lock()
	atomic.AddInt32(&s.runningCount, -1)
	s.wg.Done()
	s.signalDoneLocked()
	s.mu.Unlock()
	parentDone()
}

func (s *AdaptiveConcurrencySubWorker) WaitAllDone() {
	s.wg.Wait()
}

func (s *AdaptiveConcurrencySubWorker) RunningCount() int {
	return int(atomic.LoadInt32(&s.runningCount))
}

func (s *AdaptiveConcurrencySubWorker) WaitForADone() bool {
	return s.WaitForADoneWithContext(context.Background())
}

func (s *AdaptiveConcurrencySubWorker) WaitForADoneWithContext(ctx context.Context) bool {
	s.mu.Lock()
	if s.RunningCount() == 0 {
		s.mu.Unlock()
		return false
	}
	notify := s.notifyLocked()
	s.mu.Unlock()

	select {
	case <-ctx.Done():
		return false
	case <-notify:
		return true
	}
}

func (s *AdaptiveConcurrencySubWorker) notifyLocked() chan struct{} {
	if s.notify == nil {
		s.notify = make(chan struct{})
	}
	return s.notify
}

func (s *AdaptiveConcurrencySubWorker) signalDoneLocked() {
	close(s.notifyLocked())
	s.notify = make(chan struct{})
}
