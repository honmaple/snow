package writer

import (
	"context"
	"io"
	"io/fs"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/spf13/afero"
)

type MemoryWriter struct {
	fs  afero.Fs
	ctx *core.Context
}

func (m *MemoryWriter) Reset() {
	m.fs = afero.NewMemMapFs()
}

func (m *MemoryWriter) Open(file string) (fs.File, error) {
	return m.fs.Open(file)
}

func (m *MemoryWriter) WriteFile(ctx context.Context, file string, r io.Reader) error {
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	return afero.WriteReader(m.fs, file, r)
}

func NewMemoryWriter(ctx *core.Context) *MemoryWriter {
	return &MemoryWriter{ctx: ctx, fs: afero.NewMemMapFs()}
}
