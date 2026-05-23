package content

import (
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/site/content/parser"
)

type (
	Node struct {
		File        *File
		FrontMatter *FrontMatter

		Toc        []*Heading
		Lang       string
		Slug       string
		Title      string
		Summary    string
		Content    string
		RawContent string
	}
	Heading = parser.Heading
)

func (d *Processor) parseNode(fullpath string, isPage bool) (*Node, error) {
	file, err := d.parseFile(fullpath)
	if err != nil {
		return nil, err
	}

	// TODO: 增加缓存
	result, err := d.parser.Parse(fullpath)
	if err != nil {
		return nil, err
	}

	meta := NewFrontMatter(result.FrontMatter)
	// 合并配置
	if isPage {
		meta.MergeFrom(d.ctx.GetPageConfig(file.Dir))
	} else {
		meta.MergeFrom(d.ctx.GetSectionConfig(file.Dir))
	}

	lang := meta.GetString("lang")
	if lang == "" {
		langExt := stdpath.Ext(file.BaseName)
		if langExt != "" {
			lang = strings.TrimPrefix(langExt, ".")
		}
	}
	if !d.ctx.VerifyLanguage(lang) {
		lang = d.ctx.GetDefaultLanguage()
	}

	if ext := "." + lang; strings.HasSuffix(file.BaseName, ext) {
		file.BaseName = strings.TrimSuffix(file.BaseName, ext)
		file.LanguageName = lang
	}

	node := &Node{
		FrontMatter: meta,
		File:        file,
		Lang:        lang,
		Slug:        meta.GetString("slug"),
		Title:       meta.GetString("title"),
		Content:     result.Content,
		Summary:     result.Summary,
		Toc:         result.Toc,
	}
	lctx := d.ctx.For(lang)
	if node.Summary == "" {
		node.Summary = lctx.GetSummary(result.Content)
	}
	return node, nil
}
