package content

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/honmaple/snow/internal/writer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assetTestParser struct {
	result *parser.Result
}

func (p *assetTestParser) Parse(string) (*parser.Result, error) {
	return p.result, nil
}

func (p *assetTestParser) SupportedExtensions() []string {
	return []string{".md"}
}

type assetTestTemplateSet struct{}

func (assetTestTemplateSet) Lookup(...string) template.Template                 { return nil }
func (assetTestTemplateSet) FromFile(string) (template.Template, error)         { return nil, nil }
func (assetTestTemplateSet) FromBytes([]byte) (template.Template, error)        { return nil, nil }
func (assetTestTemplateSet) FromString(string) (template.Template, error)       { return nil, nil }
func (assetTestTemplateSet) Register(string, any) error                         { return nil }
func (assetTestTemplateSet) RegisterTag(string, pongo2.TagParser) error         { return nil }
func (assetTestTemplateSet) RegisterFilter(string, pongo2.FilterFunction) error { return nil }
func (assetTestTemplateSet) RegisterTransient(string, template.TransientFunction) error {
	return nil
}

func newAssetTestProcessor(t *testing.T, root string, result *parser.Result) *Processor {
	t.Helper()

	conf := core.DefaultConfig()
	conf.Set("content_dir", filepath.Join(root, "content"))
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	return NewProcessor(ctx, WithParser(&assetTestParser{result: result}))
}

func readMemoryFile(t *testing.T, w *writer.MemoryWriter, path string) string {
	t.Helper()

	f, err := w.Open(path)
	require.NoError(t, err)
	defer f.Close()

	b, err := io.ReadAll(f)
	require.NoError(t, err)
	return string(b)
}

func TestRenderPageBundleAssetsCopiesFiles(t *testing.T) {
	root := t.TempDir()
	bundleDir := filepath.Join(root, "content", "posts", "hello")
	require.NoError(t, os.MkdirAll(filepath.Join(bundleDir, "images"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "index.md"), []byte("# Hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "images", "cover.txt"), []byte("cover"), 0644))

	processor := newAssetTestProcessor(t, root, &parser.Result{
		FrontMatter: map[string]any{
			"asset_path": "assets/{path}",
		},
		Content: "<p>Hello</p>",
	})
	page, err := processor.ParsePage(filepath.Join(bundleDir, "index.md"), true)
	require.NoError(t, err)

	assert.Len(t, page.Assets, 1)
	assert.Equal(t, "/assets/posts/hello/images/cover.txt", page.Assets[0].Path)

	w := writer.NewMemoryWriter()
	require.NoError(t, processor.RenderPage(page, assetTestTemplateSet{}, w))
	assert.Equal(t, "cover", readMemoryFile(t, w, "/assets/posts/hello/images/cover.txt"))
}

func TestPageBundleAssetsUseDirectoryForHTMLAssetPath(t *testing.T) {
	root := t.TempDir()
	bundleDir := filepath.Join(root, "content", "posts", "hello")
	require.NoError(t, os.MkdirAll(filepath.Join(bundleDir, "images"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "index.md"), []byte("# Hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(bundleDir, "images", "cover.txt"), []byte("cover"), 0644))

	processor := newAssetTestProcessor(t, root, &parser.Result{
		FrontMatter: map[string]any{
			"asset_path": "assets/{path}.html",
		},
		Content: "<p>Hello</p>",
	})
	page, err := processor.ParsePage(filepath.Join(bundleDir, "index.md"), true)
	require.NoError(t, err)

	require.Len(t, page.Assets, 1)
	assert.Equal(t, "/assets/posts/hello/images/cover.txt", page.Assets[0].Path)

	w := writer.NewMemoryWriter()
	require.NoError(t, processor.RenderPage(page, assetTestTemplateSet{}, w))
	assert.Equal(t, "cover", readMemoryFile(t, w, "/assets/posts/hello/images/cover.txt"))
}

func TestRenderSectionAssetsCopiesFilesRelativeToSection(t *testing.T) {
	root := t.TempDir()
	sectionDir := filepath.Join(root, "content", "blog")
	require.NoError(t, os.MkdirAll(filepath.Join(sectionDir, "images"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sectionDir, "_index.md"), []byte("# Blog"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sectionDir, "cover.txt"), []byte("cover"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sectionDir, "images", "banner.txt"), []byte("banner"), 0644))

	processor := newAssetTestProcessor(t, root, &parser.Result{
		FrontMatter: map[string]any{
			"assets":     []string{"cover.txt", "images/banner.txt"},
			"asset_path": "{path}/media",
		},
		Content: "<p>Blog</p>",
	})
	section, err := processor.ParseSection(filepath.Join(sectionDir, "_index.md"))
	require.NoError(t, err)

	require.Len(t, section.Assets, 2)
	assert.Equal(t, filepath.Join(sectionDir, "cover.txt"), section.Assets[0].File)
	assert.Equal(t, "/blog/media/cover.txt", section.Assets[0].Path)
	assert.Equal(t, filepath.Join(sectionDir, "images", "banner.txt"), section.Assets[1].File)
	assert.Equal(t, "/blog/media/images/banner.txt", section.Assets[1].Path)

	w := writer.NewMemoryWriter()
	require.NoError(t, processor.RenderSection(section, assetTestTemplateSet{}, w))
	assert.Equal(t, "cover", readMemoryFile(t, w, "/blog/media/cover.txt"))
	assert.Equal(t, "banner", readMemoryFile(t, w, "/blog/media/images/banner.txt"))
}

func TestSectionAssetsUseDirectoryForHTMLPath(t *testing.T) {
	root := t.TempDir()
	sectionDir := filepath.Join(root, "content", "blog")
	require.NoError(t, os.MkdirAll(sectionDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sectionDir, "_index.md"), []byte("# Blog"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sectionDir, "cover.txt"), []byte("cover"), 0644))

	processor := newAssetTestProcessor(t, root, &parser.Result{
		FrontMatter: map[string]any{
			"assets": []string{"cover.txt"},
			"path":   "{path}.html",
		},
		Content: "<p>Blog</p>",
	})
	section, err := processor.ParseSection(filepath.Join(sectionDir, "_index.md"))
	require.NoError(t, err)

	require.Len(t, section.Assets, 1)
	assert.Equal(t, "/blog/cover.txt", section.Assets[0].Path)

	w := writer.NewMemoryWriter()
	require.NoError(t, processor.RenderSection(section, assetTestTemplateSet{}, w))
	assert.Equal(t, "cover", readMemoryFile(t, w, "/blog/cover.txt"))
}
