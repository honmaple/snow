package static

import (
	"fmt"
	"path/filepath"
	"strings"
	// "github.com/honmaple/snow/utils"
)

func (b *Builder) Write(files []*StaticFile) error {
	for _, static := range files {
		b.write(static.File, static.URL)
	}
	return nil
}

func (b *Builder) write(file string, dest string) error {
	output := b.conf.GetString("output_dir")

	if dest == "" {
		dest = filepath.Join(output, filepath.Base(file))
	} else if strings.HasSuffix(dest, "/") {
		dest = filepath.Join(output, dest, filepath.Base(file))
	} else {
		dest = filepath.Join(output, dest)
	}

	fmt.Println(fmt.Sprintf("Copying %s to %s", file, dest))
	// _, err := utils.CopyFile(file, dest)
	return nil
}
