package template

import (
	"bytes"
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
	fr, err := l.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	buf, err := io.ReadAll(fr)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(buf), nil
}

func newLoader(fs fs.FS) *loader {
	return &loader{fs: fs}
}
