package page

import (
	"path/filepath"
)

func (b *Builder) insertAsset(file string) {
	dir := filepath.Dir(file)
	section := b.ctx.findSection(dir)
	if section == nil {
		return
	}
	b.ctx.withLock(func() {
		section.Assets = append(section.Assets, file)
	})
}

func (b *Builder) writeAsset(file string) {
}
