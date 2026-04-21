package content

import (
	"os"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/site/content/types"
)

type (
	ContentParser struct {
		ctx        *core.Context
		parser     parser.Parser
		parserExts map[string]bool
	}
	ContentParserOption func(*ContentParser)
)

func (d *ContentParser) findIndexFiles(fullpath string, prefix string) []string {
	// 如果有多个扩展: index.md, index.org只返回第一个
	allowedFiles := make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		allowedFiles[prefix+ext] = true
		for lang := range d.ctx.OtherLanguages {
			allowedFiles[prefix+lang+ext] = true
		}
	}

	files, err := os.ReadDir(fullpath)
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
		nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
		if allowedFiles[name] && !resultMap[nameWithoutExt] {
			results = append(results, name)
			resultMap[nameWithoutExt] = true
		}
	}
	return results
}

func (d *ContentParser) parseFile(fullpath string) (*types.File, error) {
	relPath, err := filepath.Rel(d.ctx.GetContentDir(), fullpath)
	if err != nil {
		return nil, &core.Error{
			Op:   "parse file",
			Err:  err,
			Path: "relpath",
		}
	}
	relPath = filepath.ToSlash(relPath)

	ext := stdpath.Ext(relPath)
	nameWithExt := stdpath.Base(relPath)
	nameWithoutExt := strings.TrimSuffix(nameWithExt, ext)

	dir := stdpath.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	return &types.File{
		Path:     relPath,
		Dir:      dir,
		Ext:      ext,
		Name:     nameWithExt,
		BaseName: nameWithoutExt,
	}, nil
}

func (d *ContentParser) parseNode(fullpath string) (*types.Node, error) {
	file, err := d.parseFile(fullpath)
	if err != nil {
		return nil, err
	}

	result, err := d.parser.Parse(fullpath)
	if err != nil {
		return nil, err
	}

	meta := types.NewFrontMatter(result.FrontMatter)

	lang := meta.GetString("lang")
	if lang == "" {
		langExt := stdpath.Ext(file.BaseName)
		if langExt != "" {
			lang = strings.TrimPrefix(langExt, ".")
		}
	}
	if lang == "" || (lang != d.ctx.GetDefaultLanguage() && !d.ctx.Config.IsSet("languages."+lang)) {
		lang = d.ctx.GetDefaultLanguage()
	}

	if ext := "." + lang; strings.HasSuffix(file.BaseName, ext) {
		file.BaseName = strings.TrimSuffix(file.BaseName, ext)
		file.LanguageName = lang
	}
	lctx := d.ctx.For(lang)

	node := &types.Node{
		FrontMatter: meta,
		File:        file,
		Lang:        lang,
		Slug:        meta.GetString("slug"),
		Title:       meta.GetString("title"),
		Description: meta.GetString("desc"),
		Content:     result.Content,
		Summary:     meta.GetString("summary"),
	}
	if node.Summary == "" {
		node.Summary = lctx.GetSummary(result.Content)
	}
	return node, nil
}

func WithParser(p parser.Parser) ContentParserOption {
	return func(d *ContentParser) {
		d.parser = p
	}
}

func NewContentParser(ctx *core.Context, opts ...ContentParserOption) *ContentParser {
	d := &ContentParser{
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
