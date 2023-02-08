package static

import (
	"os"
	"path/filepath"

	"github.com/honmaple/snow/utils"
)

func (b *Builder) copyFile(static *Static) error {
	src, dst := static.Name(), filepath.Join(b.conf.GetOutput(), static.URL)
	b.conf.Log.Debugln("Copying", src, "to", dst)

	if dir := filepath.Dir(dst); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}
	return static.CopyTo(dst)
	// return nil
}

func (b *Builder) Write(files []*Static) error {
	for _, static := range files {
		if static.URL == "" {
			continue
		}
		if err := b.copyFile(static); err != nil {
			return err
		}
	}
	return nil
}
