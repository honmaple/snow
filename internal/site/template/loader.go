package template

import (
	"io"
	"io/fs"
	"path/filepath"
)

type loader struct {
	fs fs.FS
}

func (l *loader) Abs(base, name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return name
	// return filepath.Join(filepath.Dir(base), name)
}

func (l *loader) Get(path string) (io.Reader, error) {
	return l.fs.Open(path)
}

func newLoader(fs fs.FS) *loader {
	return &loader{fs: fs}
}
