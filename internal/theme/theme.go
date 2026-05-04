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
	//go:embed static templates
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

func New(name string) (*Theme, error) {
	var (
		root fs.FS
	)
	if name == "" {
		root = internalFS
	} else {
		path := filepath.Join("themes", name)
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
