package template

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type loader struct {
	theme    fs.FS
	override string
}

func (l *loader) Abs(base, name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return name
	// return filepath.Join(filepath.Dir(base), name)
}

func (l *loader) Get(path string) (io.Reader, error) {
	buf, err := l.GetBytes(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func (l *loader) GetBytes(path string) ([]byte, error) {
	if strings.HasPrefix(path, "internal/") {
		if !strings.HasPrefix(path, "internal/templates/") {
			path = filepath.Join("internal/templates", path[8:])
		}
	} else if !strings.HasPrefix(path, "templates/") {
		path = filepath.Join("templates", path)
	}

	if strings.HasPrefix(path, "templates/") {
		buf, err := os.ReadFile(path)
		if err == nil {
			return buf, nil
		}
	}

	f, err := l.theme.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func newLoader(theme fs.FS) *loader {
	return &loader{theme: theme}
}
