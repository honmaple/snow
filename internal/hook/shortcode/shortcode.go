package shortcode

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/content"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/hook"
	"github.com/honmaple/snow/internal/theme/template"
	"github.com/honmaple/snow/internal/utils"
	"golang.org/x/net/html"
)

type shortcode struct {
	hook.HookImpl
	ctx    *core.Context
	tpls   map[string]func(map[string]any) string
	tplset template.TemplateSet
}

func (h *shortcode) lookupTemplate(lookups ...string) (func(map[string]any) string, error) {
	tpl := h.tplset.Lookup(lookups...)
	if tpl == nil {
		return nil, fmt.Errorf("Lookup %s but not found or some error happen", lookups)
	}
	return func(vars map[string]any) string {
		out, err := tpl.Execute(h.ctx, vars)
		if err != nil {
			h.ctx.Logger.Error(err.Error())
			return ""
		}
		return out
	}, nil
}

func (h *shortcode) renderNext(page *content.Page, w *bytes.Buffer, z *html.Tokenizer, startToken *html.Token, counter map[string]int) bool {
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
			shortcode, ok := h.tpls[name]
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
						return h.render(page, s)
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
				w.WriteString(shortcode(vars))
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

func (h *shortcode) render(page *content.Page, content string) string {
	var (
		w       bytes.Buffer
		z       = html.NewTokenizer(strings.NewReader(content))
		counter = make(map[string]int)
	)
	h.renderNext(page, &w, z, nil, counter)
	return w.String()
}

func (h *shortcode) HandlePage(page *content.Page) *content.Page {
	if len(h.tpls) == 0 {
		return page
	}
	page.Summary = h.render(page, page.Summary)
	page.Content = h.render(page, page.Content)
	return page
}

func (h *shortcode) load() error {
	if err := fs.WalkDir(h.ctx.Theme, "internal/templates/shortcodes", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "internal/templates/shortcodes" {
			return nil
		}

		lookups := []string{path}
		if d.IsDir() {
			lookups = []string{
				filepath.Join(path, "index.tpl"),
				filepath.Join(path, "index.html"),
			}
		}

		tpl, err := h.lookupTemplate(lookups...)
		if err != nil {
			h.ctx.Logger.Errorln(path, err.Error())
			return nil
		}
		h.tpls[utils.FileBaseName(path)] = tpl
		return nil
	}); err != nil {
		return err
	}

	if err := fs.WalkDir(h.ctx.Theme, "templates/shortcodes", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "templates/shortcodes" {
			return nil
		}

		lookups := []string{path}
		if d.IsDir() {
			lookups = []string{
				filepath.Join(path, "index.tpl"),
				filepath.Join(path, "index.html"),
			}
		}

		tpl, err := h.lookupTemplate(lookups...)
		if err != nil {
			h.ctx.Logger.Errorln(path, err.Error())
			return nil
		}
		h.tpls[utils.FileBaseName(path)] = tpl
		return nil
	}); err != nil {
		return err
	}

	if err := fs.WalkDir(os.DirFS("."), "templates/shortcodes", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "templates/shortcodes" {
			return nil
		}

		lookups := []string{path}
		if d.IsDir() {
			lookups = []string{
				filepath.Join(path, "index.tpl"),
				filepath.Join(path, "index.html"),
			}
		}

		tpl, err := h.lookupTemplate(lookups...)
		if err != nil {
			h.ctx.Logger.Errorln(path, err.Error())
			return nil
		}
		h.tpls[utils.FileBaseName(path)] = tpl
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func New(ctx *core.Context) hook.Hook {
	h := &shortcode{
		ctx:    ctx,
		tpls:   make(map[string]func(map[string]any) string),
		tplset: template.NewSet(ctx, ctx.Theme),
	}

	h.load()
	return h
}

func init() {
	hook.Register("shortcode", New)
}
