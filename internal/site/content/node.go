package content

import (
	"golang.org/x/net/html"
	stdpath "path"
	"strings"
	"unicode"

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

		Weight      int64
		WordCount   int64
		ReadingTime int64
	}
	Heading = parser.Heading
)

func (n *Node) Param(key string, defaults ...any) any {
	return n.FrontMatter.Get("params."+key, defaults...)
}

func (d *Processor) countReadingTime(wordCount int64) int64 {
	if wordCount <= 0 {
		return 0
	}
	const wordsPerMinute int64 = 200
	return (wordCount + wordsPerMinute - 1) / wordsPerMinute
}

func (d *Processor) countReadingWords(text string) int64 {
	var count int64
	inWord := false
	for _, r := range text {
		if unicode.In(r, unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana) {
			if inWord {
				inWord = false
			}
			count++
			continue
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if !inWord {
				count++
				inWord = true
			}
			continue
		}
		inWord = false
	}
	return count
}

func (d *Processor) countReadingStats(content string) (int64, int64) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		wordCount := d.countReadingWords(content)
		return wordCount, d.countReadingTime(wordCount)
	}

	var walk func(*html.Node) int64
	walk = func(node *html.Node) int64 {
		if node.Type == html.ElementNode && (node.Data == "script" || node.Data == "style") {
			return 0
		}
		if node.Type == html.TextNode {
			return d.countReadingWords(node.Data)
		}

		var count int64
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			count += walk(child)
		}
		return count
	}
	wordCount := walk(doc)
	return wordCount, d.countReadingTime(wordCount)
}

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
		Weight:      fm.GetInt64("weight"),
		Toc:         result.Toc,
		Content:     result.Content,
		RawContent:  result.RawContent,
		Summary:     result.Summary,
	}
	node.WordCount, node.ReadingTime = d.countReadingStats(node.Content)

	lctx := d.ctx.For(lang)
	if node.Summary == "" && node.Content != "" {
		node.Summary = lctx.GetSummary(node.Content)
	}
	return node, nil
}
