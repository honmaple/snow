package links

import (
	"bytes"
	stdpath "path"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func testLinkPage(filePath, fileDir, outputPath, body string) *content.Page {
	return &content.Page{
		Node: &content.Node{
			File: &content.File{
				Path: filePath,
				Dir:  fileDir,
				Ext:  stdpath.Ext(filePath),
			},
			Content: body,
			Summary: body,
		},
		Path: outputPath,
	}
}

func testLinkSection(filePath, fileDir, outputPath, body string) *content.Section {
	return &content.Section{
		Node: &content.Node{
			File: &content.File{
				Path: filePath,
				Dir:  fileDir,
				Ext:  stdpath.Ext(filePath),
			},
			Content: body,
			Summary: body,
		},
		Path: outputPath,
	}
}

type testContentStore struct {
	pages    content.Pages
	hidden   content.Pages
	sections content.Sections
}

func (s testContentStore) Pages(string) content.Pages {
	return s.pages
}

func (s testContentStore) HiddenPages(string) content.Pages {
	return s.hidden
}

func (s testContentStore) Sections(string) content.Sections {
	return s.sections
}

func testLinkContext() (*core.Context, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.Out = &buf
	logger.Formatter = &logrus.TextFormatter{DisableTimestamp: true}
	return &core.Context{Logger: logger}, &buf
}

func TestLinksHookKeepsAbsolutePathLinks(t *testing.T) {
	target := testLinkPage("pages/hello.md", "pages", "/custom/hello/", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<p><a href="/pages/hello.md">hello</a></p>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source, target}}, "zh")

	assert.Contains(t, source.Content, `<a href="/pages/hello.md">hello</a>`)
	assert.Contains(t, source.Summary, `<a href="/pages/hello.md">hello</a>`)
}

func TestLinksHookRewritesRelativeContentLinks(t *testing.T) {
	sibling := testLinkPage("posts/hello.md", "posts", "/posts/hello/", "")
	parent := testLinkPage("pages/hello.md", "pages", "/pages/hello/", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="hello.md">sibling</a><a href="../pages/hello.md">parent</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source, sibling, parent}}, "zh")

	assert.Contains(t, source.Content, `<a href="/posts/hello/">sibling</a>`)
	assert.Contains(t, source.Content, `<a href="/pages/hello/">parent</a>`)
}

func TestLinksHookRewritesCurrentDirectoryOrgLinks(t *testing.T) {
	target := testLinkPage("posts/another.org", "posts", "/posts/another/", "")
	source := testLinkPage("posts/source.org", "posts", "/posts/source/", `<a href="./another.org">aaa</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source, target}}, "zh")

	assert.Contains(t, source.Content, `<a href="/posts/another/">aaa</a>`)
}

func TestLinksHookRewritesContentRootLinks(t *testing.T) {
	target := testLinkPage("pages/test.md", "pages", "/pages/test/", "")
	source := testLinkPage("posts/deep/source.md", "posts/deep", "/posts/deep/source/", `<a href="@/pages/test.md">test</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source, target}}, "zh")

	assert.Contains(t, source.Content, `<a href="/pages/test/">test</a>`)
}

func TestLinksHookKeepsUnchangedTagsRaw(t *testing.T) {
	target := testLinkPage("pages/test.md", "pages", "/pages/test/", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<p>before<br><hr><img src="cover.png"><a href="@/pages/test.md">test</a></p>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source, target}}, "zh")

	assert.Equal(t, `<p>before<br><hr><img src="cover.png"><a href="/pages/test/">test</a></p>`, source.Content)
}

func TestLinksHookPreservesQueryAndFragment(t *testing.T) {
	target := testLinkPage("pages/hello.md", "pages", "/custom/hello/", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="@/pages/hello.md?ref=1#intro">hello</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source, target}}, "zh")

	assert.Contains(t, source.Content, `<a href="/custom/hello/?ref=1#intro">hello</a>`)
}

func TestLinksHookIgnoresExternalAndProtocolLinks(t *testing.T) {
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="https://example.com/a.md">external</a><a href="mailto:test@example.com">mail</a><a href="#local">local</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{source}}, "zh")

	assert.Contains(t, source.Content, `href="https://example.com/a.md"`)
	assert.Contains(t, source.Content, `href="mailto:test@example.com"`)
	assert.Contains(t, source.Content, `href="#local"`)
}

func TestLinksHookWarnsAndKeepsMissingContentLinks(t *testing.T) {
	ctx, buf := testLinkContext()
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="@/pages/missing.md">missing</a>`)

	h := &LinksHook{ctx: ctx}
	h.HandleContent(testContentStore{pages: content.Pages{source}}, "zh")

	assert.Contains(t, source.Content, `href="@/pages/missing.md"`)
	assert.Contains(t, buf.String(), "content link not found")
	assert.Contains(t, buf.String(), "page=posts/source.md")
	assert.Contains(t, buf.String(), "target=pages/missing.md")
}

func TestLinksHookRewritesHiddenPages(t *testing.T) {
	target := testLinkPage("pages/hello.md", "pages", "/custom/hello/", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="@/pages/hello.md">hello</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{pages: content.Pages{target}, hidden: content.Pages{source}}, "zh")

	assert.Contains(t, source.Content, `<a href="/custom/hello/">hello</a>`)
	assert.Contains(t, source.Summary, `<a href="/custom/hello/">hello</a>`)
}

func TestLinksHookKeepsAlreadyResolvedOutputPath(t *testing.T) {
	target := testLinkPage("pages/hello.md", "pages", "/pages/hello.html", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="/pages/hello.html">hello</a>`)
	ctx, buf := testLinkContext()

	h := &LinksHook{ctx: ctx}
	h.HandleContent(testContentStore{pages: content.Pages{source, target}}, "zh")

	assert.Contains(t, source.Content, `href="/pages/hello.html"`)
	assert.NotContains(t, buf.String(), "content link not found")
}

func TestLinksHookRewritesSectionTargets(t *testing.T) {
	target := testLinkSection("docs/_index.md", "docs", "/docs/", "")
	source := testLinkPage("posts/source.md", "posts", "/posts/source/", `<a href="@/docs/_index.md">docs</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{
		pages:    content.Pages{source},
		sections: content.Sections{target},
	}, "zh")

	assert.Contains(t, source.Content, `<a href="/docs/">docs</a>`)
}

func TestLinksHookRewritesSectionContent(t *testing.T) {
	target := testLinkPage("posts/hello.md", "posts", "/posts/hello/", "")
	source := testLinkSection("docs/_index.md", "docs", "/docs/", `<a href="../posts/hello.md">hello</a>`)

	h := &LinksHook{}
	h.HandleContent(testContentStore{
		pages:    content.Pages{target},
		sections: content.Sections{source},
	}, "zh")

	assert.Contains(t, source.Content, `<a href="/posts/hello/">hello</a>`)
	assert.Contains(t, source.Summary, `<a href="/posts/hello/">hello</a>`)
}
