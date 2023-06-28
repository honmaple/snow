package static

import (
	"io/fs"
	"strings"
)

type (
	Static struct {
		Root      fs.FS
		Name      string
		Path      string
		Permalink string
	}
	Statics []*Static
)

func (file Static) IsTheme() bool {
	return strings.HasPrefix(file.Name, "@theme/")
}

func (file Static) Open() (fs.File, error) {
	return file.Root.Open(file.Name)
}

func (statics Statics) Lookup(files []string) Statics {
	m := make(map[string]bool)
	for _, file := range files {
		m[file] = true
	}

	newstatics := make(Statics, 0)
	for _, static := range statics {
		if m[static.Name] {
			newstatics = append(newstatics, static)
		}
	}
	return newstatics
}
