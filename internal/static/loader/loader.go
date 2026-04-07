package loader

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/static/types"
	"github.com/honmaple/snow/internal/utils"
)

type (
	DiskLoader struct {
		ctx *core.Context

		hook    types.Hook
		statics *utils.Slice[*types.Static]
	}
	LoaderOption func(*DiskLoader)
)

func (d *DiskLoader) Load() (types.Store, error) {
	root := d.ctx.GetStaticDir()
	if root == "" {
		return nil, fmt.Errorf("The static dir is null")
	}

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		for _, pattern := range d.ctx.Config.GetStringSlice("ignored_static") {
			if strings.HasPrefix(pattern, "@theme/") {
				continue
			}
			matchPath := path
			if info.IsDir() {
				matchPath = matchPath + "/"
			}
			matched, err := filepath.Match(pattern, matchPath)
			if err != nil {
				d.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			}
			if matched {
				if info.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		return d.insertStatic(path)
	}
	if err := filepath.WalkDir(d.ctx.GetStaticDir(), walkDir); err != nil {
		return nil, err
	}

	walkThemeDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "static" {
			return nil
		}
		for _, pattern := range d.ctx.Config.GetStringSlice("ignored_static") {
			if !strings.HasPrefix(pattern, "@theme/") {
				continue
			}
			matchPath := path
			if info.IsDir() {
				matchPath = matchPath + "/"
			}
			matched, err := filepath.Match(pattern[7:], matchPath)
			if err != nil {
				d.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			}
			if matched {
				if info.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		return d.insertThemeStatic(path)
	}
	if err := fs.WalkDir(d.ctx.Theme, "static", walkThemeDir); err != nil {
		return nil, err
	}
	return d, nil
}

func New(ctx *core.Context, opts ...LoaderOption) *DiskLoader {
	d := &DiskLoader{
		ctx:     ctx,
		statics: utils.NewSlice[*types.Static](),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}
