package site

import (
	"context"
	"io/fs"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/core"
)

func (site *Site) IsIgnoredStatic(path string, isDir bool) bool {
	// path 是 static FS 内部路径，不包含 static 前缀。
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

func (site *Site) BuildStatic(ctx context.Context, writer core.Writer) error {
	site.ctx.Logger.Infof("Copying static...")

	now := time.Now()

	staticFS, err := site.ctx.GetFS(core.MountStatic, true, true)
	if err != nil {
		return err
	}
	if err := fs.WalkDir(staticFS, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if site.IsIgnoredStatic(path, info.IsDir()) {
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

		site.ctx.Logger.Debugf("copy static file [%s] -> %s", path, path)
		return writer.WriteFile(ctx, path, src)
	}); err != nil {
		return err
	}
	site.ctx.Logger.Infof("Done: in %v",
		time.Since(now),
	)
	return nil
}
