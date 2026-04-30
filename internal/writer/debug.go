package writer

import (
	"context"
	"io"
	"path/filepath"

	"github.com/honmaple/snow/internal/core"
)

type DebugWriter struct {
	ctx *core.Context
}

func (w *DebugWriter) Write(ctx context.Context, file string, r io.Reader) error {
	if file == "" {
		return nil
	}
	output := filepath.Join(w.ctx.Config.GetString("output_dir"), file)

	w.ctx.Logger.Infoln("Writing", output)
	return nil
}

func NewDebugWriter(ctx *core.Context) *DebugWriter {
	return &DebugWriter{ctx: ctx}
}
