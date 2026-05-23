package writer

import (
	"context"
	"io"
	"io/fs"
	"strings"

	"github.com/spf13/afero"
)

type MemoryWriter struct {
	fs afero.Fs
}

func (m *MemoryWriter) Reset() {
	afero.Walk(m.fs, ".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			m.fs.RemoveAll(path)
			return fs.SkipDir
		}
		m.fs.Remove(path)
		return nil
	})
}

func (m *MemoryWriter) Open(file string) (fs.File, error) {
	return m.fs.Open(file)
}

func (m *MemoryWriter) Create(file string) (io.WriteCloser, error) {
	return m.fs.Create(file)
}

func (m *MemoryWriter) WriteFile(ctx context.Context, file string, r io.Reader) error {
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	return afero.WriteReader(m.fs, file, r)
}

func NewMemoryWriter() *MemoryWriter {
	return &MemoryWriter{fs: afero.NewMemMapFs()}
}
