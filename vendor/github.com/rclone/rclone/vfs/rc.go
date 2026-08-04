package vfs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/cache"
	"github.com/rclone/rclone/fs/rc"
	"github.com/rclone/rclone/vfs/vfscache/writeback"
)

const getVFSHelp = ` 
This command takes an "fs" parameter. If this parameter is not
supplied and if there is only one VFS in use then that VFS will be
used. If there is more than one VFS in use then the "fs" parameter
must be supplied.`

// GetVFS gets a VFS with config name "fs" from the cache or returns an error.
//
// If "fs" is not set and there is one and only one VFS in the active
// cache then it returns it. This is for backwards compatibility.
//
// This deletes the "fs" parameter from in if it is valid
func getVFS(in rc.Params) (vfs *VFS, err error) {
	fsString, err := in.GetString("fs")
	if rc.IsErrParamNotFound(err) {
		var count int
		vfs, count = activeCacheEntries()
		if count == 1 {
			return vfs, nil
		} else if count == 0 {
			return nil, errors.New(`no VFS active and "fs" parameter not supplied`)
		}
		return nil, errors.New(`more than one VFS active - need "fs" parameter`)
	} else if err != nil {
		return nil, err
	}
	activeMu.Lock()
	defer activeMu.Unlock()
	fsString = cache.Canonicalize(fsString)
	activeVFS := active[fsString]
	if len(activeVFS) == 0 {
		return nil, fmt.Errorf("no VFS found with name %q", fsString)
	} else if len(activeVFS) > 1 {
		return nil, fmt.Errorf("more than one VFS active with name %q", fsString)
	}
	delete(in, "fs") // delete the fs parameter
	return activeVFS[0], nil
}

// getVFSes gets all VFSes with config name "fs" from the cache, or returns an
// error.
//
// Unlike getVFS it does NOT error when more than one VFS shares the same name.
// The CSI driver disables VFS reuse (issue #69), so an RWX volume mounted into
// several pods on one node creates one VFS per mount for the same remote.
// Directory-cache operations like vfs/refresh and vfs/forget must therefore
// apply to every matching VFS, otherwise stale caches remain on the other
// mounts (and getVFS would reject the call outright).
//
// If "fs" is not set, all active VFSes are returned. This deletes the "fs"
// parameter from in if it is valid.
// See: https://github.com/veloxpack/csi-driver-rclone/issues/88
func getVFSes(in rc.Params) (vfses []*VFS, err error) {
	fsString, err := in.GetString("fs")
	if rc.IsErrParamNotFound(err) {
		activeMu.Lock()
		defer activeMu.Unlock()
		for _, activeVFS := range active {
			vfses = append(vfses, activeVFS...)
		}
		if len(vfses) == 0 {
			return nil, errors.New(`no VFS active and "fs" parameter not supplied`)
		}
		return vfses, nil
	} else if err != nil {
		return nil, err
	}
	activeMu.Lock()
	defer activeMu.Unlock()
	fsString = cache.Canonicalize(fsString)
	vfses = active[fsString]
	if len(vfses) == 0 {
		return nil, fmt.Errorf("no VFS found with name %q", fsString)
	}
	delete(in, "fs") // delete the fs parameter
	return vfses, nil
}

func init() {
	rc.Add(rc.Call{
		Path:  "vfs/refresh",
		Fn:    rcRefresh,
		Title: "Refresh the directory cache.",
		Help: `
This reads the directories for the specified paths and freshens the
directory cache.

If no paths are passed in then it will refresh the root directory.

    rclone rc vfs/refresh

Otherwise pass directories in as dir=path. Any parameter key
starting with dir will refresh that directory, e.g.

    rclone rc vfs/refresh dir=home/junk dir2=data/misc

If the parameter recursive=true is given the whole directory tree
will get refreshed. This refresh will use --fast-list if enabled.
` + getVFSHelp,
	})
}

// mergeResult records status for path in result, preferring an error status
// over "OK" so that a failure on any one VFS is not masked by success on another.
func mergeResult(result map[string]string, path, status string) {
	if prev, ok := result[path]; ok && prev != "OK" {
		return // an error was already recorded for this path; keep it
	}
	result[path] = status
}

// refreshVFS applies vfs/refresh to a single VFS, merging its per-path status
// into result. Parameters in `in` are assumed to have been validated already.
func refreshVFS(vfs *VFS, in rc.Params, recursive bool, result map[string]string) error {
	root, err := vfs.Root()
	if err != nil {
		return err
	}
	getDir := func(path string) (*Dir, error) {
		var err error
		path = strings.Trim(path, "/")
		segments := strings.Split(path, "/")
		var node Node = root
		for _, s := range segments {
			if dir, ok := node.(*Dir); ok {
				node, err = dir.stat(s)
				if err != nil {
					return nil, err
				}
			}
		}
		if dir, ok := node.(*Dir); ok {
			return dir, nil
		}
		return nil, EINVAL
	}

	if len(in) == 0 {
		if recursive {
			err = root.readDirTree()
		} else {
			err = root.readDir()
		}
		if err != nil {
			mergeResult(result, "", err.Error())
		} else {
			mergeResult(result, "", "OK")
		}
		return nil
	}
	for _, v := range in {
		path := v.(string) // validated by caller
		dir, err := getDir(path)
		if err != nil {
			mergeResult(result, path, err.Error())
			continue
		}
		if recursive {
			err = dir.readDirTree()
		} else {
			err = dir.readDir()
		}
		if err != nil {
			mergeResult(result, path, err.Error())
		} else {
			mergeResult(result, path, "OK")
		}
	}
	return nil
}

func rcRefresh(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	vfses, err := getVFSes(in)
	if err != nil {
		return nil, err
	}

	recursive := false
	{
		const k = "recursive"

		if v, ok := in[k]; ok {
			s, ok := v.(string)
			if !ok {
				return out, fmt.Errorf("value must be string %q=%v", k, v)
			}
			recursive, err = strconv.ParseBool(s)
			if err != nil {
				return out, fmt.Errorf("invalid value %q=%v", k, v)
			}
			delete(in, k)
		}
	}

	// Validate the remaining dir* parameters up front so the call fails
	// consistently before touching any VFS.
	for k, v := range in {
		if _, ok := v.(string); !ok {
			return out, fmt.Errorf("value must be string %q=%v", k, v)
		}
		if !strings.HasPrefix(k, "dir") {
			return out, fmt.Errorf("unknown key %q", k)
		}
	}

	result := map[string]string{}
	for _, vfs := range vfses {
		if err := refreshVFS(vfs, in, recursive, result); err != nil {
			return out, err
		}
	}
	out = rc.Params{
		"result": result,
	}
	return out, nil
}

// Add remote control for the VFS
func init() {
	rc.Add(rc.Call{
		Path:  "vfs/forget",
		Fn:    rcForget,
		Title: "Forget files or directories in the directory cache.",
		Help: `
This forgets the paths in the directory cache causing them to be
re-read from the remote when needed.

If no paths are passed in then it will forget all the paths in the
directory cache.

    rclone rc vfs/forget

Otherwise pass files or dirs in as file=path or dir=path.  Any
parameter key starting with file will forget that file and any
starting with dir will forget that dir, e.g.

    rclone rc vfs/forget file=hello file2=goodbye dir=home/junk
` + getVFSHelp,
	})
}

func rcForget(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	vfses, err := getVFSes(in)
	if err != nil {
		return nil, err
	}

	// Validate the parameters up front so the call fails consistently before
	// touching any VFS.
	for k, v := range in {
		if _, ok := v.(string); !ok {
			return out, fmt.Errorf("value must be string %q=%v", k, v)
		}
		if !strings.HasPrefix(k, "file") && !strings.HasPrefix(k, "dir") {
			return out, fmt.Errorf("unknown key %q", k)
		}
	}

	forgotten := []string{}
	seen := map[string]bool{}
	addForgotten := func(path string) {
		if !seen[path] {
			seen[path] = true
			forgotten = append(forgotten, path)
		}
	}
	for _, vfs := range vfses {
		root, err := vfs.Root()
		if err != nil {
			return nil, err
		}
		if len(in) == 0 {
			root.ForgetAll()
			continue
		}
		for k, v := range in {
			path := strings.Trim(v.(string), "/") // validated above
			if strings.HasPrefix(k, "file") {
				root.ForgetPath(path, fs.EntryObject)
			} else {
				root.ForgetPath(path, fs.EntryDirectory)
			}
			addForgotten(path)
		}
	}
	out = rc.Params{
		"forgotten": forgotten,
	}
	return out, nil
}

func getDuration(k string, v any) (time.Duration, error) {
	s, ok := v.(string)
	if !ok {
		return 0, fmt.Errorf("value must be string %q=%v", k, v)
	}
	interval, err := fs.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return interval, nil
}

func getInterval(in rc.Params) (time.Duration, bool, error) {
	k := "interval"
	v, ok := in[k]
	if !ok {
		return 0, false, nil
	}
	interval, err := getDuration(k, v)
	if err != nil {
		return 0, true, err
	}
	if interval < 0 {
		return 0, true, errors.New("interval must be >= 0")
	}
	delete(in, k)
	return interval, true, nil
}

func getTimeout(in rc.Params) (time.Duration, error) {
	k := "timeout"
	v, ok := in[k]
	if !ok {
		return 10 * time.Second, nil
	}
	timeout, err := getDuration(k, v)
	if err != nil {
		return 0, err
	}
	delete(in, k)
	return timeout, nil
}

func getStatus(vfs *VFS, in rc.Params) (out rc.Params, err error) {
	for k, v := range in {
		return nil, fmt.Errorf("invalid parameter: %s=%s", k, v)
	}
	return rc.Params{
		"enabled":   vfs.Opt.PollInterval != 0,
		"supported": vfs.pollChan != nil,
		"interval": map[string]any{
			"raw":     vfs.Opt.PollInterval,
			"seconds": time.Duration(vfs.Opt.PollInterval) / time.Second,
			"string":  vfs.Opt.PollInterval.String(),
		},
	}, nil
}

func init() {
	rc.Add(rc.Call{
		Path:  "vfs/poll-interval",
		Fn:    rcPollInterval,
		Title: "Get the status or update the value of the poll-interval option.",
		Help: `
Without any parameter given this returns the current status of the
poll-interval setting.

When the interval=duration parameter is set, the poll-interval value
is updated and the polling function is notified.
Setting interval=0 disables poll-interval.

    rclone rc vfs/poll-interval interval=5m

The timeout=duration parameter can be used to specify a time to wait
for the current poll function to apply the new value.
If timeout is less or equal 0, which is the default, wait indefinitely.

The new poll-interval value will only be active when the timeout is
not reached.

If poll-interval is updated or disabled temporarily, some changes
might not get picked up by the polling function, depending on the
used remote.
` + getVFSHelp,
	})
}

func rcPollInterval(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	vfs, err := getVFS(in)
	if err != nil {
		return nil, err
	}

	interval, intervalPresent, err := getInterval(in)
	if err != nil {
		return nil, err
	}
	timeout, err := getTimeout(in)
	if err != nil {
		return nil, err
	}
	for k, v := range in {
		return nil, fmt.Errorf("invalid parameter: %s=%s", k, v)
	}
	if vfs.pollChan == nil {
		return nil, errors.New("poll-interval is not supported by this remote")
	}

	if !intervalPresent {
		return getStatus(vfs, in)
	}
	var timeoutHit bool
	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timeoutChan = timer.C
	}
	select {
	case vfs.pollChan <- interval:
		vfs.Opt.PollInterval = fs.Duration(interval)
	case <-timeoutChan:
		timeoutHit = true
	}
	out, err = getStatus(vfs, in)
	if out != nil {
		out["timeout"] = timeoutHit
	}
	return
}

func init() {
	rc.Add(rc.Call{
		Path:   "vfs/list",
		NoAuth: true,
		Title:  "List active VFSes.",
		Help: `
This lists the active VFSes.

It returns a list under the key "vfses" where the values are the VFS
names that could be passed to the other VFS commands in the "fs"
parameter.`,
		Fn: rcList,
	})
}

func rcList(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	activeMu.Lock()
	defer activeMu.Unlock()
	var names = []string{}
	for name, vfses := range active {
		if len(vfses) == 1 {
			names = append(names, name)
		} else {
			for i := range vfses {
				names = append(names, fmt.Sprintf("%s[%d]", name, i))
			}
		}
	}
	out = rc.Params{}
	out["vfses"] = names
	return out, nil
}

func init() {
	rc.Add(rc.Call{
		Path:   "vfs/stats",
		NoAuth: true,
		Title:  "Stats for a VFS.",
		Help: `
This returns stats for the selected VFS.

    {
        // Status of the disk cache - only present if --vfs-cache-mode > off
        "diskCache": {
            "bytesUsed": 0,
            "erroredFiles": 0,
            "files": 0,
            "hashType": 1,
            "outOfSpace": false,
            "path": "/home/user/.cache/rclone/vfs/local/mnt/a",
            "pathMeta": "/home/user/.cache/rclone/vfsMeta/local/mnt/a",
            "uploadsInProgress": 0,
            "uploadsQueued": 0
        },
        "fs": "/mnt/a",
        "inUse": 1,
        // Status of the in memory metadata cache
        "metadataCache": {
            "dirs": 1,
            "files": 0
        },
        // Options as returned by options/get
        "opt": {
            "CacheMaxAge": 3600000000000,
            // ...
            "WriteWait": 1000000000
        }
    }

` + getVFSHelp,
		Fn: rcStats,
	})
}

func rcStats(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	vfs, err := getVFS(in)
	if err != nil {
		return nil, err
	}
	return vfs.Stats(), nil
}

func init() {
	rc.Add(rc.Call{
		Path:   "vfs/queue",
		NoAuth: true,
		Title:  "Queue info for a VFS.",
		Help: strings.ReplaceAll(`
This returns info about the upload queue for the selected VFS.

This is only useful if |--vfs-cache-mode| > off. If you call it when
the |--vfs-cache-mode| is off, it will return an empty result.

    {
        "queue": // an array of files queued for upload
        [
            {
                "name":      "file",   // string: name (full path) of the file,
                "id":        123,      // integer: id of this item in the queue,
                "size":      79,       // integer: size of the file in bytes
                "expiry":    1.5       // float: time until file is eligible for transfer, lowest goes first
                "tries":     1,        // integer: number of times we have tried to upload
                "delay":     5.0,      // float: seconds between upload attempts
                "uploading": false,    // boolean: true if item is being uploaded
            },
       ],
    }

The |expiry| time is the time until the file is eligible for being
uploaded in floating point seconds. This may go negative. As rclone
only transfers |--transfers| files at once, only the lowest
|--transfers| expiry times will have |uploading| as |true|. So there
may be files with negative expiry times for which |uploading| is
|false|.

`, "|", "`") + getVFSHelp,
		Fn: rcQueue,
	})
}

func rcQueue(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	vfs, err := getVFS(in)
	if err != nil {
		return nil, err
	}
	if vfs.cache == nil {
		return nil, nil
	}
	return vfs.cache.Queue(), nil
}

func init() {
	rc.Add(rc.Call{
		Path:  "vfs/queue-set-expiry",
		Title: "Set the expiry time for an item queued for upload.",
		Help: strings.ReplaceAll(`

Use this to adjust the |expiry| time for an item in the upload queue.
You will need to read the |id| of the item using |vfs/queue| before
using this call.

You can then set |expiry| to a floating point number of seconds from
now when the item is eligible for upload. If you want the item to be
uploaded as soon as possible then set it to a large negative number (eg
-1000000000). If you want the upload of the item to be delayed
for a long time then set it to a large positive number.

Setting the |expiry| of an item which has already has started uploading
will have no effect - the item will carry on being uploaded.

This will return an error if called with |--vfs-cache-mode| off or if
the |id| passed is not found.

This takes the following parameters

- |fs| - select the VFS in use (optional)
- |id| - a numeric ID as returned from |vfs/queue|
- |expiry| - a new expiry time as floating point seconds
- |relative| - if set, expiry is to be treated as relative to the current expiry (optional, boolean)

This returns an empty result on success, or an error.

`, "|", "`") + getVFSHelp,
		Fn: rcQueueSetExpiry,
	})
}

func rcQueueSetExpiry(ctx context.Context, in rc.Params) (out rc.Params, err error) {
	vfs, err := getVFS(in)
	if err != nil {
		return nil, err
	}
	if vfs.cache == nil {
		return nil, rc.NewErrParamInvalid(errors.New("can't call this unless using the VFS cache"))
	}

	// Read input values
	id, err := in.GetInt64("id")
	if err != nil {
		return nil, err
	}
	expiry, err := in.GetFloat64("expiry")
	if err != nil {
		return nil, err
	}
	relative, err := in.GetBool("relative")
	if err != nil && !rc.IsErrParamNotFound(err) {
		return nil, err
	}

	// Set expiry
	var refTime time.Time
	if !relative {
		refTime = time.Now()
	}
	err = vfs.cache.QueueSetExpiry(writeback.Handle(id), refTime, time.Duration(float64(time.Second)*expiry))
	return nil, err
}
