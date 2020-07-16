package static

import (
	"path/filepath"

	"github.com/honmaple/snow/utils"
)

func (b *Builder) Write(files []*StaticFile) error {
	output := b.conf.GetString("output")
	for _, file := range files {
		utils.CopyFile(file.File, filepath.Join(output, "static"))
	}
	return nil
}
