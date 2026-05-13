package shortcode

import (
	"bytes"
	"io/fs"
	stdpath "path"
	"slices"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
	"golang.org/x/net/html"
)

type ShortcodeSet struct {
	ctx    *core.Context
	tpls   map[string]template.Template
	tplset template.TemplateSet
}

func isSelfClosing(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

func (h *ShortcodeSet) renderNext(z *html.Tokenizer, out *bytes.Buffer, stopToken string, page *content.Page, counter map[string]int) bool {
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			return false
		}

		token := z.Token()
		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			name := token.Data
			if token.Data == "shortcode" {
				for _, attr := range token.Attr {
					if attr.Key == "_name" {
						name = attr.Val
						break
					}
				}
				if name == "" {
					h.ctx.Logger.Warnf("%s: shortcode no name", page.File.Path)
					break
				}
			}
			tpl, ok := h.tpls[name]
			if ok {
				params := make(map[string]any)
				for _, attr := range token.Attr {
					params[attr.Key] = attr.Val
				}

				vars := map[string]any{
					"page":     page,
					"name":     name,
					"body":     "",
					"params":   params,
					"counter":  counter[name],
					"_name":    name,
					"_counter": counter[name],
					"_shortcode": func(s string) string {
						return h.Render(page, s)
					},
				}
				counter[name]++

				if tokenType == html.StartTagToken && !isSelfClosing(token.Data) {
					var buf bytes.Buffer

					if !h.renderNext(z, &buf, token.Data, page, counter) {
						h.ctx.Logger.Warnf("%s for %s: closing delimiter '</%s>' is missing", page.File.Path, name, token.Data)
					}
					vars["body"] = buf.String()
				}

				result, err := tpl.Execute(vars)
				if err != nil {
					out.WriteString("")
				} else {
					out.WriteString(result)
				}
				continue
			}
		case html.EndTagToken:
			if stopToken != "" && stopToken == token.Data {
				return true
			}
		}
		out.WriteString(token.String())
	}
}

func (h *ShortcodeSet) Render(page *content.Page, content string) string {
	if len(h.tpls) == 0 {
		return content
	}

	var (
		w       bytes.Buffer
		z       = html.NewTokenizer(strings.NewReader(content))
		counter = make(map[string]int)
	)
	h.renderNext(z, &w, "", page, counter)
	return w.String()
}

func (h *ShortcodeSet) Load() (map[string]template.Template, error) {
	subFS, err := h.ctx.GetFS("templates", true)
	if err != nil {
		return nil, err
	}

	exts := []string{".tpl", ".html"}
	files, err := fs.ReadDir(subFS, "shortcodes")
	if err != nil {
		return nil, err
	}

	results := make(map[string]template.Template)
	for _, file := range files {
		filename := file.Name()
		basename := strings.TrimSuffix(filename, stdpath.Ext(filename))
		if _, ok := results[basename]; ok {
			continue
		}

		tplFiles := make([]string, 0)
		if file.IsDir() {
			for _, ext := range exts {
				tplFiles = append(tplFiles, stdpath.Join("shortcodes", filename, "index"+ext))
			}
		} else {
			if slices.Contains(exts, stdpath.Ext(filename)) {
				tplFiles = []string{
					stdpath.Join("shortcodes", filename),
				}
			}
		}

		for _, tplFile := range tplFiles {
			buf, err := fs.ReadFile(subFS, tplFile)
			if err != nil {
				continue
			}
			tpl, err := h.tplset.FromBytes(buf)
			if err != nil {
				h.ctx.Logger.Warnf("compile tpl %s err: %s", tplFile, err.Error())
				continue
			}
			results[basename] = tpl
		}
	}
	return results, nil
}

func NewShortcodeSet(ctx *core.Context, tplset template.TemplateSet) (*ShortcodeSet, error) {
	h := &ShortcodeSet{
		ctx:    ctx,
		tpls:   make(map[string]template.Template),
		tplset: tplset,
	}
	tpls, err := h.Load()
	if err != nil {
		return nil, err
	}
	h.tpls = tpls
	return h, nil
}
