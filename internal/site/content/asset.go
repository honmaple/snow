package content

import (
	"context"
	"fmt"
	"io/fs"
	stdpath "path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/core"
)

type (
	Asset struct {
		FS        fs.FS
		File      string
		Path      string
		Permalink string
	}
	Assets []*Asset
)

func (d *Processor) validateAssetPath(assetPath string) error {
	if stdpath.IsAbs(assetPath) {
		return &core.Error{
			Op:   "parse content asset",
			Err:  fmt.Errorf("absolute asset path is not allowed"),
			Path: assetPath,
		}
	}
	cleanPath := stdpath.Clean(assetPath)
	if cleanPath != assetPath || strings.HasPrefix(cleanPath, "..") {
		return &core.Error{
			Op:   "parse content asset",
			Err:  fmt.Errorf("asset path must be a clean relative path"),
			Path: assetPath,
		}
	}
	return nil
}

func (d *Processor) parseAssetPaths(fsys fs.FS, root string, files []string) ([]string, error) {
	rootFS, err := fs.Sub(fsys, root)
	if err != nil {
		return nil, &core.Error{
			Op:   "parse content asset",
			Err:  err,
			Path: root,
		}
	}

	assetPaths := make([]string, 0)
	for _, assetPath := range files {
		if err := d.validateAssetPath(assetPath); err != nil {
			return nil, err
		}

		matches, err := doublestar.Glob(rootFS, assetPath)
		if err != nil {
			return nil, &core.Error{
				Op:   "parse content asset",
				Err:  err,
				Path: assetPath,
			}
		}
		for _, match := range matches {
			info, err := fs.Stat(rootFS, match)
			if err != nil {
				return nil, &core.Error{
					Op:   "parse content asset",
					Err:  err,
					Path: match,
				}
			}
			if info.IsDir() {
				continue
			}
			assetPaths = append(assetPaths, match)
		}
	}
	return assetPaths, nil
}

func (d *Processor) resolveAssetPath(basePath string, assetPath string) string {
	if !strings.HasSuffix(basePath, "/") {
		basePath = stdpath.Dir(basePath)
	}
	return stdpath.Join(basePath, assetPath)
}

func (d *Processor) RenderAsset(asset *Asset, writer core.Writer) error {
	if asset == nil || asset.FS == nil || asset.File == "" || asset.Path == "" {
		return nil
	}

	src, err := asset.FS.Open(asset.File)
	if err != nil {
		return &core.Error{
			Op:   "open asset",
			Err:  err,
			Path: asset.File,
		}
	}
	defer src.Close()

	d.ctx.Logger.Debugf("copy content asset [%s] -> %s", asset.File, asset.Path)
	if err := writer.WriteFile(context.TODO(), asset.Path, src); err != nil {
		return &core.Error{
			Op:   "write asset",
			Err:  err,
			Path: asset.Path,
		}
	}
	return nil
}
