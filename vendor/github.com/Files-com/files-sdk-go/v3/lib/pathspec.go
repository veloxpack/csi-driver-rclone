package lib

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PathSpecEntry struct {
	Dir         bool
	path        string
	Preexisting bool
	Mtime       time.Time
}
type PathSpecArgs struct {
	Src           string
	Dest          string
	PreserveTimes bool
}

type PathSpecTest struct {
	Name string
	Args PathSpecArgs
	Dest []PathSpecEntry
	Src  []PathSpecEntry
}

func BuildPathSpecTest(t *testing.T, mutex *sync.Mutex, tt PathSpecTest, sourceFs ReadWriteFs, destinationFs ReadWriteFs, cmdBuilder func(PathSpecArgs) Cmd) {
	t.Helper()
	t.Log(tt.Name)

	sourceTmpDir := sourceFs.PathJoin(sourceFs.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))
	err := sourceFs.MkdirAll(sourceTmpDir, 0750)
	assert.NoError(t, err)
	sourceRoot, err := sourceFs.MkdirTemp(sourceTmpDir, "src-root")
	assert.NoError(t, err)

	destTmpDir := destinationFs.PathJoin(destinationFs.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))
	err = destinationFs.MkdirAll(destTmpDir, 0750)
	assert.NoError(t, err)
	destRoot, err := destinationFs.MkdirTemp(destTmpDir, "dest-root")
	assert.NoError(t, err)

	for _, e := range tt.Src {
		if e.Dir {
			require.NoError(t, sourceFs.MkdirAll(sourceFs.PathJoin(sourceRoot, e.path), 0750))
		} else {
			var f io.WriteCloser
			f, err = sourceFs.Create(sourceFs.PathJoin(sourceRoot, e.path))
			require.NoError(t, err)
			require.NoError(t, f.Close())
			if tt.Args.PreserveTimes && !e.Mtime.IsZero() {
				require.NoError(t, sourceFs.Chtimes(sourceFs.PathJoin(sourceRoot, e.path), e.Mtime, e.Mtime))
			}
		}
	}
	for _, e := range tt.Dest {
		if !e.Preexisting {
			continue
		}
		if e.Dir {
			require.NoError(t, destinationFs.MkdirAll(destinationFs.PathJoin(destRoot, e.path), 0750))
		} else {
			var f io.WriteCloser
			f, err = destinationFs.Create(destinationFs.PathJoin(destRoot, e.path))
			require.NoError(t, err)
			require.NoError(t, f.Close())
		}
	}

	source := join(sourceFs.PathSeparator(), sourceRoot, tt.Args.Src)
	destination := join(destinationFs.PathSeparator(), destRoot, tt.Args.Dest)

	restoreDirSource, sourceDirChanged, err := ChangeDir(sourceFs, tt.Args.Src, source, mutex)
	require.NoError(t, err)

	restoreDirDest, destDirChanged, err := ChangeDir(destinationFs, tt.Args.Dest, destination, mutex)
	require.NoError(t, err)

	if tt.Args.Dest == "" && destDirChanged {
		destination = ""
	}

	relativePathCommand := cmdBuilder(PathSpecArgs{Src: tt.Args.Src, Dest: tt.Args.Dest, PreserveTimes: tt.Args.PreserveTimes})
	fullPathCommand := cmdBuilder(PathSpecArgs{Src: source, Dest: destination, PreserveTimes: tt.Args.PreserveTimes})

	if _, ok := os.LookupEnv("FULL_PATHS"); ok {
		t.Log(fullPathCommand.Args())
	}
	t.Log(relativePathCommand.Args())
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	fullPathCommand.SetErr(stderr)
	fullPathCommand.SetOut(stdout)
	err = fullPathCommand.Run()
	restoreDirSource()
	restoreDirDest()
	assert.NoError(t, err)
	if stderr.String() != "" {
		t.Log(strings.TrimSuffix(stderr.String(), "\n"))
	}
	t.Log(strings.TrimSuffix(stdout.String(), "\n"))
	sourceFile, err := sourceFs.Open(source)
	require.NoError(t, err)
	sourceStat, err := sourceFile.Stat()
	require.NoError(t, err)
	var sourceDir string
	if sourceStat.IsDir() {
		sourceDir = source
	} else {
		sourceDir, _ = sourceFs.SplitPath(source)
	}
	require.NoError(t, sourceFile.Close())
	t.Log("\tSource")
	if _, ok := os.LookupEnv("FULL_PATHS"); ok && (destDirChanged || sourceDirChanged) {
		t.Logf("\t\tPWD: %v", sourceRoot)
	}
	err = fs.WalkDir(sourceFs, sourceDir, func(path string, d fs.DirEntry, err error) error {
		sourceRel, _ := sourceFs.RelPath(sourceRoot, path)
		if sourceRel == "." {
			return nil
		}
		t.Logf("\t\t%v", strings.TrimPrefix(sourceRel, sourceFs.PathSeparator()))
		return nil
	})
	require.NoError(t, err)

	destFile, err := destinationFs.Open(destinationFs.PathJoin(destRoot, tt.Args.Dest))
	require.NoError(t, err)
	destStat, err := destFile.Stat()
	require.NoError(t, err)
	var destinationDir string
	if destStat.IsDir() {
		destinationDir = join(destinationFs.PathSeparator(), destRoot, tt.Args.Dest)
	} else {
		destinationDir, _ = destinationFs.SplitPath(destination)
	}
	require.NoError(t, destFile.Close())

	t.Log("\tDestination")
	err = fs.WalkDir(destinationFs, destinationDir, func(path string, d fs.DirEntry, err error) error {
		destRel, _ := destinationFs.RelPath(destRoot, path)
		if destRel == "." {
			return nil
		}
		var entry *PathSpecEntry
		for _, e := range tt.Dest {
			if e.path == destRel {
				entry = &e
				break
			}
		}

		if entry == nil || entry.Preexisting {
			t.Logf("\t\t%v", strings.TrimPrefix(destRel, sourceFs.PathSeparator()))
		} else {
			t.Logf("\t\t\u001B[32m%v\u001B[0m", strings.TrimPrefix(destRel, sourceFs.PathSeparator()))
		}

		return nil
	})
	require.NoError(t, err)

	for _, e := range tt.Dest {
		file, err := destinationFs.Open(destinationFs.PathJoin(destRoot, e.path))
		require.NoError(t, err, e.path)
		fileStat, err := file.Stat()
		require.NoError(t, err, e.path)
		assert.Equal(t, e.Dir, fileStat.IsDir(), e.path)
		if tt.Args.PreserveTimes && !e.Mtime.IsZero() {
			assert.Equal(t, e.Mtime, fileStat.ModTime().UTC())
		}
		require.NoError(t, file.Close())
	}

	assert.NoError(t, sourceFs.RemoveAll(sourceRoot))
	assert.NoError(t, destinationFs.RemoveAll(destRoot))
}

func PathSpec(t *testing.T, srcPathSeparator string, destPathSeparator string) []PathSpecTest {
	t.Helper()
	return []PathSpecTest{
		{
			Name: "copy foo to Dest",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo"),
				Dest: "Dest",
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "Dest", "foo")},
				{Dir: false, path: join(destPathSeparator, "Dest", "foo", "baz.txt")},
			},
		},
		{
			Name: "copy foo to Dest with PreserveTimes",
			Args: PathSpecArgs{
				Src:           join(srcPathSeparator, "Src", "foo"),
				Dest:          "Dest",
				PreserveTimes: true,
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt"), Mtime: time.Date(2010, 11, 17, 20, 34, 58, 0, time.UTC)},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "Dest", "foo")},
				{Dir: false, path: join(destPathSeparator, "Dest", "foo", "baz.txt"), Mtime: time.Date(2010, 11, 17, 20, 34, 58, 0, time.UTC)},
			},
		},
		{
			Name: "copy contents of foo to foo",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo") + srcPathSeparator,
				Dest: join(destPathSeparator, "Dest", "foo"),
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "Dest", "foo"), Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "Dest", "foo", "baz.txt")},
			},
		},
		{
			Name: "copy contents of foo to bar",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo") + srcPathSeparator,
				Dest: join(destPathSeparator, "Dest", "bar"),
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "Dest", "bar"), Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "Dest", "bar", "baz.txt")},
			},
		},
		{
			Name: "copy foo to dest (with destination slash)",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "src", "foo"),
				Dest: join(destPathSeparator, "dest") + destPathSeparator,
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "src"},
				{Dir: true, path: join(srcPathSeparator, "src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "dest", "foo")},
				{Dir: false, path: join(destPathSeparator, "dest", "foo", "baz.txt")},
			},
		},
		{
			Name: "copy baz.txt to Dest",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo", "baz.txt"),
				Dest: join(destPathSeparator, "Dest", "baz.txt"),
			},
			Src: []PathSpecEntry{
				{Dir: true, path: join(srcPathSeparator, "Src")},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "Dest", "baz.txt")},
			},
		},
		{
			Name: "copy baz.txt to Dest without Name",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo", "baz.txt"),
				Dest: "Dest" + destPathSeparator,
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "Dest", "baz.txt")},
			},
		},
		{
			Name: "copy baz.txt to Dest with rename",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo", "baz.txt"),
				Dest: join(destPathSeparator, "Dest", "taz.txt"),
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "Dest", Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "Dest", "taz.txt")},
			},
		},
		{
			Name: "copy baz.txt to current working directory",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "Src", "foo", "baz.txt"),
				Dest: "",
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "Src"},
				{Dir: true, path: join(srcPathSeparator, "Src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "Src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: false, path: "baz.txt"},
			},
		},
		{
			Name: "copy contents of foo to foo (with destination slash)",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "src", "foo") + srcPathSeparator,
				Dest: join(destPathSeparator, "dest", "foo") + destPathSeparator,
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "src"},
				{Dir: true, path: join(srcPathSeparator, "src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "dest", "foo"), Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "dest", "foo", "baz.txt")},
			},
		},
		{
			Name: "copy contents of foo to bar (with destination slash)",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "src", "foo") + srcPathSeparator,
				Dest: join(destPathSeparator, "dest", "bar") + destPathSeparator,
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "src"},
				{Dir: true, path: join(srcPathSeparator, "src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "src", "foo", "baz.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "dest", "bar"), Preexisting: true},
				{Dir: false, path: join(destPathSeparator, "dest", "bar", "baz.txt")},
			},
		},
		{
			Name: "copy foo to dest with nested directories and files",
			Args: PathSpecArgs{
				Src:  join(srcPathSeparator, "src", "foo"),
				Dest: "dest",
			},
			Src: []PathSpecEntry{
				{Dir: true, path: "src"},
				{Dir: true, path: join(srcPathSeparator, "src", "foo")},
				{Dir: false, path: join(srcPathSeparator, "src", "foo", "baz.txt")},
				{Dir: true, path: join(srcPathSeparator, "src", "foo", "bar")},
				{Dir: false, path: join(srcPathSeparator, "src", "foo", "bar", "lo.txt")},
			},
			Dest: []PathSpecEntry{
				{Dir: true, path: "dest", Preexisting: true},
				{Dir: true, path: join(destPathSeparator, "dest", "foo")},
				{Dir: false, path: join(destPathSeparator, "dest", "foo", "baz.txt")},
				{Dir: true, path: join(destPathSeparator, "dest", "foo", "bar")},
				{Dir: false, path: join(destPathSeparator, "dest", "foo", "bar", "lo.txt")},
			},
		},
	}
}

func join(pathSeparator string, parts ...string) string {
	parts = lo.FilterMap[string, string](parts, func(item string, index int) (string, bool) {
		if item == "" {
			return item, false
		}
		return item, true
	})
	if parts[len(parts)-1] == pathSeparator {
		return strings.Join(parts, pathSeparator) + pathSeparator
	} else {
		return strings.Join(parts, pathSeparator)
	}
}

func ChangeDir(fs ReadWriteFs, relativePath string, fullPath string, mutex *sync.Mutex) (func(), bool, error) {
	if relativePath != "" {
		return func() {}, false, nil
	}
	statefulFs, ok := fs.(StatefulDirectory)
	if !ok {
		return func() {}, false, nil
	}

	originalDir, err := statefulFs.Getwd()
	if err != nil {
		return func() {}, false, err
	}
	mutex.Lock()
	err = statefulFs.Chdir(fullPath)
	if err != nil {
		mutex.Unlock()
		return func() {}, false, err
	}

	return func() {
		statefulFs.Chdir(originalDir)
		mutex.Unlock()
	}, true, nil
}

type Cmd interface {
	SetOut(io.Writer)
	SetErr(newErr io.Writer)
	Run() error
	Args() []string
}

type ExeCmd struct {
	*exec.Cmd
}

func (e ExeCmd) SetOut(w io.Writer) {
	e.Stdout = w
}

func (e ExeCmd) SetErr(w io.Writer) {
	e.Stderr = w
}

func (e ExeCmd) Run() error {
	return e.Cmd.Run()
}

func (e ExeCmd) Args() []string {
	return e.Cmd.Args
}
