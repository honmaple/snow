package shortcode

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
	"golang.org/x/net/html"
)

type shortcode struct {
	hook.BaseHook
	conf  config.Config
	theme theme.Theme
	tpls  map[string]func(map[string]interface{}) string
}

func (self *shortcode) Name() string {
	return "shortcode"
}

func (self *shortcode) template(lookups ...string) (func(map[string]interface{}) string, error) {
	tpl := self.theme.LookupTemplate(lookups...)
	if tpl == nil {
		return nil, fmt.Errorf("Lookup %s but not found or some error happen", lookups)
	}
	return func(vars map[string]interface{}) string {
		out, err := tpl.Execute(vars)
		if err != nil {
			self.conf.Log.Error(err.Error())
			return ""
		}
		return out
	}, nil
}

func (self *shortcode) renderNext(page *page.Page, w *bytes.Buffer, z *html.Tokenizer, startToken *html.Token, counter map[string]int) bool {
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
					self.conf.Log.Warnf("%s: shortcode no name", page.File)
					break
				}
			}
			shortcode, ok := self.tpls[name]
			if ok {
				params := make(map[string]interface{})
				for _, attr := range token.Attr {
					params[attr.Key] = attr.Val
				}

				vars := map[string]interface{}{
					"page":     page,
					"body":     "",
					"params":   params,
					"_name":    name,
					"_counter": counter[name],
					"_shortcode": func(s string) string {
						return self.render(page, s)
					},
				}
				counter[name]++

				if next == html.StartTagToken {
					var buf bytes.Buffer

					if !self.renderNext(page, &buf, z, &token, counter) {
						self.conf.Log.Warnf("%s: closing delimiter '</%s>' is missing", page.File, token.Data)
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
	return false
}

func (self *shortcode) render(page *page.Page, content string) string {
	var (
		w       bytes.Buffer
		z       = html.NewTokenizer(strings.NewReader(content))
		counter = make(map[string]int)
	)
	self.renderNext(page, &w, z, nil, counter)
	return w.String()
}

func (self *shortcode) Page(page *page.Page) *page.Page {
	if len(self.tpls) == 0 {
		return page
	}
	page.Content = self.render(page, page.Content)
	page.Summary = self.render(page, page.Summary)
	return page
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	h := &shortcode{conf: conf, theme: theme}
	h.tpls = make(map[string]func(map[string]interface{}) string)

	roots := []string{
		"_internal/templates/shortcodes",
		"templates/shortcodes",
	}

	rootFunc := func(path string) bool {
		for _, root := range roots {
			if root == path {
				return true
			}
		}
		return false
	}

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if rootFunc(path) {
			return nil
		}

		lookups := []string{path}
		if d.IsDir() {
			lookups = []string{
				filepath.Join(path, "index.tpl"),
				filepath.Join(path, "index.html"),
			}
		}

		tpl, err := h.template(lookups...)
		if err != nil {
			conf.Log.Errorln(path, err.Error())
			return nil
		}
		h.tpls[utils.FileBaseName(path)] = tpl
		return nil
	}

	for _, root := range roots {
		fs.WalkDir(theme, root, walkFunc)
	}
	return h
}

func init() {
	hook.Register("shortcode", New)
}
