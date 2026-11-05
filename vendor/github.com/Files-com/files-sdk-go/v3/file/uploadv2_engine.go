package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/hashicorp/go-retryablehttp"
)

type uploadV2Engine struct {
	u                         *uploadIO
	plan                      uploadV2PartPlan
	manager                   *lib.AdaptiveConcurrencyManager
	globalManager             lib.ConcurrencyManager
	etags                     map[int]files_sdk.EtagsParam
	stats                     uploadV2SchedulerStats
	readDuration              time.Duration
	usePartOffsets            bool
	httpClientLimits          uploadV2HTTPClientLimits
	checksumTrailerEnabled    bool
	checksumTrailerSkipReason string
}

type uploadV2HTTPClientLimits struct {
	adjusted            bool
	available           bool
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
}

type uploadV2PartDescriptor struct {
	number int
	offset OffSet
	final  bool
	upload files_sdk.FileUploadPart
	legacy *Part
}

type uploadV2Part struct {
	uploadV2PartDescriptor
	reader   ProxyReader
	progress *uploadV2ProgressBatcher
}

type uploadV2PreparedPart struct {
	part        *uploadV2Part
	memoryBytes int64
}

const uploadV2ProgressBatchSize = int64(1024 * 1024)
const uploadV2PreallocateConcurrencyMultiplier = 4
const uploadV2DefaultSeekableS3ReadyRunwayMinParts = 32
const uploadV2DefaultSeekableS3ReadyRunwayMaxParts = 128
const uploadV2DefaultSeekableS3ReadyRunwayTargetDivisor = 2
const uploadV2DefaultHTTPIdleConnectionCap = 128
const uploadV2SmallS3KnownSizeConcurrencyCutoff = uploadV2GiB

type uploadV2ProgressBatcher struct {
	progress func(int64)
	pending  int64
}

func newUploadV2ProgressBatcher(progress func(int64)) *uploadV2ProgressBatcher {
	if progress == nil {
		progress = func(int64) {}
	}
	return &uploadV2ProgressBatcher{progress: progress}
}

func (b *uploadV2ProgressBatcher) Add(delta int64) {
	if b == nil || delta == 0 {
		return
	}
	b.pending += delta
	if b.pending >= uploadV2ProgressBatchSize || b.pending <= -uploadV2ProgressBatchSize || delta < 0 {
		b.Flush()
	}
}

func (b *uploadV2ProgressBatcher) Flush() {
	if b == nil || b.pending == 0 {
		return
	}
	b.progress(b.pending)
	b.pending = 0
}

func (p *uploadV2Part) flushProgress() {
	if p != nil && p.progress != nil {
		p.progress.Flush()
	}
}

func (p *uploadV2Part) closeReader() {
	if p != nil && p.reader != nil {
		p.reader.Close()
	}
}

type uploadV2PartResult struct {
	part         *uploadV2Part
	etag         files_sdk.EtagsParam
	bytes        int64
	duration     time.Duration
	readDuration time.Duration
	statusCode   int
	backPressure bool
	retryAfter   time.Duration
	err          error
}

type uploadV2PartConcurrencyGate struct {
	parent lib.ConcurrencyManager
	local  *lib.ConstrainedWorkGroup
}

func newUploadV2PartConcurrencyGate(parent lib.ConcurrencyManager, limit int) *uploadV2PartConcurrencyGate {
	return &uploadV2PartConcurrencyGate{
		parent: parent,
		local:  lib.NewConstrainedWorkGroup(max(1, limit)),
	}
}

func (g *uploadV2PartConcurrencyGate) Wait() {
	g.local.Wait()
	g.parent.Wait()
}

func (g *uploadV2PartConcurrencyGate) WaitWithContext(ctx context.Context) bool {
	if !g.local.WaitWithContext(ctx) {
		return false
	}
	if g.parent.WaitWithContext(ctx) {
		return true
	}
	g.local.Done()
	return false
}

func (g *uploadV2PartConcurrencyGate) Done() {
	g.DoneWithSample(lib.AdaptiveConcurrencySample{Success: true})
}

func (g *uploadV2PartConcurrencyGate) DoneWithSample(sample lib.AdaptiveConcurrencySample) {
	if sampler, ok := g.parent.(lib.AdaptiveConcurrencyManagerWithSample); ok {
		sampler.DoneWithSample(sample)
	} else {
		g.parent.Done()
	}
	g.local.Done()
}

func (g *uploadV2PartConcurrencyGate) WaitAllDone() {
	g.local.WaitAllDone()
	g.parent.WaitAllDone()
}

func (g *uploadV2PartConcurrencyGate) RunningCount() int {
	return g.local.RunningCount()
}

func (g *uploadV2PartConcurrencyGate) WaitForADone() bool {
	return g.local.WaitForADone()
}

func (g *uploadV2PartConcurrencyGate) WaitForADoneWithContext(ctx context.Context) bool {
	return g.local.WaitForADoneWithContext(ctx)
}

func newUploadV2Engine(u *uploadIO, plan uploadV2PartPlan) *uploadV2Engine {
	if u.transferStarted == nil {
		// Production uploads initialize this in Run. Some focused engine tests
		// construct uploadIO directly, so keep the engine helper self-contained.
		u.transferStarted = &atomic.Bool{}
	}
	maxConcurrency := u.uploadV2MaxConcurrency()
	manager := lib.NewAdaptiveConcurrencyManagerWithConfig(uploadV2AdaptiveConcurrencyConfigWithInitial(plan, maxConcurrency, uploadV2InitialConcurrencyForPlan(plan, maxConcurrency, u.uploadV2Tuning), u.uploadV2Tuning))
	if u.uploadV2ManagerProvider != nil {
		manager = u.uploadV2ManagerProvider(plan, maxConcurrency, u.uploadV2Tuning)
	}
	transportMaxConcurrency := uploadV2HTTPMaxConnsPerHost(plan, u.uploadV2Tuning, manager.Max())
	idleConnectionCap := uploadV2HTTPIdleConnectionCap(plan, transportMaxConcurrency)
	httpClientLimits := uploadV2HTTPClientLimits{}
	if u.uploadV2HTTPClientProvider != nil {
		if client, limits, ok := u.uploadV2HTTPClientProvider(u.Client, plan, transportMaxConcurrency, idleConnectionCap); ok {
			u.Client = client
			httpClientLimits = limits
		} else {
			httpClientLimits = u.configureUploadV2HTTPClient(transportMaxConcurrency, idleConnectionCap)
		}
	} else {
		httpClientLimits = u.configureUploadV2HTTPClient(transportMaxConcurrency, idleConnectionCap)
	}
	var globalManager lib.ConcurrencyManager
	if u.managerSet && !u.uploadV2UseSDKDefaultCaps {
		globalManager = u.Manager
	}
	engine := &uploadV2Engine{
		u:                         u,
		plan:                      plan,
		manager:                   manager,
		globalManager:             globalManager,
		etags:                     make(map[int]files_sdk.EtagsParam, enginePreallocatePartCapacity(plan, manager.Max())),
		usePartOffsets:            uploadV2UsesPartOffsets(plan.target),
		httpClientLimits:          httpClientLimits,
		checksumTrailerEnabled:    false,
		checksumTrailerSkipReason: "",
	}
	engine.checksumTrailerEnabled, engine.checksumTrailerSkipReason = uploadV2ChecksumTrailerDecision(u, plan)
	engine.prepareKnownSizeStorage()
	return engine
}

func (u *uploadIO) configureUploadV2HTTPClient(maxConnsPerHost int, maxIdleConnsPerHost int) uploadV2HTTPClientLimits {
	client, limits := configuredUploadV2HTTPClient(u.Client, maxConnsPerHost, maxIdleConnsPerHost)
	if client != nil {
		u.Client = client
	}
	return limits
}

func configuredUploadV2HTTPClient(client *Client, maxConnsPerHost int, maxIdleConnsPerHost int) (*Client, uploadV2HTTPClientLimits) {
	if client == nil || client.Config.Client == nil || client.Config.Client.HTTPClient == nil {
		return client, uploadV2HTTPClientLimits{}
	}
	httpClient, ok := lib.CloneHTTPClientWithExactMaxConnsPerHost(client.Config.Client.HTTPClient, maxConnsPerHost)
	if !ok {
		limits := uploadV2HTTPClientLimitsForClient(client.Config.Client.HTTPClient)
		limits.adjusted = false
		return client, limits
	}
	limitUploadV2HTTPClientIdleConns(httpClient, maxIdleConnsPerHost)

	retryClient := *client.Config.Client
	retryClient.HTTPClient = httpClient

	adjustedClient := *client
	adjustedClient.Config.Client = &retryClient

	limits := uploadV2HTTPClientLimitsForClient(httpClient)
	limits.adjusted = true
	return &adjustedClient, limits
}

func uploadV2HTTPIdleConnectionCap(plan uploadV2PartPlan, maxConnsPerHost int) int {
	if maxConnsPerHost <= 0 {
		return 0
	}
	switch plan.target {
	case uploadV2TargetS3:
		return min(maxConnsPerHost, uploadV2DefaultHTTPIdleConnectionCap)
	default:
		return maxConnsPerHost
	}
}

func uploadV2HTTPMaxConnsPerHost(plan uploadV2PartPlan, tuning UploadV2Tuning, maxConnsPerHost int) int {
	if maxConnsPerHost <= 0 || plan.target != uploadV2TargetS3 {
		return maxConnsPerHost
	}
	ceiling := AdaptiveTransferS3GrowthCeiling
	if tuning.InitialTarget > 0 {
		ceiling = tuning.InitialTarget
	}
	if tuning.S3GrowthCeiling > 0 {
		ceiling = tuning.S3GrowthCeiling
	}
	if ceiling <= 0 || ceiling >= maxConnsPerHost {
		return maxConnsPerHost
	}
	probeBytes := AdaptiveTransferS3GrowthCeilingProbeBytes
	if tuning.S3GrowthCeilingProbeBytes > 0 {
		probeBytes = tuning.S3GrowthCeilingProbeBytes
	}
	workloadBytes := tuning.S3WorkloadBytes
	if workloadBytes <= 0 && plan.totalSize != nil {
		workloadBytes = *plan.totalSize
	}
	// The transport can open based on workload bytes before the adaptive
	// manager unlocks growth above the S3 soft ceiling. It can also open when
	// the workload implies enough scheduled parts for many-file probing. That
	// keeps connection headroom ready, while actual concurrency still has to
	// prove sustained throughput before it probes above the default plateau.
	if probeBytes > 0 && workloadBytes >= probeBytes {
		return maxConnsPerHost
	}
	probeSuccesses := AdaptiveTransferS3GrowthCeilingProbeSuccesses
	if tuning.S3GrowthCeilingProbeSuccesses > 0 {
		probeSuccesses = tuning.S3GrowthCeilingProbeSuccesses
	}
	if probeSuccesses > 0 && workloadBytes > 0 && plan.partSize > 0 && ceilDiv(workloadBytes, plan.partSize) >= int64(probeSuccesses) {
		return maxConnsPerHost
	}
	return ceiling
}

func limitUploadV2HTTPClientIdleConns(client *http.Client, maxIdleConnsPerHost int) {
	if client == nil || maxIdleConnsPerHost <= 0 {
		return
	}
	switch transport := client.Transport.(type) {
	case *lib.Transport:
		if transport.Transport != nil {
			limitUploadV2HTTPTransportIdleConns(transport.Transport, maxIdleConnsPerHost)
		}
	case *http.Transport:
		limitUploadV2HTTPTransportIdleConns(transport, maxIdleConnsPerHost)
	}
}

func limitUploadV2HTTPTransportIdleConns(transport *http.Transport, maxIdleConnsPerHost int) {
	if transport == nil || maxIdleConnsPerHost <= 0 {
		return
	}
	if transport.MaxIdleConns == 0 || transport.MaxIdleConns > maxIdleConnsPerHost {
		transport.MaxIdleConns = maxIdleConnsPerHost
	}
	if transport.MaxIdleConnsPerHost == 0 || transport.MaxIdleConnsPerHost > maxIdleConnsPerHost {
		transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func uploadV2HTTPClientLimitsForClient(client *http.Client) uploadV2HTTPClientLimits {
	if client == nil || client.Transport == nil {
		return uploadV2HTTPClientLimits{}
	}
	switch transport := client.Transport.(type) {
	case *lib.Transport:
		if transport.Transport == nil {
			return uploadV2HTTPClientLimits{}
		}
		return uploadV2HTTPClientLimits{
			available:           true,
			maxIdleConns:        transport.MaxIdleConns,
			maxIdleConnsPerHost: transport.MaxIdleConnsPerHost,
			maxConnsPerHost:     transport.MaxConnsPerHost,
		}
	case *http.Transport:
		return uploadV2HTTPClientLimits{
			available:           true,
			maxIdleConns:        transport.MaxIdleConns,
			maxIdleConnsPerHost: transport.MaxIdleConnsPerHost,
			maxConnsPerHost:     transport.MaxConnsPerHost,
		}
	default:
		return uploadV2HTTPClientLimits{}
	}
}

func (e *uploadV2Engine) prepareKnownSizeStorage() {
	partCount := max(len(e.u.Parts), enginePreallocatePartCapacity(e.plan, e.manager.Max()))
	if partCount <= 0 || cap(e.u.Parts) >= partCount {
		return
	}
	if _, readerAtOk := e.u.ReaderAt(); !readerAtOk || e.u.Size == nil {
		return
	}
	parts := make(Parts, len(e.u.Parts), partCount)
	copy(parts, e.u.Parts)
	e.u.Parts = parts
}

func enginePreallocatePartCapacity(plan uploadV2PartPlan, maxConcurrency int) int {
	partCount := plan.estimatedPartCount()
	if partCount <= 0 {
		return 0
	}
	maxPreallocate := max(1, maxConcurrency) * uploadV2PreallocateConcurrencyMultiplier
	return min(partCount, maxPreallocate)
}

func sortedUploadV2Parts(parts Parts) Parts {
	if len(parts) < 2 {
		return parts
	}

	sorted := true
	for index := 1; index < len(parts); index++ {
		previous, current := parts[index-1], parts[index]
		if previous == nil || current == nil || previous.number > current.number {
			sorted = false
			break
		}
	}
	if sorted {
		return parts
	}

	sortedParts := append(Parts(nil), parts...)
	sort.SliceStable(sortedParts, func(i, j int) bool {
		if sortedParts[i] == nil {
			return false
		}
		if sortedParts[j] == nil {
			return true
		}
		return sortedParts[i].number < sortedParts[j].number
	})
	return sortedParts
}

func (e *uploadV2Engine) resultsBufferSize() int {
	return max(1, e.partConcurrencyLimit())
}

func (e *uploadV2Engine) partConcurrencyManager() lib.ConcurrencyManager {
	parent := e.manager.NewSubWorker()
	limit := e.partConcurrencyLimit()
	if limit >= e.manager.Max() {
		return parent
	}
	return newUploadV2PartConcurrencyGate(parent, limit)
}

func (e *uploadV2Engine) partConcurrencyLimit() int {
	limit := max(1, e.manager.Max())
	if e.plan.target != uploadV2TargetS3 || e.plan.totalSize == nil || *e.plan.totalSize >= uploadV2SmallS3KnownSizeConcurrencyCutoff {
		return limit
	}
	partCount := e.plan.estimatedPartCount()
	if partCount <= 0 {
		return limit
	}
	return min(limit, uploadV2SmallS3KnownSizeConcurrency(partCount))
}

func uploadV2SmallS3KnownSizeConcurrency(partCount int) int {
	if partCount <= 0 {
		return 0
	}
	if partCount <= 12 {
		return partCount
	}
	return min(partCount, max(4, min(12, ceilDivInt(partCount, 2))))
}

func ceilDivInt(n int, d int) int {
	if n <= 0 {
		return 0
	}
	return (n + d - 1) / d
}

func (e *uploadV2Engine) resumeResetReason() string {
	if len(e.u.Parts) == 0 {
		return ""
	}

	parts := sortedUploadV2Parts(e.u.Parts)

	var expectedOffset int64
	for index, part := range parts {
		expectedNumber := index + 1
		if part == nil {
			return "nil_part"
		}
		if part.number != expectedNumber {
			return "non_contiguous_part_number"
		}
		if part.off != expectedOffset {
			return "non_contiguous_part_offset"
		}
		if part.len < 0 {
			return "negative_part_size"
		}
		if part.len == 0 && !e.isZeroSizePart(part) {
			return "zero_length_part"
		}
		if e.u.Size != nil && part.off+part.len > *e.u.Size {
			return "part_beyond_known_size"
		}
		if !e.partSizeMatchesPlan(part, index) {
			return "part_size_plan_mismatch"
		}
		if part.Successful() {
			if part.Etag == "" || part.Part == "" {
				return "successful_part_missing_etag"
			}
			if e.usesPartOffsets() && !uploadURIHasPartOffset(part.FileUploadPart.UploadUri, part.number, part.off) {
				return "successful_offset_part_missing_part_offset"
			}
		}
		expectedOffset += part.len
	}
	return ""
}

func (e *uploadV2Engine) isZeroSizePart(part *Part) bool {
	return e.u.Size != nil && *e.u.Size == 0 && part.number == 1 && part.off == 0 && part.len == 0
}

func (e *uploadV2Engine) partSizeMatchesPlan(part *Part, index int) bool {
	if e.u.Size != nil && part.off+part.len == *e.u.Size {
		return true
	}
	return part.len == e.plan.partSizeForIndex(index)
}

func (e *uploadV2Engine) run(ctx context.Context) (UploadResumable, error) {
	e.u.Manager = e.manager
	readyRunway := e.readyRunwayConfig()
	e.u.logUploadV2(e.uploadV2EnabledLogAttrs(readyRunway))

	partCtx, cancelParts := context.WithCancelCause(ctx)
	defer cancelParts(nil)

	wait := e.partConcurrencyManager()
	results := make(chan uploadV2PartResult, e.resultsBufferSize())
	producerDone := make(chan struct{})

	go func() {
		defer close(producerDone)
		e.scheduleParts(partCtx, wait, results)
	}()
	go func() {
		<-producerDone
		wait.WaitAllDone()
		close(results)
	}()

	var allErrors error
	for result := range results {
		e.applyResult(result)
		if result.err != nil {
			allErrors = errors.Join(allErrors, result.err)
			cancelParts(result.err)
		}
	}
	if allErrors == nil && partCtx.Err() != nil {
		allErrors = context.Cause(partCtx)
	}

	if allErrors != nil {
		e.u.rewindSuccessfulParts()
		e.logUploadV2SchedulerSummary(allErrors)
		e.u.logUploadV2Complete(allErrors, e.u.bytesWritten)
		return e.u.UploadResumable(), allErrors
	}

	path, ref := e.u.Path, e.u.FileUploadPart.Ref
	if e.u.renamedCallback != nil {
		path, ref = e.u.renamedCallback()
	}
	if err := ctx.Err(); err != nil {
		e.logUploadV2SchedulerSummary(err)
		e.u.logUploadV2Complete(err, e.u.bytesWritten)
		return e.u.UploadResumable(), err
	}
	e.u.logUploadV2(map[string]any{
		"timestamp":        time.Now(),
		"event":            "upload v2 finalize",
		"bytes_written":    e.u.bytesWritten,
		"etag_count":       len(e.etags),
		"read_duration_ms": e.readDuration.Milliseconds(),
		"read_duration_ns": e.readDuration.Nanoseconds(),
	})
	file, err := e.u.completeUpload(ctx, e.u.ProvidedMtime, e.etagsForComplete(), e.u.bytesWritten, path, ref)
	e.u.file = file
	if err != nil {
		e.u.LogPath(e.u.Path, map[string]any{
			"timestamp": time.Now(),
			"error":     err.Error(),
			"event":     "complete upload",
			"message":   "rewindSuccessfulParts",
		})
		e.u.rewindSuccessfulParts()
		if files_sdk.IsNotExist(err) {
			e.u.notResumable.Store(true)
		}
	}
	e.logUploadV2SchedulerSummary(err)
	e.u.logUploadV2Complete(err, e.u.bytesWritten)
	return e.u.UploadResumable(), err
}

func (e *uploadV2Engine) uploadV2EnabledLogAttrs(readyRunway uploadV2ReadyRunwayConfig) map[string]any {
	attrs := map[string]any{
		"timestamp":                   time.Now(),
		"event":                       "upload v2 enabled",
		"target_class":                string(e.plan.target),
		"part_size":                   e.plan.partSize,
		"part_size_cap":               e.plan.unknownCap,
		"part_size_mode":              e.plan.mode,
		"known_size":                  e.u.Size != nil,
		"offset_upload":               e.usesPartOffsets(),
		"adaptive_initial_target":     e.manager.Target(),
		"adaptive_max_target":         e.manager.Max(),
		"adaptive_part_target":        e.partConcurrencyLimit(),
		"adaptive_planned_parts":      e.plan.estimatedPartCount(),
		"ready_runway_parts":          readyRunway.parts,
		"ready_runway_bytes":          readyRunway.bytes,
		"upload_http_client_adjusted": e.httpClientLimits.adjusted,
	}
	if e.httpClientLimits.available {
		attrs["upload_max_idle_conns_per_host"] = e.httpClientLimits.maxIdleConnsPerHost
		attrs["upload_max_idle_conns"] = e.httpClientLimits.maxIdleConns
		attrs["upload_max_conns_per_host"] = e.httpClientLimits.maxConnsPerHost
	}
	return attrs
}

func (e *uploadV2Engine) scheduleParts(ctx context.Context, wait lib.ConcurrencyManager, results chan<- uploadV2PartResult) {
	runway := e.readyRunwayConfig()
	nextOffset, nextIndex := e.scheduleResumedParts(ctx, wait, results, runway)
	if ctx.Err() != nil {
		return
	}

	iterator := e.plan.resume(nextOffset, nextIndex)
	if e.plan.done(nextOffset, nextIndex) {
		return
	}
	nextDescriptor := func() (uploadV2PartDescriptor, bool) {
		if iterator == nil {
			return uploadV2PartDescriptor{}, false
		}
		offset, next, index := iterator()
		descriptor := uploadV2PartDescriptor{
			number: index + 1,
			offset: offset,
			final:  next == nil && e.u.Size != nil,
			upload: e.uploadForPart(index+1, offset),
		}
		iterator = next
		return descriptor, true
	}
	e.schedulePartDescriptors(ctx, wait, results, runway, nextDescriptor)
}

func (e *uploadV2Engine) scheduleResumedParts(ctx context.Context, wait lib.ConcurrencyManager, results chan<- uploadV2PartResult, runway uploadV2ReadyRunwayConfig) (int64, int) {
	parts := sortedUploadV2Parts(e.u.Parts)

	var nextOffset int64
	var nextIndex int
	for _, part := range parts {
		if part.off+part.len > nextOffset {
			nextOffset = part.off + part.len
		}
		if part.number > nextIndex {
			nextIndex = part.number
		}
		if part.Successful() {
			e.u.Progress(part.bytes)
			e.recordSuccess(part.number, part.EtagsParam, part.bytes)
		}
	}

	partIndex := 0
	nextDescriptor := func() (uploadV2PartDescriptor, bool) {
		for partIndex < len(parts) && parts[partIndex].Successful() {
			partIndex++
		}
		if partIndex >= len(parts) {
			return uploadV2PartDescriptor{}, false
		}
		part := parts[partIndex]
		partIndex++

		part.Clear()
		upload := part.FileUploadPart
		if upload.Path == "" {
			upload.Path = e.u.Path
		}
		if upload.Ref == "" {
			upload.Ref = e.u.FileUploadPart.Ref
		}
		if upload.HttpMethod == "" {
			upload.HttpMethod = e.u.FileUploadPart.HttpMethod
		}
		upload.ParallelParts = e.u.FileUploadPart.ParallelParts
		upload.PartNumber = int64(part.number)
		if e.usePartOffsets {
			decorateUploadURLWithPartOffset(&upload, part.number, part.off)
		}

		return uploadV2PartDescriptor{
			number: part.number,
			offset: part.OffSet,
			final:  part.final,
			upload: upload,
			legacy: part,
		}, true
	}
	e.schedulePartDescriptors(ctx, wait, results, runway, nextDescriptor)
	return nextOffset, nextIndex
}

func (e *uploadV2Engine) schedulePartDescriptors(ctx context.Context, wait lib.ConcurrencyManager, results chan<- uploadV2PartResult, runway uploadV2ReadyRunwayConfig, nextDescriptor func() (uploadV2PartDescriptor, bool)) {
	if runway.parts <= 0 {
		for {
			descriptor, ok := nextDescriptor()
			if !ok {
				return
			}
			if !e.schedulePart(ctx, wait, results, descriptor) {
				return
			}
		}
	}

	var ready []uploadV2PreparedPart
	var readyBytes int64
	var pendingDescriptor uploadV2PartDescriptor
	var hasPendingDescriptor bool
	sourceDone := false
	abortScheduling := false
	defer func() {
		e.closePreparedParts(ready)
	}()
	takeNextDescriptor := func() (uploadV2PartDescriptor, bool) {
		if hasPendingDescriptor {
			hasPendingDescriptor = false
			return pendingDescriptor, true
		}
		if sourceDone {
			return uploadV2PartDescriptor{}, false
		}
		descriptor, ok := nextDescriptor()
		if !ok {
			sourceDone = true
			return uploadV2PartDescriptor{}, false
		}
		return descriptor, true
	}
	prepareDescriptor := func(descriptor uploadV2PartDescriptor) bool {
		prepared, keepGoing, ok := e.preparePart(ctx, results, descriptor)
		if !ok {
			sourceDone = true
			abortScheduling = true
			return false
		}
		ready = append(ready, prepared)
		readyBytes += prepared.memoryBytes
		e.stats.recordReadyDepth(len(ready), readyBytes)
		if !keepGoing {
			sourceDone = true
		}
		return true
	}
	fillRunway := func() {
		for !sourceDone && !abortScheduling {
			descriptor, ok := takeNextDescriptor()
			if !ok {
				return
			}
			nextMemoryBytes := e.readyRunwayMemoryCost(descriptor.offset.len)
			if !e.readyRunwayCanPrepare(ready, readyBytes, nextMemoryBytes, runway) {
				if len(ready) >= runway.parts {
					e.stats.recordRunwayPartCapBlock()
				} else {
					e.stats.recordRunwayByteCapBlock()
				}
				pendingDescriptor = descriptor
				hasPendingDescriptor = true
				return
			}
			if !prepareDescriptor(descriptor) {
				return
			}
		}
	}

	for {
		if abortScheduling {
			return
		}

		if len(ready) == 0 {
			descriptor, ok := takeNextDescriptor()
			if !ok {
				return
			}
			if !e.schedulePart(ctx, wait, results, descriptor) {
				return
			}
		} else {
			prepared := ready[0]
			copy(ready, ready[1:])
			ready[len(ready)-1] = uploadV2PreparedPart{}
			ready = ready[:len(ready)-1]
			readyBytes -= prepared.memoryBytes
			e.stats.recordReadyDepth(len(ready), readyBytes)
			if !e.dispatchPreparedPart(ctx, wait, results, prepared.part) {
				prepared.part.closeReader()
				return
			}
		}

		fillRunway()
	}
}

func (e *uploadV2Engine) schedulePart(ctx context.Context, wait lib.ConcurrencyManager, results chan<- uploadV2PartResult, descriptor uploadV2PartDescriptor) bool {
	if !e.waitForPartCapacity(ctx, wait) {
		return false
	}

	reader, progress, err := e.buildReader(descriptor.offset)
	if err != nil {
		results <- uploadV2PartResult{part: &uploadV2Part{uploadV2PartDescriptor: descriptor}, err: err}
		e.donePart(wait, lib.AdaptiveConcurrencySample{Success: false})
		return false
	}
	if reader.Len() == 0 && e.u.Size == nil {
		reader.Close()
		e.donePart(wait, lib.AdaptiveConcurrencySample{Success: true})
		return false
	}

	if descriptor.legacy == nil {
		descriptor.legacy = e.newResumablePart(descriptor, reader)
	}
	part := &uploadV2Part{uploadV2PartDescriptor: descriptor, reader: reader, progress: progress}
	e.u.logUploadV2(map[string]any{
		"timestamp":   time.Now(),
		"event":       "upload v2 part scheduled",
		"part":        descriptor.number,
		"part_offset": descriptor.offset.off,
		"part_size":   descriptor.offset.len,
		"final":       descriptor.final,
	})
	go func() {
		result := e.runPart(ctx, part)
		results <- result
		e.donePart(wait, result.sample())
	}()
	e.stats.recordDirectScheduled()

	return int64(reader.Len()) == descriptor.offset.len
}

func (e *uploadV2Engine) preparePart(ctx context.Context, results chan<- uploadV2PartResult, descriptor uploadV2PartDescriptor) (uploadV2PreparedPart, bool, bool) {
	if ctx.Err() != nil {
		return uploadV2PreparedPart{}, false, false
	}

	reader, progress, err := e.buildReader(descriptor.offset)
	if err != nil {
		results <- uploadV2PartResult{part: &uploadV2Part{uploadV2PartDescriptor: descriptor}, err: err}
		return uploadV2PreparedPart{}, false, false
	}
	if reader.Len() == 0 && e.u.Size == nil {
		reader.Close()
		return uploadV2PreparedPart{}, false, false
	}

	if descriptor.legacy == nil {
		descriptor.legacy = e.newResumablePart(descriptor, reader)
	}
	part := &uploadV2Part{uploadV2PartDescriptor: descriptor, reader: reader, progress: progress}
	prepared := uploadV2PreparedPart{
		part:        part,
		memoryBytes: e.readyRunwayMemoryCost(int64(reader.Len())),
	}
	e.stats.recordPrepared(int64(reader.Len()), prepared.memoryBytes)
	return prepared, int64(reader.Len()) == descriptor.offset.len, true
}

func (e *uploadV2Engine) dispatchPreparedPart(ctx context.Context, wait lib.ConcurrencyManager, results chan<- uploadV2PartResult, part *uploadV2Part) bool {
	if !e.waitForPartCapacity(ctx, wait) {
		return false
	}
	e.u.logUploadV2(map[string]any{
		"timestamp":     time.Now(),
		"event":         "upload v2 part scheduled",
		"part":          part.number,
		"part_offset":   part.offset.off,
		"part_size":     part.offset.len,
		"final":         part.final,
		"ready_runway":  true,
		"prepared_size": part.reader.Len(),
	})
	go func() {
		result := e.runPart(ctx, part)
		results <- result
		e.donePart(wait, result.sample())
	}()
	e.stats.recordDispatchedPrepared()
	return true
}

func (e *uploadV2Engine) readyRunwayConfig() uploadV2ReadyRunwayConfig {
	config := e.u.uploadV2ReadyRunway.resolved()
	if !e.u.uploadV2ReadyRunway.configured && (e.plan.target == uploadV2TargetS3 || e.plan.target == uploadV2TargetDirect) && !e.readyRunwayBuffersPartBytes() {
		parts := max(uploadV2DefaultSeekableS3ReadyRunwayMinParts, e.manager.Target()/uploadV2DefaultSeekableS3ReadyRunwayTargetDivisor)
		parts = min(parts, uploadV2DefaultSeekableS3ReadyRunwayMaxParts)
		if partCount := e.plan.estimatedPartCount(); partCount > 0 {
			parts = min(parts, partCount)
		}
		config.parts = max(config.parts, parts)
	}
	if config.parts < 0 {
		config.parts = 0
	}
	if config.bytes < 0 {
		config.bytes = 0
	}
	return config
}

func (e *uploadV2Engine) readyRunwayCanPrepare(ready []uploadV2PreparedPart, readyBytes int64, nextMemoryBytes int64, runway uploadV2ReadyRunwayConfig) bool {
	if len(ready) >= runway.parts {
		return false
	}
	if runway.bytes == 0 || nextMemoryBytes <= 0 {
		return true
	}
	return readyBytes+nextMemoryBytes <= runway.bytes
}

func (e *uploadV2Engine) readyRunwayMemoryCost(bytes int64) int64 {
	if !e.readyRunwayBuffersPartBytes() {
		return 0
	}
	return max(bytes, 0)
}

func (e *uploadV2Engine) readyRunwayBuffersPartBytes() bool {
	_, readerAtOk := e.u.ReaderAt()
	return e.u.Size == nil || !readerAtOk
}

func (e *uploadV2Engine) closePreparedParts(parts []uploadV2PreparedPart) {
	for _, prepared := range parts {
		if prepared.part != nil {
			prepared.part.closeReader()
		}
	}
}

func (e *uploadV2Engine) waitForPartCapacity(ctx context.Context, wait lib.ConcurrencyManager) bool {
	start := time.Now()
	if !wait.WaitWithContext(ctx) {
		e.stats.recordAdaptiveWait(time.Since(start), false)
		return false
	}
	e.stats.recordAdaptiveWait(time.Since(start), true)
	if e.globalManager == nil {
		e.u.markTransferStarted()
		return true
	}
	start = time.Now()
	if e.globalManager.WaitWithContext(ctx) {
		e.stats.recordGlobalWait(time.Since(start), true)
		e.u.markTransferStarted()
		return true
	}
	e.stats.recordGlobalWait(time.Since(start), false)
	e.done(wait, lib.AdaptiveConcurrencySample{Success: false})
	return false
}

func (e *uploadV2Engine) newResumablePart(descriptor uploadV2PartDescriptor, reader ProxyReader) *Part {
	part := &Part{
		OffSet:         descriptor.offset,
		number:         descriptor.number,
		final:          descriptor.final,
		ProxyReader:    reader,
		FileUploadPart: descriptor.upload,
	}
	if _, readerAtOk := e.u.ReaderAt(); readerAtOk && e.u.Size != nil {
		e.u.Parts = append(e.u.Parts, part)
	}
	return part
}

func (e *uploadV2Engine) buildReader(offset OffSet) (reader ProxyReader, progress *uploadV2ProgressBatcher, err error) {
	start := time.Now()
	defer func() {
		bytes := offset.len
		if reader != nil {
			bytes = int64(reader.Len())
		}
		e.stats.recordReaderBuild(bytes, time.Since(start), err)
	}()

	progress = newUploadV2ProgressBatcher(e.u.Progress)
	readerAt, readerAtOk := e.u.ReaderAt()

	if e.u.Size != nil && readerAtOk {
		return newProxySectionReader(readerAt, offset.off, offset.len, progress.Add, true), progress, nil
	}

	if e.u.Size == nil || lib.UnWrapBool(e.u.FileUploadPart.ParallelParts) {
		source, _ := e.u.Reader()
		if readerAtOk {
			source = io.NewSectionReader(readerAt, offset.off, offset.len)
		}

		buf := new(bytes.Buffer)
		n, err := io.CopyN(buf, source, offset.len)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		return newProxySectionReader(bytes.NewReader(buf.Bytes()), 0, n, progress.Add, true), progress, nil
	}

	source, _ := e.u.Reader()
	return &ProxyRead{
		Reader:            source,
		len:               offset.len,
		onRead:            progress.Add,
		trackReadDuration: true,
	}, progress, nil
}

func (e *uploadV2Engine) runPart(ctx context.Context, part *uploadV2Part) uploadV2PartResult {
	defer part.flushProgress()
	result := uploadV2PartResult{part: part}
	for attempt := 1; attempt <= maxUploadPartAttempts; attempt++ {
		start := time.Now()
		result.etag, result.statusCode, result.backPressure, result.retryAfter, result.err = e.uploadPart(ctx, part)
		result.duration = time.Since(start)
		if result.err != nil && uploadV2NetworkBackPressure(result.err) {
			result.backPressure = true
		}
		result.bytes = int64(part.reader.BytesRead())
		result.readDuration = part.reader.ReadDuration()
		part.legacy.bytes = result.bytes
		part.legacy.EtagsParam = result.etag
		part.legacy.error = result.err
		if result.err == nil {
			return result
		}
		if !(lib.S3ErrorIsRequestHasExpired(result.err) || files_sdk.IsExpired(result.err)) {
			return result
		}
		part.upload.Expires = ""
		part.upload.UploadUri = ""
		if attempt == maxUploadPartAttempts || !e.rewindPartForRetry(part, attempt, result.err, "clearing upload_uri and fetching new one") {
			return result
		}
	}
	return result
}

func (e *uploadV2Engine) uploadPart(ctx context.Context, part *uploadV2Part) (files_sdk.EtagsParam, int, bool, time.Duration, error) {
	if part.upload.PartNumber != 1 || part.upload.UploadUri == "" {
		if time.Now().After(part.upload.ExpiresTime()) {
			if !part.upload.ExpiresTime().IsZero() {
				e.u.LogPath(e.u.Path, map[string]any{
					"timestamp":           time.Now(),
					"part":                part.upload.PartNumber,
					"part_offset":         part.offset.off,
					"event":               "upload v2 part",
					"previous_upload_uri": part.upload.UploadUri != "",
					"expired":             time.Now().After(part.upload.ExpiresTime()),
				})
			}
			params := files_sdk.FileBeginUploadParams{
				Path:         part.upload.Path,
				Ref:          part.upload.Ref,
				Part:         part.upload.PartNumber,
				MkdirParents: lib.Bool(true),
			}

			start := time.Now()
			upload, err := e.u.startUpload(ctx, params)
			e.stats.recordUploadURLRefresh(time.Since(start), err)
			if err != nil && files_sdk.IsNotExist(err) && params.Ref != "" {
				params.Ref = ""
				start = time.Now()
				upload, err = e.u.startUpload(ctx, params)
				e.stats.recordUploadURLRefresh(time.Since(start), err)
			}
			if err != nil {
				return files_sdk.EtagsParam{}, 0, false, 0, err
			}
			part.upload = upload
			part.upload.PartNumber = int64(part.number)
			if e.usePartOffsets {
				decorateUploadURLWithPartOffset(&part.upload, part.number, part.offset.off)
			}
			part.legacy.FileUploadPart = part.upload
		}
	}

	headers := http.Header{}
	headers.Add("Content-Length", strconv.FormatInt(int64(part.reader.Len()), 10))
	body := io.ReadCloser(part.reader)
	trailer, err := e.prepareChecksumTrailer(part)
	if err != nil {
		return files_sdk.EtagsParam{}, 0, false, 0, err
	}
	if trailer.enabled {
		headers = trailer.headers
		trailerBody, err := trailer.newReader(part.reader)
		if err != nil {
			return files_sdk.EtagsParam{}, 0, false, 0, err
		}
		body = uploadV2ReadCloser{Reader: trailerBody, closer: part.reader}
	}
	params := &files_sdk.CallParams{
		Method:  part.upload.HttpMethod,
		Config:  e.u.Config,
		Uri:     part.upload.UploadUri,
		BodyIo:  body,
		Headers: &headers,
		Context: ctx,
	}
	canReplay := func() bool {
		if part.reader.Len() == 0 {
			return true
		}
		return part.reader.Rewind()
	}
	backPressureSeen := false
	retryAfterSeen := time.Duration(0)
	observer := func(attempt lib.UploadRetryAttempt) {
		if uploadV2BackPressureStatus(attempt.StatusCode) {
			backPressureSeen = true
			if attempt.RetryAfter > retryAfterSeen {
				retryAfterSeen = attempt.RetryAfter
			}
		}
	}
	params.Client = lib.UploadRetryableHttpWithObserver(e.u.Config.Client, canReplay, observer)
	if part.reader.Len() > 0 {
		params.RetryableBody = retryablehttp.ReaderFunc(func() (io.Reader, error) {
			if !part.reader.Rewind() {
				return nil, errors.New("upload part body not rewindable")
			}
			if trailer.enabled {
				trailerBody, err := trailer.newReader(part.reader)
				if err != nil {
					return nil, err
				}
				return uploadPartRetryReader{Reader: trailerBody, len: int(trailer.encodedLength)}, nil
			}
			return uploadPartRetryReader{Reader: part.reader, len: part.reader.Len()}, nil
		})
	} else if trailer.enabled {
		params.RetryableBody = retryablehttp.ReaderFunc(func() (io.Reader, error) {
			if !part.reader.Rewind() {
				return nil, errors.New("upload part body not rewindable")
			}
			trailerBody, err := trailer.newReader(part.reader)
			if err != nil {
				return nil, err
			}
			return uploadPartRetryReader{Reader: trailerBody, len: int(trailer.encodedLength)}, nil
		})
	}
	callStart := time.Now()
	res, callErr := e.callUploadV2Part(ctx, part, params, canReplay, observer)
	callDuration := time.Since(callStart)
	if callErr != nil {
		e.stats.recordHTTPCall(callDuration, int64(part.reader.Len()), callErr)
		return files_sdk.EtagsParam{}, 0, backPressureSeen, retryAfterSeen, callErr
	}
	statusCode := 0
	retryAfter := time.Duration(0)
	if res != nil {
		statusCode = res.StatusCode
		retryAfter = parseUploadV2RetryAfter(res.Header.Get("Retry-After"))
		if res.Body != nil {
			defer res.Body.Close()
		}
		if uploadV2BackPressureStatus(statusCode) {
			backPressureSeen = true
		}
		if retryAfter > retryAfterSeen {
			retryAfterSeen = retryAfter
		}
	}
	if err := lib.ResponseErrors(res, files_sdk.APIError(), lib.S3XMLError, lib.NonOkError); err != nil {
		e.stats.recordHTTPCall(callDuration, int64(part.reader.Len()), err)
		return files_sdk.EtagsParam{}, statusCode, backPressureSeen, retryAfterSeen, err
	}
	e.stats.recordHTTPCall(callDuration, int64(part.reader.Len()), nil)
	etag := strings.Trim(res.Header.Get("Etag"), "\"")
	if etag == "" {
		etag = "null"
	}
	return files_sdk.EtagsParam{
		Etag: etag,
		Part: strconv.Itoa(part.number),
	}, statusCode, backPressureSeen, retryAfterSeen, nil
}

func (e *uploadV2Engine) callUploadV2Part(ctx context.Context, part *uploadV2Part, params *files_sdk.CallParams, canReplay func() bool, observer func(lib.UploadRetryAttempt)) (*http.Response, error) {
	if res, ok := e.u.callRawWithDirectTransfer(ctx, part.upload.DirectConnectionInfo, params, directTransferUploadOptions{
		partNumber: int64(part.number),
		canReplay:  canReplay,
		observer:   observer,
		stats:      &e.stats,
	}); ok {
		return res, nil
	}

	return files_sdk.CallRaw(params)
}

func (e *uploadV2Engine) rewindPartForRetry(part *uploadV2Part, attempt int, err error, message string) bool {
	if !part.reader.Rewind() {
		return false
	}
	e.u.LogPath(e.u.Path, map[string]any{
		"timestamp":   time.Now(),
		"event":       "upload v2 part retry",
		"error":       uploadRetryLogError(err),
		"part":        part.number,
		"part_offset": part.offset.off,
		"run_count":   attempt,
		"message":     message,
	})
	return true
}

func (e *uploadV2Engine) uploadForPart(number int, offset OffSet) files_sdk.FileUploadPart {
	if number == 1 {
		upload := e.u.FileUploadPart
		upload.PartNumber = int64(number)
		if e.usePartOffsets {
			decorateUploadURLWithPartOffset(&upload, number, offset.off)
		}
		return upload
	}

	upload := files_sdk.FileUploadPart{
		HttpMethod:    e.u.FileUploadPart.HttpMethod,
		Path:          e.u.FileUploadPart.Path,
		Ref:           e.u.FileUploadPart.Ref,
		PartNumber:    int64(number),
		ParallelParts: e.u.ParallelParts,
	}
	if e.u.usesSameUrl(e.u.FileUploadPart) {
		upload.UploadUri = e.u.FileUploadPart.UploadUri
		upload.DirectConnectionInfo = e.u.FileUploadPart.DirectConnectionInfo
		upload.Expires = e.u.FileUploadPart.Expires
		if upload.PartNumber > 1 {
			upload.Expires = e.u.FileUploadPart.ExpiresTime().Add(3 * time.Minute).Format(time.RFC3339)
		}
		if e.usePartOffsets {
			decorateUploadURLWithPartOffset(&upload, number, offset.off)
		}
	}
	return upload
}

func (e *uploadV2Engine) decorateUploadURL(upload *files_sdk.FileUploadPart, partNumber int, partOffset int64) {
	if !e.usePartOffsets {
		return
	}
	decorateUploadURLWithPartOffset(upload, partNumber, partOffset)
}

func decorateUploadURLWithPartOffset(upload *files_sdk.FileUploadPart, partNumber int, partOffset int64) {
	if upload.UploadUri == "" {
		return
	}
	upload.UploadUri = uploadURLWithPartOffset(upload.UploadUri, partNumber, partOffset)
	if upload.DirectConnectionInfo.DirectUri != "" {
		upload.DirectConnectionInfo.DirectUri = uploadURLWithPartOffset(upload.DirectConnectionInfo.DirectUri, partNumber, partOffset)
	}
}

func uploadURLWithPartOffset(uploadURI string, partNumber int, partOffset int64) string {
	uriWithoutFragment := uploadURI
	fragment := ""
	if fragmentIndex := strings.IndexByte(uploadURI, '#'); fragmentIndex >= 0 {
		uriWithoutFragment = uploadURI[:fragmentIndex]
		fragment = uploadURI[fragmentIndex:]
	}

	base := uriWithoutFragment
	query := ""
	if queryIndex := strings.IndexByte(uriWithoutFragment, '?'); queryIndex >= 0 {
		base = uriWithoutFragment[:queryIndex]
		query = uriWithoutFragment[queryIndex+1:]
	}

	var partNumberBytes [20]byte
	var partOffsetBytes [20]byte
	partNumberValue := strconv.AppendInt(partNumberBytes[:0], int64(partNumber), 10)
	partOffsetValue := strconv.AppendInt(partOffsetBytes[:0], partOffset, 10)

	var builder strings.Builder
	builder.Grow(len(uploadURI) + len("part_number=&part_offset=") + len(partNumberValue) + len(partOffsetValue) + 1)
	builder.WriteString(base)
	separator := byte('?')
	if query != "" {
		for len(query) > 0 {
			param := query
			if separatorIndex := strings.IndexByte(query, '&'); separatorIndex >= 0 {
				param = query[:separatorIndex]
				query = query[separatorIndex+1:]
			} else {
				query = ""
			}
			if param == "" || uploadV2OffsetQueryParam(param) {
				continue
			}
			builder.WriteByte(separator)
			builder.WriteString(param)
			separator = '&'
		}
	}
	builder.WriteByte(separator)
	builder.WriteString("part_number=")
	builder.Write(partNumberValue)
	builder.WriteByte('&')
	builder.WriteString("part_offset=")
	builder.Write(partOffsetValue)
	builder.WriteString(fragment)
	return builder.String()
}

func uploadV2OffsetQueryParam(param string) bool {
	key := param
	if equalIndex := strings.IndexByte(param, '='); equalIndex >= 0 {
		key = param[:equalIndex]
	}
	return key == "part_number" || key == "partNumber" || key == "part_offset" || key == "offset"
}

func uploadURIHasPartOffset(uploadURI string, partNumber int, partOffset int64) bool {
	uri, err := url.Parse(uploadURI)
	if err != nil {
		return false
	}
	q := uri.Query()
	gotPartNumber := q.Get("part_number")
	if gotPartNumber == "" {
		gotPartNumber = q.Get("partNumber")
	}
	gotOffset := q.Get("part_offset")
	if gotOffset == "" {
		gotOffset = q.Get("offset")
	}
	return gotPartNumber == strconv.Itoa(partNumber) && gotOffset == strconv.FormatInt(partOffset, 10)
}

func (e *uploadV2Engine) usesPartOffsets() bool {
	return e.usePartOffsets
}

func uploadV2UsesPartOffsets(target TransferV2TargetClass) bool {
	return target != uploadV2TargetS3
}

func (e *uploadV2Engine) done(wait lib.ConcurrencyManager, sample lib.AdaptiveConcurrencySample) {
	if sampler, ok := wait.(lib.AdaptiveConcurrencyManagerWithSample); ok {
		sampler.DoneWithSample(sample)
		return
	}
	wait.Done()
}

func (e *uploadV2Engine) donePart(wait lib.ConcurrencyManager, sample lib.AdaptiveConcurrencySample) {
	if e.globalManager != nil {
		e.globalManager.Done()
	}
	e.done(wait, sample)
}

func (e *uploadV2Engine) applyResult(result uploadV2PartResult) {
	e.readDuration += result.readDuration
	e.stats.recordPartComplete(result.bytes, result.duration, result.err)
	if result.err != nil {
		e.u.Progress(-result.bytes)
		e.u.LogPath(e.u.Path, map[string]any{
			"timestamp":                        time.Now(),
			"error":                            result.err.Error(),
			"part":                             result.part.number,
			"part_offset":                      result.part.offset.off,
			"duration_ms":                      result.duration.Milliseconds(),
			"read_duration_ms":                 result.readDuration.Milliseconds(),
			"read_duration_ns":                 result.readDuration.Nanoseconds(),
			"read_throughput_bytes_per_second": uploadV2RateBytesPerSecond(result.bytes, result.readDuration),
			"read_duration_ratio":              uploadV2DurationRatio(result.readDuration, result.duration),
			"status_code":                      result.statusCode,
			"back_pressure":                    result.backPressure,
			"event":                            "upload v2 part complete",
		})
		return
	}
	e.u.logUploadV2(map[string]any{
		"timestamp":                        time.Now(),
		"event":                            "upload v2 part complete",
		"part":                             result.part.number,
		"part_offset":                      result.part.offset.off,
		"bytes":                            result.bytes,
		"duration_ms":                      result.duration.Milliseconds(),
		"read_duration_ms":                 result.readDuration.Milliseconds(),
		"read_duration_ns":                 result.readDuration.Nanoseconds(),
		"read_throughput_bytes_per_second": uploadV2RateBytesPerSecond(result.bytes, result.readDuration),
		"read_duration_ratio":              uploadV2DurationRatio(result.readDuration, result.duration),
		"status_code":                      result.statusCode,
		"back_pressure":                    result.backPressure,
		"success":                          true,
	})
	e.recordSuccess(result.part.number, result.etag, result.bytes)
}

func uploadV2RateBytesPerSecond(bytes int64, duration time.Duration) float64 {
	if bytes <= 0 || duration <= 0 {
		return 0
	}
	return float64(bytes) / duration.Seconds()
}

func uploadV2DurationRatio(part time.Duration, total time.Duration) float64 {
	if part <= 0 || total <= 0 {
		return 0
	}
	return part.Seconds() / total.Seconds()
}

func (e *uploadV2Engine) recordSuccess(partNumber int, etag files_sdk.EtagsParam, bytes int64) {
	e.etags[partNumber] = etag
	e.u.bytesWritten += bytes
}

func (e *uploadV2Engine) etagsForComplete() []files_sdk.EtagsParam {
	if len(e.etags) == 0 {
		return nil
	}
	etags := make([]files_sdk.EtagsParam, len(e.etags))
	contiguous := true
	for partNumber, etag := range e.etags {
		if partNumber < 1 || partNumber > len(etags) {
			contiguous = false
			break
		}
		etags[partNumber-1] = etag
	}
	if contiguous {
		for _, etag := range etags {
			if etag.Part == "" {
				contiguous = false
				break
			}
		}
		if contiguous {
			return etags
		}
	}

	partNumbers := make([]int, 0, len(e.etags))
	for partNumber := range e.etags {
		partNumbers = append(partNumbers, partNumber)
	}
	sort.Ints(partNumbers)

	sortedEtags := make([]files_sdk.EtagsParam, 0, len(partNumbers))
	for _, partNumber := range partNumbers {
		sortedEtags = append(sortedEtags, e.etags[partNumber])
	}
	return sortedEtags
}

func (r uploadV2PartResult) sample() lib.AdaptiveConcurrencySample {
	backPressure := uploadV2NetworkBackPressure(r.err) || uploadV2BackPressureStatus(r.statusCode)
	retryAfter := time.Duration(0)
	if r.err != nil || uploadV2BackPressureStatus(r.statusCode) {
		backPressure = backPressure || r.backPressure
		retryAfter = r.retryAfter
	}
	return lib.AdaptiveConcurrencySample{
		Success:      r.err == nil,
		Duration:     r.duration,
		Bytes:        r.bytes,
		StatusCode:   r.statusCode,
		BackPressure: backPressure,
		RetryAfter:   retryAfter,
	}
}

func uploadV2BackPressureStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

func uploadV2NetworkBackPressure(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "broken pipe") ||
		strings.Contains(message, "connection reset by peer") ||
		strings.Contains(message, "tls handshake timeout") ||
		strings.Contains(message, "timeout awaiting response headers")
}

func parseUploadV2RetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		if wait := time.Until(retryAt); wait > 0 {
			return wait
		}
	}
	return 0
}
