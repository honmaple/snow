package page

import (
	"path/filepath"
)

func (b *Builder) insertAsset(file string) {
	dir := filepath.Dir(file)
	for lang := range b.sections {
		section := b.findSection(dir, lang)
		if section != nil {
			section.Assets = append(section.Assets, file)
		}
	}
}
