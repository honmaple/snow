package parser

import (
	"fmt"
	"io"
	"io/fs"
	stdpath "path"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
)

type (
	Parser interface {
		Parse(fs.FS, string) (*Result, error)
		SupportedExtensions() []string
	}
	MarkupOption struct {
		Style           string
		ShowToc         bool
		ShowLineNumbers bool
		PreventPreCode  bool
	}
	MarkupParser interface {
		Parse(io.Reader) (*Result, error)
		SupportedExtensions() []string
	}
)

type parserImpl struct {
	ps      map[string]MarkupParser
	exts    []string
	formats map[string]MarkupParser
}

func (d *parserImpl) ParseString(content string, format string) (*Result, error) {
	p, ok := d.formats[format]
	if !ok {
		return nil, fmt.Errorf("no %s parser", format)
	}
	result, err := p.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *parserImpl) Parse(fsys fs.FS, file string) (*Result, error) {
	p, ok := d.ps[stdpath.Ext(file)]
	if !ok {
		return nil, fmt.Errorf("no parser for %s", file)
	}
	f, err := fsys.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result, err := p.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("Read file %s err: %s", file, err.Error())
	}
	return result, nil
}

func (d *parserImpl) SupportedExtensions() []string {
	return d.exts
}

func New(ctx *core.Context) Parser {
	d := &parserImpl{
		ps:      make(map[string]MarkupParser),
		exts:    make([]string, 0),
		formats: make(map[string]MarkupParser),
	}
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !ctx.Config.GetBool(fmt.Sprintf("markups.%s.enabled", name)) {
			continue
		}
		p := factories[name](ctx)
		for _, ext := range p.SupportedExtensions() {
			if _, ok := d.ps[ext]; ok {
				continue
			}
			d.exts = append(d.exts, ext)
			d.ps[ext] = p
		}
		d.formats[name] = p
	}
	return d
}

func NewMarkupOption(ctx *core.Context, name string) MarkupOption {
	opt := MarkupOption{
		Style:           ctx.GetMarkupConfig(name, "style").String(),
		ShowToc:         ctx.GetMarkupConfig(name, "show_toc").Bool(),
		ShowLineNumbers: ctx.GetMarkupConfig(name, "show_line_numbers").Bool(),
		PreventPreCode:  ctx.GetMarkupConfig(name, "prevent_pre_code").Bool(),
	}
	if opt.Style == "" {
		opt.Style = "monokai"
	}
	return opt
}

type Factory func(*core.Context) MarkupParser

func Register(name string, c Factory) {
	factories[name] = c
}

var factories map[string]Factory

func init() {
	factories = make(map[string]Factory)
}
