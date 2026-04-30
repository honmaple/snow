package shortcode

import (
	"bytes"
	"io/fs"
	"os"
	stdpath "path"
	"slices"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
	"golang.org/x/net/html"
)

type Shortcode struct {
	ctx    *core.Context
	tpls   map[string]template.Template
	tplset template.TemplateSet
}

func (h *Shortcode) renderNext(page *content.Page, w *bytes.Buffer, z *html.Tokenizer, startToken *html.Token, counter map[string]int) bool {
	for {
		next := z.Next()
		if next == html.ErrorToken {
			return false
		}

		token := z.Token()
		switch next {
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
					h.ctx.Logger.Warnf("%s: shortcode no name", page.File)
					break
				}
			}
			tpl, ok := h.tpls[name]
			if ok {
				attrs := make(map[string]any)
				for _, attr := range token.Attr {
					attrs[attr.Key] = attr.Val
				}

				vars := map[string]any{
					"page":     page,
					"body":     "",
					"attrs":    attrs,
					"params":   attrs,
					"_name":    name,
					"_counter": counter[name],
					"_shortcode": func(s string) string {
						return h.Render(page, s)
					},
				}
				counter[name]++

				if next == html.StartTagToken {
					var buf bytes.Buffer

					if !h.renderNext(page, &buf, z, &token, counter) {
						h.ctx.Logger.Warnf("%s: closing delimiter '</%s>' is missing", page.File, token.Data)
					}
					vars["body"] = buf.String()
				}

				result, err := tpl.Execute(vars)
				if err != nil {
					w.WriteString("")
				} else {
					w.WriteString(result)
				}
				continue
			}
		case html.EndTagToken:
			if startToken != nil && startToken.Data == token.Data {
				return true
			}
		}
		w.WriteString(token.String())
	}
}

func (h *Shortcode) Render(page *content.Page, content string) string {
	if len(h.tpls) == 0 {
		return content
	}

	var (
		w       bytes.Buffer
		z       = html.NewTokenizer(strings.NewReader(content))
		counter = make(map[string]int)
	)
	h.renderNext(page, &w, z, nil, counter)
	return w.String()
}

func (h *Shortcode) Load() error {
	exts := []string{".tpl", ".html"}

	sub1, _ := fs.Sub(os.DirFS("."), "templates")
	sub2, _ := fs.Sub(h.ctx.Theme, "templates")
	sub3, _ := fs.Sub(h.ctx.Theme, "internal/templates")
	for _, subFS := range []fs.FS{sub1, sub2, sub3} {
		if subFS == nil {
			continue
		}

		files, err := fs.ReadDir(subFS, "shortcodes")
		if err != nil {
			continue
		}
		for _, file := range files {
			filename := file.Name()
			basename := strings.TrimSuffix(filename, stdpath.Ext(filename))
			if _, ok := h.tpls[basename]; ok {
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
				h.tpls[basename] = tpl
			}
		}
	}
	return nil
}

func NewShortcode(ctx *core.Context, tplset template.TemplateSet) (*Shortcode, error) {
	h := &Shortcode{
		ctx:    ctx,
		tpls:   make(map[string]template.Template),
		tplset: tplset,
	}

	if err := h.Load(); err != nil {
		return nil, err
	}
	return h, nil
}
