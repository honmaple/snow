package site

import (
	"context"
	"io/fs"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

func (site *Site) isIgnoredStatic(path string, isDir bool) bool {
	matchPath := path
	if isDir {
		matchPath = matchPath + "/"
	}
	for _, pattern := range site.ctx.Config.GetStringSlice("ignored_static") {
		matched, err := doublestar.Match(pattern, matchPath)
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
	site.ctx.Logger.Infof("Copying static...")

	now := time.Now()

	staticFS, err := site.ctx.GetFS("static", true)
	if err != nil {
		return err
	}

	if err := site.copyStaticDir(staticFS); err != nil {
		return err
	}
	site.ctx.Logger.Infof("Done: in %v",
		time.Since(now),
	)
	return nil
}
