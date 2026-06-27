package links

import (
	"bytes"
	"net/url"
	stdpath "path"
	"strings"

	"golang.org/x/net/html"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
)

type LinksHook struct {
	hook.HookImpl

	ctx *core.Context
}

func init() {
	hook.Register("links", New)
}

func New(ctx *core.Context) (hook.Hook, error) {
	return &LinksHook{ctx: ctx}, nil
}

func (h *LinksHook) HandleContent(store hook.ContentStore, lang string) {
	pages := store.Pages(lang)
	hidden := store.HiddenPages(lang)
	sections := store.Sections(lang)

	rewriter := newContentLinkRewriter(h.ctx, pages, hidden, sections)
	rewriter.Rewrite(pages)
	rewriter.Rewrite(hidden)
	rewriter.RewriteSections(sections)
}

type contentLinkRewriter struct {
	ctx        *core.Context
	targets    map[string]string
	outputs    map[string]struct{}
	extensions map[string]struct{}
}

func newContentLinkRewriter(ctx *core.Context, pages content.Pages, hidden content.Pages, sections content.Sections) *contentLinkRewriter {
	rewriter := &contentLinkRewriter{
		ctx:        ctx,
		targets:    make(map[string]string),
		outputs:    make(map[string]struct{}),
		extensions: make(map[string]struct{}),
	}
	rewriter.addPages(pages)
	rewriter.addPages(hidden)
	rewriter.addSections(sections)
	return rewriter
}

func (r *contentLinkRewriter) addPages(pages content.Pages) {
	for _, page := range pages {
		if page == nil || page.File == nil {
			continue
		}
		r.addTarget(page.File, page.Path)
	}
}

func (r *contentLinkRewriter) addSections(sections content.Sections) {
	for _, section := range sections {
		if section == nil || section.File == nil {
			continue
		}
		r.addTarget(section.File, section.Path)
	}
}

func (r *contentLinkRewriter) addTarget(file *content.File, outputPath string) {
	if file.Path != "" && outputPath != "" {
		r.targets[file.Path] = outputPath
	}
	if outputPath != "" {
		r.outputs[outputPath] = struct{}{}
	}
	if file.Ext != "" {
		r.extensions[normalizeExtension(file.Ext)] = struct{}{}
	}
}

func (r *contentLinkRewriter) Rewrite(pages content.Pages) {
	for _, page := range pages {
		if page == nil {
			continue
		}
		page.Content = r.rewriteHTML(page.Node, page.Content)
		page.Summary = r.rewriteHTML(page.Node, page.Summary)
	}
}

func (r *contentLinkRewriter) RewriteSections(sections content.Sections) {
	for _, section := range sections {
		if section == nil {
			continue
		}
		section.Content = r.rewriteHTML(section.Node, section.Content)
		section.Summary = r.rewriteHTML(section.Node, section.Summary)
	}
}

func (r *contentLinkRewriter) rewriteHTML(node *content.Node, data string) string {
	if data == "" {
		return data
	}

	doc, err := html.Parse(strings.NewReader("<div>" + data + "</div>"))
	if err != nil {
		return data
	}

	var walk func(*html.Node)
	walk = func(htmlNode *html.Node) {
		if htmlNode.Type == html.ElementNode && htmlNode.Data == "a" {
			for i, attr := range htmlNode.Attr {
				if attr.Key != "href" {
					continue
				}
				if href, ok := r.resolveHref(node, attr.Val); ok {
					htmlNode.Attr[i].Val = href
				}
				break
			}
		}
		for child := htmlNode.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	container := findFirstElement(doc, "div")
	if container == nil {
		return data
	}

	var buf bytes.Buffer
	for child := container.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&buf, child); err != nil {
			return data
		}
	}
	return buf.String()
}

func (r *contentLinkRewriter) resolveHref(node *content.Node, href string) (string, bool) {
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "/") || hasURLScheme(href) {
		return "", false
	}

	u, err := url.Parse(href)
	if err != nil || u.Path == "" {
		return "", false
	}
	if _, ok := r.outputs[u.Path]; ok {
		return "", false
	}

	targetPath := normalizeContentPath(node, u.Path)
	if !r.isContentPath(targetPath) {
		return "", false
	}

	targetPathOutput, ok := r.targets[targetPath]
	if !ok {
		if r.ctx != nil && r.ctx.Logger != nil {
			source := ""
			if node != nil && node.File != nil {
				source = node.File.Path
			}
			r.ctx.Logger.Warnf("content link not found: page=%s href=%s target=%s", source, href, targetPath)
		}
		return "", false
	}

	return joinTargetHref(targetPathOutput, u), true
}

func (r *contentLinkRewriter) isContentPath(path string) bool {
	if path == "" || strings.HasPrefix(path, "..") {
		return false
	}
	ext := normalizeExtension(stdpath.Ext(path))
	if ext == "" {
		return false
	}
	_, ok := r.extensions[ext]
	return ok
}

func normalizeExtension(ext string) string {
	return strings.TrimPrefix(ext, ".")
}

func normalizeContentPath(node *content.Node, hrefPath string) string {
	if strings.HasPrefix(hrefPath, "@/") {
		return strings.TrimPrefix(stdpath.Clean(strings.TrimPrefix(hrefPath, "@/")), "/")
	}

	base := ""
	if node != nil && node.File != nil {
		base = node.File.Dir
	}
	return stdpath.Clean(stdpath.Join(base, hrefPath))
}

func findFirstElement(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && node.Data == name {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findFirstElement(child, name); result != nil {
			return result
		}
	}
	return nil
}

func joinTargetHref(targetPath string, u *url.URL) string {
	var result strings.Builder
	result.WriteString(targetPath)
	if u.RawQuery != "" {
		result.WriteByte('?')
		result.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		result.WriteByte('#')
		result.WriteString(u.EscapedFragment())
	}
	return result.String()
}

func hasURLScheme(value string) bool {
	for i, r := range value {
		switch {
		case r == ':':
			return i > 0
		case r == '/', r == '?', r == '#':
			return false
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '+' || r == '-' || r == '.':
			continue
		default:
			return false
		}
	}
	return false
}
