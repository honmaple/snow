package core

import (
	"io"
	"io/fs"
	stdpath "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/honmaple/snow/internal/utils/mergefs"
)

const (
	MountContent   = "content"
	MountStatic    = "static"
	MountTemplates = "templates"
	MountThemes    = "themes"
)

type (
	VirtualFS interface {
		fs.FS
		Mount(fs.FS, string) error
	}
	virtualFS struct {
		base   fs.FS
		mounts []*mountEntry
	}
)

func newVirtualFS(base fs.FS) VirtualFS {
	return &virtualFS{base: base}
}

func normalizeVirtualPath(name string) string {
	return strings.Trim(strings.TrimSpace(filepath.ToSlash(name)), "/")
}

func normalizeMountTarget(target string) (string, error) {
	target = strings.TrimSpace(filepath.ToSlash(target))
	if target == "" || stdpath.IsAbs(target) {
		return "", fs.ErrInvalid
	}

	clean := stdpath.Clean(target)
	if target != clean || clean == "." || strings.HasPrefix(clean, "..") {
		return "", fs.ErrInvalid
	}
	return clean, nil
}

func subDirFSIfExists(root fs.FS, name string) fs.FS {
	name = normalizeVirtualPath(name)
	if root == nil || name == "" {
		return nil
	}
	subFS, err := fs.Sub(root, name)
	if err != nil {
		return nil
	}
	info, err := fs.Stat(subFS, ".")
	if err != nil || !info.IsDir() {
		return nil
	}
	return subFS
}

func (v *virtualFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	fsys := make([]fs.FS, 0, len(v.mounts)+1)
	for _, item := range v.mounts {
		fsys = append(fsys, item)
	}
	if v.base != nil {
		fsys = append(fsys, v.base)
	}
	return mergefs.Merge(fsys...).Open(name)
}

func (v *virtualFS) Mount(source fs.FS, target string) error {
	item, err := newMountEntry(source, target)
	if err != nil {
		return err
	}
	v.mounts = append(v.mounts, item)
	return nil
}

type (
	mountEntry struct {
		source fs.FS
		target string
		isDir  bool
	}
	mountedDir struct {
		name  string
		entry fs.DirEntry
		read  bool
	}
	mountedDirInfo struct {
		name string
	}
)

func newMountEntry(source fs.FS, target string) (*mountEntry, error) {
	rawTarget := target

	target, err := normalizeMountTarget(target)
	if err != nil {
		return nil, &fs.PathError{Op: "mount", Path: rawTarget, Err: fs.ErrInvalid}
	}
	if source == nil {
		return nil, &fs.PathError{Op: "mount", Path: target, Err: fs.ErrInvalid}
	}
	info, err := fs.Stat(source, ".")
	if err != nil {
		return nil, err
	}
	return &mountEntry{
		source: source,
		target: target,
		isDir:  info.IsDir(),
	}, nil
}

func (e *mountEntry) mountedChild(name string, target string) (string, bool) {
	if name == "." {
		if target == "." {
			return "", false
		}
		return target, true
	}
	if strings.HasPrefix(target, name+"/") {
		return strings.TrimPrefix(target, name+"/"), true
	}
	return "", false
}

func (e *mountEntry) mountedDirWith(name string, entry fs.DirEntry) fs.File {
	return &mountedDir{
		name:  stdpath.Base(name),
		entry: entry,
	}
}

func (e *mountEntry) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	name = normalizeVirtualPath(name)
	if name == "" {
		name = "."
	}

	if !e.isDir && name == e.target {
		return e.source.Open(".")
	}
	if e.isDir {
		if name == e.target {
			return e.source.Open(".")
		}
		if strings.HasPrefix(name, e.target+"/") {
			return e.source.Open(strings.TrimPrefix(name, e.target+"/"))
		}
	}

	child, ok := e.mountedChild(name, e.target)
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	if strings.Contains(child, "/") {
		child, _, _ = strings.Cut(child, "/")
		return e.mountedDirWith(name, &mountedDirInfo{name: child}), nil
	}
	if e.isDir {
		return e.mountedDirWith(name, &mountedDirInfo{name: child}), nil
	}
	info, err := fs.Stat(e.source, ".")
	if err != nil {
		return nil, err
	}
	return e.mountedDirWith(name, fs.FileInfoToDirEntry(info)), nil
}

func (d *mountedDir) Stat() (fs.FileInfo, error) {
	return mountedDirInfo{name: d.name}, nil
}

func (d *mountedDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}

func (d *mountedDir) Close() error {
	return nil
}

func (d *mountedDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.read {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	d.read = true
	if n <= 0 {
		return []fs.DirEntry{d.entry}, nil
	}
	return []fs.DirEntry{d.entry}, io.EOF
}

func (i mountedDirInfo) Sys() any           { return nil }
func (i mountedDirInfo) Name() string       { return i.name }
func (i mountedDirInfo) Size() int64        { return 0 }
func (i mountedDirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (i mountedDirInfo) ModTime() time.Time { return time.Time{} }
func (i mountedDirInfo) IsDir() bool        { return true }
func (i mountedDirInfo) Type() fs.FileMode  { return fs.ModeDir }
func (i mountedDirInfo) Info() (fs.FileInfo, error) {
	return i, nil
}
