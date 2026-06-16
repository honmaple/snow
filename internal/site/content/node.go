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

		Toc         []*Heading
		Lang        string
		Slug        string
		Title       string
		Summary     string
		Content     string
		RawContent  string
		Description string
	}
	Heading = parser.Heading
)

func (d *Processor) parseNode(fullpath string) (*Node, error) {
	file, err := d.parseFile(fullpath)
	if err != nil {
		return nil, err
	}

	// TODO: 增加缓存
	result, err := d.parser.Parse(d.contentFS, fullpath)
	if err != nil {
		return nil, err
	}

	fm := NewFrontMatter(result.FrontMatter)
	// 合并配置
	if strings.HasPrefix(file.Name, "_index.") {
		fm.MergeFrom(d.ctx.GetSectionConfig(file.Dir))
	} else {
		fm.MergeFrom(d.ctx.GetPageConfig(file.Dir))
	}

	lang := fm.GetString("lang")
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
		FrontMatter: fm,
		File:        file,
		Lang:        lang,
		Slug:        fm.GetString("slug"),
		Title:       fm.GetString("title"),
		Description: fm.GetString("description"),
		Content:     result.Content,
		RawContent:  result.RawContent,
		Summary:     result.Summary,
		Toc:         result.Toc,
	}
	lctx := d.ctx.For(lang)
	if node.Summary == "" && node.Content != "" {
		node.Summary = lctx.GetSummary(node.Content)
	}
	return node, nil
}
