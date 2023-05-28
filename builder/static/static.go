package static

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
)

type (
	Static struct {
		File interface {
			io.Reader
			Name() string
		}
		Path string
	}
	Statics []*Static
)

type localFile struct {
	mu      sync.RWMutex
	root    fs.FS
	file    string
	buff    io.Reader
	isTheme bool
}

func (l *localFile) Name() string {
	if l.isTheme {
		return filepath.Join("@theme", l.file)
	}
	return l.file
}

func (l *localFile) Read(p []byte) (int, error) {
	l.mu.RLock()
	buff := l.buff
	l.mu.RUnlock()

	if buff == nil {
		l.mu.Lock()
		defer l.mu.Unlock()

		f, err := l.root.Open(l.file)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		var b bytes.Buffer
		if _, err := io.Copy(&b, f); err != nil {
			return 0, err
		}
		l.buff = &b
		return l.buff.Read(p)
	}
	return buff.Read(p)
}

func (statics Statics) Lookup(files []string) Statics {
	m := make(map[string]bool)
	for _, file := range files {
		m[file] = true
	}

	newstatics := make(Statics, 0)
	for _, static := range statics {
		if m[static.File.Name()] {
			newstatics = append(newstatics, static)
		}
	}
	return newstatics
}
