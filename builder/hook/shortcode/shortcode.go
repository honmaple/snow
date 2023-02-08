package shortcode

import (
	"bytes"
	"io/fs"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
	"github.com/honmaple/snow/utils"
	"golang.org/x/net/html"
)

type shortcode struct {
	hook.BaseHook
	conf       config.Config
	theme      theme.Theme
	shortcodes map[string]func(map[string]interface{}) string
}

func (s *shortcode) Name() string {
	return "shortcode"
}

func (s *shortcode) template(path string) func(map[string]interface{}) string {
	return func(vars map[string]interface{}) string {
		buf, err := fs.ReadFile(s.theme, path)
		if err != nil {
			s.conf.Log.Error(err.Error())
			return ""
		}
		tpl, err := pongo2.FromBytes(buf)
		if err != nil {
			s.conf.Log.Error(err.Error())
			return ""
		}
		out, err := tpl.Execute(pongo2.Context(vars))
		if err != nil {
			s.conf.Log.Error(err.Error())
			return ""
		}
		return out
	}
}

func (s *shortcode) shortcode(content string) string {
	var (
		w    bytes.Buffer
		z    = html.NewTokenizer(strings.NewReader(content))
		once = make(map[string]int)
	)
	for {
		next := z.Next()
		if next == html.ErrorToken {
			break
		}

		token := z.Token()
		switch next {
		case html.StartTagToken:
			if token.Data == "shortcode" {
				vars := make(map[string]interface{})
				for _, attr := range token.Attr {
					vars[attr.Key] = attr.Val
				}
				name := vars["_name"].(string)
				shortcode, ok := s.shortcodes[name]
				if !ok {
					w.WriteString(token.String())
					break
				}
				var (
					end = false
					buf bytes.Buffer
				)
				for {
					next = z.Next()
					if next == html.ErrorToken {
						break
					}
					ton := z.Token()
					if next == html.EndTagToken && ton.Data == token.Data {
						end = true
						break
					}
					buf.WriteString(ton.String())
				}
				if !end {
					s.conf.Log.Warnln(token.String(), "closing delimiter '</shortcode>' is missing")
				}
				vars["once"] = once[name]
				vars["body"] = buf.String()
				once[name]++
				w.WriteString(shortcode(vars))
			} else {
				w.WriteString(token.String())
			}
		case html.SelfClosingTagToken:
			if token.Data == "shortcode" {
				vars := make(map[string]interface{})
				for _, attr := range token.Attr {
					vars[attr.Key] = attr.Val
				}
				name := vars["_name"].(string)
				shortcode, ok := s.shortcodes[name]
				if !ok {
					w.WriteString(token.String())
					break
				}
				vars["once"] = once[name]
				vars["body"] = ""
				once[name]++
				w.WriteString(shortcode(vars))
			} else {
				w.WriteString(token.String())
			}
		default:
			w.WriteString(token.String())
		}
	}
	return w.String()
}

func (s *shortcode) AfterPageParse(page *page.Page) *page.Page {
	page.Content = s.shortcode(page.Content)
	page.Summary = s.shortcode(page.Summary)
	return page
}

func newShortcode(conf config.Config, theme theme.Theme) hook.Hook {
	h := &shortcode{conf: conf, theme: theme}
	h.shortcodes = make(map[string]func(map[string]interface{}) string)
	fs.WalkDir(theme, "_internal/templates/shortcodes", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		h.shortcodes[utils.FileBaseName(path)] = h.template(path)
		return nil
	})
	fs.WalkDir(theme, "templates/shortcodes", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		h.shortcodes[utils.FileBaseName(path)] = h.template(path)
		return nil
	})
	return h
}

func init() {
	hook.Register("shortcode", newShortcode)
}
