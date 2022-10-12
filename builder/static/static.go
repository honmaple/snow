package static

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/config"
	"sort"
)

type (
	StaticFile struct {
		URL  string
		File string
	}
	Builder struct {
		conf *config.Config
	}
)

func (b *Builder) parse(file string, meta map[string]string) *StaticFile {
	metaList := make([]string, 0)
	for m := range meta {
		metaList = append(metaList, m)
	}
	sort.SliceStable(metaList, func(i, j int) bool {
		return len(metaList[i]) > len(metaList[j])
	})
	for _, m := range metaList {
		if strings.HasPrefix(file, m) {
			return &StaticFile{URL: meta[m]}
		}
	}
	return &StaticFile{URL: file, File: file}
}

func (b *Builder) Read() ([]*StaticFile, error) {
	exts := make(map[string]bool)
	for _, ext := range b.conf.GetStringSlice("static_exts") {
		exts[ext] = true
	}
	extsIsSet := b.conf.IsSet("static_exts")

	meta := b.conf.GetStringMapString("static_meta")
	dirs := b.conf.GetStringSlice("static_dirs")
	// url := b.conf.GetStringSlice("static_url")
	files := make([]*StaticFile, 0)
	for _, dir := range dirs {
		filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if !extsIsSet || exts[filepath.Ext(file)] {
				files = append(files, b.parse(file, meta))
			}
			return nil
		})
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
