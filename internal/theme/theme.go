package theme

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	//go:embed templates
	internalFS embed.FS
)

type Theme struct {
	root fs.FS
}

func (t *Theme) Open(name string) (fs.File, error) {
	if name == "internal" {
		return internalFS.Open(".")
	}
	if strings.HasPrefix(name, "internal/") {
		return internalFS.Open(name[9:])
	}
	return t.root.Open(name)
}

func (t *Theme) Sub(name string) (fs.FS, error) {
	var (
		err   error
		subFS fs.FS
	)

	if name == "internal" {
		subFS = internalFS
	} else if strings.HasPrefix(name, "internal/") {
		subFS, err = fs.Sub(internalFS, name[9:])
	} else {
		subFS, err = fs.Sub(t.root, name)
	}
	if err != nil {
		return nil, err
	}
	if _, err := fs.Stat(subFS, "."); err != nil {
		return nil, err
	}
	return subFS, nil
}

func New(dir string, name string) (*Theme, error) {
	if dir == "" {
		dir = "themes"
	}

	var (
		root fs.FS
	)
	if name == "" {
		root = os.DirFS(".")
	} else {
		path := filepath.Join(dir, name)
		_, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("The theme %s not found: %s", name, err.Error())
		}
		root = os.DirFS(path)
	}

	t := &Theme{
		root: root,
	}
	return t, nil
}
