package content

import (
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/core"
)

type File struct {
	Path         string // tech/develop/rust/rust-note.fr.md
	Dir          string // tech/develop/rust
	Name         string // rust-note.fr.md
	BaseName     string // rust-note
	LanguageName string // fr
	Ext          string // md
}

func (file *File) GetFullPath() string {
	return filepath.FromSlash(file.Path)
}

func (file *File) GetSectionPath() string {
	if file.BaseName == "index" {
		return stdpath.Dir(file.Dir)
	}
	return file.Dir
}

func (d *Processor) parseFile(fullpath string) (*File, error) {
	relPath, err := filepath.Rel(d.ctx.GetContentDir(), fullpath)
	if err != nil {
		return nil, &core.Error{
			Op:   "parse file",
			Err:  err,
			Path: "relpath",
		}
	}
	relPath = filepath.ToSlash(relPath)

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
