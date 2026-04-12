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
	//go:embed internal
	internalFS embed.FS
)

type Theme struct {
	name string
	root fs.ReadDirFS
}

func (t *Theme) Name() string {
	return t.name
}

// @theme/statics/css/main.css
// @theme/templates/page.html
// @theme/shortcodes/img/index.html
func (t *Theme) Open(file string) (fs.File, error) {
	if strings.HasPrefix(file, "internal") {
		return internalFS.Open(file)
	}
	return t.root.Open(file)
}

func (t *Theme) ReadDir(file string) ([]fs.DirEntry, error) {
	if strings.HasPrefix(file, "internal") {
		return internalFS.ReadDir(file)
	}
	return t.root.ReadDir(file)
}

func New(name string) (*Theme, error) {
	var (
		root fs.FS
	)
	if name == "" {
		root, _ = fs.Sub(internalFS, "internal")
	} else {
		path := filepath.Join("themes", name)
		_, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("The theme %s not found: %s", name, err.Error())
		}
		root = os.DirFS(path)
	}

	t := &Theme{
		name: name,
		root: root.(fs.ReadDirFS),
	}
	return t, nil
}
