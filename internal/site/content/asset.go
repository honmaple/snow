package content

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honmaple/snow/internal/core"
)

type (
	Asset struct {
		File      string
		Path      string
		Permalink string
	}
	Assets []*Asset
)

func (d *Processor) validateAssetPath(file string, assetPath string) error {
	if filepath.IsAbs(file) {
		return &core.Error{
			Op:   "parse content asset",
			Err:  fmt.Errorf("absolute asset path is not allowed"),
			Path: file,
		}
	}
	cleanPath := filepath.Clean(file)
	if filepath.ToSlash(cleanPath) != assetPath ||
		strings.HasPrefix(cleanPath, "..") {
		return &core.Error{
			Op:   "parse content asset",
			Err:  fmt.Errorf("asset path must be a clean relative path"),
			Path: file,
		}
	}
	return nil
}

func (d *Processor) parseAssetPaths(root string, files []string) ([]string, error) {
	rootFS := os.DirFS(root)

	assetPaths := make([]string, 0)
	for _, file := range files {
		assetPath := filepath.ToSlash(file)
		if err := d.validateAssetPath(file, assetPath); err != nil {
			return nil, err
		}

		matches, err := doublestar.Glob(rootFS, assetPath)
		if err != nil {
			return nil, &core.Error{
				Op:   "parse content asset",
				Err:  err,
				Path: file,
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
	if asset == nil || asset.File == "" || asset.Path == "" {
		return nil
	}

	src, err := os.Open(asset.File)
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
