package writer

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"strings"
	"sync"
	"time"

	"github.com/honmaple/snow/internal/core"
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

func (m *MemoryWriter) Reset() {
	m.files.Range(func(k, v interface{}) bool {
		m.files.Delete(k)
		return true
	})
}

func (m *MemoryWriter) Open(file string) (fs.File, error) {
	return nil, nil
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
