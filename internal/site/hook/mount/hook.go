package mount

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
)

type (
	MountHook struct {
		hook.HookImpl
	}
	Option struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
)

func New(ctx *core.Context) (hook.Hook, error) {
	var opts []*Option
	if err := hook.Unmarshal(ctx.Config.Get("hooks.mount.option"), &opts); err != nil {
		return nil, err
	}
	for i, opt := range opts {
		if opt.Source == "" {
			return nil, fmt.Errorf("hooks.mount.option[%d].source is required", i)
		}
		if opt.Target == "" {
			return nil, fmt.Errorf("hooks.mount.option[%d].target is required", i)
		}
		source, err := newLocalFS(opt.Source)
		if err != nil {
			return nil, fmt.Errorf("hooks.mount.option[%d]: %w", i, err)
		}
		if err := ctx.FS.Mount(source, opt.Target); err != nil {
			return nil, fmt.Errorf("hooks.mount.option[%d]: %w", i, err)
		}
	}
	return &MountHook{}, nil
}

type localFS struct {
	path  string
	isDir bool
}

func newLocalFS(path string) (fs.FS, error) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &localFS{
		path:  path,
		isDir: info.IsDir(),
	}, nil
}

func (l *localFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if l.isDir {
		return os.DirFS(l.path).Open(name)
	}
	if name != "." {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return os.Open(l.path)
}

func init() {
	hook.Register("mount", New)
}
