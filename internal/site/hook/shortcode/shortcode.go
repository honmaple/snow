package shortcode

import (
	"io"
	"io/fs"
	stdpath "path"
	"slices"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/honmaple/snow/internal/utils"
	"github.com/spf13/cast"
	"golang.org/x/net/html"
)

type ShortcodeSet struct {
	ctx    *core.Context
	tpls   map[string]template.Template
	tplset template.TemplateSet
}

type (
	shortcodeToken struct {
		tpl  template.Template
		name string
		attr []html.Attribute
	}
	shortcodeFrame struct {
		tag      string
		body     strings.Builder
		rawToken string
		newToken *shortcodeToken
	}
)

func shortcodeName(token html.Token) string {
	for _, attr := range token.Attr {
		if attr.Key == "_name" && attr.Val != "" {
			return attr.Val
		}
	}
	for _, attr := range token.Attr {
		if attr.Key != "_name" && attr.Val == "" {
			return attr.Key
		}
	}
	return ""
}

func (h *ShortcodeSet) warnf(format string, args ...any) {
	if h.ctx != nil && h.ctx.Logger != nil {
		h.ctx.Logger.Warnf(format, args...)
	}
}

func (h *ShortcodeSet) resolveShortcode(source Source, token html.Token) (*shortcodeToken, bool) {
	name := token.Data
	if name == "shortcode" {
		name = shortcodeName(token)
		if name == "" {
			h.warnf("%s: shortcode no name", source.Id())
			return nil, false
		}
	}

	tpl, ok := h.tpls[name]
	if !ok {
		return nil, false
	}
	return &shortcodeToken{tpl: tpl, name: name, attr: token.Attr}, true
}

func (h *ShortcodeSet) executeShortcode(source Source, token *shortcodeToken, body string) (string, error) {
	params := make(Params)
	for _, attr := range token.attr {
		if attr.Key == "_name" || (attr.Key == token.name && attr.Val == "") {
			continue
		}
		params[attr.Key] = attr.Val
	}

	counter := 0
	counterKey := "counter:" + token.name
	if value, ok := source.Get(counterKey); ok {
		counter = cast.ToInt(value)
	}
	source.Set(counterKey, counter+1)

	vars := map[string]any{
		"name":     token.name,
		"body":     body,
		"params":   params,
		"counter":  counter,
		"_name":    token.name,
		"_counter": counter,
	}
	for k, v := range source.Context() {
		vars[k] = v
	}
	return token.tpl.Execute(vars)
}

func (h *ShortcodeSet) render(source Source) (string, error) {
	var (
		stack     = []*shortcodeFrame{{}}
		tokenizer = html.NewTokenizer(strings.NewReader(source.Content()))
	)

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return source.Content(), err
		}

		top := stack[len(stack)-1]
		raw := string(tokenizer.Raw())
		token := tokenizer.Token()
		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			if newToken, ok := h.resolveShortcode(source, token); ok {
				// 如果是自闭合标签则直接替换
				if tokenType == html.SelfClosingTagToken || utils.IsHTMLVoidElement(token.Data) {
					newHTML, err := h.executeShortcode(source, newToken, "")
					if err != nil {
						h.warnf("%s: %s", source.Id(), err)
						top.body.WriteString(raw)
					} else {
						top.body.WriteString(newHTML)
					}
				} else {
					stack = append(stack, &shortcodeFrame{
						tag:      token.Data,
						rawToken: raw,
						newToken: newToken,
					})
				}
			} else {
				top.body.WriteString(raw)
			}
		case html.EndTagToken:
			if len(stack) > 1 && top.tag == token.Data {
				stack = stack[:len(stack)-1]
				parent := stack[len(stack)-1]

				newHTML, err := h.executeShortcode(source, top.newToken, top.body.String())
				if err != nil {
					h.warnf("%s: %s", source.Id(), err)
					fallback := top.rawToken + top.body.String() + raw
					parent.body.WriteString(fallback)
				} else {
					parent.body.WriteString(newHTML)
				}
			} else {
				top.body.WriteString(raw)
			}
		default:
			top.body.WriteString(raw)
		}
	}

	for len(stack) > 1 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		parent := stack[len(stack)-1]

		h.warnf("%s: closing delimiter '</%s>' is missing", source.Id(), top.tag)
		parent.body.WriteString(top.rawToken + top.body.String())
	}
	return stack[0].body.String(), nil
}

func (h *ShortcodeSet) Render(id string, content string, context map[string]any) string {
	if len(h.tpls) == 0 {
		return content
	}
	s := &source{id: id, content: content, context: context}
	return h.RenderSource(s)
}

func (h *ShortcodeSet) RenderSource(source Source) string {
	if len(h.tpls) == 0 {
		return source.Content()
	}
	result, err := h.render(source)
	if err != nil {
		h.warnf("parse html err: %s", err.Error())
		return source.Content()
	}
	return result
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
