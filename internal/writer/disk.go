package writer

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/utils"
)

type DiskWriter struct {
	ctx *core.Context
}

func (w *DiskWriter) Write(ctx context.Context, file string, r io.Reader) error {
	if file == "" {
		return nil
	}
	output := filepath.Join(w.ctx.Config.GetString("output_dir"), filepath.FromSlash(file))

	w.ctx.Logger.Debugln("Writing", output)
	if dir := filepath.Dir(output); !utils.FileExists(output) {
		os.MkdirAll(dir, 0755)
	}
	dstFile, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, r)
	return err
}

func NewDiskWriter(ctx *core.Context) *DiskWriter {
	return &DiskWriter{ctx: ctx}
}
