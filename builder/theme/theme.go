package theme

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/honmaple/snow/builder/theme/template/html"
	"github.com/honmaple/snow/builder/theme/template/pongo2"
	"github.com/honmaple/snow/config"
)

type (
	Theme struct {
		root     fs.FS
		template Template
	}
	Template interface {
		Write([]string, string, map[string]interface{}) error
	}
)

var (
	//go:embed internal
	themeFS embed.FS
)

func (t *Theme) Root() fs.FS {
	return t.root
}

func (t *Theme) WriteTemplate(names []string, file string, context map[string]interface{}) error {
	return t.template.Write(names, file, context)
}

func New(conf *config.Config) (*Theme, error) {
	var (
		err  error
		root fs.FS
		tmpl Template
	)

	name := conf.GetString("theme.path")
	switch name {
	case "simple":
		root, err = fs.Sub(themeFS, filepath.Join("internal", name))
		if err != nil {
			return nil, err
		}
	default:
		root = os.DirFS(name)
	}
	templateFS, err := fs.Sub(root, "templates")
	if err != nil {
		return nil, err
	}

	engine := conf.GetString("theme.engine")
	switch engine {
	case "pongo2":
		tmpl = pongo2.New(conf, templateFS)
	default:
		tmpl = html.New(conf)
	}
	return &Theme{
		root:     root,
		template: tmpl,
	}, nil
}
