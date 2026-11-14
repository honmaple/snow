package site

import (
	"context"
	"io/fs"
	stdpath "path"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/core"
)

func (site *Site) resolveStaticPath(path string, customPaths map[string]string) string {
	if len(customPaths) == 0 {
		return path
	}
	if target := customPaths[strings.ToLower(path)]; target != "" {
		return target
	}

	for parent := stdpath.Dir(path); parent != "."; parent = stdpath.Dir(parent) {
		if target := customPaths[strings.ToLower(parent)]; target != "" {
			return stdpath.Join(target, strings.TrimPrefix(path, parent+"/"))
		}
	}
	return path
}

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

	paths := make(map[string]string)
	if sc := site.ctx.Config.Sub("statics"); sc != nil {
		for _, k := range sc.AllKeys() {
			paths[k] = sc.GetString(k)
		}
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

		customPath := site.resolveStaticPath(path, paths)

		site.ctx.Logger.Debugf("copy static file [%s] -> %s", path, customPath)
		return writer.WriteFile(ctx, customPath, src)
	}); err != nil {
		return err
	}
	site.ctx.Logger.Infof("Done: in %v",
		time.Since(now),
	)
	return nil
}
