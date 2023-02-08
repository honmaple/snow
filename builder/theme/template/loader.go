package template

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
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
	if l.override != "" {
		buf, err := ioutil.ReadFile(filepath.Join(l.override, path))
		if err == nil {
			return buf, nil
		}
	}

	if strings.HasPrefix(path, "_internal") {
		path = filepath.Join("_internal/templates", path[9:])
	} else {
		path = filepath.Join("templates", path)
	}

	f, err := l.theme.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func newLoader(theme fs.FS, override string) *loader {
	return &loader{theme: theme, override: override}
}
