package lib

import (
	"context"
	"errors"
	"io/fs"
)

type Walk[T any] struct {
	fs.FS
	Queue[string]
	IterChan[T]
	ConcurrencyManager ConcurrencyManagerWithSubWorker
	Root               string
	WalkFile           func(d fs.DirEntry, path string, err error) (T, error)
	ListDirectories    bool
}

type DirEntry struct {
	fs.DirEntry
	fs.FileInfo
	path string
	error
}

func (d DirEntry) Err() error {
	return d.error
}

func (d DirEntry) Path() string {
	return d.path
}

func (w *Walk[T]) Walk(parentCtx context.Context) *IterChan[T] {
	w.Queue.Init(1)
	it := (&IterChan[T]{}).Init(parentCtx)
	if w.Root == "" {
		w.Queue.Push(".")
	} else {
		w.Queue.Push(w.Root)
	}

	waitGroup := w.ConcurrencyManager.NewSubWorker()

	go func() {
		defer it.Stop()
		for {
			if it.Context.Err() != nil {
				break
			}
			if dir := w.Queue.Pop(); dir != "" {
				if waitGroup.WaitWithContext(parentCtx) {
					go func() {
						err := w.walkDir(it.Context, dir, it)
						if err != nil && !errors.Is(err, context.Canceled) {
							it.SendError <- err
						}
						waitGroup.Done()
					}()
				}
			}

			if w.Len() == 0 {
				for {
					if waitGroup.WaitForADone() {
						if w.Len() != 0 {
							break
						}
					} else {
						return
					}
				}
			}
		}
	}()
	return it
}

func (w *Walk[T]) walkDir(ctx context.Context, dir string, it *IterChan[T]) error {
	return fs.WalkDir(w.FS, dir, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			if err := w.send(ctx, d, path, it, err); err != nil {
				return err
			}
			return fs.SkipDir
		}

		if NormalizeForComparison(path) == NormalizeForComparison(dir) && d.IsDir() {
			if NormalizeForComparison(path) == NormalizeForComparison(w.Root) && w.ListDirectories {
				if err := w.send(ctx, d, path, it, nil); err != nil {
					return err
				}
			}
			return nil
		}

		if d.IsDir() && path != "." {
			if w.ListDirectories {
				if err := w.send(ctx, d, path, it, nil); err != nil {
					return err
				}
			}
			w.Queue.Push(path)
			return fs.SkipDir
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if err := w.send(ctx, d, path, it, nil); err != nil {
			return err
		}

		return nil
	})
}

func (w *Walk[T]) send(ctx context.Context, d fs.DirEntry, path string, it *IterChan[T], err error) error {
	toSend, err := w.WalkFile(d, path, err)
	if err == nil {
		select {
		case <-ctx.Done():
		default:
			it.Send <- toSend
		}
		return nil
	} else {
		return err
	}
}

func DirEntryWalkFile(d fs.DirEntry, path string, err error) (DirEntry, error) {
	if err != nil {
		return DirEntry{DirEntry: d, FileInfo: nil, path: path, error: err}, nil
	}
	info, err := d.Info()
	return DirEntry{DirEntry: d, FileInfo: info, path: path, error: err}, nil
}
