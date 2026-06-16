package content

import (
	"context"
	"io/fs"
	stdpath "path"
	"regexp"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/site/template"
)

var (
	doubleSlashRe = regexp.MustCompile(`/{2,}`) // 匹配两个或更多的斜杠
)

type (
	Processor struct {
		ctx        *core.Context
		parser     parser.Parser
		parserExts map[string]bool
	}
	ProcessorOption func(*Processor)
)

func (d *Processor) resolvePath(path string, vars map[string]string) string {
	if vars == nil || path == "" {
		return path
	}
	args := make([]string, 0)
	for k, v := range vars {
		args = append(args, k)
		args = append(args, v)
	}
	r := strings.NewReplacer(args...)
	return doubleSlashRe.ReplaceAllString(r.Replace(path), "/")
}

func (d *Processor) findIndexFiles(fsys fs.FS, fullpath string, prefix string) []string {
	// 如果有多个扩展: index.md, index.org只返回第一个
	allowedFiles := make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		allowedFiles[prefix+ext] = true
		for lang := range d.ctx.OtherLanguages {
			allowedFiles[prefix+lang+ext] = true
		}
	}

	files, err := fs.ReadDir(fsys, fullpath)
	if err != nil {
		return nil
	}

	results := make([]string, 0)
	resultMap := make(map[string]bool)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		nameWithoutExt := strings.TrimSuffix(name, stdpath.Ext(name))
		if allowedFiles[name] && !resultMap[nameWithoutExt] {
			results = append(results, name)
			resultMap[nameWithoutExt] = true
		}
	}
	return results
}

func (d *Processor) RenderTemplate(path string, tpl template.Template, vars map[string]any, writer core.Writer) error {
	if path == "" {
		return nil
	}
	// 支持uglyurls和非uglyurls形式
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}

	lang := d.ctx.GetDefaultLanguage()
	if l, ok := vars["current_lang"]; ok {
		lang = l.(string)
	}
	lctx := d.ctx.For(lang)

	commonVars := map[string]any{
		"current_url":      lctx.GetURL(path),
		"current_path":     path,
		"current_lang":     lang,
		"current_template": tpl.Name(),
	}
	for k, v := range commonVars {
		if _, ok := vars[k]; !ok {
			vars[k] = v
		}
	}

	result, err := tpl.Execute(vars)
	if err != nil {
		return &core.Error{
			Op:   "execute tpl",
			Err:  err,
			Path: tpl.Name(),
		}
	}
	if err := writer.WriteFile(context.TODO(), path, strings.NewReader(result)); err != nil {
		return &core.Error{
			Op:   "write tpl",
			Err:  err,
			Path: path,
		}
	}
	return nil
}

func WithParser(p parser.Parser) ProcessorOption {
	return func(d *Processor) {
		d.parser = p
	}
}

func NewProcessor(ctx *core.Context, opts ...ProcessorOption) *Processor {
	d := &Processor{
		ctx: ctx,
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.parser == nil {
		d.parser = parser.New(ctx)
	}

	d.parserExts = make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		d.parserExts[ext] = true
	}
	return d
}
