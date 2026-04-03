package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/honmaple/snow/config"
)

type (
	Reader interface {
		Read(io.Reader) (*Result, error)
	}
	Parser interface {
		Parse(string) (*Result, error)
		IsSupport(string) bool
		SupportExtensions() []string
	}
)

type parserImpl struct {
	cache   sync.Map
	exts    []string
	readers map[string]Reader
}

func (d *parserImpl) IsSupport(ext string) bool {
	_, ok := d.readers[ext]
	return ok
}

func (d *parserImpl) SupportExtensions() []string {
	return d.exts
}

func (d *parserImpl) Parse(file string) (*Result, error) {
	v, ok := d.cache.Load(file)
	if ok {
		return v.(*Result), nil
	}

	reader, ok := d.readers[filepath.Ext(file)]
	if !ok {
		return nil, fmt.Errorf("no reader for %s", file)
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result, err := reader.Read(f)
	if err != nil {
		return nil, fmt.Errorf("Read file %s err: %s", file, err.Error())
	}
	d.cache.Store(file, result)
	return result, nil
}

func New(conf config.Config) Parser {
	d := &parserImpl{
		exts:    make([]string, 0),
		readers: make(map[string]Reader),
	}
	for ext, foctory := range factories {
		d.exts = append(d.exts, ext)
		d.readers[ext] = foctory(conf)
	}
	return d
}

type Factory func(config.Config) Reader

func Register(ext string, c Factory) {
	factories[ext] = c
}

var factories map[string]Factory

func init() {
	factories = make(map[string]Factory)
}
