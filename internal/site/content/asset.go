package content

import (
	"context"
	"os"
	stdpath "path"
	"strings"

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

func assetOutputPath(basePath string, assetPath string) string {
	if !strings.HasSuffix(basePath, "/") {
		ext := stdpath.Ext(basePath)
		if ext != "" {
			basePath = strings.TrimSuffix(basePath, ext)
		}
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
