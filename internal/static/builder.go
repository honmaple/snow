package static

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
)

type (
	Builder struct {
		ctx    *core.Context
		writer core.Writer
	}
	BuilderOption func(*Builder)
)

func (b *Builder) isIgnored(path string, isDir bool) bool {
	matchPath := path
	if isDir {
		matchPath = matchPath + "/"
	}
	for _, pattern := range b.ctx.Config.GetStringSlice("ignored_static") {
		matched, err := filepath.Match(pattern, matchPath)
		if err != nil {
			b.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (b *Builder) copyDir(ctx context.Context, staticFS fs.FS) error {
	return fs.WalkDir(staticFS, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if b.isIgnored(path, info.IsDir()) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}
		src, err := staticFS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		return b.writer.Write(ctx, path, src)
	})
}

func (b *Builder) Build(ctx context.Context) error {
	if _, err := fs.Stat(b.ctx.Theme, "static"); err == nil {
		b.ctx.Logger.Debug("Copy theme static")

		subFS, err := fs.Sub(b.ctx.Theme, "static")
		if err != nil {
			return err
		}
		if err := b.copyDir(ctx, subFS); err != nil {
			return err
		}
	}

	staticDir := b.ctx.GetStaticDir()
	if _, err := os.Stat(staticDir); err == nil {
		b.ctx.Logger.Debugf("Copy static: %s", staticDir)

		if err := b.copyDir(ctx, os.DirFS(staticDir)); err != nil {
			return err
		}
	}
	return nil
}

func WithWriter(w core.Writer) BuilderOption {
	return func(b *Builder) {
		b.writer = w
	}
}

func New(ctx *core.Context, opts ...BuilderOption) (*Builder, error) {
	b := &Builder{
		ctx: ctx,
	}
	for _, opt := range opts {
		opt(b)
	}
	if b.writer == nil {
		return nil, errors.New("static writer is required")
	}
	return b, nil
}

func Build(ctx *core.Context, opts ...BuilderOption) error {
	b, err := New(ctx, opts...)
	if err != nil {
		return err
	}
	return b.Build(context.TODO())
}
