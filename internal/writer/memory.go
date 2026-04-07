package writer

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	"context"
	"github.com/honmaple/snow/internal/core"
	"io/fs"
)

type (
	memoryFile struct {
		reader  io.ReadSeeker
		modTime time.Time
	}
	MemoryWriter struct {
		ctx   *core.Context
		files sync.Map
	}
)

func (m *MemoryWriter) Find(file string) (fs.File, bool) {
	return nil, false
}

func (m *MemoryWriter) Write(ctx context.Context, file string, r io.Reader) error {
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	// TODO: large file handle
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.files.Store(file, &memoryFile{bytes.NewReader(buf), time.Now()})
	return nil
}

func NewMemoryWriter(ctx *core.Context) *MemoryWriter {
	return &MemoryWriter{ctx: ctx}
}
