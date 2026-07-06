package shortcode

import (
	"bytes"
	"io/fs"
	stdpath "path"
	"slices"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/spf13/cast"
	"golang.org/x/net/html"
)

type ShortcodeSet struct {
	ctx    *core.Context
	tpls   map[string]template.Template
	tplset template.TemplateSet
}

type shortcodeFrame struct {
	token html.Token
	name  string
	vars  map[string]any
	body  bytes.Buffer
}

func isSelfClosing(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

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

func shortcodeParams(token html.Token, name string) Params {
	params := make(Params)
	for _, attr := range token.Attr {
		if attr.Key == "_name" || (attr.Key == name && attr.Val == "") {
			continue
		}
		params[attr.Key] = attr.Val
	}
	return params
}

func writeShortcodeOutput(out *bytes.Buffer, stack []*shortcodeFrame, result string) {
	if len(stack) == 0 {
		out.WriteString(result)
		return
	}
	stack[len(stack)-1].body.WriteString(result)
}

func nextCounter(source Source, name string) int {
	key := "counter:" + name
	counter := 0
	if value, ok := source.Get(key); ok {
		counter = cast.ToInt(value)
	}
	source.Set(key, counter+1)
	return counter
}

func (h *ShortcodeSet) renderContext(source Source, name string, params Params, counter int) map[string]any {
	vars := map[string]any{
		"name":     name,
		"body":     "",
		"params":   params,
		"counter":  counter,
		"_name":    name,
		"_counter": counter,
	}
	for k, v := range source.Context() {
		vars[k] = v
	}
	return vars
}

func (h *ShortcodeSet) warnf(format string, args ...any) {
	if h.ctx != nil && h.ctx.Logger != nil {
		h.ctx.Logger.Warnf(format, args...)
	}
}

func (h *ShortcodeSet) renderFrame(frame *shortcodeFrame, source Source) string {
	frame.vars["body"] = frame.body.String()

	tpl, ok := h.tpls[frame.name]
	if !ok {
		return frame.token.String()
	}
	result, err := tpl.Execute(frame.vars)
	if err != nil {
		h.warnf("%s: %s", source.Id(), err)
		return frame.token.String()
	}
	return result
}

func (h *ShortcodeSet) render(source Source) string {
	var (
		out   bytes.Buffer
		stack []*shortcodeFrame
		z     = html.NewTokenizer(strings.NewReader(source.Content()))
	)

	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			for len(stack) > 0 {
				frame := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				h.warnf("%s: closing delimiter '</%s>' is missing", source.Id(), frame.token.Data)
				writeShortcodeOutput(&out, stack, h.renderFrame(frame, source))
			}
			return out.String()
		}

		token := z.Token()
		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			name := token.Data
			if token.Data == "shortcode" {
				name = shortcodeName(token)
				if name == "" {
					h.warnf("%s: shortcode no name", source.Id())
					break
				}
			}
			tpl, ok := h.tpls[name]
			if ok {
				vars := h.renderContext(source, name, shortcodeParams(token, name), nextCounter(source, name))

				if tokenType == html.StartTagToken && !isSelfClosing(token.Data) {
					stack = append(stack, &shortcodeFrame{
						token: token,
						name:  name,
						vars:  vars,
					})
					continue
				}

				result, err := tpl.Execute(vars)
				if err != nil {
					h.warnf("%s: %s", source.Id(), err)
					break
				}
				writeShortcodeOutput(&out, stack, result)
				continue
			}
		case html.EndTagToken:
			if len(stack) > 0 && stack[len(stack)-1].token.Data == token.Data {
				frame := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				writeShortcodeOutput(&out, stack, h.renderFrame(frame, source))
				continue
			}
		}
		writeShortcodeOutput(&out, stack, token.String())
	}
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
	return h.render(source)
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
