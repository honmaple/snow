package site

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
)

func (site *Site) isIgnoredStatic(path string, isDir bool) bool {
	matchPath := path
	if isDir {
		matchPath = matchPath + "/"
	}
	for _, pattern := range site.ctx.Config.GetStringSlice("ignored_static") {
		matched, err := filepath.Match(pattern, matchPath)
		if err != nil {
			site.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (site *Site) copyStaticDir(staticFS fs.FS) error {
	return fs.WalkDir(staticFS, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if site.isIgnoredStatic(path, info.IsDir()) {
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

		return site.writer.Write(context.TODO(), path, src)
	})
}

func (site *Site) buildStatic(ctx context.Context) error {
	if _, err := fs.Stat(site.ctx.Theme, "static"); err == nil {
		site.ctx.Logger.Debug("Copy theme static")

		subFS, err := fs.Sub(site.ctx.Theme, "static")
		if err != nil {
			return err
		}
		if err := site.copyStaticDir(subFS); err != nil {
			return err
		}
	}

	staticDir := site.ctx.GetStaticDir()
	if _, err := os.Stat(staticDir); err == nil {
		site.ctx.Logger.Debugf("Copy static: %s", staticDir)

		if err := site.copyStaticDir(os.DirFS(staticDir)); err != nil {
			return err
		}
	}
	return nil
}
