package static

import (
	"io/ioutil"
	"path/filepath"

	"github.com/honmaple/snow/config"
)

type (
	StaticFile struct {
		File string
		Type string
	}
	Builder struct {
		conf *config.Config
	}
)

func (b *Builder) Read() ([]*StaticFile, error) {
	dirs := b.conf.GetStringSlice("static_dirs")

	files := make([]*StaticFile, 0)
	for _, path := range dirs {
		rd, err := ioutil.ReadDir(path)
		if err != nil {
			continue
		}
		for _, file := range rd {
			filename := filepath.Join(path, file.Name())
			if !file.IsDir() {
				continue
			}
			files = append(files, &StaticFile{File: filename})
		}
	}
	return files, nil
}

func (b *Builder) Build() error {
	files, err := b.Read()
	if err != nil {
		return err
	}
	return b.Write(files)
}

func NewBuilder(conf *config.Config) *Builder {
	return &Builder{conf: conf}
}
