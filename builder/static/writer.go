package static

import (
	"fmt"
	"github.com/honmaple/snow/utils"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func (b *Builder) copyFile(static *Static) error {
	var (
		err     error
		srcFile fs.File
		output  = b.conf.GetOutput()
	)

	src, dst := static.File, filepath.Join(output, static.URL)
	b.conf.Log.Debugln("Copying", src, "to", dst)

	if static.IsTheme {
		srcFile, err = b.theme.Root().Open(src)
	} else {
		if stat, err := os.Stat(src); err != nil {
			return err
		} else if !stat.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", src)
		}
		srcFile, err = os.Open(src)
	}
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if dir := filepath.Dir(dst); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
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
