package shortcode

import (
	"bytes"
	"errors"
	"io/fs"
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

func (self *shortcode) template(path string) (func(map[string]interface{}) string, error) {
	tpl := self.theme.LookupTemplate(path)
	if tpl == nil {
		return nil, errors.New("not found")
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

func (self *shortcode) renderNext(page *page.Page, w *bytes.Buffer, z *html.Tokenizer, startToken *html.Token, times map[string]int) bool {
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
				vars := make(map[string]interface{})

				attrs := make(map[string]interface{})
				for _, attr := range token.Attr {
					vars[attr.Key] = attr.Val
					attrs[attr.Key] = attr.Val
				}
				vars["page"] = page
				vars["attr"] = attrs
				vars["body"] = ""
				vars["times"] = times[name]
				times[name]++

				if next == html.StartTagToken {
					var buf bytes.Buffer

					if !self.renderNext(page, &buf, z, &token, times) {
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

func (self *shortcode) shortcode(page *page.Page, content string) string {
	var (
		w     bytes.Buffer
		z     = html.NewTokenizer(strings.NewReader(content))
		times = make(map[string]int)
	)
	self.renderNext(page, &w, z, nil, times)
	return w.String()
}

func (self *shortcode) AfterPageParse(page *page.Page) *page.Page {
	if len(self.tpls) == 0 {
		return page
	}
	page.Content = self.shortcode(page, page.Content)
	page.Summary = self.shortcode(page, page.Summary)
	return page
}

func newShortcode(conf config.Config, theme theme.Theme) hook.Hook {
	h := &shortcode{conf: conf, theme: theme}
	h.tpls = make(map[string]func(map[string]interface{}) string)

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		tpl, err := h.template(path)
		if err != nil {
			conf.Log.Errorln(path, err.Error())
			return nil
		}
		h.tpls[utils.FileBaseName(path)] = tpl
		return nil
	}
	fs.WalkDir(theme, "_internal/templates/shortcodes", walkFunc)
	fs.WalkDir(theme, "templates/shortcodes", walkFunc)
	return h
}

func init() {
	hook.Register("shortcode", newShortcode)
}
