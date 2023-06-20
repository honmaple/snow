package page

import (
	"path/filepath"
)

func (b *Builder) insertAsset(file string) {
	dir := filepath.Dir(file)
	for lang := range b.ctx.langs {
		section := b.ctx.findSection(dir, lang)
		if section == nil {
			continue
		}
		b.ctx.withLock(func() {
			section.Assets = append(section.Assets, file)
		})
	}
}

func (b *Builder) writeAsset(file string) {
}
