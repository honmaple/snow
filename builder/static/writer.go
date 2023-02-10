package static

import (
	"path/filepath"
)

func (b *Builder) Write(files []*Static) error {
	for _, static := range files {
		if static.URL == "" {
			continue
		}
		src := static.File.Name()
		dst := filepath.Join(b.conf.GetOutput(), static.URL)
		b.conf.Log.Debugln("Copying", src, "to", dst)

		if err := b.conf.WriteOutput(static.URL, static.File); err != nil {
			return err
		}
	}
	return nil
}
