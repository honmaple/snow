package writer

import (
	"context"
	"io"

	"github.com/honmaple/snow/internal/core"
)

type NullWriter struct {
	ctx *core.Context
}

func (w *NullWriter) Write(ctx context.Context, file string, r io.Reader) error {
	return nil
}

func NewNullWriter(ctx *core.Context) *NullWriter {
	return &NullWriter{ctx: ctx}
}
