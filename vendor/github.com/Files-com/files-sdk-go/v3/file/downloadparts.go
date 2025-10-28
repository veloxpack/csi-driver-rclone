package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/file/manager"
	"github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/panjf2000/ants/v2"
	"github.com/samber/lo"
)

func init() {
	ants.Release()
}

const (
	DownloadPartChunkSize = int64(1024 * 1024 * 5)
	DownloadPartLimit     = 15
)

type DownloadParts struct {
	globalWait manager.ConcurrencyManager
	context.CancelFunc
	context.Context
	queueCancel  context.CancelFunc
	queueContext context.Context
	fs.File
	fs.FileInfo
	lib.WriterAndAt
	totalWritten  int64
	parts         []*Part
	queue         chan *Part
	finishedParts chan *Part
	CloseError    error
	files_sdk.Config
	fileManager *ants.Pool
	*sync.RWMutex
	queueLock      *sync.Mutex
	partsCompleted uint32
	path           string
}

func (d *DownloadParts) Init(file fs.File, info fs.FileInfo, globalWait manager.ConcurrencyManager, writer lib.WriterAndAt, config files_sdk.Config) *DownloadParts {
	d.File = file
	d.FileInfo = info
	d.path = info.Name()
	d.globalWait = globalWait
	d.WriterAndAt = writer
	d.Config = config
	d.RWMutex = &sync.RWMutex{}
	d.queueLock = &sync.Mutex{}
	return d
}

func (d *DownloadParts) Run(ctx context.Context) error {
	d.Context, d.CancelFunc = context.WithCancel(ctx)
	d.queueContext, d.queueCancel = context.WithCancel(d.Context)
	defer func() {
		d.Config.LogPath(
			d.path,
			map[string]interface{}{
				"message":  "Finished canceling context and closing file",
				"realSize": atomic.LoadInt64(&d.totalWritten),
			},
		)
		d.CancelFunc()
		d.CloseError = d.WriterAndAt.Close()
		d.fileManager.Release()
	}()
	var err error
	d.fileManager, err = ants.NewPool(lo.Min[int](append([]int{}, DownloadPartLimit, d.globalWait.Max())))
	if err != nil {
		return err
	}
	if d.downloadFileCutOff() {
		return d.downloadFile()
	} else {
		d.buildParts()
		d.listenOnQueue()
		d.addPartsToQueue()
		return d.waitForParts()
	}
}

func (d *DownloadParts) downloadFileCutOff() bool {
	// Don't break up file if running part serially.
	if d.fileManager.Cap() == 1 || d.globalWait.DownloadFilesAsSingleStream {
		return true
	}

	return d.FileInfo.Size() <= DownloadPartChunkSize*2
}

func (d *DownloadParts) FinalSize() int64 {
	return atomic.LoadInt64(&d.totalWritten)
}

func (d *DownloadParts) waitForParts() error {
	var err error
	for i := range d.parts {
		part := <-d.finishedParts
		if part.Err() != nil && !errors.Is(part.Err(), context.Canceled) {
			err = part.Err()
		}
		atomic.AddUint32(&d.partsCompleted, 1)
		d.Config.LogPath(
			d.path,
			map[string]interface{}{
				"RunningParts":  d.fileManager.Running(),
				"limit":         d.fileManager.Cap(),
				"parts":         len(d.parts),
				"Written":       atomic.LoadInt64(&d.totalWritten),
				"PartFinished":  part.number,
				"partBytes":     part.bytes,
				"PartsFinished": i + 1,
				"error":         part.Err(),
			},
		)
	}
	close(d.queue)
	if err != nil {
		return err
	}
	return d.realSizeOverLap()
}

func (d *DownloadParts) realSizeOverLap() error {
	lastPart := d.parts[len(d.parts)-1]
	d.Config.LogPath(
		d.path,
		map[string]interface{}{
			"message":  "starting realSizeOverLap",
			"size":     d.FileInfo.Size(),
			"realSize": atomic.LoadInt64(&d.totalWritten),
		},
	)
	defer func() {
		d.Config.LogPath(
			d.path,
			map[string]interface{}{
				"message":  "finishing realSizeOverLap",
				"size":     d.FileInfo.Size(),
				"realSize": atomic.LoadInt64(&d.totalWritten),
			},
		)
	}()
	for {
		if d.FileInfo.(UntrustedSize).UntrustedSize() && d.queueContext.Err() == nil && lastPart.bytes == lastPart.len {
			d.queueLock.Lock()
			d.queue = make(chan *Part, 1)
			d.queueLock.Unlock()
			d.finishedParts = make(chan *Part, 1)
			nextPart := &Part{number: lastPart.number + 1, OffSet: OffSet{off: lastPart.off + lastPart.bytes, len: DownloadPartChunkSize}}
			d.Config.LogPath(d.path, map[string]interface{}{"message": "Next Part for size guess", "part": nextPart.number})
			d.parts = append(d.parts, nextPart)

			go d.processPart(nextPart.Start(d.Context), true)
			select {
			case lastPart = <-d.finishedParts:
				if lastPart.error != nil {
					if lastPart.error == io.EOF || errors.Is(lastPart.error, UntrustedSizeRangeRequestSizeSentReceived) {
						return nil
					}
					return lastPart.error
				}
			case lastPart = <-d.queue:
				if lastPart.error != nil {
					if lastPart.error == io.EOF {
						return nil
					}
					return lastPart.error
				}
			}
		} else {
			if d.FileInfo.Size() != atomic.LoadInt64(&d.totalWritten) && !d.FileInfo.(UntrustedSize).UntrustedSize() {
				return fmt.Errorf("server reported size does not match downloaded file. - expected: %v, actual: %v", d.FileInfo.Size(), atomic.LoadInt64(&d.totalWritten))
			}
			return nil
		}
	}
}

func (d *DownloadParts) addPartsToQueue() {
	for _, part := range d.parts {
		d.queue <- part
	}
}

func (d *DownloadParts) listenOnQueue() {
	go func() {
		d.queueLock.Lock()
		defer d.queueLock.Unlock()
		for {
			select {
			case part := <-d.queue:
				if part == nil {
					return
				}
				if d.queueContext.Err() != nil {
					d.finishedParts <- part
					continue
				}
				if part.processing {
					panic(part)
				}
				if len(part.requests) > 3 {
					d.Config.LogPath(d.path, map[string]interface{}{"message": "Maxed out reties", "part": part.number})
					d.finishedParts <- part
				} else {
					if part.Context.Err() != nil {
						d.finishedParts <- part.Done()
						continue
					}
					part.Clear()
					d.globalWait.Wait()
					d.fileManager.Submit(func() {
						d.stateLog()
						d.processPart(part.Start(), false)
						d.globalWait.Done()
					})

					d.slowDownTellFirstPart(part)
				}
			}
		}
	}()
}

func (d *DownloadParts) slowDownTellFirstPart(part *Part) {
	// One request needs to return the header for MaxConnections.
	// Once there finishedParts can be tuned to that value. So slow down to give time to get that value.
	if atomic.LoadUint32(&d.partsCompleted) != 0 || d.parts[0].Err() != nil {
		return
	}
	startTime := time.Now()
	timeout := startTime.Add(time.Duration((part.number)*250) * time.Millisecond)
	ctx, cancel := context.WithDeadline(d.Context, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			d.Config.LogPath(d.path, map[string]interface{}{"message": fmt.Sprintf("Part1 to Finish: stopped waited %v after part %v", time.Now().Sub(startTime).Truncate(time.Microsecond), part.number)})
			return
		default:
			if atomic.LoadUint32(&d.partsCompleted) != 0 || d.parts[0].Err() != nil {
				d.Config.LogPath(d.path, map[string]interface{}{"message": fmt.Sprintf("Part1 to Finish: finish after waiting %v after part %v", time.Now().Sub(startTime).Truncate(time.Microsecond), part.number)})
				cancel()
				return
			}
		}
	}
}

func (d *DownloadParts) stateLog(extraState ...map[string]interface{}) {
	d.Config.LogPath(
		d.path,
		lo.Assign[string, interface{}](append(extraState, d.state())...),
	)
}

func (d *DownloadParts) state() map[string]interface{} {
	return map[string]interface{}{
		"RunningParts": d.fileManager.Running(),
		"limit":        d.fileManager.Cap(),
		"parts":        len(d.parts),
		"written":      atomic.LoadInt64(&d.totalWritten),
		"completed":    atomic.LoadUint32(&d.partsCompleted),
	}
}

func (d *DownloadParts) buildParts() {
	size := d.FileInfo.Size()
	iter := (ByteOffset{PartSizes: lib.PartSizes}).BySize(&size)

	for {
		offset, next, i := iter()
		d.parts = append(d.parts, (&Part{OffSet: offset, number: i + 1}).WithContext(d.Context))
		if next == nil {
			break
		}
		iter = next
	}

	d.finishedParts = make(chan *Part, len(d.parts))
	d.queue = make(chan *Part, len(d.parts))
	d.stateLog()
}

func (d *DownloadParts) processPart(part *Part, UnexpectedEOF bool) {
	d.processRanger(part, d.File.(ReaderRange), UnexpectedEOF)
}

func (d *DownloadParts) processRanger(part *Part, ranger ReaderRange, UnexpectedEOF bool) {
	withContext, ok := ranger.(lib.FileWithContext)
	if ok {
		partCtx, partCancel := context.WithCancel(part.Context)
		defer partCancel()
		ranger = withContext.WithContext(partCtx).(ReaderRange)
	}
	r, err := ranger.ReaderRange(part.off, part.len+part.off-1)
	if d.requeueOnError(part, err, UnexpectedEOF) {
		return
	}
	if f, ok := ranger.(*File); ok {
		if f.MaxConnections != 0 && d.fileManager.Cap() > f.MaxConnections {
			d.fileManager.Tune(f.MaxConnections)
			d.stateLog(map[string]interface{}{"message": "tuning pool", "cap": d.fileManager.Cap()})
		}
	}
	info, _ := ranger.Stat()
	sizeTrustInfo, ok := info.(UntrustedSize)
	if ok && sizeTrustInfo.SizeTrust() != NullSizeTrust {
		d.RWMutex.Lock()
		d.FileInfo = sizeTrustInfo
		d.RWMutex.Unlock()
	}

	wn, err := lib.CopyAt(d.WriterAndAt, part.off, r)
	part.bytes = wn

	part.SetError(r.Close())
	if sizeTrustInfo.UntrustedSize() && part.Err() != nil {
		d.verifySizeAndUpdateParts(part)
	}

	if d.requeueOnError(part, err, UnexpectedEOF) {
		return
	}

	atomic.AddInt64(&d.totalWritten, wn)
	d.finishedParts <- part.Done()
}

func (d *DownloadParts) verifySizeAndUpdateParts(part *Part) {
	if errors.Is(part.Err(), UntrustedSizeRangeRequestSizeSentLessThanExpected) {
		d.Config.LogPath(
			d.path,
			map[string]interface{}{"error": part.Err(), "part": part.number},
		)
		d.queueCancel()
		// cancelAll greater parts
		for _, p := range d.parts[part.number:] {
			d.Config.LogPath(
				d.path,
				map[string]interface{}{"message": "canceling invalid part", "part": p.number},
			)
			p.CancelFunc()
		}
		part.SetError(nil)
	}
}

func (d *DownloadParts) requeueOnError(part *Part, err error, UnexpectedEOF bool) bool {
	for _, err := range []error{err, part.Err()} {
		if err != nil && !errors.Is(err, io.EOF) {
			if strings.Contains(err.Error(), "stream error") {
				return false
			}
			if UnexpectedEOF && errors.Is(err, io.ErrUnexpectedEOF) {
				return false
			}
			if part.error == nil {
				part.SetError(err)
			}
			d.Config.LogPath(
				d.path,
				map[string]interface{}{"message": "requeuing", "error": part.Err(), "part": part.number},
			)
			progressWriter, ok := d.WriterAndAt.(lib.ProgressWriter)
			if ok {
				progressWriter.ProgressWatcher(-part.bytes)
			}
			d.queue <- part.Done() // either timeout or stream error try part again.
			return true
		}
	}

	return false
}

func (d *DownloadParts) downloadFile() error {
	withContext, ok := d.File.(lib.FileWithContext)
	if ok {
		d.File = withContext.WithContext(d.Context)
	}
	n, err := io.Copy(d.WriterAndAt, d.File)
	if n == 0 {
		d.WriterAndAt.Write([]byte{})
	}
	atomic.AddInt64(&d.totalWritten, n)
	if err != nil {
		return err
	}
	err = d.File.Close()
	if err != nil {
		return err
	}

	info, _ := d.File.Stat()
	sizeTrustInfo, ok := info.(UntrustedSize)
	if ok && sizeTrustInfo.SizeTrust() != NullSizeTrust {
		d.FileInfo = sizeTrustInfo
	}

	if d.FileInfo.Size() != atomic.LoadInt64(&d.totalWritten) {
		return fmt.Errorf("server reported size does not match downloaded file. - expected: %v, actual: %v", d.FileInfo.Size(), atomic.LoadInt64(&d.totalWritten))
	}
	return nil
}
