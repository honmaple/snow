package writer

import (
	"context"
	"io"
)

type NullWriter struct {
}

func (w *NullWriter) WriteFile(ctx context.Context, file string, r io.Reader) error {
	return nil
}

func NewNullWriter() *NullWriter {
	return &NullWriter{}
}
