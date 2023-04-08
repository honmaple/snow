package page

import (
	"path/filepath"
)

func (b *Builder) insertAsset(file string) {
	dir := filepath.Dir(file)
	for lang := range b.ctx.sections {
		section := b.ctx.findSection(dir, lang)
		if section != nil {
			section.Assets = append(section.Assets, file)
		}
	}
}

func (b *Builder) writeAsset(file string) {
}
