package content

import (
	stdpath "path"
	"path/filepath"
	"strings"
)

type File struct {
	Path         string // tech/develop/rust/rust-note.fr.md
	Dir          string // tech/develop/rust
	Name         string // rust-note.fr.md
	BaseName     string // rust-note
	LanguageName string // fr
	Ext          string // md
}

func (d *Processor) parseFile(fullpath string) (*File, error) {
	relPath := filepath.ToSlash(fullpath)
	if relPath == "." {
		relPath = ""
	}

	ext := stdpath.Ext(relPath)
	nameWithExt := stdpath.Base(relPath)
	nameWithoutExt := strings.TrimSuffix(nameWithExt, ext)

	dir := stdpath.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	return &File{
		Path:     relPath,
		Dir:      dir,
		Ext:      ext,
		Name:     nameWithExt,
		BaseName: nameWithoutExt,
	}, nil
}
