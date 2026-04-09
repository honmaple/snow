package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/honmaple/snow/internal/core"
)

type (
	Parser interface {
		Parse(string) (*Result, error)
		SupportedExtensions() []string
	}
	MarkupParser interface {
		Parse(io.Reader) (*Result, error)
	}
)

type parserImpl struct {
	cache sync.Map
	ps    map[string]MarkupParser
	exts  []string
}

func (d *parserImpl) Parse(file string) (*Result, error) {
	v, ok := d.cache.Load(file)
	if ok {
		return v.(*Result), nil
	}

	p, ok := d.ps[filepath.Ext(file)]
	if !ok {
		return nil, fmt.Errorf("no parser for %s", file)
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result, err := p.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("Read file %s err: %s", file, err.Error())
	}
	d.cache.Store(file, result)
	return result, nil
}

func (d *parserImpl) SupportedExtensions() []string {
	return d.exts
}

func New(ctx *core.Context) Parser {
	d := &parserImpl{
		ps:   make(map[string]MarkupParser),
		exts: make([]string, 0),
	}
	for ext, foctory := range factories {
		d.exts = append(d.exts, ext)

		d.ps[ext] = foctory(ctx)
	}
	return d
}

type Factory func(*core.Context) MarkupParser

func Register(ext string, c Factory) {
	factories[ext] = c
}

var factories map[string]Factory

func init() {
	factories = make(map[string]Factory)
}
