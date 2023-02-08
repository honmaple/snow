package theme

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/honmaple/snow/builder/theme/template"
	"github.com/honmaple/snow/config"
)

type (
	Theme interface {
		Name() string
		Open(string) (fs.File, error)
		LookupTemplate(...string) (template.Writer, bool)
	}
	theme struct {
		name     string
		root     fs.FS
		cache    sync.Map
		template template.Interface
	}
)

var (
	//go:embed internal
	themeFS embed.FS
)

func (t *theme) Name() string {
	return t.name
}

func (t *theme) Open(file string) (fs.File, error) {
	if strings.HasPrefix(file, "_internal") {
		return themeFS.Open(file[1:])
	}
	return t.root.Open(file)
}

func (t *theme) LookupTemplate(names ...string) (template.Writer, bool) {
	for _, name := range names {
		v, ok := t.cache.Load(name)
		if ok {
			return v.(template.Writer), true
		}
		// 模版未找到不输出日志, 编译模版有问题才输出
		w, err := t.template.Lookup(name)
		if err != nil {
			continue
		}
		t.cache.Store(name, w)
		return w, true
	}
	return nil, false
}

func New(conf config.Config) (Theme, error) {
	var (
		root fs.FS
	)
	name := conf.GetString("theme.name")
	if name == "" {
		root, _ = fs.Sub(themeFS, "internal")
	} else {
		path := filepath.Join("themes", name)
		_, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		root = os.DirFS(path)
	}
	t := &theme{
		name: name,
		root: root,
	}
	t.template = template.New(conf, t)
	return t, nil
}
