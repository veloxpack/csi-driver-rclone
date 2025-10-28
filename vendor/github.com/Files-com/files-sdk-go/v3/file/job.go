package file

import (
	"context"
	"io/fs"
	"strings"
	"sync"
	"time"

	filesSDK "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/directory"
	"github.com/Files-com/files-sdk-go/v3/file/manager"
	"github.com/Files-com/files-sdk-go/v3/file/status"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/Files-com/files-sdk-go/v3/lib/direction"
	libLog "github.com/Files-com/files-sdk-go/v3/lib/logpath"
	"github.com/Files-com/files-sdk-go/v3/lib/timer"
	"github.com/bradfitz/iter"
	"github.com/chilts/sid"
	"github.com/hashicorp/go-retryablehttp"
	ignore "github.com/sabhiram/go-gitignore"
)

type EventsReporter map[status.Status][]reporterSettings

type reporterSettings struct {
	OnBytesChange bool
	Reporter
}

type IFile interface {
	SetStatus(status.Status, error)
	StatusChanges() status.Changes
	TransferBytes() int64
	IncrementTransferBytes(int64)
	File() filesSDK.File
	Size() int64
	Id() string
	LocalPath() string
	RemotePath() string
	Status() status.Status
	EndedAt() time.Time
	StartedAt() time.Time
	Err() error
	Job() *Job
}

func ToStatusFile(f IFile) JobFile {
	return JobFile{
		TransferBytes: f.TransferBytes(),
		File:          f.File(),
		LocalPath:     f.LocalPath(),
		RemotePath:    f.RemotePath(),
		Id:            f.Id(),
		Job:           f.Job(),
		Status:        f.Status(),
		EndedAt:       f.EndedAt(),
		StartedAt:     f.StartedAt(),
		Err:           MashableError{error: f.Err()}.Err(),
		Size:          f.Size(),
		StatusName:    f.Status().Name,
		Attempts:      f.StatusChanges().Count(status.Queued),
	}
}

type Subscriptions struct {
	Started     chan time.Time
	Finished    chan time.Time
	Canceled    chan time.Time
	Scanning    chan time.Time
	EndScanning chan time.Time
}

type Job struct {
	Id string
	*timer.Timer
	Statuses      []IFile
	Direction     direction.Direction
	statusesMutex *sync.RWMutex
	LocalPath     string
	RemotePath    string
	Sync          bool
	CodeStart     func()
	context.CancelFunc
	cancelMutex *sync.Mutex
	Params      interface{}
	Client      interface{}
	Config      filesSDK.Config
	EventsReporter
	directory.Type
	*manager.Manager
	RetryPolicy interface{}
	Ignore      *ignore.GitIgnore
	Include     *ignore.GitIgnore
	Started     *lib.Signal
	Finished    *lib.Signal
	Canceled    *lib.Signal
	Scanning    *lib.Signal
	EndScanning *lib.Signal
	retryablehttp.Logger
	RemoteFs fs.FS
	*lib.Meter
}

func (r *Job) Init() *Job {
	r.statusesMutex = &sync.RWMutex{}
	r.cancelMutex = &sync.Mutex{}
	r.Id = sid.IdBase64()
	r.EventsReporter = make(map[status.Status][]reporterSettings)
	r.Timer = timer.New()
	r.Started = (&lib.Signal{}).Init()
	r.Finished = (&lib.Signal{}).Init()
	r.Canceled = (&lib.Signal{}).Init()
	r.Scanning = (&lib.Signal{}).Init()
	r.EndScanning = (&lib.Signal{}).Init()
	r.Meter, _ = lib.NewMeter(time.Millisecond*250, time.Second*5)
	return r
}

func (r *Job) SetManager(m *manager.Manager) {
	if m == nil {
		r.Manager = manager.Default()
	} else {
		r.Manager = m
	}
}

func (r *Job) SetEventsReporter(e EventsReporter) {
	if len(e) > 0 {
		r.EventsReporter = e
	}
}

func (r *Job) ClearCalled() {
	r.Started.Clear()
	r.Finished.Clear()
	r.Canceled.Clear()
	r.Scanning.Clear()
	r.EndScanning.Clear()
}

func (r *Job) ClearStatuses() Job {
	newJob := *r
	newJob.Reset()
	newJob.Statuses = []IFile{}
	return newJob
}

func (r *Job) Scan() {
	r.Scanning.Call()
}

func (r *Job) EndScan() {
	r.EndScanning.Call()
}

func (r *Job) Start(ignoreCodeStart ...bool) {
	r.Timer.Start()
	r.Started.Call()
	r.Meter.Start(time.Now())
	if r.CodeStart != nil && len(ignoreCodeStart) == 0 {
		r.CodeStart()
	}
}

func (r *Job) Finish() {
	r.Meter.Close(time.Now())
	r.Finished.Call()
}

func (r *Job) Cancel() {
	r.Meter.Close(time.Now())
	r.Canceled.Call()
	r.cancelMutex.Lock()
	r.CancelFunc()
	r.cancelMutex.Unlock()
}

func (r *Job) Reset() {
	r.Timer = timer.New()
}

func (r *Job) Wait() {
	select {
	case <-r.Finished.C:
	case <-r.Canceled.C:
	}
}

func (r *Job) Job() *Job {
	return r
}

func (r *Job) WithContext(ctx context.Context) context.Context {
	jobCtx, cancel := context.WithCancel(ctx)
	r.cancelMutex.Lock()
	r.CancelFunc = cancel
	r.cancelMutex.Unlock()
	return jobCtx
}

func OnBytesChange(event status.GetStatus) status.GetStatus {
	return onBytesChange{GetStatus: event}
}

type onBytesChange struct {
	status.GetStatus
}

func CreateFileEvents(callback Reporter, events ...status.GetStatus) EventsReporter {
	eventsReporter := make(EventsReporter)
	for _, event := range events {
		s := event.Status()
		switch event.(type) {
		case onBytesChange:
			eventsReporter[s] = append(eventsReporter[s], reporterSettings{Reporter: callback, OnBytesChange: true})
		default:
			eventsReporter[s] = append(eventsReporter[s], reporterSettings{Reporter: callback})
		}
	}
	return eventsReporter
}

func (r *Job) RegisterFileEvent(callback Reporter, events ...status.GetStatus) {
	for k, v := range CreateFileEvents(callback, events...) {
		r.EventsReporter[k] = append(r.EventsReporter[k], v...)
	}
}

type UnwrappedError struct {
	error
	OriginalError error
}

func (r *Job) UpdateStatusWithBytes(status status.GetStatus, file IFile, bytesCount int64) {
	file.IncrementTransferBytes(bytesCount)

	if bytesCount > 0 {
		r.Meter.Record(time.Now(), uint64(bytesCount))
	}

	if status != file.Status() {
		r.UpdateStatus(status, file, nil)
	} else {
		callbacks, ok := r.EventsReporter[status.Status()]
		if ok {
			for _, callback := range callbacks {
				if callback.OnBytesChange {
					callback.Reporter(ToStatusFile(file))
				}
			}
		}
	}
}

func (r *Job) UpdateStatus(s status.GetStatus, file IFile, err error) {
	if s == file.Status() && err == nil && file.Err() == nil {
		return
	}
	if err != nil && strings.Contains(err.Error(), "context canceled") {
		err = nil
		s = status.Canceled
	}
	r.Logger.Printf(libLog.New(
		file.File().Path,
		map[string]interface{}{
			"status": s,
			"error":  err,
		}))
	file.SetStatus(s.Status(), err)
	r.statusCallbacks(s, file)
}

func (r *Job) statusCallbacks(status status.GetStatus, file IFile) {
	callbacks, ok := r.EventsReporter[status.Status()]
	if ok {
		for _, callback := range callbacks {
			callback.Reporter(ToStatusFile(file))
		}
	}
}

func (r *Job) Count(t ...status.GetStatus) int {
	r.statusesMutex.RLock()
	defer r.statusesMutex.RUnlock()
	if len(t) == 0 {
		return len(r.Statuses)
	}
	var total int
	for _, s := range r.Statuses {
		if s.Status().Any(t...) {
			total += 1
		}
	}
	return total
}

func (r *Job) Add(report IFile) {
	r.statusesMutex.Lock()
	if r.EndScanning.Called() {
		panic("adding new file after Scanning is complete")
	}
	r.Statuses = append(r.Statuses, report)
	r.statusesMutex.Unlock()
}

func (r *Job) TotalBytes(t ...status.GetStatus) int64 {
	var total int64
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		if s.Status().Any(t...) {
			total += s.Size()
		}
	}
	r.statusesMutex.RUnlock()
	return total
}

func (r *Job) RemainingBytes(t ...status.GetStatus) int64 {
	return r.TotalBytes(t...) - r.TransferBytes(t...)
}

func (r *Job) TransferBytes(t ...status.GetStatus) int64 {
	var transferBytes int64
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		if s.Status().Any(t...) {
			transferBytes += s.TransferBytes()
		}
	}
	r.statusesMutex.RUnlock()
	return transferBytes
}

func (r *Job) CountFunc(call func(IFile) bool, t ...status.GetStatus) int {
	var count int
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		if s.Status().Any(t...) && call(s) {
			count++
		}
	}
	r.statusesMutex.RUnlock()
	return count
}

func (r *Job) Idle() bool {
	ago := time.Second * 3
	now := r.Meter.BitRate(time.Now())
	since := r.Meter.BitRate(time.Now().Add(-ago))
	return now == 0 && since == 0
}

func (r *Job) TransferRate() int64 {
	return int64(r.Meter.BitRate(time.Now()) / 8)
}

func (r *Job) FinalTransferRate() int64 {
	_, _, b := r.Meter.Total(time.Now())
	return int64(b / 8)
}

func (r *Job) FilesRate() float64 {
	duration := time.Second * 3
	since := time.Now().Add(-duration)
	fileCount := r.CountFunc(func(file IFile) bool {
		return file.EndedAt().After(since)
	}, status.Complete)

	filesPerMilliseconds := float64(fileCount) / float64(duration.Milliseconds())
	filesPerSecond := filesPerMilliseconds * float64(1000)

	if filesPerSecond < 0 {
		return 0
	}

	return filesPerSecond
}

func (r *Job) ETA() time.Duration {
	transferRate := r.TransferRate()
	if transferRate == 0 {
		return 0
	}
	seconds := time.Duration(r.RemainingBytes() / transferRate)
	eta := seconds * time.Second
	if eta < 0 {
		return 0
	}
	return eta
}

func (r *Job) ElapsedTime() time.Duration {
	return r.Elapsed()
}

func (r *Job) All(t ...status.GetStatus) bool {
	allEnded := true
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		if !s.Status().Any(t...) {
			allEnded = false
			break
		}
	}
	r.statusesMutex.RUnlock()
	return allEnded
}

func (r *Job) Any(t ...status.GetStatus) (b bool) {
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		if s.Status().Any(t...) {
			b = true
			break
		}
	}
	r.statusesMutex.RUnlock()
	return
}

func (r *Job) Find(t status.GetStatus) (IFile, bool) {
	r.statusesMutex.RLock()
	defer r.statusesMutex.RUnlock()

	for _, s := range r.Statuses {
		if s.Status().Any(t) {
			return s, true
		}
	}

	return nil, false
}

func (r *Job) EnqueueNext() (f IFile, ok bool) {
	r.statusesMutex.Lock()
	defer func() {
		r.statusesMutex.Unlock()
		if f != nil {
			// Call statusCallbacks to run event callbacks, which needs to be done outside the mutex.
			r.statusCallbacks(status.Queued, f)
		}
	}()

	for _, s := range r.Statuses {
		if s.Status().Any(status.Indexed) {
			f = s
			ok = true
			// The status must be changed within the mutex in order that it's not reused.
			s.SetStatus(status.Queued, nil)
			break
		}
	}

	return
}

func (r *Job) Sub(t ...status.GetStatus) *Job {
	var sub []IFile
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		if s.Status().Any(t...) {
			sub = append(sub, s)
		}
	}
	r.statusesMutex.RUnlock()
	newJob := Job{Statuses: sub, statusesMutex: &sync.RWMutex{}}
	return &newJob
}

func (r *Job) Files() []filesSDK.File {
	var files []filesSDK.File
	r.statusesMutex.RLock()
	for _, s := range r.Statuses {
		files = append(files, s.File())
	}
	r.statusesMutex.RUnlock()
	return files
}

func (r *Job) Percentage(t ...status.GetStatus) int {
	p := int((float64(r.TransferBytes(t...)) / float64(r.TotalBytes(t...))) * float64(100))
	if p < 0 {
		return 0
	}
	return p
}

func (r *Job) StatusFromError(s IFile, err error) {
	if r.Canceled.Called() {
		r.UpdateStatus(status.Canceled, s, nil)
	} else {
		r.UpdateStatus(status.Errored, s, err)
	}
}

func (r *Job) FindRemoteFile(file IFile) (filesSDK.File, bool, error) {
	if r.Type == directory.File {
		entry, err := r.RemoteFs.Open(file.RemotePath())
		if err != nil {
			return filesSDK.File{}, false, err
		}
		info, err := entry.Stat()
		if err != nil {
			return filesSDK.File{}, false, err
		}

		return info.Sys().(filesSDK.File), true, nil
	} else {
		dir, _ := lib.UrlLastSegment(file.RemotePath())
		entries, err := fs.ReadDir(r.RemoteFs, lib.Path{Path: dir}.ConvertEmptyToRoot().String())
		if err != nil {
			return filesSDK.File{}, false, err
		}

		for _, entry := range entries {
			if !entry.IsDir() && lib.NormalizeForComparison(lib.UrlJoinNoEscape(dir, entry.Name())) == lib.NormalizeForComparison(file.RemotePath()) {
				info, err := entry.Info()
				if err != nil {
					panic(err)
					return filesSDK.File{}, false, err
				}

				return info.Sys().(filesSDK.File), true, nil
			}
		}
	}

	return filesSDK.File{}, false, nil
}

func WaitTellFinished[T any](job *Job, onStatusComplete chan T, beforeCallingFinish func()) {
	go func() {
		wait := waitForAndCount(job.EndScanning.C, onStatusComplete)
		n := len(job.Statuses) - wait
		for range iter.N(n) {
			<-onStatusComplete
		}
		close(onStatusComplete)
		if !job.Canceled.Called() {
			beforeCallingFinish()
		}
		job.Finish()
	}()
}

func waitForAndCount[T any, F any](wait chan T, onComplete chan F) int {
	completed := 0
	for {
		select {
		case <-wait:
			return completed
		case <-onComplete:
			completed += 1
		}
	}
}
