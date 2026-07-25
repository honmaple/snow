package parser

import (
	"fmt"
	"io"
	"io/fs"
	stdpath "path"
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/utils/slugify"
)

type (
	Parser interface {
		Parse(fs.FS, string) (*Result, error)
		SupportedExtensions() []string
	}
	MarkupParser interface {
		Parse(io.Reader) (*Result, error)
		SupportedExtensions() []string
	}
	MarkupOption struct {
		Style           string
		ShowToc         bool
		TocId           string
		ShowLineNumbers bool
		PreventPreCode  bool
	}
)

type parserImpl struct {
	exts      []string
	extMap    map[string]MarkupParser
	formatMap map[string]MarkupParser
}

func (d *parserImpl) Parse(fsys fs.FS, file string) (*Result, error) {
	markup, ok := d.extMap[stdpath.Ext(file)]
	if !ok {
		return nil, fmt.Errorf("no parser for %s", file)
	}
	f, err := fsys.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result, err := markup.Parse(f)
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
		exts:      make([]string, 0),
		extMap:    make(map[string]MarkupParser),
		formatMap: make(map[string]MarkupParser),
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
			if _, ok := d.extMap[ext]; ok {
				continue
			}
			d.exts = append(d.exts, ext)
			d.extMap[ext] = p
		}
		d.formatMap[name] = p
	}
	return d
}

const (
	TocIdIndex = "index"
	TocIdTitle = "title"
)

func NewMarkupOption(ctx *core.Context, name string) MarkupOption {
	opt := MarkupOption{
		Style:           ctx.GetMarkupConfig(name, "style").String(),
		TocId:           ctx.GetMarkupConfig(name, "toc_id").String(),
		ShowToc:         ctx.GetMarkupConfig(name, "show_toc").Bool(),
		ShowLineNumbers: ctx.GetMarkupConfig(name, "show_line_numbers").Bool(),
		PreventPreCode:  ctx.GetMarkupConfig(name, "prevent_pre_code").Bool(),
	}
	if opt.Style == "" {
		opt.Style = "monokai"
	}
	return opt
}

func (opt MarkupOption) HeadingID(title string, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(opt.TocId)) {
	case TocIdTitle, "":
		if id := slugify.Make(title, slugify.WithPreserveUnicode(true)); id != "" {
			return id
		}
	}
	return fallback
}

type Factory func(*core.Context) MarkupParser

func Register(name string, c Factory) {
	factories[name] = c
}

var factories map[string]Factory

func init() {
	factories = make(map[string]Factory)
}
