package types

import (
	stdpath "path"
	"path/filepath"
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
