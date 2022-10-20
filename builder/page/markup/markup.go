package markup

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/honmaple/snow/builder/page/markup/html"
	"github.com/honmaple/snow/builder/page/markup/markdown"
	"github.com/honmaple/snow/builder/page/markup/orgmode"
	"github.com/honmaple/snow/config"
)

type (
	MarkupReader interface {
		Read(io.Reader) (map[string]string, error)
		Exts() []string
	}
	Markup struct {
		conf config.Config
		exts map[string]MarkupReader
	}
)

func (m *Markup) Read(file string) (map[string]string, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(file)
	reader, ok := m.exts[ext]
	if !ok {
		return nil, fmt.Errorf("no reader for %s", file)
	}
	return reader.Read(bytes.NewBuffer(buf))
}

func New(conf config.Config) *Markup {
	rs := []MarkupReader{
		html.New(conf),
		orgmode.New(conf),
		markdown.New(conf),
	}
	exts := make(map[string]MarkupReader)
	for _, r := range rs {
		for _, ext := range r.Exts() {
			exts[ext] = r
		}
	}
	return &Markup{conf: conf, exts: exts}
}
