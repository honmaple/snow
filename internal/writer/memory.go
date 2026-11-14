package writer

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type MemoryWriter struct {
	fs afero.Fs
}

func (m *MemoryWriter) Reset() {
	root := string(filepath.Separator)
	entries, _ := afero.ReadDir(m.fs, root)
	for _, entry := range entries {
		m.fs.RemoveAll(filepath.Join(root, entry.Name()))
	}
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
