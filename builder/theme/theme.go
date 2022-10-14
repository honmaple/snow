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
	Theme interface {
		Root() fs.FS
		WriteTemplate([]string, string, map[string]interface{}) error
	}
	Template interface {
		Write([]string, string, map[string]interface{}) error
	}
	theme struct {
		root     fs.FS
		template Template
	}
)

var (
	//go:embed internal
	themeFS embed.FS
)

func (t *theme) Root() fs.FS {
	return t.root
}

func (t *theme) WriteTemplate(names []string, file string, context map[string]interface{}) error {
	return t.template.Write(names, file, context)
}

func New(conf config.Config) (Theme, error) {
	var (
		err  error
		root fs.FS
		tmpl Template
	)

	name := conf.GetString("theme.path")
	switch name {
	case "simple":
		root, _ = fs.Sub(themeFS, filepath.Join("internal", name))
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
	return &theme{
		root:     root,
		template: tmpl,
	}, nil
}
