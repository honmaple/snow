package mergefs

import (
	"io"
	"io/fs"
	"path"
	"slices"
	"time"
)

// Merge returns a new fs.FS that overlays the given filesystems.
// For file reads: the first FS that contains the file wins.
// For directory reads: entries are merged across all FSes (first FS wins on name conflict).
func Merge(fsys ...fs.FS) fs.FS {
	return &mergedFS{fsys: fsys}
}

type mergedFS struct {
	fsys []fs.FS
}

// Open opens the named file.
// It tries each FS in order and returns the first success.
// For directories, it returns a merged directory that combines entries from all FSes.
func (m *mergedFS) Open(name string) (fs.File, error) {
	// Validate path per fs.FS contract
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	if len(m.fsys) == 0 && name == "." {
		return &emptyDir{}, nil
	}

	var dirs []fs.File

	for _, fsys := range m.fsys {
		f, err := fsys.Open(name)
		if err != nil {
			continue
		}

		info, err := f.Stat()
		if err != nil {
			f.Close()
			continue
		}

		if !info.IsDir() {
			// Regular file: first match wins
			// Close any dirs we already opened
			for _, d := range dirs {
				d.Close()
			}
			return f, nil
		}

		// It's a directory: collect from all FSes for merging
		dirs = append(dirs, f)
	}

	if len(dirs) == 0 {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	// Return a merged directory view
	return &mergedDir{
		name: path.Base(name),
		dirs: dirs,
	}, nil
}

// mergedDir is an fs.File representing a merged directory.
type mergedDir struct {
	name    string
	dirs    []fs.File
	entries []fs.DirEntry
	offset  int
}

func (d *mergedDir) Close() error {
	var lastErr error
	for _, dir := range d.dirs {
		if err := dir.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (d *mergedDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}

func (d *mergedDir) Stat() (fs.FileInfo, error) {
	// Use the first directory's stat info
	return d.dirs[0].Stat()
}

// ReadDir implements fs.ReadDirFile.
// Merges entries from all underlying directories; first FS wins on name conflicts.
func (d *mergedDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.entries == nil {
		merged, err := d.mergeEntries()
		if err != nil {
			return nil, err
		}
		d.entries = merged
	}

	if n <= 0 {
		result := d.entries[d.offset:]
		d.offset = len(d.entries)
		return result, nil
	}

	if d.offset >= len(d.entries) {
		return nil, io.EOF
	}

	end := d.offset + n
	if end > len(d.entries) {
		end = len(d.entries)
	}
	result := d.entries[d.offset:end]
	d.offset = end
	if d.offset >= len(d.entries) {
		return result, io.EOF
	}
	return result, nil
}

func (d *mergedDir) mergeEntries() ([]fs.DirEntry, error) {
	seen := make(map[string]bool)
	var result []fs.DirEntry

	for _, dir := range d.dirs {
		rdf, ok := dir.(fs.ReadDirFile)
		if !ok {
			continue
		}
		entries, err := rdf.ReadDir(-1)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !seen[e.Name()] {
				seen[e.Name()] = true
				result = append(result, e)
			}
		}
	}

	// Sort alphabetically for deterministic output
	slices.SortFunc(result, func(a, b fs.DirEntry) int {
		if a.Name() < b.Name() {
			return -1
		}
		if a.Name() > b.Name() {
			return 1
		}
		return 0
	})

	return result, nil
}

// mergedDirInfo is a synthetic fs.FileInfo for the merged root directory.
type mergedDirInfo struct {
	name string
}

func (i *mergedDirInfo) Name() string       { return i.name }
func (i *mergedDirInfo) Size() int64        { return 0 }
func (i *mergedDirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (i *mergedDirInfo) ModTime() time.Time { return time.Time{} }
func (i *mergedDirInfo) IsDir() bool        { return true }
func (i *mergedDirInfo) Sys() any           { return nil }

type emptyDir struct{}

func (d *emptyDir) Stat() (fs.FileInfo, error) {
	return &emptyDirInfo{}, nil
}

func (d *emptyDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: ".", Err: io.EOF}
}

func (d *emptyDir) Close() error { return nil }

func (d *emptyDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 {
		return nil, nil
	}
	return nil, io.EOF
}

type emptyDirInfo struct{}

func (i *emptyDirInfo) Name() string       { return "." }
func (i *emptyDirInfo) Size() int64        { return 0 }
func (i *emptyDirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (i *emptyDirInfo) ModTime() time.Time { return time.Time{} }
func (i *emptyDirInfo) IsDir() bool        { return true }
func (i *emptyDirInfo) Sys() any           { return nil }
